import request from './request';

// MatchType mirrors the proto MatchType enum.
export const MatchType = {
	UNSPECIFIED: 0,
	SEMANTIC: 1,
	KEYWORD: 2
} as const;

export type MatchTypeValue = (typeof MatchType)[keyof typeof MatchType];

export interface SearchResultItem {
	doc_id: number;
	title: string;
	original_file_name: string;
	summary: string;
	authors: string[];
	tags: string[];
	similarity: number;
	match_type: MatchTypeValue;
}

export interface SearchDocumentsResponse {
	results: SearchResultItem[];
}

export interface SearchHistoryItem {
	id: number;
	query: string;
	lab_id?: number | null;
	result_count: number;
	last_used_at: string;
}

export interface ListSearchHistoryResponse {
	items: SearchHistoryItem[];
}

export interface DeleteSearchHistoryResponse {
	deleted: number;
}

const searchApi = {
	searchDocuments(query: string, labId?: number, limit?: number): Promise<SearchDocumentsResponse> {
		return request.get<SearchDocumentsResponse>('/search', {
			params: { query, lab_id: labId || undefined, limit: limit || undefined },
			timeout: 30000
		}) as unknown as Promise<SearchDocumentsResponse>;
	},

	listHistory(limit?: number): Promise<ListSearchHistoryResponse> {
		return request.get<ListSearchHistoryResponse>('/search/history', {
			params: { limit: limit || undefined }
		}) as unknown as Promise<ListSearchHistoryResponse>;
	},

	clearHistory(): Promise<DeleteSearchHistoryResponse> {
		return request.delete('/search/history') as unknown as Promise<DeleteSearchHistoryResponse>;
	}
};

export default searchApi;
