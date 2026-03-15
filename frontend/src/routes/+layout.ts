import { browser } from '$app/environment';
import '$lib/i18n';
import { waitLocale } from 'svelte-i18n';

// Use SSG only
export const prerender = true;

export const load = async () => {
	if (browser) {
		await waitLocale();
	}
};
