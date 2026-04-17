from __future__ import annotations

import logging
import os
from typing import Optional

from fastapi import FastAPI, Form, Request
from fastapi.responses import PlainTextResponse
from twilio.twiml.voice_response import Gather, VoiceResponse

logger = logging.getLogger("skipcalls")
logging.basicConfig(level=logging.INFO, format="%(levelname)s %(name)s %(message)s")

app = FastAPI(title="SkipCalls MVP")

PRIORITY_KEYWORDS = frozenset(
    ["urgent", "emergency", "lawyer", "school", "family", "hospital"]
)


def build_base_url(request: Request) -> str:
    forwarded_proto = request.headers.get("x-forwarded-proto")
    forwarded_host = request.headers.get("x-forwarded-host")
    host = forwarded_host or request.headers.get("host", "localhost:8000")
    scheme = forwarded_proto or request.url.scheme
    return f"{scheme}://{host}"


@app.get("/health")
async def health() -> dict:
    return {"ok": True, "service": "skipcalls"}


@app.post("/voice/incoming", response_class=PlainTextResponse)
async def incoming_call(request: Request, CallSid: str = Form(...), From: str = Form(...)) -> str:
    """
    Twilio voice webhook entrypoint.
    """
    base_url = build_base_url(request)

    vr = VoiceResponse()
    vr.pause(length=1)
    vr.say(
        "Hello. You have reached Skip Calls automated reception. "
        "Please say your name and the reason for your call after the tone.",
        voice="alice",
    )

    gather = Gather(
        input="speech",
        action=f"{base_url}/voice/process",
        method="POST",
        speech_timeout="auto",
        language="en-US",
    )
    vr.append(gather)

    # Fallback if speech is not captured.
    vr.say("I did not catch that.")
    vr.redirect(f"{base_url}/voice/retry", method="POST")

    return str(vr)


@app.post("/voice/retry", response_class=PlainTextResponse)
async def retry_call(request: Request) -> str:
    base_url = build_base_url(request)

    vr = VoiceResponse()
    vr.say(
        "Please state your name and reason for calling clearly after the tone.",
        voice="alice",
    )

    gather = Gather(
        input="speech",
        action=f"{base_url}/voice/process",
        method="POST",
        speech_timeout="auto",
        language="en-US",
    )
    vr.append(gather)

    vr.say("No response received. Goodbye.", voice="alice")
    vr.hangup()
    return str(vr)


@app.post("/voice/process", response_class=PlainTextResponse)
async def process_call(
    request: Request,
    CallSid: str = Form(...),
    From: str = Form(...),
    SpeechResult: Optional[str] = Form(default=None),
    Confidence: Optional[str] = Form(default=None),
) -> str:
    """
    Process caller speech.
    In production, store this in a database and optionally trigger downstream actions.
    """
    transcript = (SpeechResult or "").strip()
    confidence = Confidence or "unknown"

    # Basic placeholder routing logic.
    normalized = transcript.lower()
    likely_priority = any(word in normalized for word in PRIORITY_KEYWORDS)

    logger.info(
        "call_processed call_sid=%s from=%s confidence=%s priority=%s transcript=%r",
        CallSid,
        From,
        confidence,
        likely_priority,
        transcript,
    )

    vr = VoiceResponse()

    if not transcript:
        vr.say(
            "I was not able to capture your message. Please call back and try again.",
            voice="alice",
        )
        vr.hangup()
        return str(vr)

    if likely_priority:
        vr.say(
            "Thank you. Your message has been marked high priority and will be reviewed promptly.",
            voice="alice",
        )
    else:
        vr.say(
            "Thank you. Your message has been received and logged.",
            voice="alice",
        )

    vr.hangup()
    return str(vr)
