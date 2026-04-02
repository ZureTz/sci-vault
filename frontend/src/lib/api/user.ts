import request, { type DefaultResponse } from './request';

// Type definitions for request payloads and responses

export interface SendEmailCodeRequest {
	email: string;
}

export interface LoginRequest {
	// Mutex validation: either username or email must be provided, but not both
	// If both are provided, the username will be used for lookup. If only email is provided, it will be used for lookup.
	username?: string;
	email?: string;
	password: string;
}

export interface LoginResponse {
	user_id: string;
	username: string;
	email: string;
	token: string;
}

export interface RegisterRequest {
	username: string;
	email: string;
	password: string;
	confirmed_password: string;
	email_code: string;
}

export interface ResetPasswordRequest {
	email: string;
	email_code: string;
	password: string;
	confirmed_password: string;
}

export interface AvatarResponse {
	avatar_url: string;
}

export interface ProfileResponse {
	user_id: number;
	nickname: string | null;
	bio: string | null;
	avatar_url: string | null;
	website: string | null;
	location: string | null;
}

export interface UpdateProfileRequest {
	nickname: string | null;
	bio: string | null;
	website: string | null;
	location: string | null;
}

export interface UploadAvatarResponse {
	avatar_url: string;
}

// ==== API functions ====

const userApi = {
	/**
	 * Send an email verification code for registration or password reset
	 */
	sendEmailCode(data: SendEmailCodeRequest) {
		return request.post<SendEmailCodeRequest, DefaultResponse>('/user/send_email_code', data);
	},

	/**
	 * Log in a user
	 */
	login(data: LoginRequest) {
		return request.post<LoginRequest, LoginResponse>('/user/login', data);
	},

	/**
	 * Register a new user
	 */
	register(data: RegisterRequest) {
		return request.post<RegisterRequest, DefaultResponse>('/user/register', data);
	},

	/**
	 * Reset the user's password
	 */
	resetPassword(data: ResetPasswordRequest) {
		return request.post<ResetPasswordRequest, DefaultResponse>('/user/reset_password', data);
	},

	/**
	 * Get the avatar URL for the given user, or the current authenticated user if omitted.
	 */
	getAvatar(userId?: string | number): Promise<AvatarResponse> {
		const url = userId ? `/user/avatar/${userId}` : '/user/avatar';
		return request.get<AvatarResponse>(url) as unknown as Promise<AvatarResponse>;
	},

	/**
	 * Get the user's profile information. If userId is not provided, it will return the profile of the currently authenticated user.
	 */
	getProfile(userId?: string | number): Promise<ProfileResponse> {
		const url = userId ? `/user/profile/${userId}` : `/user/profile`;
		return request.get<ProfileResponse>(url) as unknown as Promise<ProfileResponse>;
	},

	/**
	 * Update the user's profile information
	 */
	updateProfile(data: UpdateProfileRequest) {
		return request.put<UpdateProfileRequest, DefaultResponse>('/user/profile', data);
	},

	/**
	 * Upload the user's avatar
	 */
	uploadAvatar(file: File) {
		const formData = new FormData();
		formData.append('avatar', file);
		return request.post<FormData, UploadAvatarResponse>('/user/upload_avatar', formData, {
			headers: {
				'Content-Type': 'multipart/form-data'
			}
		});
	}
};

export default userApi;
