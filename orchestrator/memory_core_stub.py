# minimal memory core receiver (for local dev)
from flask import Flask, request, jsonify
app = Flask(__name__)

@app.route("/store", methods=["POST"])
def store():
    data = request.json or {}
    # in prod: validate + persist
    print("Stored:", data.get("scroll", {}).get("source"))
    return jsonify({"status":"ok","received": True})

if __name__ == "__main__":
    app.run(port=3000)
