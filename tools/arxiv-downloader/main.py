"""Download PDF versions of publications from arXiv.

Single-paper mode:
    uv run main.py --id 2401.12345 2402.54321
    uv run main.py --query "retrieval augmented generation" --max 10

Batch mode (across multiple arXiv categories):
    uv run main.py --fields cs.AI cs.LG stat.ML math.CO physics.gen-ph --total 100
    uv run main.py --fields cs.AI cs.CV --per-field 25 --max-size-mb 50

Common flags:
    --out ./papers        Output directory (papers are grouped by field in batch mode).
    --max-size-mb 50      Skip any single PDF larger than this (default 50 MB).
    --overwrite           Re-download even if the file exists.
"""

from __future__ import annotations

import argparse
import re
import sys
import time
from pathlib import Path

import feedparser
import httpx

ARXIV_API = "https://export.arxiv.org/api/query"
ARXIV_PDF = "https://arxiv.org/pdf/{id}.pdf"

# Polite delay between arXiv API calls (they request >=3s).
API_DELAY_SEC = 3.0
# Smaller delay between PDF downloads from the mirror.
PDF_DELAY_SEC = 1.0
USER_AGENT = "sci-vault-arxiv-downloader/0.2"

DEFAULT_MAX_SIZE_MB = 50

# Matches new-style (2401.12345 / 2401.12345v2) and old-style (math.GT/0309136) IDs.
ARXIV_ID_RE = re.compile(r"^(?:[a-z\-]+(?:\.[A-Z]{2})?/\d{7}|\d{4}\.\d{4,5})(v\d+)?$")


def normalize_id(raw: str) -> str:
    s = raw.strip()
    s = re.sub(r"^https?://arxiv\.org/(abs|pdf)/", "", s)
    s = re.sub(r"\.pdf$", "", s)
    s = re.sub(r"^arXiv:", "", s, flags=re.IGNORECASE)
    return s


def safe_filename(title: str, arxiv_id: str) -> str:
    cleaned = re.sub(r"\s+", " ", title).strip()
    cleaned = re.sub(r"[^A-Za-z0-9 \-_.]", "", cleaned)
    cleaned = cleaned[:120].rstrip()
    slug_id = arxiv_id.replace("/", "_")
    return f"{slug_id} - {cleaned}.pdf" if cleaned else f"{slug_id}.pdf"


def _parse_feed(xml: str) -> list[dict]:
    feed = feedparser.parse(xml)
    papers = []
    for entry in feed.entries:
        if entry is None:
            continue

        entry_id = str(entry.get("id", ""))
        raw_id = entry_id.rsplit("/abs/", 1)[-1]
        bare_id = re.sub(r"v\d+$", "", raw_id)
        pdf_link = next(
            (
                link.href
                for link in entry.links
                if link.get("type") == "application/pdf"
            ),
            ARXIV_PDF.format(id=bare_id),
        )
        papers.append(
            {
                "id": bare_id,
                "title": entry.title,
                "pdf_url": pdf_link,
                "authors": [a.name for a in (entry.get("authors") or [])],
            }
        )
    return papers


def fetch_metadata(client: httpx.Client, arxiv_ids: list[str]) -> list[dict]:
    params = {"id_list": ",".join(arxiv_ids), "max_results": str(len(arxiv_ids))}
    resp = client.get(ARXIV_API, params=params, timeout=30.0)
    resp.raise_for_status()
    return _parse_feed(resp.text)


def search(
    client: httpx.Client,
    search_query: str,
    max_results: int,
    sort_by: str = "submittedDate",
) -> list[dict]:
    params = {
        "search_query": search_query,
        "start": "0",
        "max_results": str(max_results),
        "sortBy": sort_by,
        "sortOrder": "descending",
    }
    resp = client.get(ARXIV_API, params=params, timeout=60.0)
    resp.raise_for_status()
    return _parse_feed(resp.text)


class OversizeError(Exception):
    """Raised when a PDF exceeds the configured size limit."""


def download_pdf(
    client: httpx.Client, url: str, dest: Path, max_size_bytes: int
) -> int:
    """Stream a PDF to disk, aborting if it exceeds max_size_bytes.

    Returns the byte count written. Raises OversizeError if the limit is hit
    (either via Content-Length or during streaming). On any failure, the
    partial file is removed.
    """
    tmp = dest.with_suffix(dest.suffix + ".part")
    written = 0
    try:
        with client.stream("GET", url, timeout=180.0, follow_redirects=True) as resp:
            resp.raise_for_status()

            content_length = resp.headers.get("content-length")
            if content_length and int(content_length) > max_size_bytes:
                raise OversizeError(
                    f"advertised {int(content_length) / 1_048_576:.1f} MB"
                )

            with tmp.open("wb") as f:
                for chunk in resp.iter_bytes(chunk_size=64 * 1024):
                    written += len(chunk)
                    if written > max_size_bytes:
                        raise OversizeError(
                            f"exceeded limit mid-stream at {written / 1_048_576:.1f} MB"
                        )
                    f.write(chunk)
        tmp.rename(dest)
        return written
    except BaseException:
        if tmp.exists():
            tmp.unlink()
        raise


def download_batch(
    client: httpx.Client,
    papers: list[dict],
    out_dir: Path,
    max_size_bytes: int,
    overwrite: bool,
    label: str,
) -> tuple[int, int, int]:
    """Download a list of papers into out_dir. Returns (ok, skipped, failed)."""
    out_dir.mkdir(parents=True, exist_ok=True)
    ok = skipped = failed = 0

    for i, paper in enumerate(papers):
        dest = out_dir / safe_filename(paper["title"], paper["id"])
        prefix = f"  [{label}] [{i + 1}/{len(papers)}] {paper['id']}"

        if dest.exists() and not overwrite:
            print(f"{prefix} skip (exists): {dest.name}")
            skipped += 1
            continue

        title_snip = paper["title"][:80].replace("\n", " ")
        print(f"{prefix} {title_snip}")
        try:
            if i > 0:
                time.sleep(PDF_DELAY_SEC)
            size = download_pdf(client, paper["pdf_url"], dest, max_size_bytes)
            print(f"           -> {dest.name} ({size / 1_048_576:.1f} MB)")
            ok += 1
        except OversizeError as e:
            print(f"           ! oversize ({e}), skipped", file=sys.stderr)
            skipped += 1
        except httpx.HTTPError as e:
            print(f"           ! http error: {e}", file=sys.stderr)
            failed += 1

    return ok, skipped, failed


def distribute_counts(total: int, n_fields: int) -> list[int]:
    """Split `total` across `n_fields` as evenly as possible."""
    base, extra = divmod(total, n_fields)
    return [base + (1 if i < extra else 0) for i in range(n_fields)]


def run_batch(
    client: httpx.Client,
    fields: list[str],
    total: int | None,
    per_field: int | None,
    out_dir: Path,
    max_size_bytes: int,
    overwrite: bool,
) -> int:
    if per_field is not None:
        counts = [per_field] * len(fields)
    else:
        assert total is not None
        counts = distribute_counts(total, len(fields))

    print(f"Batch: {len(fields)} field(s), targeting {sum(counts)} paper(s) total.")
    for field, n in zip(fields, counts, strict=True):
        print(f"  - {field}: {n}")

    grand_ok = grand_skip = grand_fail = 0
    for idx, (field, n) in enumerate(zip(fields, counts, strict=True)):
        if n <= 0:
            continue
        print(f"\n[{idx + 1}/{len(fields)}] Fetching {n} paper(s) from {field!r}")
        if idx > 0:
            time.sleep(API_DELAY_SEC)

        try:
            papers = search(client, f"cat:{field}", n)
        except httpx.HTTPError as e:
            print(f"  ! metadata query failed for {field}: {e}", file=sys.stderr)
            grand_fail += n
            continue

        if not papers:
            print(f"  (no results for {field})")
            continue

        field_dir = out_dir / field.replace("/", "_")
        ok, skipped, failed = download_batch(
            client, papers, field_dir, max_size_bytes, overwrite, label=field
        )
        grand_ok += ok
        grand_skip += skipped
        grand_fail += failed

    print(f"\nDone. downloaded={grand_ok} skipped={grand_skip} failed={grand_fail}")
    return 0 if grand_fail == 0 else 2


def run_single(
    client: httpx.Client,
    ids: list[str],
    query: str | None,
    max_results: int,
    out_dir: Path,
    max_size_bytes: int,
    overwrite: bool,
) -> int:
    if query:
        print(f"Searching arXiv for: {query!r} (max {max_results})")
        papers = search(client, f"all:{query}", max_results)
    else:
        normalized = [normalize_id(i) for i in ids]
        for nid in normalized:
            if not ARXIV_ID_RE.match(nid):
                print(
                    f"Warning: {nid!r} does not look like an arXiv ID",
                    file=sys.stderr,
                )
        papers = fetch_metadata(client, normalized)

    if not papers:
        print("No results.", file=sys.stderr)
        return 1

    print(f"Found {len(papers)} paper(s).")
    ok, skipped, failed = download_batch(
        client, papers, out_dir, max_size_bytes, overwrite, label="single"
    )
    print(f"\nDone. downloaded={ok} skipped={skipped} failed={failed}")
    return 0 if failed == 0 else 2


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Download arXiv papers as PDFs (single or batched across fields).",
    )
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument(
        "--id",
        nargs="+",
        help="One or more arXiv IDs (e.g. 2401.12345, math.GT/0309136, or a full URL).",
    )
    group.add_argument("--query", help="Free-text search query (single-batch mode).")
    group.add_argument(
        "--fields",
        nargs="+",
        help="arXiv category codes (e.g. cs.AI cs.LG math.CO). Enables batch mode.",
    )

    parser.add_argument(
        "--total",
        type=int,
        help="Total papers to fetch across --fields (distributed evenly).",
    )
    parser.add_argument(
        "--per-field",
        type=int,
        help="Fetch this many papers per field (overrides --total).",
    )
    parser.add_argument(
        "--max",
        type=int,
        default=10,
        help="Max results when using --query (default: 10).",
    )
    parser.add_argument(
        "--out",
        type=Path,
        default=Path("./papers"),
        help="Output directory (default: ./papers).",
    )
    parser.add_argument(
        "--max-size-mb",
        type=float,
        default=DEFAULT_MAX_SIZE_MB,
        help=f"Per-file size limit in MB (default: {DEFAULT_MAX_SIZE_MB}).",
    )
    parser.add_argument(
        "--overwrite",
        action="store_true",
        help="Re-download even if the file exists.",
    )
    args = parser.parse_args()

    if args.fields and not (args.total or args.per_field):
        parser.error("--fields requires either --total or --per-field")
    if args.fields and args.total and args.per_field:
        parser.error("use either --total or --per-field, not both")

    max_size_bytes = int(args.max_size_mb * 1_048_576)
    args.out.mkdir(parents=True, exist_ok=True)

    with httpx.Client(headers={"User-Agent": USER_AGENT}) as client:
        if args.fields:
            return run_batch(
                client=client,
                fields=args.fields,
                total=args.total,
                per_field=args.per_field,
                out_dir=args.out,
                max_size_bytes=max_size_bytes,
                overwrite=args.overwrite,
            )
        return run_single(
            client=client,
            ids=args.id or [],
            query=args.query,
            max_results=args.max,
            out_dir=args.out,
            max_size_bytes=max_size_bytes,
            overwrite=args.overwrite,
        )


if __name__ == "__main__":
    sys.exit(main())
