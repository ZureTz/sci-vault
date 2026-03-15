<script lang="ts">
	import { browser } from '$app/environment';
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { Button } from '$lib/components/ui/button';
	import { Languages } from 'lucide-svelte';
	import { tick } from 'svelte';
	import { _, locale, waitLocale } from 'svelte-i18n';

	let { children } = $props();

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

<div class="fixed top-4 right-4 z-50">
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
