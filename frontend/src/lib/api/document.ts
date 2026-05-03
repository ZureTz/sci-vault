import request, { type DefaultResponse } from './request';

export type DocumentVisibility = 'private' | 'lab';

export interface UploadDocumentRequest {
	file: File;
	title: string | null;
	year: number | null;
	doi: string | null;
	visibility?: DocumentVisibility;
	lab_id?: number | null;
}

export interface DocumentListItem {
	id: number;
	title: string | null;
	original_file_name: string;
	file_size: number;
	enrich_status: string;
	visibility: DocumentVisibility;
	lab_id: number | null;
	lab_name: string | null;
	uploaded_by: number;
	uploaded_by_username?: string | null;
	created_at: string;
}

export interface ListDocumentsResponse {
	documents: DocumentListItem[];
	total: number;
	page: number;
	page_size: number;
}

export interface EnrichStatusResponse {
	doc_id: number;
	status: string;
}

export interface DocumentResponse {
	id: number;
	title: string | null;
	original_file_name: string;
	file_size: number;
	content_type: string;
	year: number | null;
	doi: string | null;
	enrich_status: string;
	visibility: DocumentVisibility;
	lab_id: number | null;
	lab_name: string | null;
	authors: string[];
	summary: string | null;
	tags: string[];
	view_count: number;
	like_count: number;
	liked_by_me: boolean;
	uploaded_by: number;
	download_url: string;
	created_at: string;
}

export interface BatchUploadDocumentRequest {
	files: File[];
	visibility?: DocumentVisibility;
	lab_id?: number | null;
}

export interface BatchUploadItemResult {
	filename: string;
	doc_id?: number;
	error?: string;
}

export interface BatchUploadDocumentResponse {
	results: BatchUploadItemResult[];
	succeeded: number;
	failed: number;
}

export interface UpdateVisibilityRequest {
	visibility: DocumentVisibility;
	lab_id?: number | null;
}

export interface BatchUpdateVisibilityRequest {
	doc_ids: number[];
	visibility: DocumentVisibility;
	lab_id?: number | null;
}

export interface BatchUpdateVisibilityResponse {
	updated: number;
}

export interface UpdateMetadataRequest {
	title?: string | null;
	year?: number | null;
	doi?: string | null;
}

export interface ListMyDocumentsParams {
	page?: number;
	page_size?: number;
	search?: string;
	status?: 'not_started' | 'pending' | 'processing' | 'done' | 'failed';
	visibility?: DocumentVisibility;
	lab_id?: number;
	sort_by?: 'created_at' | 'title' | 'file_size' | 'view_count';
	sort_order?: 'asc' | 'desc';
}

export interface ListLabDocumentsParams {
	page?: number;
	page_size?: number;
	search?: string;
	status?: 'not_started' | 'pending' | 'processing' | 'done' | 'failed';
	sort_by?: 'created_at' | 'title' | 'file_size' | 'view_count';
	sort_order?: 'asc' | 'desc';
}

const documentApi = {
	listMyDocuments(params: ListMyDocumentsParams = {}): Promise<ListDocumentsResponse> {
		const query: Record<string, string | number> = {
			page: params.page ?? 1,
			page_size: params.page_size ?? 20
		};
		if (params.search) query.search = params.search;
		if (params.status) query.status = params.status;
		if (params.visibility) query.visibility = params.visibility;
		if (params.lab_id != null) query.lab_id = params.lab_id;
		if (params.sort_by) query.sort_by = params.sort_by;
		if (params.sort_order) query.sort_order = params.sort_order;
		return request.get<ListDocumentsResponse>('/docs/mine', {
			params: query
		}) as unknown as Promise<ListDocumentsResponse>;
	},

	listLabDocuments(
		labId: number,
		params: ListLabDocumentsParams = {}
	): Promise<ListDocumentsResponse> {
		const query: Record<string, string | number> = {
			page: params.page ?? 1,
			page_size: params.page_size ?? 20
		};
		if (params.search) query.search = params.search;
		if (params.status) query.status = params.status;
		if (params.sort_by) query.sort_by = params.sort_by;
		if (params.sort_order) query.sort_order = params.sort_order;
		return request.get<ListDocumentsResponse>(`/labs/${labId}/documents`, {
			params: query
		}) as unknown as Promise<ListDocumentsResponse>;
	},

	listPendingDocuments(): Promise<ListDocumentsResponse> {
		return request.get<ListDocumentsResponse>(
			'/docs/pending'
		) as unknown as Promise<ListDocumentsResponse>;
	},

	getEnrichStatus(docId: number): Promise<EnrichStatusResponse> {
		return request.get<EnrichStatusResponse>(
			`/docs/${docId}/enrich_status`
		) as unknown as Promise<EnrichStatusResponse>;
	},

	uploadDocument(
		data: UploadDocumentRequest,
		onProgress?: (percent: number) => void
	): Promise<DocumentResponse> {
		const formData = new FormData();
		formData.append('file', data.file);
		if (data.title !== null) formData.append('title', data.title);
		if (data.year !== null) formData.append('year', String(data.year));
		if (data.doi !== null) formData.append('doi', data.doi);
		if (data.visibility) formData.append('visibility', data.visibility);
		if (data.lab_id != null) formData.append('lab_id', String(data.lab_id));
		return request.post<FormData, DocumentResponse>('/docs/upload', formData, {
			headers: { 'Content-Type': 'multipart/form-data' },
			timeout: 0,
			onUploadProgress: onProgress
				? (e) => {
						if (e.total) onProgress(Math.round((e.loaded / e.total) * 100));
					}
				: undefined
		}) as unknown as Promise<DocumentResponse>;
	},

	batchUploadDocuments(
		data: BatchUploadDocumentRequest,
		onProgress?: (percent: number) => void
	): Promise<BatchUploadDocumentResponse> {
		const formData = new FormData();
		for (const f of data.files) formData.append('files', f);
		if (data.visibility) formData.append('visibility', data.visibility);
		if (data.lab_id != null) formData.append('lab_id', String(data.lab_id));
		return request.post<FormData, BatchUploadDocumentResponse>('/docs/upload/batch', formData, {
			headers: { 'Content-Type': 'multipart/form-data' },
			timeout: 0,
			onUploadProgress: onProgress
				? (e) => {
						if (e.total) onProgress(Math.round((e.loaded / e.total) * 100));
					}
				: undefined
		}) as unknown as Promise<BatchUploadDocumentResponse>;
	},

	getDocument(docId: number): Promise<DocumentResponse> {
		return request.get<DocumentResponse>(`/docs/${docId}`) as unknown as Promise<DocumentResponse>;
	},

	restartEnrichment(docId: number): Promise<DefaultResponse> {
		return request.post<DefaultResponse>(
			`/docs/${docId}/restart_enrichment`
		) as unknown as Promise<DefaultResponse>;
	},

	updateVisibility(docId: number, data: UpdateVisibilityRequest): Promise<DefaultResponse> {
		return request.patch<UpdateVisibilityRequest, DefaultResponse>(
			`/docs/${docId}/visibility`,
			data
		) as unknown as Promise<DefaultResponse>;
	},

	batchUpdateVisibility(
		data: BatchUpdateVisibilityRequest
	): Promise<BatchUpdateVisibilityResponse> {
		return request.post<BatchUpdateVisibilityRequest, BatchUpdateVisibilityResponse>(
			'/docs/visibility/batch',
			data
		) as unknown as Promise<BatchUpdateVisibilityResponse>;
	},

	updateMetadata(docId: number, data: UpdateMetadataRequest): Promise<DefaultResponse> {
		return request.patch<UpdateMetadataRequest, DefaultResponse>(
			`/docs/${docId}`,
			data
		) as unknown as Promise<DefaultResponse>;
	},

	deleteDocument(docId: number): Promise<DefaultResponse> {
		return request.delete<DefaultResponse>(`/docs/${docId}`) as unknown as Promise<DefaultResponse>;
	}
};

export default documentApi;
