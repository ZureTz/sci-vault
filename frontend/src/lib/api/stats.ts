import request from './request';

export interface StatusBreakdown {
	not_started: number;
	pending: number;
	processing: number;
	done: number;
	failed: number;
}

export interface RecentDocument {
	id: number;
	title: string | null;
	original_file_name: string;
	file_size: number;
	enrich_status: string;
	created_at: string;
}

export interface DashboardStatsResponse {
	total_documents: number;
	total_storage: number;
	total_views: number;
	status_breakdown: StatusBreakdown;
	recent_documents: RecentDocument[];
}

const statsApi = {
	getDashboardStats(): Promise<DashboardStatsResponse> {
		return request.get<DashboardStatsResponse>(
			'/stats/dashboard'
		) as unknown as Promise<DashboardStatsResponse>;
	}
};

export default statsApi;
