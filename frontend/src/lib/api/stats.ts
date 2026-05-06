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

export interface DayCount {
	date: string;
	count: number;
}

export interface FormatBucket {
	content_type: string;
	count: number;
}

export interface TopDocument {
	id: number;
	title: string | null;
	original_file_name: string;
	view_count: number;
	like_count: number;
}

export interface Contributor {
	user_id: number;
	username: string;
	nickname: string | null;
	avatar_url: string | null;
	doc_count: number;
}

export interface DashboardStatsResponse {
	total_documents: number;
	total_storage: number;
	total_views: number;
	total_likes: number;
	status_breakdown: StatusBreakdown;
	recent_documents: RecentDocument[];
	uploads_by_day: DayCount[];
	views_by_day: DayCount[];
	likes_by_day: DayCount[];
	format_distribution: FormatBucket[];
	top_viewed: TopDocument[];
}

export interface LabDashboardStatsResponse {
	total_documents: number;
	total_storage: number;
	total_views: number;
	total_likes: number;
	member_count: number;
	status_breakdown: StatusBreakdown;
	recent_documents: RecentDocument[];
	uploads_by_day: DayCount[];
	views_by_day: DayCount[];
	likes_by_day: DayCount[];
	format_distribution: FormatBucket[];
	top_contributors: Contributor[];
}

const statsApi = {
	getDashboardStats(): Promise<DashboardStatsResponse> {
		return request.get<DashboardStatsResponse>(
			'/stats/mine/dashboard'
		) as unknown as Promise<DashboardStatsResponse>;
	},
	getLabDashboardStats(labID: number): Promise<LabDashboardStatsResponse> {
		return request.get<LabDashboardStatsResponse>(
			`/stats/labs/${labID}/dashboard`
		) as unknown as Promise<LabDashboardStatsResponse>;
	}
};

export default statsApi;
