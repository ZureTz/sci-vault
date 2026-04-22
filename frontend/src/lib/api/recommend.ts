import request from './request';

export interface SimilarDocumentItem {
	doc_id: number;
	title: string;
	original_file_name: string;
	summary: string;
	authors: string[];
	tags: string[];
	similarity: number;
}

export interface RecommendSimilarResponse {
	results: SimilarDocumentItem[];
}

const recommendApi = {
	getSimilar(
		docId: number,
		opts: { lab_id?: number; limit?: number } = {}
	): Promise<RecommendSimilarResponse> {
		const params: Record<string, string | number> = {};
		if (opts.lab_id != null) params.lab_id = opts.lab_id;
		if (opts.limit != null) params.limit = opts.limit;
		return request.get<RecommendSimilarResponse>(`/docs/${docId}/similar`, {
			params
		}) as unknown as Promise<RecommendSimilarResponse>;
	}
};

export default recommendApi;
