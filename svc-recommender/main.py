"""Entry point for svc-recommender.

Uses environment configuration (PYTHONPATH) to correctly resolve the flat module layout
and the buf-generated stubs (src/pb/recommender/) without requiring an installed package.
"""

from server import serve

if __name__ == "__main__":
    serve()
