"""Entry point for svc-recommender.

Adds src/ and src/pb/ to sys.path so that the flat module layout and the
buf-generated stubs (src/pb/recommender/) are both importable without
requiring an installed package or a PYTHONPATH export.
"""

import sys
from pathlib import Path

_ROOT = Path(__file__).parent
sys.path.insert(0, str(_ROOT / "src" / "pb"))  # generated stubs
sys.path.insert(0, str(_ROOT / "src"))  # application modules

from server import serve

if __name__ == "__main__":
    serve()
