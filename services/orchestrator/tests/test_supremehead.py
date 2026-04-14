"""
Tests for orchestrator/supremehead.py

Run with:
    cd services/orchestrator && python -m pytest tests/ -v
"""
import json
import os
import sys
import pytest

# Add parent dir so we can import supremehead directly
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from supremehead import (
    now_iso,
    safe_write_json,
    SupremeHead,
    MemoryCoreClient,
    MindNexusClient,
)


# ── Utility helpers ────────────────────────────────────────────────────────────

class TestNowIso:
    def test_returns_string(self):
        assert isinstance(now_iso(), str)

    def test_ends_with_z(self):
        assert now_iso().endswith("Z")

    def test_is_valid_iso_format(self):
        from datetime import datetime
        ts = now_iso().rstrip("Z")
        dt = datetime.fromisoformat(ts)
        assert dt is not None


class TestSafeWriteJson:
    def test_writes_valid_json(self, tmp_path):
        path = str(tmp_path / "out.json")
        payload = {"key": "value", "num": 42}
        safe_write_json(path, payload)
        with open(path) as f:
            assert json.load(f) == payload

    def test_no_tmp_file_left_behind(self, tmp_path):
        path = str(tmp_path / "atomic.json")
        safe_write_json(path, {"ok": True})
        assert os.path.exists(path)
        assert not os.path.exists(path + ".tmp")

    def test_overwrites_existing(self, tmp_path):
        path = str(tmp_path / "overwrite.json")
        safe_write_json(path, {"v": 1})
        safe_write_json(path, {"v": 2})
        with open(path) as f:
            assert json.load(f) == {"v": 2}


# ── SupremeHead initialization ─────────────────────────────────────────────────

class TestSupremeHeadInit:
    def test_uses_defaults_when_config_missing(self, tmp_path):
        """Should fall back to DEFAULT_CONFIG when config file doesn't exist."""
        nonexistent = str(tmp_path / "no_config.json")
        head = SupremeHead(config_path=nonexistent)
        assert head.config["memory_core_url"] == "http://localhost:3000"
        assert head.config["nft_threshold"] == 85

    def test_loads_custom_config(self, tmp_path):
        cfg_path = str(tmp_path / "cfg.json")
        custom = {"memory_core_url": "http://mem:9000", "nft_threshold": 70}
        with open(cfg_path, "w") as f:
            json.dump(custom, f)
        head = SupremeHead(config_path=cfg_path)
        assert head.config["memory_core_url"] == "http://mem:9000"
        assert head.config["nft_threshold"] == 70

    def test_merges_config_with_defaults(self, tmp_path):
        """Partial config should be merged — not replace — the defaults."""
        cfg_path = str(tmp_path / "partial.json")
        with open(cfg_path, "w") as f:
            json.dump({"nft_threshold": 99}, f)
        head = SupremeHead(config_path=cfg_path)
        assert head.config["nft_threshold"] == 99
        # Default keys still present
        assert "memory_core_url" in head.config


# ── SupremeHead.ingest_scroll ──────────────────────────────────────────────────

class TestIngestScroll:
    @pytest.fixture
    def head(self, tmp_path):
        """SupremeHead wired to a temp ledger; no real network services."""
        cfg_path = str(tmp_path / "cfg.json")
        cfg = {
            "memory_core_url": "http://localhost:19999",
            "mind_nexus_url": "http://localhost:19998",
            "nft_threshold": 85,
            "codex_ledger_path": str(tmp_path / "ledger.log"),
            "retries": 1,
            "retry_delay_seconds": 0,
        }
        with open(cfg_path, "w") as f:
            json.dump(cfg, f)
        return SupremeHead(config_path=cfg_path)

    def test_returns_dict_with_expected_keys(self, head):
        result = head.ingest_scroll("test raw data", "pytest")
        assert isinstance(result, dict)
        for key in ("status", "action", "score", "source", "analysis"):
            assert key in result, f"Missing key: {key}"

    def test_source_is_preserved(self, head):
        result = head.ingest_scroll("some content", "test-source")
        assert result["source"] == "test-source"

    def test_status_is_processed(self, head):
        result = head.ingest_scroll("data", "pytest")
        assert result["status"] == "Processed"

    def test_score_is_numeric(self, head):
        result = head.ingest_scroll("data", "pytest")
        assert isinstance(result["score"], (int, float))

    def test_analysis_contains_fallback_on_no_service(self, head):
        """With no Mind Nexus running, analysis should be a fallback dict."""
        result = head.ingest_scroll("test", "pytest")
        assert "sentiment" in result["analysis"] or "notes" in result["analysis"]


# ── MemoryCoreClient / MindNexusClient ────────────────────────────────────────

class TestMemoryCoreClient:
    def test_instantiates(self):
        client = MemoryCoreClient("http://localhost:3000")
        assert client.base_url == "http://localhost:3000"

    def test_store_returns_error_dict_without_service(self):
        client = MemoryCoreClient("http://localhost:19999")
        result = client.store({"scroll": "test"})
        # No service running — should return {"error": ...} gracefully
        assert "error" in result


class TestMindNexusClient:
    def test_instantiates(self):
        client = MindNexusClient("http://localhost:3001")
        assert client.base_url == "http://localhost:3001"

    def test_analyze_returns_fallback_without_service(self):
        client = MindNexusClient("http://localhost:19998")
        result = client.analyze("some text")
        # Fallback path should return a dict with known keys
        assert isinstance(result, dict)
        assert "sentiment" in result
        assert "value_score" in result
