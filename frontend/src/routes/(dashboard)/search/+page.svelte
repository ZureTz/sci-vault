<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { Search, FileText, Tag, Users, LoaderCircle, Info } from 'lucide-svelte';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import documentApi, { MatchType } from '$lib/api/document';
	import { getActiveLab } from '$lib/stores/lab.svelte';
	import { getSearchState, setSearchState, setSearchQuery } from '$lib/stores/search.svelte';
	import { showApiErrors } from '$lib/utils/api-error';
	import { SvelteSet } from 'svelte/reactivity';

	let searchState = getSearchState();
	let isSearching = $state(false);
	let expandedTags = new SvelteSet<number>();

	let activeLab = $derived(getActiveLab());

	async function handleSearch() {
		const trimmed = searchState.query.trim();
		if (!trimmed) return;

		isSearching = true;
		try {
			const resp = await documentApi.searchDocuments(trimmed, activeLab?.id ?? undefined);
			setSearchState(trimmed, resp.results ?? [], true);
		} catch (err) {
			showApiErrors(err, $_('search.no_results'));
			setSearchState(searchState.query, [], true);
		} finally {
			isSearching = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			handleSearch();
		}
	}

	function toggleTags(docId: number, e: MouseEvent) {
		e.stopPropagation();
		if (expandedTags.has(docId)) {
			expandedTags.delete(docId);
		} else {
			expandedTags.add(docId);
		}
	}

	function formatSimilarity(score: number): string {
		return `${Math.round(score * 100)}%`;
	}
</script>

<div class="flex flex-col gap-6">
	<div>
		<h1 class="text-2xl font-bold tracking-tight">{$_('search.title')}</h1>
		<p class="text-muted-foreground">{$_('search.description')}</p>
	</div>

	<!-- Search input -->
	<div class="flex gap-2">
		<div class="relative flex-1">
			<Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="text"
				placeholder={$_('search.placeholder')}
				class="pl-10"
				value={searchState.query}
				oninput={(e) => setSearchQuery(e.currentTarget.value)}
				onkeydown={handleKeydown}
				disabled={isSearching}
			/>
		</div>
		<Button onclick={handleSearch} disabled={isSearching || !searchState.query.trim()}>
			{#if isSearching}
				<LoaderCircle class="size-4 animate-spin" />
				{$_('search.searching')}
			{:else}
				<Search class="size-4" />
				{$_('search.submit')}
			{/if}
		</Button>
	</div>

	<!-- Lab scope indicator -->
	{#if activeLab}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<Info class="size-4" />
			<span>{$_('search.lab_scope', { values: { lab: activeLab.name } })}</span>
		</div>
	{:else}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<Info class="size-4" />
			<span>{$_('search.no_lab_hint')}</span>
		</div>
	{/if}

	<!-- Results -->
	{#if isSearching}
		<div class="relative flex flex-col gap-3">
			<div
				class="absolute inset-0 z-10 flex items-center justify-center rounded-xl bg-background/40 backdrop-blur-[2px]"
			>
				<div class="flex items-center gap-3 rounded-lg border bg-background px-5 py-3 shadow-lg">
					<LoaderCircle class="size-5 animate-spin text-primary" />
					<p class="text-sm font-medium">{$_('search.searching')}</p>
				</div>
			</div>
			{#each Array(3).keys() as index (index)}
				<Card.Root>
					<Card.Header class="pb-2">
						<div
							class="flex min-w-0 flex-col gap-2 overflow-hidden sm:flex-row sm:items-center sm:justify-between sm:gap-4"
						>
							<div class="flex min-w-0 items-center gap-2 overflow-hidden">
								<Skeleton class="size-5 shrink-0 rounded-full" />
								<Skeleton class="h-6 w-3/4 max-w-100" />
							</div>
							<div class="flex shrink-0 items-center gap-1.5">
								<Skeleton class="h-5 w-24" />
								<Skeleton class="h-5 w-16" />
							</div>
						</div>
						<Skeleton class="mt-2 h-4 w-1/4 max-w-50" />
					</Card.Header>
					<Card.Content class="pb-3">
						<div class="flex flex-col gap-2">
							<Skeleton class="h-4 w-full" />
							<Skeleton class="h-4 w-5/6" />
						</div>
						<div class="mt-4 flex flex-wrap items-center gap-2">
							<Skeleton class="h-5 w-24" />
							<Skeleton class="h-5 w-16" />
							<Skeleton class="h-5 w-16" />
						</div>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{:else if searchState.hasSearched && searchState.results.length === 0}
		<div class="flex flex-col items-center justify-center gap-2 py-16 text-center">
			<Search class="size-10 text-muted-foreground" />
			<p class="font-medium">{$_('search.no_results')}</p>
			<p class="text-sm text-muted-foreground">{$_('search.no_results_hint')}</p>
		</div>
	{:else if searchState.results.length > 0}
		<p class="text-sm text-muted-foreground">
			{$_('search.result_count', { values: { count: searchState.results.length } })}
		</p>
		<div class="flex flex-col gap-3">
			{#each searchState.results as result (result.doc_id)}
				<Card.Root
					class="cursor-pointer transition-colors hover:bg-muted/50"
					onclick={() => goto(resolve(`/documents/${result.doc_id}`))}
				>
					<Card.Header class="pb-2">
						<div
							class="flex min-w-0 flex-col gap-2 overflow-hidden sm:flex-row sm:items-center sm:justify-between sm:gap-4"
						>
							<div class="flex min-w-0 items-center gap-2 overflow-hidden">
								<FileText class="size-5 shrink-0 text-primary" />
								<Card.Title
									class="truncate text-base"
									title={result.title || result.original_file_name}
								>
									{result.title || result.original_file_name}
								</Card.Title>
							</div>
							<div class="flex shrink-0 flex-wrap items-center gap-1.5">
								{#if result.match_type === MatchType.KEYWORD}
									<Badge
										variant="outline"
										class="border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-400"
									>
										{$_('search.match_keyword')}
									</Badge>
								{:else}
									<Badge
										variant="outline"
										class="border-green-500/30 bg-green-500/10 text-green-700 dark:text-green-400"
									>
										{$_('search.match_semantic')}
									</Badge>
								{/if}
								{#if result.match_type !== MatchType.KEYWORD}
									<Badge variant="secondary">
										{$_('search.similarity')}
										{formatSimilarity(result.similarity)}
									</Badge>
								{/if}
							</div>
						</div>
						{#if result.title}
							<p class="min-w-0 truncate text-xs text-muted-foreground">
								{result.original_file_name}
							</p>
						{/if}
					</Card.Header>
					<Card.Content class="pb-3">
						{#if result.summary}
							<p class="line-clamp-2 text-sm text-muted-foreground">{result.summary}</p>
						{/if}
						<div class="mt-2 flex flex-wrap items-center gap-2">
							{#if result.authors && result.authors.length > 0}
								<div class="flex items-center gap-1 text-xs text-muted-foreground">
									<Users class="size-3" />
									<span>{result.authors.join(', ')}</span>
								</div>
							{/if}
							{#if result.tags && result.tags.length > 0}
								<div class="mt-1 flex flex-wrap items-center gap-1.5 sm:mt-0">
									<Tag class="mr-1 size-3 text-muted-foreground" />
									{#if expandedTags.has(result.doc_id)}
										{#each result.tags as tag (tag)}
											<Badge variant="outline" class="text-xs">{tag}</Badge>
										{/each}
										<button
											class="ml-1 text-xs text-muted-foreground hover:text-foreground hover:underline"
											onclick={(e) => toggleTags(result.doc_id, e)}
										>
											{$_('search.show_less', { default: 'Show less' })}
										</button>
									{:else}
										{#each result.tags.slice(0, 5) as tag (tag)}
											<Badge variant="outline" class="text-xs">{tag}</Badge>
										{/each}
										{#if result.tags.length > 5}
											<button
												class="ml-1 text-xs text-muted-foreground hover:text-foreground hover:underline"
												onclick={(e) => toggleTags(result.doc_id, e)}
											>
												+{result.tags.length - 5}
												{$_('search.more', { default: 'more' })}
											</button>
										{/if}
									{/if}
								</div>
							{/if}
						</div>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{/if}
</div>
