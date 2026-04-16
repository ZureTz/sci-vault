<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { Search, FileText, Tag, Users, LoaderCircle, Info } from 'lucide-svelte';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import documentApi from '$lib/api/document';
	import { getActiveLab } from '$lib/stores/lab.svelte';
	import { getSearchState, setSearchState, setSearchQuery } from '$lib/stores/search.svelte';
	import { showApiErrors } from '$lib/utils/api-error';

	let searchState = getSearchState();
	let isSearching = $state(false);

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
		<div class="flex flex-col items-center justify-center gap-3 py-16">
			<LoaderCircle class="size-8 animate-spin text-muted-foreground" />
			<p class="text-sm text-muted-foreground">{$_('search.searching')}</p>
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
						<div class="flex items-start justify-between gap-4">
							<div class="flex min-w-0 items-center gap-2">
								<FileText class="size-5 shrink-0 text-primary" />
								<Card.Title class="truncate text-base">
									{result.title || result.original_file_name}
								</Card.Title>
							</div>
							<Badge variant="secondary" class="shrink-0">
								{$_('search.similarity')}
								{formatSimilarity(result.similarity)}
							</Badge>
						</div>
						{#if result.title}
							<p class="truncate text-xs text-muted-foreground">{result.original_file_name}</p>
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
								<div class="flex items-center gap-1">
									<Tag class="size-3 text-muted-foreground" />
									{#each result.tags.slice(0, 5) as tag (tag)}
										<Badge variant="outline" class="text-xs">{tag}</Badge>
									{/each}
									{#if result.tags.length > 5}
										<span class="text-xs text-muted-foreground">
											+{result.tags.length - 5}
										</span>
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
