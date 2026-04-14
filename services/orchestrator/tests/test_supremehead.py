"""
Tests for orchestrator/supremehead.py

Run with:
    cd services/orchestrator && python -m pytest tests/ -v
"""
import asyncio
import sys
import os
import json
import pytest

# Add parent dir to path so we can import supremehead
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from supremehead import (
    now_iso,
    safe_write_json,
    SupremeHeadOrchestrator,
    OrchestratorConfig,
)


# ── Utility helpers ────────────────────────────────────────────────────────────

class TestNowIso:
    def test_returns_string(self):
        result = now_iso()
        assert isinstance(result, str)

    def test_ends_with_z(self):
        result = now_iso()
        assert result.endswith("Z")

    def test_is_iso_format(self):
        result = now_iso()
        # Should parse cleanly as ISO 8601
        from datetime import datetime
        dt = datetime.fromisoformat(result.rstrip("Z"))
        assert dt is not None


class TestSafeWriteJson:
    def test_writes_valid_json(self, tmp_path):
        path = str(tmp_path / "out.json")
        payload = {"key": "value", "num": 42}
        safe_write_json(path, payload)
        with open(path) as f:
            result = json.load(f)
        assert result == payload

    def test_atomic_write_creates_file(self, tmp_path):
        path = str(tmp_path / "atomic.json")
        safe_write_json(path, {"ok": True})
        assert os.path.exists(path)
        assert not os.path.exists(path + ".tmp")


# ── OrchestratorConfig ─────────────────────────────────────────────────────────

class TestOrchestratorConfig:
    def test_loads_from_dict(self):
        cfg = OrchestratorConfig(
            memory_core_url="http://localhost:3000",
            mind_nexus_url="http://localhost:3001",
            nft_threshold=85,
            codex_ledger_path="logs/ledger.log",
            retries=2,
            retry_delay_seconds=1,
        )
        assert cfg.memory_core_url == "http://localhost:3000"
        assert cfg.nft_threshold == 85

    def test_default_retries(self):
        cfg = OrchestratorConfig(
            memory_core_url="http://a",
            mind_nexus_url="http://b",
        )
        assert cfg.retries >= 1


# ── SupremeHeadOrchestrator ────────────────────────────────────────────────────

class TestSupremeHeadOrchestrator:
    @pytest.fixture
    def config(self, tmp_path):
        return OrchestratorConfig(
            memory_core_url="http://localhost:3000",
            mind_nexus_url="http://localhost:3001",
            nft_threshold=85,
            codex_ledger_path=str(tmp_path / "ledger.log"),
            retries=1,
            retry_delay_seconds=0,
        )

    def test_instantiates(self, config):
        orch = SupremeHeadOrchestrator(config)
        assert orch is not None

    @pytest.mark.asyncio
    async def test_ingest_scroll_returns_result(self, config, tmp_path):
        """Smoke test: ingest a minimal scroll payload without network calls."""
        orch = SupremeHeadOrchestrator(config)
        scroll = {
            "id": "test-001",
            "content": "This is a test scroll.",
            "timestamp": now_iso(),
        }
        # With no live services, this should either succeed or raise a
        # connection error — not crash with a programming error.
        try:
            result = await orch.ingest_scroll(scroll)
            assert result is not None
        except (ConnectionError, OSError):
            pass  # expected: no real services running in test environment
