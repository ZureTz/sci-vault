import request from './request';

export interface CreateLabRequest {
	name: string;
	description?: string;
}

export interface JoinLabByCodeRequest {
	invite_code: string;
}

export interface JoinLabResponse {
	id: number;
	name: string;
	description?: string;
	invite_code: string;
	owner_id: number;
	member_count: number;
}

export interface LabListItem {
	id: number;
	name: string;
	description?: string;
	owner_id: number;
	member_count: number;
	role: 'owner' | 'member';
}

const labApi = {
	createLab(data: CreateLabRequest): Promise<JoinLabResponse> {
		return request.post<JoinLabResponse>('/labs', data) as unknown as Promise<JoinLabResponse>;
	},

	getMyLabs(): Promise<LabListItem[]> {
		return request.get<LabListItem[]>('/labs') as unknown as Promise<LabListItem[]>;
	},

	joinLabByCode(data: JoinLabByCodeRequest): Promise<JoinLabResponse> {
		return request.post<JoinLabResponse>('/labs/join', data) as unknown as Promise<JoinLabResponse>;
	}
};

export default labApi;
