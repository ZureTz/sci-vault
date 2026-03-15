import { goto } from '$app/navigation';
import { resolve } from '$app/paths';
import axios, { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios';

// Configure the base URL, which can be read from environment variables
const baseURL = '/api/v1'; // Default base URL for API requests

const request = axios.create({
	baseURL,
	timeout: 10000,
	headers: {
		'Content-Type': 'application/json'
	}
});

// Request interceptor: Add token
request.interceptors.request.use(
	(config: InternalAxiosRequestConfig) => {
		// Retrieve JWT token from localStorage
		const token = localStorage.getItem('token');
		if (token) {
			// If a token exists, include it in the Authorization header as a Bearer token
			config.headers.Authorization = `Bearer ${token}`;
		}
		return config;
	},
	(error) => {
		return Promise.reject(error);
	}
);

// Response interceptor: Handle unified error responses, etc.
request.interceptors.response.use(
	(response: AxiosResponse) => {
		// Process the response data based on the backend's data structure
		return response.data;
	},
	(error) => {
		// Unified error handling, e.g., redirect to login page for 401 Unauthorized
		if (error.response?.status === 401) {
			localStorage.removeItem('token');
			goto(resolve('/login'));
		}
		return Promise.reject(error);
	}
);

// ==== API request/response types ====
export interface DefaultResponse {
	message: string;
}

export default request;
