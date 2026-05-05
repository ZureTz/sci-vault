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

export interface PersonalizedRecommendationsResponse {
	results: SimilarDocumentItem[];
}

export interface CollaboratorItem {
	user_id: number;
	username: string;
	nickname: string;
	avatar_url: string | null;
	similarity: number;
	signal_count: number;
}

export interface CollaboratorRecommendationsResponse {
	results: CollaboratorItem[];
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
	},
	getForUser(
		opts: { lab_id?: number; limit?: number } = {}
	): Promise<PersonalizedRecommendationsResponse> {
		const params: Record<string, string | number> = {};
		if (opts.lab_id != null) params.lab_id = opts.lab_id;
		if (opts.limit != null) params.limit = opts.limit;
		return request.get<PersonalizedRecommendationsResponse>('/mine/recommendations', {
			params: params,
			timeout: 30000
		}) as unknown as Promise<PersonalizedRecommendationsResponse>;
	},
	getCollaborators(opts: {
		lab_id: number;
		limit?: number;
	}): Promise<CollaboratorRecommendationsResponse> {
		const params: Record<string, string | number> = { lab_id: opts.lab_id };
		if (opts.limit != null) params.limit = opts.limit;
		return request.get<CollaboratorRecommendationsResponse>('/mine/collaborators', {
			params,
			timeout: 30000
		}) as unknown as Promise<CollaboratorRecommendationsResponse>;
	}
};

export default recommendApi;
