import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer

import numpy as np
from catboost import CatBoostClassifier

MODEL_FILE_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), "models", "best_catboost_model_new.cbm")
HOST = os.getenv("PREDICT_HOST", "127.0.0.1")
PORT = int(os.getenv("PREDICT_PORT", "8087"))

class MedicalRiskPredictor:
    def __init__(self):
        self.model = None
        self.features = ["Age", "Number of polyps", "Long diameter", "Short diameter", "Base"]
        self.bins = [
            (0.00, 0.25, "Low Risk", "Low Risk"),
            (0.25, 0.50, "Moderate Risk", "Moderate Risk"),
            (0.50, 0.75, "High Risk", "High Risk"),
            (0.75, 1.00, "Very High Risk", "Very High Risk"),
        ]
        self.advice = {
            "Low Risk": "Follow-up is not required",
            "Moderate Risk": "Follow-up ultrasound is recommended at 6 months, 1 year, and 2 years; Follow-up should be discontinued after 2 years in the absence of growth.",
            "High Risk": "Cholecystectomy is recommended if the patient is fit for, and accepts, surgery; MDT discussion may be considered",
            "Very High Risk": "Cholecystectomy is strongly recommended if the patient is fit for, and accepts, surgery",
        }
        self.load_model()

    def load_model(self):
        if not os.path.exists(MODEL_FILE_PATH):
            return False
        self.model = CatBoostClassifier()
        self.model.load_model(MODEL_FILE_PATH)
        return True

    def predict_risk(self, age, polyps, long_diameter, short_diameter, base):
        if self.model is None:
            return None, "Model not loaded"
        input_data = np.array([[age, polyps, long_diameter, short_diameter, base]])
        probability = self.model.predict_proba(input_data)[0][1]
        risk_level = None
        risk_level_cn = None
        for min_prob, max_prob, level, level_cn in self.bins:
            if min_prob <= probability < max_prob:
                risk_level = level
                risk_level_cn = level_cn
                break
        advice = self.advice.get(risk_level, "No advice")
        return {
            "probability": probability,
            "risk_level": risk_level,
            "risk_level_cn": risk_level_cn,
            "advice": advice,
        }, None

predictor = MedicalRiskPredictor()


class Handler(BaseHTTPRequestHandler):
    def _send_json(self, status_code, payload):
        data = json.dumps(payload, ensure_ascii=True).encode("utf-8")
        self.send_response(status_code)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_POST(self):
        if self.path != "/api/predict":
            self.send_error(404, "not found")
            return
        content_length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(content_length)
        try:
            payload = json.loads(body.decode("utf-8"))
        except json.JSONDecodeError:
            self.send_error(400, "invalid json")
            return

        age = payload.get("age")
        polyps = payload.get("polyps")
        long_diameter = payload.get("long_diameter")
        short_diameter = payload.get("short_diameter")
        base = payload.get("base", payload.get("Base"))

        if any(v is None for v in [age, polyps, long_diameter, short_diameter, base]):
            self.send_error(400, "missing required fields")
            return

        try:
            age = int(age)
            polyps = int(polyps)
            long_diameter = float(long_diameter)
            short_diameter = float(short_diameter)
            base = int(base)
        except (TypeError, ValueError):
            self.send_error(400, "invalid field types")
            return

        if base not in (1, 2):
            self.send_error(400, "base must be 1 or 2")
            return

        result, error = predictor.predict_risk(age, polyps, long_diameter, short_diameter, base)
        if error:
            self._send_json(503, {"error": error})
            return
        self._send_json(200, {"result": result})

    def do_GET(self):
        if self.path != "/api/health":
            self.send_error(404, "not found")
            return
        self._send_json(
            200,
            {
                "status": "healthy",
                "model_loaded": predictor.model is not None,
                "model_path": MODEL_FILE_PATH,
            },
        )

    def log_message(self, format, *args):
        return


def main():
    server = HTTPServer((HOST, PORT), Handler)
    print(f"prediction service running on http://{HOST}:{PORT}/api")
    server.serve_forever()


if __name__ == "__main__":
    main()
