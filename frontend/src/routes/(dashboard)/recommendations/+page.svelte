<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { Sparkles, RefreshCw, Inbox } from 'lucide-svelte';

	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import recommendApi, { type SimilarDocumentItem } from '$lib/api/recommend';
	import { getActiveLab } from '$lib/stores/lab.svelte';
	import { showApiErrors } from '$lib/utils/api-error';

	const PAGE_LIMIT = 30;

	let items = $state<SimilarDocumentItem[]>([]);
	let isLoading = $state(true);
	let loadedOnce = $state(false);

	// Track the active lab by primitive id so re-equal-but-new lab objects
	// (e.g. after a labs reload) don't trigger spurious refetches.
	let activeLabId = $derived(getActiveLab()?.id ?? null);

	$effect(() => {
		const id = activeLabId;
		void load(id);
	});

	async function load(labId: number | null) {
		isLoading = true;
		try {
			const res = await recommendApi.getForUser({
				lab_id: labId ?? undefined,
				limit: PAGE_LIMIT
			});
			items = res.results;
		} catch (error: unknown) {
			items = [];
			showApiErrors(error, $_('recommendations.error.failed'));
		} finally {
			isLoading = false;
			loadedOnce = true;
		}
	}

	function refresh() {
		void load(activeLabId);
	}
</script>

<div class="container mx-auto flex flex-col gap-6 px-4 py-6 lg:px-8">
	<header class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
		<div class="flex flex-col gap-1">
			<h1 class="flex items-center gap-2 text-2xl font-semibold tracking-tight">
				<Sparkles class="h-5 w-5 text-primary" />
				{$_('recommendations.title')}
			</h1>
			<p class="text-sm text-muted-foreground">
				{$_('recommendations.subtitle')}
			</p>
		</div>
		<Button variant="outline" size="sm" onclick={refresh} disabled={isLoading}>
			<RefreshCw class={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
			{$_('recommendations.refresh')}
		</Button>
	</header>

	{#if isLoading && !loadedOnce}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each Array.from({ length: 6 }, (_, i) => i) as i (i)}
				<div class="space-y-3 rounded-lg border bg-card p-4">
					<Skeleton class="h-4 w-3/4" />
					<Skeleton class="h-3 w-full" />
					<Skeleton class="h-3 w-full" />
					<Skeleton class="h-3 w-5/6" />
				</div>
			{/each}
		</div>
	{:else if items.length === 0}
		<Card.Root>
			<Card.Content class="flex flex-col items-center gap-3 py-16 text-center">
				<Inbox class="h-10 w-10 text-muted-foreground" />
				<h2 class="text-lg font-semibold">{$_('recommendations.empty.title')}</h2>
				<p class="max-w-md text-sm text-muted-foreground">
					{$_('recommendations.empty.description')}
				</p>
			</Card.Content>
		</Card.Root>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each items as item (item.doc_id)}
				<a
					href={resolve(`/documents/${item.doc_id}`)}
					class="group flex flex-col gap-2 rounded-lg border bg-card p-4 transition-colors hover:border-primary/50 hover:bg-muted/40"
				>
					<div class="flex items-start justify-between gap-2">
						<h3 class="line-clamp-2 text-sm leading-snug font-semibold group-hover:text-primary">
							{item.title || item.original_file_name}
						</h3>
						<Badge variant="outline" class="shrink-0 gap-1 font-mono text-[10px]">
							{Math.round(item.similarity * 100)}%
						</Badge>
					</div>
					{#if item.summary}
						<p class="line-clamp-3 text-xs text-muted-foreground">
							{item.summary}
						</p>
					{/if}
					{#if item.authors && item.authors.length > 0}
						<p class="line-clamp-1 text-[11px] text-muted-foreground">
							{item.authors.slice(0, 3).join(', ')}
						</p>
					{/if}
					{#if item.tags && item.tags.length > 0}
						<div class="flex flex-wrap gap-1 pt-1">
							{#each item.tags.slice(0, 4) as tag (tag)}
								<Badge variant="secondary" class="text-[10px] font-normal">{tag}</Badge>
							{/each}
						</div>
					{/if}
				</a>
			{/each}
		</div>
	{/if}
</div>
