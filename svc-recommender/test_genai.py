"""Smoke test for the GenAI client. Run with: uv run test_genai.py"""

import logging
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent / "src"))

from config import Config
from infrastructure.genai import DEFAULT_MODEL, GenAI

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(name)s: %(message)s")

cfg = Config.load()
print(f"Vertex AI: {cfg.google_genai_use_vertexai}")
print(f"API key set: {bool(cfg.google_genai_api_key)}")

client = GenAI(cfg)

# Test embedding client
print("\n--- embedding client ---")
embed_resp = client.embedding_client.models.embed_content(
    model="gemini-embedding-001",
    contents="Hello, world!",
)
assert embed_resp.embeddings, "no embeddings returned"
vec = embed_resp.embeddings[0].values
assert vec, "embedding values are empty"
print(f"embedding dim={len(vec)}  first3={vec[:3]}")

# Test metadata client
print("\n--- metadata client ---")
gen_resp = client.metadata_client.models.generate_content(
    model=DEFAULT_MODEL,
    contents="How is the furry culture in Japan?",
)
assert gen_resp.text, "no text in response"
print(f"response: {gen_resp.text.strip()!r}")

print("\nAll checks passed.")
