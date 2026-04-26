import request from './request';

export interface LikeStateResponse {
	doc_id: number;
	liked: boolean;
	like_count: number;
}

export interface HistoryItem {
	interaction_id: number;
	interacted_at: string;
	doc_id: number;
	title: string | null;
	original_file_name: string;
	visibility: 'private' | 'lab';
	lab_id: number | null;
	lab_name: string | null;
	enrich_status: string;
}

export interface ListHistoryResponse {
	items: HistoryItem[];
	total: number;
	page: number;
	page_size: number;
}

export interface ListHistoryParams {
	page?: number;
	page_size?: number;
}

const interactionApi = {
	like(docId: number): Promise<LikeStateResponse> {
		return request.post<undefined, LikeStateResponse>(
			`/docs/${docId}/like`
		) as unknown as Promise<LikeStateResponse>;
	},

	unlike(docId: number): Promise<LikeStateResponse> {
		return request.delete<LikeStateResponse>(
			`/docs/${docId}/like`
		) as unknown as Promise<LikeStateResponse>;
	},

	listViewHistory(params: ListHistoryParams = {}): Promise<ListHistoryResponse> {
		return request.get<ListHistoryResponse>('/mine/history/views', {
			params: { page: params.page ?? 1, page_size: params.page_size ?? 20 }
		}) as unknown as Promise<ListHistoryResponse>;
	},

	listLikeHistory(params: ListHistoryParams = {}): Promise<ListHistoryResponse> {
		return request.get<ListHistoryResponse>('/mine/history/likes', {
			params: { page: params.page ?? 1, page_size: params.page_size ?? 20 }
		}) as unknown as Promise<ListHistoryResponse>;
	}
};

export default interactionApi;
