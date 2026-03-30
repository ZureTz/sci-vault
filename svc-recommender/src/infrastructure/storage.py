"""S3-compatible object storage client factory."""

import logging
from typing import TYPE_CHECKING

import boto3
from botocore.client import Config as BotoConfig

from config import Config

if TYPE_CHECKING:
    from mypy_boto3_s3 import S3Client

log = logging.getLogger(__name__)


def build_s3_client(cfg: Config) -> "S3Client":
    """Create and return a configured boto3 S3 client."""
    log.info("building S3 client (endpoint=%s)", cfg.s3_endpoint)
    return boto3.client(
        "s3",
        endpoint_url=cfg.s3_endpoint,
        aws_access_key_id=cfg.s3_access_key,
        aws_secret_access_key=cfg.s3_secret_key,
        config=BotoConfig(signature_version="s3v4"),
        region_name="ap-east-1",
    )
