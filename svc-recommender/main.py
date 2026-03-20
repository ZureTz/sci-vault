"""Entry point for svc-recommender."""

import sys
from pathlib import Path

# Add src/ and generated stubs to sys.path
base_dir = Path(__file__).resolve().parent
src_dir = base_dir / "src"
pb_dir = src_dir / "pb"

if not src_dir.exists():
    raise RuntimeError(
        f"Source directory not found at {src_dir}. "
        "Ensure the project is laid out correctly."
    )
sys.path.insert(0, str(src_dir))

if not pb_dir.exists():
    raise RuntimeError(
        f"Generated protobuf stubs not found at {pb_dir}. "
        "Run `buf generate` to generate Python stubs before starting the service."
    )
sys.path.insert(0, str(pb_dir))

from server import serve

if __name__ == "__main__":
    serve()
