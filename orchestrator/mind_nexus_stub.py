# minimal mind nexus (for local dev)
from flask import Flask, request, jsonify
import random
app = Flask(__name__)

@app.route("/analyze", methods=["POST"])
def analyze():
    payload = request.json or {}
    text = payload.get("raw", "")
    score = min(99, max(1, 50 + len(text) % 50 + (1 if "fire" in text else 0)))
    return jsonify({
        "patterns": ["temporal"],
        "sentiment": "positive" if "good" in text else "neutral",
        "value_score": score,
        "timestamp": "placeholder"
    })

if __name__ == "__main__":
    app.run(port=3001)
