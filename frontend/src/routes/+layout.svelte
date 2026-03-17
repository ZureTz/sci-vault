<script lang="ts">
	import { ModeWatcher } from 'mode-watcher';

	import favicon from '$lib/assets/favicon.svg';
	import ThemeToggle from '$lib/components/layout/ThemeToggle.svelte';
	import LanguageToggle from '$lib/components/layout/LanguageToggle.svelte';
	import { Toaster } from '$lib/components/ui/sonner';
	import './layout.css';

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<ModeWatcher />
<Toaster position="bottom-center" richColors />

<div class="fixed top-4 right-4 z-50 flex items-center gap-2">
	<ThemeToggle />
	<LanguageToggle />
</div>

{@render children()}
