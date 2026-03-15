import request, { type DefaultResponse } from './request';

const authApi = {
	/**
	 * Test authenticated route (requires Bearer Token in the Authorization header)
	 */
	testAuth() {
		// The request interceptor in request.ts automatically attaches the Bearer token,
		// so no manual header configuration is needed here.
		return request.get<null, DefaultResponse>('/auth/test');
	}
};

export default authApi;
