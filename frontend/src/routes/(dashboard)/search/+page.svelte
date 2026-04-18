<script lang="ts">
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';
	import { Search, FileText, Tag, Users, LoaderCircle, Info, Clock, Trash2 } from 'lucide-svelte';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import searchApi, { MatchType, type SearchHistoryItem } from '$lib/api/search';
	import { getActiveLab } from '$lib/stores/lab.svelte';
	import {
		getSearchState,
		setSearchState,
		setSearchQuery,
		setSearchHistory
	} from '$lib/stores/search.svelte';
	import { showApiErrors } from '$lib/utils/api-error';
	import { SvelteSet } from 'svelte/reactivity';

	let searchState = getSearchState();
	let isSearching = $state(false);
	let expandedTags = new SvelteSet<number>();

	let clearDialogOpen = $state(false);
	let clearing = $state(false);

	// Autocomplete dropdown state
	let dropdownOpen = $state(false);
	let highlightedIndex = $state(-1);
	let inputEl = $state<HTMLInputElement | null>(null);

	let activeLab = $derived(getActiveLab());

	onMount(() => {
		void loadHistory();
	});

	async function loadHistory() {
		try {
			const resp = await searchApi.listHistory(10);
			setSearchHistory(resp.items ?? []);
		} catch (err) {
			showApiErrors(err, $_('search.history_empty'));
		}
	}

	async function runSearch(query: string) {
		const trimmed = query.trim();
		if (!trimmed) return;

		closeDropdown();
		isSearching = true;
		try {
			const resp = await searchApi.searchDocuments(trimmed, activeLab?.id ?? undefined);
			setSearchState(trimmed, resp.results ?? [], true);
			void loadHistory();
		} catch (err) {
			showApiErrors(err, $_('search.no_results'));
			setSearchState(trimmed, [], true);
		} finally {
			isSearching = false;
		}
	}

	function handleSearch() {
		void runSearch(searchState.query);
	}

	function openDropdown() {
		dropdownOpen = true;
		highlightedIndex = -1;
	}

	function closeDropdown() {
		dropdownOpen = false;
		highlightedIndex = -1;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			if (dropdownOpen) {
				e.preventDefault();
				closeDropdown();
			}
			return;
		}
		if (e.key === 'ArrowDown') {
			if (!dropdownOpen && searchState.filteredHistory.length > 0) {
				openDropdown();
				return;
			}
			if (searchState.filteredHistory.length === 0) return;
			e.preventDefault();
			highlightedIndex = (highlightedIndex + 1) % searchState.filteredHistory.length;
			return;
		}
		if (e.key === 'ArrowUp') {
			if (!dropdownOpen || searchState.filteredHistory.length === 0) return;
			e.preventDefault();
			highlightedIndex =
				highlightedIndex <= 0 ? searchState.filteredHistory.length - 1 : highlightedIndex - 1;
			return;
		}
		if (e.key === 'Enter') {
			if (
				dropdownOpen &&
				highlightedIndex >= 0 &&
				highlightedIndex < searchState.filteredHistory.length
			) {
				e.preventDefault();
				applyHistory(searchState.filteredHistory[highlightedIndex]);
				return;
			}
			handleSearch();
		}
	}

	function applyHistory(item: SearchHistoryItem) {
		setSearchQuery(item.query);
		void runSearch(item.query);
	}

	async function confirmClearHistory() {
		clearing = true;
		try {
			await searchApi.clearHistory();
			setSearchHistory([]);
			toast.success($_('search.history_cleared'));
			clearDialogOpen = false;
		} catch (err) {
			showApiErrors(err, $_('search.history_cleared'));
		} finally {
			clearing = false;
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

	// Compact relative time: "just now" / "5m" / "2h" / "3d" / else date.
	function formatRelative(iso: string): string {
		const then = new Date(iso).getTime();
		if (Number.isNaN(then)) return '';
		const diffSec = Math.max(0, Math.floor((Date.now() - then) / 1000));
		if (diffSec < 45) return $_('search.history_just_now', { default: 'just now' });
		if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m`;
		if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h`;
		if (diffSec < 604800) return `${Math.floor(diffSec / 86400)}d`;
		return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
	}
</script>

<svelte:head>
	<title>{$_('search.title')} | Sci-Vault</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div>
		<h1 class="text-2xl font-bold tracking-tight">{$_('search.title')}</h1>
		<p class="text-muted-foreground">{$_('search.description')}</p>
	</div>

	<!-- Search input with autocomplete dropdown -->
	<div class="flex gap-2">
		<div class="relative flex-1">
			<Search class="absolute top-1/2 left-3 z-10 size-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				bind:ref={inputEl}
				type="text"
				placeholder={$_('search.placeholder')}
				class="pl-10"
				value={searchState.query}
				oninput={(e) => {
					setSearchQuery(e.currentTarget.value);
					highlightedIndex = -1;
					if (!dropdownOpen) openDropdown();
				}}
				onfocus={() => openDropdown()}
				onclick={() => openDropdown()}
				onblur={() => {
					// Defer so a mousedown on a dropdown item still registers.
					setTimeout(closeDropdown, 120);
				}}
				onkeydown={handleKeydown}
				disabled={isSearching}
				autocomplete="off"
			/>

			{#if dropdownOpen && searchState.filteredHistory.length > 0}
				<div
					class="absolute top-full right-0 left-0 z-20 mt-1 overflow-hidden rounded-md border bg-popover text-popover-foreground shadow-md"
					role="listbox"
				>
					<ul class="max-h-80 overflow-y-auto py-1">
						{#each searchState.filteredHistory as item, i (item.id)}
							<li>
								<button
									type="button"
									role="option"
									aria-selected={highlightedIndex === i}
									class={`flex w-full items-center justify-between gap-3 px-3 py-2 text-left text-sm transition-colors ${highlightedIndex === i ? 'bg-accent text-accent-foreground' : 'hover:bg-accent/60'}`}
									onmousedown={(e) => {
										e.preventDefault();
										applyHistory(item);
									}}
									onmouseenter={() => (highlightedIndex = i)}
								>
									<div class="flex min-w-0 items-center gap-2">
										<Clock class="size-3.5 shrink-0 text-muted-foreground" />
										<span class="truncate" title={item.query}>{item.query}</span>
									</div>
									<div
										class="flex shrink-0 items-center gap-2 text-xs text-muted-foreground tabular-nums"
									>
										<span
											>{$_('search.history_result_count', {
												values: { count: item.result_count }
											})}</span
										>
										<span>{formatRelative(item.last_used_at)}</span>
									</div>
								</button>
							</li>
						{/each}
					</ul>
					<div class="flex items-center justify-end border-t bg-muted/30 px-2 py-1">
						<button
							type="button"
							class="flex items-center gap-1.5 rounded px-2 py-1 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-destructive"
							onmousedown={(e) => {
								e.preventDefault();
								closeDropdown();
								clearDialogOpen = true;
							}}
						>
							<Trash2 class="size-3" />
							{$_('search.history_clear')}
						</button>
					</div>
				</div>
			{/if}
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

<AlertDialog.Root bind:open={clearDialogOpen}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>{$_('search.history_clear_confirm_title')}</AlertDialog.Title>
			<AlertDialog.Description>
				{$_('search.history_clear_confirm_description')}
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={clearing}
				>{$_('search.history_clear_cancel')}</AlertDialog.Cancel
			>
			<AlertDialog.Action
				disabled={clearing}
				onclick={(e: MouseEvent) => {
					e.preventDefault();
					void confirmClearHistory();
				}}
			>
				<Trash2 class="size-3.5" />
				{$_('search.history_clear_confirm_action')}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
