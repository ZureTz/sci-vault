from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class MatchType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    MATCH_TYPE_UNSPECIFIED: _ClassVar[MatchType]
    MATCH_TYPE_SEMANTIC: _ClassVar[MatchType]
    MATCH_TYPE_KEYWORD: _ClassVar[MatchType]
MATCH_TYPE_UNSPECIFIED: MatchType
MATCH_TYPE_SEMANTIC: MatchType
MATCH_TYPE_KEYWORD: MatchType

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
    __slots__ = ("doc_id", "file_key", "content_type")
    DOC_ID_FIELD_NUMBER: _ClassVar[int]
    FILE_KEY_FIELD_NUMBER: _ClassVar[int]
    CONTENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    doc_id: int
    file_key: str
    content_type: str
    def __init__(self, doc_id: _Optional[int] = ..., file_key: _Optional[str] = ..., content_type: _Optional[str] = ...) -> None: ...

class EnrichDocumentResponse(_message.Message):
    __slots__ = ("accepted",)
    ACCEPTED_FIELD_NUMBER: _ClassVar[int]
    accepted: bool
    def __init__(self, accepted: _Optional[bool] = ...) -> None: ...

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

class SemanticSearchRequest(_message.Message):
    __slots__ = ("query", "user_id", "lab_id", "limit")
    QUERY_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    LAB_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    query: str
    user_id: int
    lab_id: int
    limit: int
    def __init__(self, query: _Optional[str] = ..., user_id: _Optional[int] = ..., lab_id: _Optional[int] = ..., limit: _Optional[int] = ...) -> None: ...

class ScoredDocument(_message.Message):
    __slots__ = ("doc_id", "title", "original_file_name", "summary", "authors", "tags", "similarity", "match_type")
    DOC_ID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    ORIGINAL_FILE_NAME_FIELD_NUMBER: _ClassVar[int]
    SUMMARY_FIELD_NUMBER: _ClassVar[int]
    AUTHORS_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    SIMILARITY_FIELD_NUMBER: _ClassVar[int]
    MATCH_TYPE_FIELD_NUMBER: _ClassVar[int]
    doc_id: int
    title: str
    original_file_name: str
    summary: str
    authors: _containers.RepeatedScalarFieldContainer[str]
    tags: _containers.RepeatedScalarFieldContainer[str]
    similarity: float
    match_type: MatchType
    def __init__(self, doc_id: _Optional[int] = ..., title: _Optional[str] = ..., original_file_name: _Optional[str] = ..., summary: _Optional[str] = ..., authors: _Optional[_Iterable[str]] = ..., tags: _Optional[_Iterable[str]] = ..., similarity: _Optional[float] = ..., match_type: _Optional[_Union[MatchType, str]] = ...) -> None: ...

class SemanticSearchResponse(_message.Message):
    __slots__ = ("results",)
    RESULTS_FIELD_NUMBER: _ClassVar[int]
    results: _containers.RepeatedCompositeFieldContainer[ScoredDocument]
    def __init__(self, results: _Optional[_Iterable[_Union[ScoredDocument, _Mapping]]] = ...) -> None: ...

class RecommendSimilarRequest(_message.Message):
    __slots__ = ("doc_id", "user_id", "lab_id", "limit")
    DOC_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    LAB_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    doc_id: int
    user_id: int
    lab_id: int
    limit: int
    def __init__(self, doc_id: _Optional[int] = ..., user_id: _Optional[int] = ..., lab_id: _Optional[int] = ..., limit: _Optional[int] = ...) -> None: ...

class RecommendSimilarResponse(_message.Message):
    __slots__ = ("results",)
    RESULTS_FIELD_NUMBER: _ClassVar[int]
    results: _containers.RepeatedCompositeFieldContainer[ScoredDocument]
    def __init__(self, results: _Optional[_Iterable[_Union[ScoredDocument, _Mapping]]] = ...) -> None: ...

class RecommendForUserRequest(_message.Message):
    __slots__ = ("user_id", "lab_id", "limit", "liked_doc_ids", "viewed_doc_ids", "recent_queries")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    LAB_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    LIKED_DOC_IDS_FIELD_NUMBER: _ClassVar[int]
    VIEWED_DOC_IDS_FIELD_NUMBER: _ClassVar[int]
    RECENT_QUERIES_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    lab_id: int
    limit: int
    liked_doc_ids: _containers.RepeatedScalarFieldContainer[int]
    viewed_doc_ids: _containers.RepeatedScalarFieldContainer[int]
    recent_queries: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, user_id: _Optional[int] = ..., lab_id: _Optional[int] = ..., limit: _Optional[int] = ..., liked_doc_ids: _Optional[_Iterable[int]] = ..., viewed_doc_ids: _Optional[_Iterable[int]] = ..., recent_queries: _Optional[_Iterable[str]] = ...) -> None: ...

class RecommendForUserResponse(_message.Message):
    __slots__ = ("results",)
    RESULTS_FIELD_NUMBER: _ClassVar[int]
    results: _containers.RepeatedCompositeFieldContainer[ScoredDocument]
    def __init__(self, results: _Optional[_Iterable[_Union[ScoredDocument, _Mapping]]] = ...) -> None: ...
