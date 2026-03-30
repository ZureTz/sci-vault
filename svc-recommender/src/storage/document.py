"""Document storage — S3-compatible object store access for PDFs."""

import logging
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from mypy_boto3_s3 import S3Client

log = logging.getLogger(__name__)


class DocumentStorage:
    """Wraps an S3 client and bucket name for document file operations."""

    def __init__(self, s3_client: "S3Client", bucket: str) -> None:
        self._s3 = s3_client
        self._bucket = bucket

    def download_pdf(self, file_key: str) -> bytes:
        """Download a PDF from the private bucket and return raw bytes."""
        response = self._s3.get_object(Bucket=self._bucket, Key=file_key)
        return response["Body"].read()
