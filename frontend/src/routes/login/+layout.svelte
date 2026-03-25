<script lang="ts">
	import { Activity } from 'lucide-svelte';
	import { _ } from 'svelte-i18n';
	import ThemeToggle from '$lib/components/layout/ThemeToggle.svelte';
	import LanguageToggle from '$lib/components/layout/LanguageToggle.svelte';

	let { children } = $props();
</script>

<div class="flex min-h-screen flex-col">
	<header
		class="sticky top-0 z-10 flex h-16 shrink-0 items-center gap-2 border-b bg-background px-4"
	>
		<div class="flex flex-1 items-center gap-2">
			<div
				class="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"
			>
				<Activity class="size-4" />
			</div>
			<div class="flex flex-col gap-0.5 leading-none">
				<span class="font-semibold">{$_('app.title')}</span>
				<span class="text-xs text-muted-foreground">{$_('app.version')}</span>
			</div>
		</div>
		<div class="flex items-center gap-2">
			<ThemeToggle />
			<LanguageToggle />
		</div>
	</header>

	{@render children()}
</div>
