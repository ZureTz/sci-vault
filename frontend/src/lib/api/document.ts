import request from './request';

export interface UploadDocumentRequest {
	file: File;
	title?: string;
	year?: number;
	doi?: string;
}

export interface DocumentListItem {
	id: number;
	title: string;
	original_file_name: string;
	file_size: number;
	enrich_status: string;
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
	title: string;
	original_file_name: string;
	file_size: number;
	content_type: string;
	year?: number;
	doi?: string;
	enrich_status: string;
	authors: string[];
	summary: string;
	tags: string[];
	view_count: number;
	like_count: number;
	uploaded_by: number;
	download_url: string;
	created_at: string;
}

const documentApi = {
	listMyDocuments(page = 1, pageSize = 20): Promise<ListDocumentsResponse> {
		return request.get<ListDocumentsResponse>('/docs/mine', {
			params: { page, page_size: pageSize }
		}) as unknown as Promise<ListDocumentsResponse>;
	},

	getEnrichStatus(docId: number): Promise<EnrichStatusResponse> {
		return request.get<EnrichStatusResponse>(
			`/docs/${docId}/enrich_status`
		) as unknown as Promise<EnrichStatusResponse>;
	},

	uploadDocument(data: UploadDocumentRequest): Promise<DocumentResponse> {
		const formData = new FormData();
		formData.append('file', data.file);
		if (data.title) formData.append('title', data.title);
		if (data.year !== undefined) formData.append('year', String(data.year));
		if (data.doi) formData.append('doi', data.doi);
		return request.post<FormData, DocumentResponse>('/docs/upload', formData, {
			headers: { 'Content-Type': 'multipart/form-data' }
		}) as unknown as Promise<DocumentResponse>;
	}
};

export default documentApi;
