import { browser } from '$app/environment';
import { init, register, getLocaleFromNavigator } from 'svelte-i18n';

register('en', () => import('./locales/en.json'));
register('zh-CN', () => import('./locales/zh-CN.json'));

const SUPPORTED_LOCALES = ['en', 'zh-CN'] as const;
type SupportedLocale = (typeof SUPPORTED_LOCALES)[number];

const normalizeLocale = (locale: string | null | undefined): SupportedLocale => {
	if (!locale) {
		return 'en';
	}

	const normalizedLocale = locale.toLowerCase();

	if (normalizedLocale.startsWith('zh')) {
		return 'zh-CN';
	}

	if (normalizedLocale.startsWith('en')) {
		return 'en';
	}

	return 'en';
};

const getInitialLocale = (): SupportedLocale => {
	if (!browser) {
		return 'en';
	}

	const savedLocale = localStorage.getItem('locale');

	if (savedLocale && SUPPORTED_LOCALES.includes(savedLocale as SupportedLocale)) {
		return savedLocale as SupportedLocale;
	}

	return normalizeLocale(getLocaleFromNavigator());
};

init({
	fallbackLocale: 'en',
	initialLocale: getInitialLocale()
});
