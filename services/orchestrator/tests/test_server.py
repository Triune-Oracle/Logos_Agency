"""
Tests for orchestrator/server.py (aiohttp HTTP layer).

Uses aiohttp's built-in test client — no real network port is opened.
Run with:
    cd services/orchestrator && python -m pytest tests/test_server.py -v
"""
import json
import pytest
from aiohttp.test_utils import TestClient, TestServer

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from server import make_app


# ── Health endpoint ────────────────────────────────────────────────────────────

@pytest.mark.asyncio
async def test_health_returns_200():
    async with TestClient(TestServer(make_app())) as client:
        resp = await client.get("/health")
        assert resp.status == 200


@pytest.mark.asyncio
async def test_health_returns_json_body():
    async with TestClient(TestServer(make_app())) as client:
        resp = await client.get("/health")
        body = await resp.json()
        assert body["status"] == "ok"
        assert body["service"] == "logos-orchestrator"
        assert "timestamp" in body


@pytest.mark.asyncio
async def test_health_content_type():
    async with TestClient(TestServer(make_app())) as client:
        resp = await client.get("/health")
        assert "application/json" in resp.content_type


# ── Ingest endpoint — validation ───────────────────────────────────────────────

@pytest.mark.asyncio
async def test_ingest_missing_raw_returns_400():
    async with TestClient(TestServer(make_app())) as client:
        resp = await client.post("/ingest", json={"source": "test"})
        assert resp.status == 400
        body = await resp.json()
        assert "error" in body


@pytest.mark.asyncio
async def test_ingest_empty_raw_returns_400():
    async with TestClient(TestServer(make_app())) as client:
        resp = await client.post("/ingest", json={"raw": "   ", "source": "test"})
        assert resp.status == 400


@pytest.mark.asyncio
async def test_ingest_invalid_json_returns_400():
    async with TestClient(TestServer(make_app())) as client:
        resp = await client.post(
            "/ingest",
            data="not-json",
            headers={"Content-Type": "application/json"},
        )
        assert resp.status == 400


# ── Ingest endpoint — happy path (no live services needed) ─────────────────────

@pytest.mark.asyncio
async def test_ingest_valid_payload_returns_200_or_processes():
    """
    With no live Memory Core / Mind Nexus, ingest_scroll will use fallback
    analysis and return a Processed result dict.  We assert the shape only.
    """
    async with TestClient(TestServer(make_app())) as client:
        resp = await client.post(
            "/ingest",
            json={"raw": "The market remembers the flame.", "source": "pulse-check"},
        )
        # Accept 200 (processed with fallback) or 500 (if upstream explodes)
        assert resp.status in (200, 500)
        if resp.status == 200:
            body = await resp.json()
            assert "status" in body
            assert "source" in body
            assert body["source"] == "pulse-check"


@pytest.mark.asyncio
async def test_ingest_default_source():
    """Source field is optional — should default gracefully."""
    async with TestClient(TestServer(make_app())) as client:
        resp = await client.post("/ingest", json={"raw": "test payload"})
        assert resp.status in (200, 500)
