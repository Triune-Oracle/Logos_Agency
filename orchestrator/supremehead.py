#!/usr/bin/env python3
"""
orchestrator/supremehead.py

Supreme Head Orchestrator (upgraded)

Responsibilities:
 - Ingest scrolls
 - Analyze via Mind Nexus
 - Decide: Memory vs NFT mint
 - Record ledger events, logs, metrics
 - Offer sync + async ingestion APIs
"""

from __future__ import annotations
import json
import os
import time
import logging
from datetime import datetime
from typing import Optional, Dict, Any
import asyncio
import functools

# try to use aiofiles for async file I/O
try:
    import aiofiles  # type: ignore
except Exception:
    aiofiles = None  # type: ignore

# try to use requests/aiohttp if available for nicer behavior; fall back to stdlib
try:
    import requests  # type: ignore
except Exception:
    requests = None  # type: ignore

try:
    import aiohttp  # type: ignore
except Exception:
    aiohttp = None  # type: ignore

# ---- Logging ----
LOG_DIR = os.environ.get("TRIUMV_LOG_DIR", ".")
os.makedirs(LOG_DIR, exist_ok=True)
LOG_FILE = os.path.join(LOG_DIR, "supremehead.log")

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
    handlers=[
        logging.FileHandler(LOG_FILE),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger("supremehead")


# ---- Utility functions ----
def now_iso() -> str:
    return datetime.utcnow().isoformat() + "Z"


def safe_write_json(path: str, obj: Any):
    tmp = f"{path}.tmp"
    with open(tmp, "w", encoding="utf-8") as f:
        json.dump(obj, f, indent=2, ensure_ascii=False)
    os.replace(tmp, path)


# ---- Integration stubs (pluggable clients) ----
class HTTPClient:
    """Simple pluggable HTTP client supporting sync and async calls."""

    @staticmethod
    def sync_post(url: str, payload: Dict[str, Any], timeout: int = 10):
        if requests:
            try:
                r = requests.post(url, json=payload, timeout=timeout)
                r.raise_for_status()
                return r.json()
            except Exception as e:
                raise
        # fallback to urllib
        import urllib.request as ur
        import urllib.error as ue
        data = json.dumps(payload).encode("utf-8")
        req = ur.Request(url, data=data, headers={"Content-Type": "application/json"})
        try:
            with ur.urlopen(req, timeout=timeout) as resp:
                body = resp.read().decode("utf-8")
                return json.loads(body) if body else {}
        except Exception:
            raise

    @staticmethod
    async def async_post(url: str, payload: Dict[str, Any], timeout: int = 10):
        if aiohttp:
            async with aiohttp.ClientSession() as sess:
                async with sess.post(url, json=payload, timeout=timeout) as resp:
                    resp.raise_for_status()
                    return await resp.json()
        # fallback: run sync in executor
        try:
            loop = asyncio.get_running_loop()
        except RuntimeError:
            loop = asyncio.get_event_loop()
        func = functools.partial(HTTPClient.sync_post, url, payload, timeout)
        return await loop.run_in_executor(None, func)


class MemoryCoreClient:
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip("/")

    def store(self, scroll: Dict[str, Any]) -> Dict[str, Any]:
        url = f"{self.base_url}/store" if not self.base_url.endswith("/store") else self.base_url
        try:
            logger.debug(f"MemoryCoreClient.store -> POST {url}")
            return HTTPClient.sync_post(url, scroll)
        except Exception as e:
            logger.exception("MemoryCore store failed")
            return {"error": str(e)}

    async def store_async(self, scroll: Dict[str, Any]) -> Dict[str, Any]:
        url = f"{self.base_url}/store" if not self.base_url.endswith("/store") else self.base_url
        try:
            return await HTTPClient.async_post(url, scroll)
        except Exception as e:
            logger.exception("MemoryCore async store failed")
            return {"error": str(e)}


class MindNexusClient:
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip("/")

    def analyze(self, raw: str, meta: Dict[str, Any] = None) -> Dict[str, Any]:
        payload = {"raw": raw, "meta": meta or {}}
        url = f"{self.base_url}/analyze" if not self.base_url.endswith("/analyze") else self.base_url
        try:
            logger.debug(f"MindNexusClient.analyze -> POST {url}")
            return HTTPClient.sync_post(url, payload)
        except Exception as e:
            logger.exception("MindNexus analyze failed")
            # fallback lightweight analysis
            return {
                "patterns": [],
                "sentiment": "neutral",
                "value_score": 50,
                "notes": f"fallback: {str(e)}",
                "timestamp": now_iso()
            }

    async def analyze_async(self, raw: str, meta: Dict[str, Any] = None) -> Dict[str, Any]:
        payload = {"raw": raw, "meta": meta or {}}
        url = f"{self.base_url}/analyze" if not self.base_url.endswith("/analyze") else self.base_url
        try:
            return await HTTPClient.async_post(url, payload)
        except Exception as e:
            logger.exception("MindNexus analyze async failed")
            return {
                "patterns": [],
                "sentiment": "neutral",
                "value_score": 50,
                "notes": f"fallback async: {str(e)}",
                "timestamp": now_iso()
            }


class SwarmEngine:
    def __init__(self, config: Dict[str, Any]):
        self.config = config or {}

    def trigger_nft_mint(self, raw: str, analysis: Dict[str, Any]) -> Dict[str, Any]:
        # Real implementation would call a minting service, sign txn, etc.
        logger.info("SwarmEngine: trigger_nft_mint called (stubbed)")
        # stubbed response:
        return {"status": "mint_triggered", "tx": None}

    async def trigger_nft_mint_async(self, raw: str, analysis: Dict[str, Any]) -> Dict[str, Any]:
        return self.trigger_nft_mint(raw, analysis)


# ---- SupremeHead orchestrator ----
class SupremeHead:
    DEFAULT_CONFIG = {
        "memory_core_url": "http://localhost:3000",
        "mind_nexus_url": "http://localhost:3001",
        "nft_threshold": 85,
        "codex_ledger_path": os.path.join(LOG_DIR, "codex_ledger.log"),
        "retries": 2,
        "retry_delay_seconds": 1
    }

    def __init__(self, config_path: str = "config.json"):
        self.config = self._load_config(config_path)
        self.memory_core = MemoryCoreClient(self.config["memory_core_url"])
        self.mind_nexus = MindNexusClient(self.config["mind_nexus_url"])
        self.swarm_engine = SwarmEngine(self.config.get("swarm_config", {}))
        self.ledger_path = self.config.get("codex_ledger_path", "codex_ledger.log")
        logger.info("Supreme Head Initialized. Awaiting Scroll Ingestion.")

    def _load_config(self, path: str) -> Dict[str, Any]:
        if not os.path.exists(path):
            logger.warning(f"Config file not found at {path}. Using defaults.")
            return dict(SupremeHead.DEFAULT_CONFIG)
        try:
            with open(path, "r", encoding="utf-8") as f:
                user = json.load(f)
            merged = dict(SupremeHead.DEFAULT_CONFIG)
            merged.update(user)
            return merged
        except Exception:
            logger.exception("Failed to load config, using defaults.")
            return dict(SupremeHead.DEFAULT_CONFIG)

    # Ledger & event recording
    def _record_event(self, event_type: str, payload: Dict[str, Any]):
        entry = {
            "event_type": event_type,
            "timestamp": now_iso(),
            "payload": payload
        }
        try:
            with open(self.ledger_path, "a", encoding="utf-8") as f:
                f.write(json.dumps(entry, ensure_ascii=False) + "\n")
            logger.debug(f"Recorded event: {event_type}")
        except Exception:
            logger.exception("Failed to write to ledger")

    # Async ledger recording (non-blocking for async paths)
    async def _record_event_async(self, event_type: str, payload: Dict[str, Any]):
        entry = {
            "event_type": event_type,
            "timestamp": now_iso(),
            "payload": payload
        }
        try:
            if aiofiles:
                # Use async file I/O if available
                async with aiofiles.open(self.ledger_path, "a", encoding="utf-8") as f:
                    await f.write(json.dumps(entry, ensure_ascii=False) + "\n")
            else:
                # Fallback: run sync I/O in executor to avoid blocking event loop
                try:
                    loop = asyncio.get_running_loop()
                except RuntimeError:
                    loop = asyncio.get_event_loop()
                await loop.run_in_executor(None, self._record_event, event_type, payload)
                return
            logger.debug(f"Recorded event async: {event_type}")
        except Exception:
            logger.exception("Failed to write to ledger async")

    # Safe call wrapper with simple retries
    def _safe_call(self, fn, *args, retries: Optional[int] = None, **kwargs):
        r = retries if retries is not None else self.config.get("retries", 2)
        delay = self.config.get("retry_delay_seconds", 1)
        last_exc = None
        for attempt in range(1, r + 1):
            try:
                return fn(*args, **kwargs)
            except Exception as e:
                last_exc = e
                logger.warning(f"Attempt {attempt}/{r} failed for {fn.__name__}: {e}")
                if attempt < r:  # Only sleep if we're going to retry
                    time.sleep(delay)
        logger.error(f"All {r} attempts failed for {fn.__name__}")
        raise last_exc

    # Async safe call wrapper with simple retries
    async def _safe_call_async(self, fn, *args, retries: Optional[int] = None, **kwargs):
        r = retries if retries is not None else self.config.get("retries", 2)
        delay = self.config.get("retry_delay_seconds", 1)
        last_exc = None
        for attempt in range(1, r + 1):
            try:
                return await fn(*args, **kwargs)
            except Exception as e:
                last_exc = e
                logger.warning(f"Attempt {attempt}/{r} failed for async {fn.__name__}: {e}")
                if attempt < r:  # Only sleep if we're going to retry
                    await asyncio.sleep(delay)
        logger.error(f"All {r} attempts failed for async {fn.__name__}")
        raise last_exc

    # canonical scroll format
    def _make_scroll(self, raw: str, source: str) -> Dict[str, Any]:
        return {
            "raw": raw,
            "source": source,
            "ingested_at": now_iso()
        }

    # Synchronous ingestion path
    def ingest_scroll(self, raw_data: str, source: str) -> Dict[str, Any]:
        logger.info(f"Ingesting scroll from {source}")
        scroll = self._make_scroll(raw_data, source)
        self._record_event("scroll_received", {"source": source, "snippet": raw_data[:160]})

        # analysis (with safe call)
        try:
            analysis = self._safe_call(self.mind_nexus.analyze, raw_data, {"source": source})
        except Exception as e:
            logger.exception("Analysis failed catastrophically")
            analysis = {
                "patterns": [],
                "sentiment": "neutral",
                "value_score": 50,
                "notes": f"analysis error: {str(e)}",
                "timestamp": now_iso()
            }

        score = analysis.get("value_score", 0)
        self._record_event("scroll_analyzed", {"score": score, "analysis_meta": analysis.get("timestamp")})

        # Decision logic
        action = None
        try:
            if score >= int(self.config.get("nft_threshold", 85)):
                logger.info(f"High Value Scroll (Score: {score}). Triggering NFT Tokenization.")
                result = self._safe_call(self.swarm_engine.trigger_nft_mint, raw_data, analysis)
                action = "NFT Mint Triggered"
                self._record_event("nft_triggered", {"score": score, "result": result})
            else:
                logger.info(f"Standard Scroll (Score: {score}). Storing in Memory Core.")
                store_payload = {"scroll": scroll, "analysis": analysis}
                result = self._safe_call(self.memory_core.store, store_payload)
                action = "Stored in Memory Core"
                self._record_event("scroll_stored", {"score": score, "result": result})
        except Exception:
            logger.exception("Action stage failed")
            action = "Action Failed"

        return {
            "status": "Processed",
            "action": action,
            "score": score,
            "source": source,
            "analysis": analysis
        }

    # Async ingestion path
    async def ingest_scroll_async(self, raw_data: str, source: str) -> Dict[str, Any]:
        logger.info(f"[async] Ingesting scroll from {source}")
        scroll = self._make_scroll(raw_data, source)
        await self._record_event_async("scroll_received_async", {"source": source, "snippet": raw_data[:160]})

        # async analyze with safe retry
        try:
            if hasattr(self.mind_nexus, "analyze_async"):
                analysis = await self._safe_call_async(self.mind_nexus.analyze_async, raw_data, {"source": source})
            else:
                try:
                    loop = asyncio.get_running_loop()
                except RuntimeError:
                    loop = asyncio.get_event_loop()
                analysis = await loop.run_in_executor(None, functools.partial(self.mind_nexus.analyze, raw_data, {"source": source}))
        except Exception:
            logger.exception("Async analysis failed")
            analysis = {
                "patterns": [],
                "sentiment": "neutral",
                "value_score": 50,
                "timestamp": now_iso()
            }

        score = analysis.get("value_score", 0)
        await self._record_event_async("scroll_analyzed_async", {"score": score})

        # decision async
        action = None
        try:
            if score >= int(self.config.get("nft_threshold", 85)):
                if hasattr(self.swarm_engine, "trigger_nft_mint_async"):
                    res = await self._safe_call_async(self.swarm_engine.trigger_nft_mint_async, raw_data, analysis)
                else:
                    try:
                        loop = asyncio.get_running_loop()
                    except RuntimeError:
                        loop = asyncio.get_event_loop()
                    res = await loop.run_in_executor(None, functools.partial(self.swarm_engine.trigger_nft_mint, raw_data, analysis))
                action = "NFT Mint Triggered"
                await self._record_event_async("nft_triggered_async", {"score": score, "result": res})
            else:
                payload = {"scroll": scroll, "analysis": analysis}
                if hasattr(self.memory_core, "store_async"):
                    res = await self._safe_call_async(self.memory_core.store_async, payload)
                else:
                    try:
                        loop = asyncio.get_running_loop()
                    except RuntimeError:
                        loop = asyncio.get_event_loop()
                    res = await loop.run_in_executor(None, functools.partial(self.memory_core.store, payload))
                action = "Stored in Memory Core"
                await self._record_event_async("scroll_stored_async", {"score": score, "result": res})
        except Exception:
            logger.exception("Async action failed")
            action = "Action Failed"

        return {
            "status": "Processed",
            "action": action,
            "score": score,
            "source": source,
            "analysis": analysis
        }


# ---- Quick CLI for manual testing ----
def _cli_demo():
    head = SupremeHead()
    test_scroll = "The flame remembers the pattern of the market's quiet laughter."
    result = head.ingest_scroll(test_scroll, "Founding Ritualist Log")
    print(json.dumps(result, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    _cli_demo()
