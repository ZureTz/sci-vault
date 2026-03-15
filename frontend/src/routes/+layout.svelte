<script lang="ts">
	import { browser } from '$app/environment';
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { Button } from '$lib/components/ui/button';
	import { Languages, Moon, Sun } from 'lucide-svelte';
	import { tick } from 'svelte';
	import { _, locale, waitLocale } from 'svelte-i18n';

	let { children } = $props();

	// Theme management

	// eslint-disable-next-line svelte/prefer-writable-derived
	let isDark = $state(false);

	$effect(() => {
		isDark = document.documentElement.classList.contains('dark');
	});

	const toggleTheme = () => {
		const nextDark = !isDark;

		const applyTheme = () => {
			isDark = nextDark;
			if (nextDark) {
				document.documentElement.classList.add('dark');
				localStorage.theme = 'dark';
			} else {
				document.documentElement.classList.remove('dark');
				localStorage.theme = 'light';
			}
		};

		if (document.startViewTransition) {
			document.startViewTransition(async () => {
				applyTheme();
				await tick();
			});
		} else {
			applyTheme();
		}
	};

	const toggleLocale = async () => {
		const nextLocale = $locale === 'zh-CN' ? 'en' : 'zh-CN';

		if (browser) {
			localStorage.setItem('locale', nextLocale);
		}

		if (document.startViewTransition) {
			document.startViewTransition(async () => {
				locale.set(nextLocale);
				await waitLocale();
				await tick();
			});
		} else {
			locale.set(nextLocale);
		}
	};
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="fixed top-4 right-4 z-50 flex items-center gap-2">
	<Button
		variant="outline"
		size="icon"
		class="bg-background/80 backdrop-blur"
		onclick={toggleTheme}
		aria-label={$_('app.toggle_theme')}
	>
		{#if isDark}
			<Moon class="size-4" />
		{:else}
			<Sun class="size-4" />
		{/if}
	</Button>

	<Button
		variant="outline"
		size="sm"
		class="gap-2 bg-background/80 backdrop-blur"
		onclick={toggleLocale}
		aria-label={$_('app.switch_language')}
	>
		<Languages class="size-4" />
		<span>{$locale === 'zh-CN' ? 'ZH' : 'EN'}</span>
	</Button>
</div>

{@render children()}
