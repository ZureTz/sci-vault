<script lang="ts">
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import {
		FileText,
		HardDrive,
		Eye,
		CircleCheck,
		LoaderCircle,
		CircleAlert,
		Clock,
		Upload,
		ArrowRight
	} from 'lucide-svelte';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Separator } from '$lib/components/ui/separator';
	import statsApi, { type DashboardStatsResponse } from '$lib/api/stats';
	import { showApiErrors } from '$lib/utils/api-error';

	let stats = $state<DashboardStatsResponse | null>(null);
	let isLoading = $state(true);

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
		return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	async function loadStats() {
		if (!localStorage.getItem('token')) return;
		isLoading = true;
		try {
			stats = await statsApi.getDashboardStats();
		} catch (error: unknown) {
			showApiErrors(error, $_('dashboard.error'));
		} finally {
			isLoading = false;
		}
	}

	onMount(loadStats);
</script>

<svelte:head>
	<title>{$_('dashboard.title')} | Sci-Vault</title>
</svelte:head>

<div class="container mx-auto max-w-5xl px-4 py-8">
	<!-- Header -->
	<div class="mb-8 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight">{$_('dashboard.title')}</h1>
			<p class="mt-1 text-sm text-muted-foreground">{$_('dashboard.description')}</p>
		</div>
		<Button onclick={() => goto(resolve('/documents/upload'))}>
			<Upload class="h-4 w-4" />
			{$_('dashboard.quick_upload')}
		</Button>
	</div>

	<!-- Stat Cards -->
	{#if isLoading}
		<div class="mb-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{#each Array.from({ length: 4 }, (_, i) => i) as i (i)}
				<Card.Root>
					<Card.Content class="p-6">
						<div class="flex items-center justify-between">
							<Skeleton class="h-4 w-24" />
							<Skeleton class="h-8 w-8 rounded-md" />
						</div>
						<Skeleton class="mt-3 h-8 w-16" />
						<Skeleton class="mt-1 h-3 w-32" />
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{:else if stats}
		<div class="mb-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<!-- Total Documents -->
			<Card.Root>
				<Card.Content class="p-6">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-muted-foreground"
							>{$_('dashboard.stats.total_documents')}</span
						>
						<div
							class="flex h-8 w-8 items-center justify-center rounded-md bg-primary/10 text-primary"
						>
							<FileText class="h-4 w-4" />
						</div>
					</div>
					<div class="mt-3 text-2xl font-bold">{stats.total_documents}</div>
					<p class="mt-1 text-xs text-muted-foreground">
						{$_('dashboard.stats.enriched', {
							values: { count: stats.status_breakdown.done }
						})}
					</p>
				</Card.Content>
			</Card.Root>

			<!-- Storage Used -->
			<Card.Root>
				<Card.Content class="p-6">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-muted-foreground"
							>{$_('dashboard.stats.storage_used')}</span
						>
						<div
							class="flex h-8 w-8 items-center justify-center rounded-md bg-blue-500/10 text-blue-600 dark:text-blue-400"
						>
							<HardDrive class="h-4 w-4" />
						</div>
					</div>
					<div class="mt-3 text-2xl font-bold">{formatFileSize(stats.total_storage)}</div>
					<p class="mt-1 text-xs text-muted-foreground">
						{$_('dashboard.stats.across_documents', {
							values: { count: stats.total_documents }
						})}
					</p>
				</Card.Content>
			</Card.Root>

			<!-- Total Views -->
			<Card.Root>
				<Card.Content class="p-6">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-muted-foreground"
							>{$_('dashboard.stats.total_views')}</span
						>
						<div
							class="flex h-8 w-8 items-center justify-center rounded-md bg-green-500/10 text-green-600 dark:text-green-400"
						>
							<Eye class="h-4 w-4" />
						</div>
					</div>
					<div class="mt-3 text-2xl font-bold">{stats.total_views}</div>
					<p class="mt-1 text-xs text-muted-foreground">{$_('dashboard.stats.all_time')}</p>
				</Card.Content>
			</Card.Root>

			<!-- Enrichment Status -->
			<Card.Root>
				<Card.Content class="p-6">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-muted-foreground"
							>{$_('dashboard.stats.enrichment')}</span
						>
						<div
							class="flex h-8 w-8 items-center justify-center rounded-md bg-yellow-500/10 text-yellow-600 dark:text-yellow-400"
						>
							<LoaderCircle class="h-4 w-4" />
						</div>
					</div>
					<div class="mt-3 text-2xl font-bold">
						{stats.total_documents > 0
							? Math.round((stats.status_breakdown.done / stats.total_documents) * 100)
							: 0}%
					</div>
					<p class="mt-1 text-xs text-muted-foreground">
						{$_('dashboard.stats.completion_rate')}
					</p>
				</Card.Content>
			</Card.Root>
		</div>
	{/if}

	<div class="grid gap-6 lg:grid-cols-3">
		<!-- Recent Documents -->
		<div class="lg:col-span-2">
			<Card.Root>
				<Card.Header class="flex flex-row items-center justify-between pb-2">
					<div>
						<Card.Title class="text-base font-semibold">{$_('dashboard.recent.title')}</Card.Title>
						<Card.Description>{$_('dashboard.recent.description')}</Card.Description>
					</div>
					<Button
						variant="ghost"
						size="sm"
						class="text-muted-foreground"
						onclick={() => goto(resolve('/documents/mine'))}
					>
						{$_('dashboard.recent.view_all')}
						<ArrowRight class="h-4 w-4" />
					</Button>
				</Card.Header>
				<Card.Content>
					{#if isLoading}
						<div class="space-y-4">
							{#each Array.from({ length: 5 }, (_, i) => i) as i (i)}
								<div class="flex items-center gap-3">
									<Skeleton class="h-9 w-9 rounded-md" />
									<div class="flex-1">
										<Skeleton class="h-4 w-48" />
										<Skeleton class="mt-1 h-3 w-24" />
									</div>
									<Skeleton class="h-5 w-16 rounded-full" />
								</div>
								{#if i < 4}<Separator />{/if}
							{/each}
						</div>
					{:else if stats && stats.recent_documents.length > 0}
						<div class="space-y-1">
							{#each stats.recent_documents as doc, i (doc.id)}
								<button
									class="flex w-full items-center gap-3 rounded-lg p-2 text-left transition-colors hover:bg-muted/50"
									onclick={() => goto(resolve(`/documents/${doc.id}`))}
								>
									<div
										class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-muted"
									>
										<FileText class="h-4 w-4 text-muted-foreground" />
									</div>
									<div class="min-w-0 flex-1">
										<p class="truncate text-sm font-medium">
											{doc.title ?? doc.original_file_name}
										</p>
										<p class="text-xs text-muted-foreground">
											{formatFileSize(doc.file_size)} &middot; {formatDate(doc.created_at)}
										</p>
									</div>
									{#if doc.enrich_status === 'done'}
										<Badge
											variant="outline"
											class="shrink-0 border-green-500/30 bg-green-500/10 text-green-700 dark:text-green-400"
										>
											<CircleCheck class="h-3 w-3" />
											{$_('document.mine.status.done')}
										</Badge>
									{:else if doc.enrich_status === 'failed'}
										<Badge variant="destructive" class="shrink-0">
											<CircleAlert class="h-3 w-3" />
											{$_('document.mine.status.failed')}
										</Badge>
									{:else if doc.enrich_status === 'processing'}
										<Badge
											variant="outline"
											class="shrink-0 border-blue-500/30 bg-blue-500/10 text-blue-700 dark:text-blue-400"
										>
											<LoaderCircle class="h-3 w-3 animate-spin" />
											{$_('document.mine.status.processing')}
										</Badge>
									{:else}
										<Badge
											variant="outline"
											class="shrink-0 border-yellow-500/30 bg-yellow-500/10 text-yellow-700 dark:text-yellow-400"
										>
											<Clock class="h-3 w-3" />
											{$_(`document.mine.status.${doc.enrich_status}`)}
										</Badge>
									{/if}
								</button>
								{#if i < stats.recent_documents.length - 1}<Separator />{/if}
							{/each}
						</div>
					{:else}
						<div class="flex flex-col items-center gap-3 py-10">
							<div class="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
								<FileText class="h-6 w-6 text-muted-foreground" />
							</div>
							<p class="text-sm text-muted-foreground">{$_('dashboard.recent.empty')}</p>
							<Button
								variant="outline"
								size="sm"
								onclick={() => goto(resolve('/documents/upload'))}
							>
								<Upload class="h-4 w-4" />
								{$_('dashboard.quick_upload')}
							</Button>
						</div>
					{/if}
				</Card.Content>
			</Card.Root>
		</div>

		<!-- Enrichment Breakdown & Quick Actions -->
		<div class="space-y-6">
			<!-- Enrichment Breakdown -->
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-base font-semibold">{$_('dashboard.breakdown.title')}</Card.Title>
				</Card.Header>
				<Card.Content>
					{#if isLoading}
						<div class="space-y-3">
							{#each Array.from({ length: 4 }, (_, i) => i) as i (i)}
								<div class="flex items-center justify-between">
									<Skeleton class="h-4 w-24" />
									<Skeleton class="h-4 w-8" />
								</div>
							{/each}
						</div>
					{:else if stats}
						<div class="space-y-3">
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<CircleCheck class="h-4 w-4 text-green-600 dark:text-green-400" />
									<span class="text-sm">{$_('document.mine.status.done')}</span>
								</div>
								<span class="text-sm font-semibold">{stats.status_breakdown.done}</span>
							</div>
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<LoaderCircle class="h-4 w-4 text-blue-600 dark:text-blue-400" />
									<span class="text-sm">{$_('document.mine.status.processing')}</span>
								</div>
								<span class="text-sm font-semibold">{stats.status_breakdown.processing}</span>
							</div>
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<Clock class="h-4 w-4 text-yellow-600 dark:text-yellow-400" />
									<span class="text-sm">{$_('dashboard.breakdown.queued')}</span>
								</div>
								<span class="text-sm font-semibold"
									>{stats.status_breakdown.pending + stats.status_breakdown.not_started}</span
								>
							</div>
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<CircleAlert class="h-4 w-4 text-red-600 dark:text-red-400" />
									<span class="text-sm">{$_('document.mine.status.failed')}</span>
								</div>
								<span class="text-sm font-semibold">{stats.status_breakdown.failed}</span>
							</div>

							<!-- Progress bar -->
							{#if stats.total_documents > 0}
								<Separator />
								<div class="space-y-1.5">
									<div class="flex justify-between text-xs text-muted-foreground">
										<span>{$_('dashboard.breakdown.progress')}</span>
										<span
											>{Math.round(
												(stats.status_breakdown.done / stats.total_documents) * 100
											)}%</span
										>
									</div>
									<div class="h-2 w-full overflow-hidden rounded-full bg-muted">
										<div
											class="h-full rounded-full bg-green-500 transition-all"
											style="width: {(stats.status_breakdown.done / stats.total_documents) * 100}%"
										></div>
									</div>
								</div>
							{/if}
						</div>
					{/if}
				</Card.Content>
			</Card.Root>

			<!-- Quick Actions -->
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-base font-semibold">{$_('dashboard.actions.title')}</Card.Title>
				</Card.Header>
				<Card.Content class="grid gap-2">
					<Button
						variant="outline"
						class="justify-start"
						onclick={() => goto(resolve('/documents/upload'))}
					>
						<Upload class="h-4 w-4" />
						{$_('dashboard.actions.upload')}
					</Button>
					<Button
						variant="outline"
						class="justify-start"
						onclick={() => goto(resolve('/documents/mine'))}
					>
						<FileText class="h-4 w-4" />
						{$_('dashboard.actions.my_documents')}
					</Button>
				</Card.Content>
			</Card.Root>
		</div>
	</div>
</div>
