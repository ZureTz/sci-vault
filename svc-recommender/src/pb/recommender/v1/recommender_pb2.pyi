from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class HealthRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class HealthResponse(_message.Message):
    __slots__ = ("status", "service")
    STATUS_FIELD_NUMBER: _ClassVar[int]
    SERVICE_FIELD_NUMBER: _ClassVar[int]
    status: str
    service: str
    def __init__(self, status: _Optional[str] = ..., service: _Optional[str] = ...) -> None: ...

class EnrichDocumentRequest(_message.Message):
    __slots__ = ("doc_id", "file_key")
    DOC_ID_FIELD_NUMBER: _ClassVar[int]
    FILE_KEY_FIELD_NUMBER: _ClassVar[int]
    doc_id: int
    file_key: str
    def __init__(self, doc_id: _Optional[int] = ..., file_key: _Optional[str] = ...) -> None: ...

class EnrichDocumentResponse(_message.Message):
    __slots__ = ("accepted",)
    ACCEPTED_FIELD_NUMBER: _ClassVar[int]
    accepted: bool
    def __init__(self, accepted: bool = ...) -> None: ...

class TranslateTextRequest(_message.Message):
    __slots__ = ("text", "target_language")
    TEXT_FIELD_NUMBER: _ClassVar[int]
    TARGET_LANGUAGE_FIELD_NUMBER: _ClassVar[int]
    text: str
    target_language: str
    def __init__(self, text: _Optional[str] = ..., target_language: _Optional[str] = ...) -> None: ...

class TranslateTextResponse(_message.Message):
    __slots__ = ("chunk",)
    CHUNK_FIELD_NUMBER: _ClassVar[int]
    chunk: str
    def __init__(self, chunk: _Optional[str] = ...) -> None: ...
