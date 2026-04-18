"""Repository layer — database access abstractions."""

# Document enrichment status (enrich_status column).
ENRICH_STATUS_NOT_STARTED = "not_started"
ENRICH_STATUS_PENDING = "pending"
ENRICH_STATUS_PROCESSING = "processing"
ENRICH_STATUS_DONE = "done"
ENRICH_STATUS_FAILED = "failed"

# Document visibility (visibility column).
DOC_VISIBILITY_PRIVATE = "private"
DOC_VISIBILITY_LAB = "lab"
