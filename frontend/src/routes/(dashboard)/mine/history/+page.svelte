<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { resolve } from '$app/paths';
	import {
		Eye,
		Heart,
		FileText,
		FlaskConical,
		Lock,
		ChevronLeft,
		ChevronRight
	} from 'lucide-svelte';

	import * as Card from '$lib/components/ui/card';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';

	import interactionApi, { type HistoryItem, type ListHistoryResponse } from '$lib/api/interaction';
	import { showApiErrors } from '$lib/utils/api-error';

	type Tab = 'views' | 'likes';

	const PAGE_SIZE = 20;

	let activeTab = $state<Tab>('views');

	// Cache per-tab so flipping back doesn't refetch unless the user pages.
	let viewsState = $state({ items: [] as HistoryItem[], total: 0, page: 1, loading: true });
	let likesState = $state({ items: [] as HistoryItem[], total: 0, page: 1, loading: true });

	let current = $derived(activeTab === 'views' ? viewsState : likesState);
	let totalPages = $derived(Math.max(1, Math.ceil(current.total / PAGE_SIZE)));

	async function load(tab: Tab, page: number) {
		const target = tab === 'views' ? viewsState : likesState;
		target.loading = true;
		try {
			const fn = tab === 'views' ? interactionApi.listViewHistory : interactionApi.listLikeHistory;
			const res: ListHistoryResponse = await fn({ page, page_size: PAGE_SIZE });
			target.items = res.items;
			target.total = res.total;
			target.page = res.page;
		} catch (error: unknown) {
			showApiErrors(error, $_('history.error'));
		} finally {
			target.loading = false;
		}
	}

	function go(delta: number) {
		const next = current.page + delta;
		if (next < 1 || next > totalPages) return;
		load(activeTab, next);
	}

	function formatDate(dateStr: string): string {
		const d = new Date(dateStr);
		return d.toLocaleString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	$effect(() => {
		const tab = activeTab;
		const target = tab === 'views' ? viewsState : likesState;
		if (target.items.length === 0 && target.loading) {
			load(tab, 1);
		}
	});
</script>

<svelte:head>
	<title>{$_('history.title')} | Sci-Vault</title>
</svelte:head>

<div class="mx-auto w-full max-w-5xl space-y-6">
	<div>
		<h1 class="text-2xl font-bold tracking-tight">{$_('history.title')}</h1>
		<p class="text-sm text-muted-foreground">{$_('history.description')}</p>
	</div>

	<Tabs.Root value={activeTab} onValueChange={(v) => (activeTab = v as Tab)}>
		<Tabs.List class="grid w-full max-w-md grid-cols-2">
			<Tabs.Trigger value="views" class="gap-2">
				<Eye class="h-4 w-4" />
				{$_('history.tab.views')}
			</Tabs.Trigger>
			<Tabs.Trigger value="likes" class="gap-2">
				<Heart class="h-4 w-4" />
				{$_('history.tab.likes')}
			</Tabs.Trigger>
		</Tabs.List>

		{#each ['views', 'likes'] as const as tab (tab)}
			<Tabs.Content value={tab} class="space-y-4">
				{@const state = tab === 'views' ? viewsState : likesState}
				<Card.Root>
					<Card.Content class="p-0">
						{#if state.loading && state.items.length === 0}
							<div class="space-y-3 p-4">
								{#each Array.from({ length: 4 }, (_, i) => i) as i (i)}
									<div class="flex items-center gap-3">
										<Skeleton class="h-10 w-10 rounded" />
										<div class="flex-1 space-y-2">
											<Skeleton class="h-4 w-3/4" />
											<Skeleton class="h-3 w-1/3" />
										</div>
									</div>
								{/each}
							</div>
						{:else if state.items.length === 0}
							<div class="px-4 py-12 text-center text-sm text-muted-foreground">
								{tab === 'views' ? $_('history.empty.views') : $_('history.empty.likes')}
							</div>
						{:else}
							<ul class="divide-y">
								{#each state.items as item (item.interaction_id)}
									<li>
										<a
											href={resolve(`/documents/${item.doc_id}`)}
											class="flex items-center gap-3 p-4 transition-colors hover:bg-muted/40"
										>
											<FileText class="h-5 w-5 shrink-0 text-muted-foreground" />
											<div class="min-w-0 flex-1">
												<div class="flex items-center gap-2">
													<span class="truncate font-medium">
														{item.title || item.original_file_name}
													</span>
													{#if item.visibility === 'lab'}
														<Badge
															variant="outline"
															class="max-w-40 gap-1 border-blue-500/30 bg-blue-500/10 text-blue-700 dark:text-blue-400"
														>
															<FlaskConical class="size-3 shrink-0" />
															<span class="truncate" title={item.lab_name ?? ''}>
																{item.lab_name ?? $_('document.mine.visibility.lab')}
															</span>
														</Badge>
													{:else}
														<Badge variant="secondary" class="gap-1">
															<Lock class="size-3" />
															{$_('document.mine.visibility.private')}
														</Badge>
													{/if}
												</div>
												<div class="mt-0.5 text-xs text-muted-foreground">
													{tab === 'views'
														? $_('history.last_viewed_at', {
																values: { time: formatDate(item.interacted_at) }
															})
														: $_('history.liked_at', {
																values: { time: formatDate(item.interacted_at) }
															})}
												</div>
											</div>
										</a>
									</li>
								{/each}
							</ul>
						{/if}
					</Card.Content>
				</Card.Root>

				{#if state.total > PAGE_SIZE}
					<div class="flex items-center justify-between text-sm text-muted-foreground">
						<span>
							{$_('history.pagination.summary', {
								values: { page: state.page, total: totalPages, count: state.total }
							})}
						</span>
						<div class="flex items-center gap-2">
							<Button
								variant="outline"
								size="sm"
								onclick={() => go(-1)}
								disabled={state.loading || state.page <= 1}
							>
								<ChevronLeft class="h-4 w-4" />
								{$_('history.pagination.prev')}
							</Button>
							<Button
								variant="outline"
								size="sm"
								onclick={() => go(1)}
								disabled={state.loading || state.page >= totalPages}
							>
								{$_('history.pagination.next')}
								<ChevronRight class="h-4 w-4" />
							</Button>
						</div>
					</div>
				{/if}
			</Tabs.Content>
		{/each}
	</Tabs.Root>
</div>
