<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { Users, RefreshCw, Inbox, FlaskConical } from 'lucide-svelte';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Avatar from '$lib/components/ui/avatar';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import recommendApi, { type CollaboratorItem } from '$lib/api/recommend';
	import { getActiveLab } from '$lib/stores/lab.svelte';
	import { showApiErrors } from '$lib/utils/api-error';

	const PAGE_LIMIT = 20;

	let items = $state<CollaboratorItem[]>([]);
	let isLoading = $state(false);
	let loadedOnce = $state(false);

	// Track the active lab by primitive id so re-equal-but-new lab objects
	// (e.g. after a labs reload) don't trigger spurious refetches.
	let activeLab = $derived(getActiveLab());
	let activeLabId = $derived(activeLab?.id ?? null);

	$effect(() => {
		const id = activeLabId;
		if (id !== null) {
			void load(id);
		} else {
			items = [];
			isLoading = false;
			loadedOnce = false;
		}
	});

	async function load(labId: number) {
		isLoading = true;
		try {
			const res = await recommendApi.getCollaborators({
				lab_id: labId,
				limit: PAGE_LIMIT
			});
			items = res.results;
		} catch (error: unknown) {
			items = [];
			showApiErrors(error, $_('collaborators.error.failed'));
		} finally {
			isLoading = false;
			loadedOnce = true;
		}
	}

	function refresh() {
		if (activeLabId !== null) void load(activeLabId);
	}

	function initials(username: string): string {
		return username.substring(0, 2).toUpperCase();
	}
</script>

<svelte:head>
	<title>{$_('collaborators.title')} | Sci-Vault</title>
</svelte:head>

<div class="container mx-auto flex flex-col gap-6 px-4 py-6 lg:px-8">
	<header class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
		<div class="flex flex-col gap-1">
			<h1 class="flex items-center gap-2 text-2xl font-semibold tracking-tight">
				<Users class="h-5 w-5 text-primary" />
				{$_('collaborators.title')}
			</h1>
			<p class="text-sm text-muted-foreground">
				{$_('collaborators.subtitle')}
			</p>
		</div>
		{#if activeLab}
			<Button variant="outline" size="sm" onclick={refresh} disabled={isLoading}>
				<RefreshCw class={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
				{$_('collaborators.refresh')}
			</Button>
		{/if}
	</header>

	{#if !activeLab}
		<Card.Root>
			<Card.Content class="flex flex-col items-center gap-3 py-16 text-center">
				<div
					class="flex size-14 items-center justify-center rounded-2xl bg-primary/10 ring-1 ring-border/50"
				>
					<FlaskConical class="size-7 text-primary" />
				</div>
				<h2 class="text-lg font-semibold">{$_('lab_dashboard.no_lab_selected')}</h2>
				<p class="max-w-md text-sm text-muted-foreground">
					{$_('collaborators.no_lab.description')}
				</p>
			</Card.Content>
		</Card.Root>
	{:else if isLoading && !loadedOnce}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each Array.from({ length: 6 }, (_, i) => i) as i (i)}
				<div class="flex items-center gap-4 rounded-lg border bg-card p-4">
					<Skeleton class="size-12 rounded-full" />
					<div class="flex-1 space-y-2">
						<Skeleton class="h-4 w-2/3" />
						<Skeleton class="h-3 w-1/2" />
					</div>
				</div>
			{/each}
		</div>
	{:else if items.length === 0}
		<Card.Root>
			<Card.Content class="flex flex-col items-center gap-3 py-16 text-center">
				<Inbox class="h-10 w-10 text-muted-foreground" />
				<h2 class="text-lg font-semibold">{$_('collaborators.empty.title')}</h2>
				<p class="max-w-md text-sm text-muted-foreground">
					{$_('collaborators.empty.description')}
				</p>
			</Card.Content>
		</Card.Root>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each items as item (item.user_id)}
				<button
					type="button"
					onclick={() => goto(resolve(`/profile/${item.user_id}`))}
					class="group flex items-center gap-4 rounded-lg border bg-card p-4 text-left transition-colors hover:border-primary/50 hover:bg-muted/40"
				>
					<Avatar.Root class="size-12 shrink-0">
						{#if item.avatar_url}
							<Avatar.Image src={item.avatar_url} alt={item.username} class="object-cover" />
						{/if}
						<Avatar.Fallback>{initials(item.username)}</Avatar.Fallback>
					</Avatar.Root>
					<div class="min-w-0 flex-1">
						<div class="flex items-center justify-between gap-2">
							<span class="truncate text-sm font-semibold group-hover:text-primary">
								{item.nickname || item.username}
							</span>
							<Badge variant="outline" class="shrink-0 gap-1 font-mono text-[10px]">
								{Math.round(item.similarity * 100)}%
							</Badge>
						</div>
						{#if item.nickname}
							<p class="truncate text-xs text-muted-foreground">@{item.username}</p>
						{/if}
						<p class="mt-1 text-[11px] text-muted-foreground">
							{$_('collaborators.signals', { values: { count: item.signal_count } })}
						</p>
					</div>
				</button>
			{/each}
		</div>
	{/if}
</div>
