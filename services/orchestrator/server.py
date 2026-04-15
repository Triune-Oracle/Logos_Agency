#!/usr/bin/env python3
"""
services/orchestrator/server.py

Minimal aiohttp HTTP server that exposes the SupremeHead orchestrator
over a REST interface so the Go backend (and tests) can call it.

Endpoints:
  GET  /health   — liveness probe; returns {"status":"ok","service":"logos-orchestrator"}
  POST /ingest   — ingest a scroll; body: {"raw": str, "source": str}

Environment variables:
  ORCHESTRATOR_PORT    TCP port to listen on (default: 5000)
  ORCHESTRATOR_CONFIG  Path to config.json (default: config.json)
"""

from __future__ import annotations

import asyncio
import logging
import os
import sys

try:
    from aiohttp import web
except ImportError:
    print("aiohttp is required: pip install aiohttp", file=sys.stderr)
    sys.exit(1)

from supremehead import SupremeHead, now_iso

logger = logging.getLogger("orchestrator.server")

CONFIG_PATH = os.environ.get("ORCHESTRATOR_CONFIG", "config.json")

# Lazily initialised — created on first request so the server starts fast.
_head: SupremeHead | None = None
_head_lock = asyncio.Lock()


async def get_head() -> SupremeHead:
    global _head
    if _head is None:
        async with _head_lock:
            if _head is None:
                _head = SupremeHead(config_path=CONFIG_PATH)
    return _head


# ── Route handlers ─────────────────────────────────────────────────────────────

async def handle_health(request: web.Request) -> web.Response:
    return web.json_response(
        {
            "status": "ok",
            "service": "logos-orchestrator",
            "timestamp": now_iso(),
        }
    )


async def handle_ingest(request: web.Request) -> web.Response:
    try:
        body = await request.json()
    except Exception:
        return web.json_response({"error": "request body must be valid JSON"}, status=400)

    raw: str = body.get("raw", "").strip()
    source: str = body.get("source", "api")

    if not raw:
        return web.json_response({"error": "'raw' field is required and must not be empty"}, status=400)

    head = await get_head()
    try:
        # ingest_scroll is synchronous — run in a thread pool so we don't
        # block the event loop while it makes HTTP calls to Memory Core etc.
        loop = asyncio.get_running_loop()
        result = await loop.run_in_executor(None, head.ingest_scroll, raw, source)
        return web.json_response(result)
    except Exception as exc:
        logger.exception("ingest_scroll raised an unexpected error")
        return web.json_response({"error": str(exc)}, status=500)


# ── App factory (used by tests and __main__) ───────────────────────────────────

def make_app() -> web.Application:
    app = web.Application()
    app.router.add_get("/health", handle_health)
    app.router.add_post("/ingest", handle_ingest)
    return app


if __name__ == "__main__":
    port = int(os.environ.get("ORCHESTRATOR_PORT", "5000"))
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(name)s %(message)s",
    )
    logger.info("Starting logos-orchestrator on port %d", port)
    web.run_app(make_app(), port=port, access_log=logger)
