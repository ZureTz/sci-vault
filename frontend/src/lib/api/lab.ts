import request from './request';
import type { DefaultResponse } from './request';

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

export interface LabDetailResponse {
	id: number;
	name: string;
	description?: string;
	invite_code: string;
	owner_id: number;
	member_count: number;
	my_role: 'owner' | 'member';
}

export interface LabMemberInfo {
	user_id: number;
	username: string;
	role: 'owner' | 'member';
	joined_at: string;
}

export interface TransferOwnershipRequest {
	target_user_id: number;
}

export interface LeaveLabRequest {
	email_code: string;
}

export interface DeleteLabRequest {
	confirm_name: string;
	email_code: string;
}

export interface ResetInviteCodeResponse {
	invite_code: string;
}

export interface UpdateLabInfoRequest {
	name: string;
	description?: string | null;
}

const labApi = {
	createLab(data: CreateLabRequest): Promise<JoinLabResponse> {
		return request.post('/labs', data) as unknown as Promise<JoinLabResponse>;
	},

	getMyLabs(): Promise<LabListItem[]> {
		return request.get('/labs') as unknown as Promise<LabListItem[]>;
	},

	joinLabByCode(data: JoinLabByCodeRequest): Promise<JoinLabResponse> {
		return request.post('/labs/join', data) as unknown as Promise<JoinLabResponse>;
	},

	getLab(labId: number): Promise<LabDetailResponse> {
		return request.get(`/labs/${labId}`) as unknown as Promise<LabDetailResponse>;
	},

	getMembers(labId: number): Promise<LabMemberInfo[]> {
		return request.get(`/labs/${labId}/members`) as unknown as Promise<LabMemberInfo[]>;
	},

	requestLeaveLab(labId: number): Promise<DefaultResponse> {
		return request.post(`/labs/${labId}/leave-request`) as unknown as Promise<DefaultResponse>;
	},

	leaveLab(labId: number, data: LeaveLabRequest): Promise<DefaultResponse> {
		return request.delete(`/labs/${labId}/members/me`, {
			data
		}) as unknown as Promise<DefaultResponse>;
	},

	kickMember(labId: number, userId: number): Promise<DefaultResponse> {
		return request.delete(
			`/labs/${labId}/members/${userId}`
		) as unknown as Promise<DefaultResponse>;
	},

	transferOwnership(labId: number, data: TransferOwnershipRequest): Promise<DefaultResponse> {
		return request.post(`/labs/${labId}/transfer`, data) as unknown as Promise<DefaultResponse>;
	},

	requestDeleteLab(labId: number): Promise<DefaultResponse> {
		return request.post(`/labs/${labId}/delete-request`) as unknown as Promise<DefaultResponse>;
	},

	deleteLab(labId: number, data: DeleteLabRequest): Promise<DefaultResponse> {
		return request.delete(`/labs/${labId}`, {
			data
		}) as unknown as Promise<DefaultResponse>;
	},

	resetInviteCode(labId: number): Promise<ResetInviteCodeResponse> {
		return request.post(
			`/labs/${labId}/invite-code/reset`
		) as unknown as Promise<ResetInviteCodeResponse>;
	},

	updateLabInfo(labId: number, data: UpdateLabInfoRequest): Promise<LabDetailResponse> {
		return request.patch(`/labs/${labId}`, data) as unknown as Promise<LabDetailResponse>;
	}
};

export default labApi;
