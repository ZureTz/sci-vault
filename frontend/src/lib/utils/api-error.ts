import axios from 'axios';
import { get } from 'svelte/store';
import { _ } from 'svelte-i18n';
import { toast } from 'svelte-sonner';
import type { ApiErrorResponse } from '$lib/api/request';

/**
 * Show a single toast for API errors.
 * Each item in `errors` is treated as a svelte-i18n locale key and translated.
 * Multiple errors are joined with newlines into one toast message.
 * Unknown keys fall back to the raw string. Falls back to `fallback` when no errors are present.
 */
export function showApiErrors(error: unknown, fallback: string): void {
	const t = get(_);
	if (axios.isAxiosError(error)) {
		const data = error.response?.data as ApiErrorResponse | undefined;
		if (data?.errors && data.errors.length > 0) {
			const msg = data.errors.map((key) => t(key, { default: key })).join('\n');
			toast.error(msg);
			return;
		}
	} else if (error instanceof Error && error.message) {
		toast.error(error.message);
		return;
	}
	toast.error(fallback);
}
