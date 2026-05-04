"""LibreOffice-headless wrapper for converting Office documents to PDF."""

import logging
import pathlib
import subprocess
import tempfile

log = logging.getLogger(__name__)

_SOFFICE = "soffice"
_TIMEOUT_SECONDS = 120


class OfficeConversionError(RuntimeError):
    """Raised when LibreOffice fails to produce a PDF."""


def office_to_pdf(raw: bytes, ext: str) -> bytes:
    """Convert DOCX/PPTX/XLSX bytes to PDF via `soffice --headless`.

    Each invocation gets its own `UserInstallation` profile dir so concurrent
    calls from the enrichment thread pool don't fight over LibreOffice's
    user-profile lock.
    """
    with tempfile.TemporaryDirectory(prefix="conv-") as workdir:
        work = pathlib.Path(workdir)
        src = work / f"input.{ext}"
        out_dir = work / "out"
        profile_dir = work / "profile"
        out_dir.mkdir()
        profile_dir.mkdir()

        src.write_bytes(raw)

        cmd = [
            _SOFFICE,
            f"-env:UserInstallation=file://{profile_dir}",
            "--headless",
            "--norestore",
            "--nolockcheck",
            "--nodefault",
            "--nofirststartwizard",
            "--convert-to",
            "pdf",
            "--outdir",
            str(out_dir),
            str(src),
        ]

        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                timeout=_TIMEOUT_SECONDS,
                check=False,
            )
        except subprocess.TimeoutExpired as exc:
            raise OfficeConversionError(
                f"LibreOffice timed out after {_TIMEOUT_SECONDS}s converting .{ext}"
            ) from exc

        if result.returncode != 0:
            raise OfficeConversionError(
                f"LibreOffice exited {result.returncode} converting .{ext}: "
                f"{result.stderr.decode(errors='replace').strip()}"
            )

        pdfs = list(out_dir.glob("*.pdf"))
        if not pdfs:
            raise OfficeConversionError(
                f"LibreOffice produced no PDF for .{ext}: "
                f"stdout={result.stdout.decode(errors='replace').strip()} "
                f"stderr={result.stderr.decode(errors='replace').strip()}"
            )

        return pdfs[0].read_bytes()
