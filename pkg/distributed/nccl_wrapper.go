// Package distributed provides NCCL-backed collective communication primitives
// for gradient synchronization across distributed training nodes.
//
// The package supports two transport modes:
//   - RDMA (InfiniBand/RoCEv2 via UCX + NCCL) — primary fast path
//   - TCP/IP (pure Go channels) — automatic fallback when RDMA is unavailable
//
// Build tags:
//   nccl  — links against libnccl.so and enables GPU-accelerated collectives
//   (none) — compiles the pure-Go TCP fallback only
package distributed

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Public types
// ─────────────────────────────────────────────────────────────────────────────

// Transport identifies the communication back-end in use.
type Transport int

const (
	// TransportRDMA uses NCCL over InfiniBand/RoCEv2 (UCX).
	TransportRDMA Transport = iota
	// TransportTCP uses a pure TCP/IP ring all-reduce (fallback).
	TransportTCP
)

func (t Transport) String() string {
	switch t {
	case TransportRDMA:
		return "RDMA/NCCL"
	case TransportTCP:
		return "TCP/IP"
	default:
		return "unknown"
	}
}

// Topology selects the collective algorithm used for all-reduce.
type Topology int

const (
	// TopologyRing performs a ring all-reduce (bandwidth-optimal for large messages).
	TopologyRing Topology = iota
	// TopologyTree performs a hierarchical tree reduction (latency-optimal for small messages).
	TopologyTree
)

// Config holds the configuration for a Communicator.
type Config struct {
	// WorldSize is the total number of training nodes (ranks).
	WorldSize int
	// Rank is the zero-based index of this node within the communicator.
	Rank int
	// Transport selects RDMA or TCP.  Set to TransportTCP when InfiniBand is
	// unavailable; the communicator will fall back automatically if RDMA init
	// fails regardless of this setting.
	Transport Transport
	// Topology selects the all-reduce algorithm.
	Topology Topology
	// RDMADeviceName is the IB/RoCE device name (e.g. "mlx5_0").
	// Ignored when Transport == TransportTCP.
	RDMADeviceName string
	// TCPAddrs lists peer TCP addresses (host:port) ordered by rank.
	// Required for the TCP fallback path.
	TCPAddrs []string
	// RendezvousAddr is the address of the bootstrap rendezvous service.
	RendezvousAddr string
	// Timeout is the per-operation deadline.
	Timeout time.Duration
}

func (c Config) validate() error {
	if c.WorldSize < 1 {
		return errors.New("distributed: WorldSize must be >= 1")
	}
	if c.Rank < 0 || c.Rank >= c.WorldSize {
		return fmt.Errorf("distributed: Rank %d out of range [0, %d)", c.Rank, c.WorldSize)
	}
	if c.Transport == TransportTCP && len(c.TCPAddrs) != c.WorldSize {
		return fmt.Errorf("distributed: TCPAddrs length %d != WorldSize %d", len(c.TCPAddrs), c.WorldSize)
	}
	return nil
}

// AllReduceResult carries timing metadata returned by AllReduce.
type AllReduceResult struct {
	// Transport reports which back-end was actually used.
	Transport Transport
	// Duration is the wall-clock time for the collective operation.
	Duration time.Duration
	// BytesTransferred is the total bytes sent/received per rank.
	BytesTransferred int64
}

// Communicator orchestrates gradient synchronization across ranks.
type Communicator struct {
	cfg       Config
	transport Transport // resolved transport (may differ from cfg if RDMA unavailable)
	mu        sync.Mutex

	// ncclHandle is a placeholder for the cgo NCCL communicator handle.
	// It is only set when built with the `nccl` build tag.
	ncclHandle interface{}

	// tcpConns holds rank → conn for the TCP fallback ring.
	tcpConns map[int]net.Conn

	initialized bool
	closed      bool
}

// ─────────────────────────────────────────────────────────────────────────────
// Constructor
// ─────────────────────────────────────────────────────────────────────────────

// NewCommunicator creates and initialises a Communicator.
//
// If RDMA initialisation fails (e.g. no InfiniBand adapter present), the
// communicator transparently falls back to TCP/IP and logs a warning.
func NewCommunicator(cfg Config) (*Communicator, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	c := &Communicator{
		cfg:      cfg,
		tcpConns: make(map[int]net.Conn),
	}

	if cfg.Transport == TransportRDMA {
		if err := c.initRDMA(); err != nil {
			log.Printf("[distributed] RDMA init failed (%v) — falling back to TCP/IP", err)
			c.transport = TransportTCP
		} else {
			c.transport = TransportRDMA
		}
	} else {
		c.transport = TransportTCP
	}

	if c.transport == TransportTCP {
		if err := c.initTCP(); err != nil {
			return nil, fmt.Errorf("distributed: TCP fallback init failed: %w", err)
		}
	}

	c.initialized = true
	log.Printf("[distributed] rank %d/%d ready, transport=%s, topology=%s",
		cfg.Rank, cfg.WorldSize, c.transport, topologyName(cfg.Topology))
	return c, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// AllReduce (sum)
// ─────────────────────────────────────────────────────────────────────────────

// AllReduce performs an in-place all-reduce SUM over the float32 gradient
// buffer buf across all ranks.  On return every rank holds the global sum.
func (c *Communicator) AllReduce(ctx context.Context, buf []float32) (AllReduceResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized || c.closed {
		return AllReduceResult{}, errors.New("distributed: communicator not ready")
	}

	start := time.Now()
	var err error

	switch c.transport {
	case TransportRDMA:
		err = c.allReduceNccl(ctx, buf)
	default:
		switch c.cfg.Topology {
		case TopologyTree:
			err = c.allReduceTreeTCP(ctx, buf)
		default:
			err = c.allReduceRingTCP(ctx, buf)
		}
	}
	if err != nil {
		return AllReduceResult{}, fmt.Errorf("distributed: AllReduce failed: %w", err)
	}

	return AllReduceResult{
		Transport:        c.transport,
		Duration:         time.Since(start),
		BytesTransferred: int64(len(buf)) * 4 * 2, // send + receive in bytes
	}, nil
}

// Transport returns the resolved transport back-end.
func (c *Communicator) Transport() Transport { return c.transport }

// Close releases all resources held by the communicator.
func (c *Communicator) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	for _, conn := range c.tcpConns {
		_ = conn.Close()
	}
	c.tcpConns = nil
	return c.destroyNccl()
}

// ─────────────────────────────────────────────────────────────────────────────
// RDMA / NCCL path
// ─────────────────────────────────────────────────────────────────────────────

// initRDMA probes for RDMA hardware and initialises the NCCL communicator.
// When compiled without the `nccl` build tag this always returns an error,
// triggering the TCP fallback.
func (c *Communicator) initRDMA() error {
	return initNCCL(c)
}

// allReduceNccl invokes the NCCL all-reduce kernel.
// The actual implementation is provided by nccl_cgo.go (nccl build tag).
func (c *Communicator) allReduceNccl(ctx context.Context, buf []float32) error {
	return ncclAllReduceSum(ctx, c.ncclHandle, buf)
}

func (c *Communicator) destroyNccl() error {
	return ncclDestroy(c.ncclHandle)
}

// ─────────────────────────────────────────────────────────────────────────────
// TCP fallback — ring all-reduce
// ─────────────────────────────────────────────────────────────────────────────

// initTCP opens connections to all peer ranks via TCP.
func (c *Communicator) initTCP() error {
	if c.cfg.WorldSize == 1 {
		return nil // single-rank: no peers to connect to
	}
	// In a production deployment the rendezvous service hands out addresses.
	// For testing, TCPAddrs is populated directly in Config.
	if len(c.cfg.TCPAddrs) == 0 {
		return nil // no peers configured — single-rank smoke test
	}
	// Connections are established lazily on first use in allReduceRingTCP.
	return nil
}

// allReduceRingTCP implements a bandwidth-optimal ring all-reduce over TCP.
//
// Algorithm (Rabenseifner):
//  1. Reduce-scatter: each rank sends/receives (N-1) times, accumulating a
//     shard of the final sum.
//  2. All-gather: each rank broadcasts its shard (N-1) times.
//
// This achieves 2*(N-1)/N * bytes transferred per rank, approaching 100%
// bandwidth utilisation for large messages.
func (c *Communicator) allReduceRingTCP(ctx context.Context, buf []float32) error {
	if c.cfg.WorldSize == 1 {
		return nil
	}
	n := c.cfg.WorldSize
	rank := c.cfg.Rank
	size := len(buf)

	// Shard boundaries
	shardSize := (size + n - 1) / n
	shards := make([][]float32, n)
	for i := 0; i < n; i++ {
		lo := i * shardSize
		hi := lo + shardSize
		if hi > size {
			hi = size
		}
		if lo >= size {
			shards[i] = nil
		} else {
			shards[i] = buf[lo:hi]
		}
	}

	sendTo := (rank + 1) % n
	recvFrom := (rank + n - 1) % n

	// ── Phase 1: Reduce-scatter ──────────────────────────────────────────
	for step := 0; step < n-1; step++ {
		sendShard := (rank - step + n) % n
		recvShard := (rank - step - 1 + n) % n

		var wg sync.WaitGroup
		var sendErr, recvErr error

		wg.Add(2)
		go func() {
			defer wg.Done()
			sendErr = c.tcpSend(ctx, sendTo, shards[sendShard])
		}()
		go func() {
			defer wg.Done()
			incoming := make([]float32, len(shards[recvShard]))
			recvErr = c.tcpRecv(ctx, recvFrom, incoming)
			if recvErr == nil {
				addFloat32(shards[recvShard], incoming)
			}
		}()
		wg.Wait()

		if sendErr != nil {
			return sendErr
		}
		if recvErr != nil {
			return recvErr
		}
	}

	// ── Phase 2: All-gather ──────────────────────────────────────────────
	for step := 0; step < n-1; step++ {
		sendShard := (rank - step + 1 + n) % n
		recvShard := (rank - step + n) % n

		var wg sync.WaitGroup
		var sendErr, recvErr error

		wg.Add(2)
		go func() {
			defer wg.Done()
			sendErr = c.tcpSend(ctx, sendTo, shards[sendShard])
		}()
		go func() {
			defer wg.Done()
			recvErr = c.tcpRecv(ctx, recvFrom, shards[recvShard])
		}()
		wg.Wait()

		if sendErr != nil {
			return sendErr
		}
		if recvErr != nil {
			return recvErr
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// TCP fallback — hierarchical tree all-reduce
// ─────────────────────────────────────────────────────────────────────────────

// allReduceTreeTCP implements a binary-tree all-reduce.
//
// Phase 1 (reduce): leaves send to parents; root accumulates the global sum.
// Phase 2 (broadcast): root sends the result down the tree to all leaves.
//
// Latency: O(log₂ N) steps — preferred for small messages (< 64 KB).
func (c *Communicator) allReduceTreeTCP(ctx context.Context, buf []float32) error {
	if c.cfg.WorldSize == 1 {
		return nil
	}
	rank := c.cfg.Rank
	n := c.cfg.WorldSize

	parent := (rank - 1) / 2
	leftChild := 2*rank + 1
	rightChild := 2*rank + 2

	// ── Phase 1: reduce toward root ──────────────────────────────────────
	// Each non-leaf receives from its children, accumulates, then sends upward.
	tmp := make([]float32, len(buf))

	if leftChild < n {
		if err := c.tcpRecv(ctx, leftChild, tmp); err != nil {
			return err
		}
		addFloat32(buf, tmp)
	}
	if rightChild < n {
		if err := c.tcpRecv(ctx, rightChild, tmp); err != nil {
			return err
		}
		addFloat32(buf, tmp)
	}
	if rank != 0 {
		if err := c.tcpSend(ctx, parent, buf); err != nil {
			return err
		}
	}

	// ── Phase 2: broadcast from root ─────────────────────────────────────
	if rank != 0 {
		if err := c.tcpRecv(ctx, parent, buf); err != nil {
			return err
		}
	}
	if leftChild < n {
		if err := c.tcpSend(ctx, leftChild, buf); err != nil {
			return err
		}
	}
	if rightChild < n {
		if err := c.tcpSend(ctx, rightChild, buf); err != nil {
			return err
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Low-level TCP helpers
// ─────────────────────────────────────────────────────────────────────────────

// tcpSend serialises buf as little-endian IEEE 754 float32 and writes it to
// the TCP connection for the given peer rank.
func (c *Communicator) tcpSend(ctx context.Context, peer int, buf []float32) error {
	conn, err := c.peerConn(ctx, peer)
	if err != nil {
		return err
	}
	if dl, ok := ctx.Deadline(); ok {
		_ = conn.SetWriteDeadline(dl)
	}
	return writeFloat32Slice(conn, buf)
}

// tcpRecv reads a float32 slice from the TCP connection for the given peer.
func (c *Communicator) tcpRecv(ctx context.Context, peer int, buf []float32) error {
	conn, err := c.peerConn(ctx, peer)
	if err != nil {
		return err
	}
	if dl, ok := ctx.Deadline(); ok {
		_ = conn.SetReadDeadline(dl)
	}
	return readFloat32Slice(conn, buf)
}

// peerConn returns (creating if needed) the TCP connection to peer rank.
func (c *Communicator) peerConn(ctx context.Context, peer int) (net.Conn, error) {
	if conn, ok := c.tcpConns[peer]; ok {
		return conn, nil
	}
	if peer >= len(c.cfg.TCPAddrs) {
		// In single-node / test mode there are no real peers.
		return nil, fmt.Errorf("distributed: no TCP address configured for rank %d", peer)
	}
	addr := c.cfg.TCPAddrs[peer]
	dialer := &net.Dialer{Timeout: c.cfg.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("distributed: dial rank %d (%s): %w", peer, addr, err)
	}
	c.tcpConns[peer] = conn
	return conn, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Serialisation helpers
// ─────────────────────────────────────────────────────────────────────────────

// writeFloat32Slice writes a []float32 to w as raw little-endian bytes.
func writeFloat32Slice(w net.Conn, buf []float32) error {
	b := make([]byte, len(buf)*4)
	for i, v := range buf {
		bits := math.Float32bits(v)
		b[i*4] = byte(bits)
		b[i*4+1] = byte(bits >> 8)
		b[i*4+2] = byte(bits >> 16)
		b[i*4+3] = byte(bits >> 24)
	}
	_, err := w.Write(b)
	return err
}

// readFloat32Slice reads len(buf) float32 values from r.
func readFloat32Slice(r net.Conn, buf []float32) error {
	b := make([]byte, len(buf)*4)
	if _, err := readFull(r, b); err != nil {
		return err
	}
	for i := range buf {
		bits := uint32(b[i*4]) | uint32(b[i*4+1])<<8 | uint32(b[i*4+2])<<16 | uint32(b[i*4+3])<<24
		buf[i] = math.Float32frombits(bits)
	}
	return nil
}

// readFull reads exactly len(b) bytes from r.
func readFull(r net.Conn, b []byte) (int, error) {
	total := 0
	for total < len(b) {
		n, err := r.Read(b[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// addFloat32 accumulates src into dst element-wise (dst[i] += src[i]).
func addFloat32(dst, src []float32) {
	n := len(dst)
	if len(src) < n {
		n = len(src)
	}
	for i := 0; i < n; i++ {
		dst[i] += src[i]
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// NCCL stub — replaced by nccl_cgo.go when built with -tags nccl
// ─────────────────────────────────────────────────────────────────────────────

// initNCCL initialises a NCCL communicator.
// This stub always returns an error so that the TCP fallback is used when the
// binary is built without the `nccl` build tag.
func initNCCL(c *Communicator) error {
	return errors.New("distributed: NCCL not compiled in (build with -tags nccl)")
}

// ncclAllReduceSum performs an in-place all-reduce SUM via NCCL.
// The stub is replaced by a cgo implementation when built with -tags nccl.
func ncclAllReduceSum(_ context.Context, _ interface{}, _ []float32) error {
	return errors.New("distributed: NCCL not available")
}

// ncclDestroy frees NCCL resources.
func ncclDestroy(_ interface{}) error { return nil }

// ─────────────────────────────────────────────────────────────────────────────
// Utility
// ─────────────────────────────────────────────────────────────────────────────

func topologyName(t Topology) string {
	switch t {
	case TopologyTree:
		return "tree"
	default:
		return "ring"
	}
}
