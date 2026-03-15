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
	}
};

export default userApi;
