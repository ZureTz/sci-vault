"""Entry point for svc-recommender."""

import sys
from pathlib import Path

# Add src/ and generated stubs to sys.path
base_dir = Path(__file__).resolve().parent
src_dir = base_dir / "src"
pb_dir = src_dir / "pb"

sys.path.insert(0, str(src_dir))
sys.path.insert(0, str(pb_dir))

from server import serve

if __name__ == "__main__":
    serve()
