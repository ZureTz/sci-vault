<script lang="ts">
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import {
		FileText,
		HardDrive,
		Eye,
		Heart,
		CircleCheck,
		LoaderCircle,
		CircleAlert,
		Clock,
		Upload,
		ArrowRight,
		TrendingUp
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
	import UploadsChart from '$lib/components/dashboard/uploads-chart.svelte';
	import EngagementChart from '$lib/components/dashboard/engagement-chart.svelte';
	import FormatDistributionChart from '$lib/components/dashboard/format-distribution-chart.svelte';

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

	const hasUploadActivity = $derived((stats?.uploads_by_day ?? []).some((d) => d.count > 0));
	const hasEngagementActivity = $derived(
		(stats?.views_by_day ?? []).some((d) => d.count > 0) ||
			(stats?.likes_by_day ?? []).some((d) => d.count > 0)
	);
</script>

<svelte:head>
	<title>{$_('dashboard.title')} | Sci-Vault</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex flex-col justify-between space-y-4 sm:flex-row sm:items-center sm:space-y-0">
		<div class="space-y-1">
			<h2 class="text-3xl font-bold tracking-tight">{$_('dashboard.title')}</h2>
			<p class="text-muted-foreground">{$_('dashboard.description')}</p>
		</div>
		<div class="flex items-center space-x-2">
			<Button onclick={() => goto(resolve('/documents/upload'))}>
				<Upload class="mr-2 h-4 w-4" />
				{$_('dashboard.quick_upload')}
			</Button>
		</div>
	</div>

	<!-- Stat Cards -->
	{#if isLoading}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
			{#each Array.from({ length: 5 }, (_, i) => i) as i (i)}
				<Card.Root>
					<Card.Content class="p-4">
						<div class="flex items-center justify-between">
							<Skeleton class="h-4 w-24" />
							<Skeleton class="h-8 w-8 rounded-lg" />
						</div>
						<Skeleton class="mt-2 h-7 w-16" />
						<Skeleton class="mt-1.5 h-3 w-32" />
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{:else if stats}
		<div class="mb-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
			<!-- Total Documents -->
			<Card.Root class="transition-shadow hover:shadow-md">
				<Card.Content class="p-4">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-muted-foreground"
							>{$_('dashboard.stats.total_documents')}</span
						>
						<div
							class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary transition-transform hover:scale-110"
						>
							<FileText class="h-4 w-4" />
						</div>
					</div>
					<div class="mt-2 text-2xl font-bold tracking-tight">{stats.total_documents}</div>
					<p class="mt-1 text-xs text-muted-foreground">
						{$_('dashboard.stats.enriched', {
							values: { count: stats.status_breakdown.done }
						})}
					</p>
				</Card.Content>
			</Card.Root>

			<!-- Storage Used -->
			<Card.Root class="transition-shadow hover:shadow-md">
				<Card.Content class="p-4">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-muted-foreground"
							>{$_('dashboard.stats.storage_used')}</span
						>
						<div
							class="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-500/10 text-blue-600 transition-transform hover:scale-110 dark:text-blue-400"
						>
							<HardDrive class="h-4 w-4" />
						</div>
					</div>
					<div class="mt-2 text-2xl font-bold tracking-tight">
						{formatFileSize(stats.total_storage)}
					</div>
					<p class="mt-1 text-xs text-muted-foreground">
						{$_('dashboard.stats.across_documents', {
							values: { count: stats.total_documents }
						})}
					</p>
				</Card.Content>
			</Card.Root>

			<!-- Total Views -->
			<Card.Root class="transition-shadow hover:shadow-md">
				<Card.Content class="p-4">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-muted-foreground"
							>{$_('dashboard.stats.total_views')}</span
						>
						<div
							class="flex h-8 w-8 items-center justify-center rounded-lg bg-green-500/10 text-green-600 transition-transform hover:scale-110 dark:text-green-400"
						>
							<Eye class="h-4 w-4" />
						</div>
					</div>
					<div class="mt-2 text-2xl font-bold tracking-tight">{stats.total_views}</div>
					<p class="mt-1 text-xs text-muted-foreground">{$_('dashboard.stats.all_time')}</p>
				</Card.Content>
			</Card.Root>

			<!-- Total Likes -->
			<Card.Root class="transition-shadow hover:shadow-md">
				<Card.Content class="p-4">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-muted-foreground"
							>{$_('dashboard.stats.total_likes')}</span
						>
						<div
							class="flex h-8 w-8 items-center justify-center rounded-lg bg-pink-500/10 text-pink-600 transition-transform hover:scale-110 dark:text-pink-400"
						>
							<Heart class="h-4 w-4" />
						</div>
					</div>
					<div class="mt-2 text-2xl font-bold tracking-tight">{stats.total_likes}</div>
					<p class="mt-1 text-xs text-muted-foreground">{$_('dashboard.stats.all_time')}</p>
				</Card.Content>
			</Card.Root>

			<!-- Enrichment Status -->
			<Card.Root class="transition-shadow hover:shadow-md">
				<Card.Content class="p-4">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-muted-foreground"
							>{$_('dashboard.stats.enrichment')}</span
						>
						<div
							class="flex h-8 w-8 items-center justify-center rounded-lg bg-yellow-500/10 text-yellow-600 transition-transform hover:scale-110 dark:text-yellow-400"
						>
							<LoaderCircle class="h-4 w-4" />
						</div>
					</div>
					<div class="mt-2 text-2xl font-bold tracking-tight">
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

	<!-- Charts row -->
	{#if isLoading}
		<div class="grid gap-6 lg:grid-cols-2">
			{#each Array.from({ length: 2 }, (_, i) => i) as i (i)}
				<Card.Root>
					<Card.Header class="pb-2">
						<Skeleton class="h-5 w-40" />
						<Skeleton class="mt-1 h-3 w-56" />
					</Card.Header>
					<Card.Content>
						<Skeleton class="h-50 w-full" />
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{:else if stats}
		<div class="grid gap-6 lg:grid-cols-2">
			<Card.Root class="transition-shadow hover:shadow-md">
				<Card.Header class="pb-2">
					<Card.Title class="text-base font-semibold">
						{$_('dashboard.charts.uploads_title')}
					</Card.Title>
					<Card.Description class="text-xs">
						{$_('dashboard.charts.uploads_desc')}
					</Card.Description>
				</Card.Header>
				<Card.Content>
					{#if hasUploadActivity}
						<UploadsChart
							data={stats.uploads_by_day}
							label={$_('dashboard.charts.uploads_label')}
						/>
					{:else}
						<div class="flex h-50 w-full items-center justify-center text-sm text-muted-foreground">
							{$_('dashboard.charts.empty')}
						</div>
					{/if}
				</Card.Content>
			</Card.Root>

			<Card.Root class="transition-shadow hover:shadow-md">
				<Card.Header class="pb-2">
					<Card.Title class="text-base font-semibold">
						{$_('dashboard.charts.engagement_title')}
					</Card.Title>
					<Card.Description class="text-xs">
						{$_('dashboard.charts.engagement_desc')}
					</Card.Description>
				</Card.Header>
				<Card.Content>
					{#if hasEngagementActivity}
						<EngagementChart
							views={stats.views_by_day}
							likes={stats.likes_by_day}
							viewsLabel={$_('dashboard.charts.views_label')}
							likesLabel={$_('dashboard.charts.likes_label')}
						/>
					{:else}
						<div class="flex h-50 w-full items-center justify-center text-sm text-muted-foreground">
							{$_('dashboard.charts.empty')}
						</div>
					{/if}
				</Card.Content>
			</Card.Root>
		</div>
	{/if}

	<div class="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
		<!-- Recent Documents -->
		<div class="md:col-span-2 xl:col-span-2">
			<Card.Root class="flex h-full flex-col transition-shadow hover:shadow-md">
				<Card.Header
					class="flex flex-col items-start gap-4 pb-2 sm:flex-row sm:items-center sm:justify-between sm:gap-0"
				>
					<div>
						<Card.Title class="text-base font-semibold">{$_('dashboard.recent.title')}</Card.Title>
						<Card.Description class="mt-1 text-xs"
							>{$_('dashboard.recent.description')}</Card.Description
						>
					</div>
					<Button
						variant="ghost"
						size="sm"
						class="h-8 w-full shrink-0 text-muted-foreground hover:bg-muted sm:w-auto"
						onclick={() => goto(resolve('/documents/mine'))}
					>
						{$_('dashboard.recent.view_all')}
						<ArrowRight class="ml-1 h-4 w-4" />
					</Button>
				</Card.Header>
				<Card.Content class="flex-1 pt-2">
					{#if isLoading}
						<div class="space-y-4">
							{#each Array.from({ length: 5 }, (_, i) => i) as i (i)}
								<div class="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:gap-4">
									<Skeleton class="hidden h-10 w-10 shrink-0 rounded-xl sm:block" />
									<div class="w-full flex-1 space-y-2">
										<Skeleton class="h-4 w-full sm:w-48" />
										<Skeleton class="h-3 w-32 sm:w-24" />
									</div>
									<Skeleton class="h-6 w-full rounded-full sm:w-20" />
								</div>
								{#if i < 4}<Separator class="my-2" />{/if}
							{/each}
						</div>
					{:else if stats && stats.recent_documents.length > 0}
						<div class="flex flex-col space-y-2">
							{#each stats.recent_documents as doc, i (doc.id)}
								<button
									class="group flex w-full flex-col items-start gap-3 rounded-xl border border-transparent p-3 text-left transition-all hover:bg-muted/50 hover:shadow-sm sm:flex-row sm:items-center sm:gap-4"
									onclick={() => goto(resolve(`/documents/${doc.id}`))}
								>
									<div
										class="hidden h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-muted transition-colors group-hover:bg-background group-hover:shadow-sm sm:flex"
									>
										<FileText class="h-5 w-5 text-muted-foreground" />
									</div>
									<div class="w-full min-w-0 flex-1">
										<p
											class="max-w-[16rem] truncate text-sm font-medium transition-colors group-hover:text-primary sm:max-w-none"
										>
											{doc.title ?? doc.original_file_name}
										</p>
										<p class="mt-0.5 text-xs text-muted-foreground">
											{formatFileSize(doc.file_size)} &middot; {formatDate(doc.created_at)}
										</p>
									</div>
									<div class="flex w-full shrink-0 justify-start sm:w-auto sm:justify-end">
										{#if doc.enrich_status === 'done'}
											<Badge
												variant="outline"
												class="border-green-500/30 bg-green-500/10 px-2 py-0.5 text-green-700 transition-colors dark:text-green-400"
											>
												<CircleCheck class="mr-1 h-3.5 w-3.5" />
												{$_('document.mine.status.done')}
											</Badge>
										{:else if doc.enrich_status === 'failed'}
											<Badge variant="destructive" class="px-2 py-0.5">
												<CircleAlert class="mr-1 h-3.5 w-3.5" />
												{$_('document.mine.status.failed')}
											</Badge>
										{:else if doc.enrich_status === 'processing'}
											<Badge
												variant="outline"
												class="border-blue-500/30 bg-blue-500/10 px-2 py-0.5 text-blue-700 transition-colors dark:text-blue-400"
											>
												<LoaderCircle class="mr-1 h-3.5 w-3.5 animate-spin" />
												{$_('document.mine.status.processing')}
											</Badge>
										{:else}
											<Badge
												variant="outline"
												class="border-yellow-500/30 bg-yellow-500/10 px-2 py-0.5 text-yellow-700 transition-colors dark:text-yellow-400"
											>
												<Clock class="mr-1 h-3.5 w-3.5" />
												{$_(`document.mine.status.${doc.enrich_status}`)}
											</Badge>
										{/if}
									</div>
								</button>
								{#if i < stats.recent_documents.length - 1}<Separator
										class="my-1 opacity-50"
									/>{/if}
							{/each}
						</div>
					{:else}
						<div
							class="flex h-full flex-col items-center justify-center gap-4 px-4 py-12 text-center"
						>
							<div
								class="flex h-16 w-16 items-center justify-center rounded-2xl bg-muted/50 ring-1 ring-border/50"
							>
								<FileText class="h-8 w-8 text-muted-foreground/50" />
							</div>
							<div>
								<p class="text-base font-medium">{$_('dashboard.recent.empty')}</p>
								<p class="mt-1 text-sm text-muted-foreground">
									Upload your first document to get started
								</p>
							</div>
							<Button
								class="mt-2 w-full transition-transform hover:scale-105 sm:w-auto"
								onclick={() => goto(resolve('/documents/upload'))}
							>
								<Upload class="mr-2 h-4 w-4" />
								{$_('dashboard.quick_upload')}
							</Button>
						</div>
					{/if}
				</Card.Content>
			</Card.Root>
		</div>

		<!-- Sidebar column: Top viewed, Format distribution, Enrichment breakdown, Quick actions -->
		<div class="col-span-1 space-y-6 md:col-span-2 xl:col-span-1">
			<!-- Top Viewed Documents -->
			<Card.Root class="transition-shadow hover:shadow-md">
				<Card.Header class="pb-2">
					<Card.Title class="text-base font-semibold">
						{$_('dashboard.charts.top_viewed_title')}
					</Card.Title>
					<Card.Description class="text-xs">
						{$_('dashboard.charts.top_viewed_desc')}
					</Card.Description>
				</Card.Header>
				<Card.Content>
					{#if isLoading}
						<div class="space-y-3">
							{#each Array.from({ length: 3 }, (_, i) => i) as i (i)}
								<div class="flex items-center gap-3">
									<Skeleton class="h-8 w-8 rounded-lg" />
									<div class="flex-1 space-y-1.5">
										<Skeleton class="h-4 w-3/4" />
										<Skeleton class="h-3 w-1/3" />
									</div>
								</div>
							{/each}
						</div>
					{:else if stats && stats.top_viewed.some((d) => d.view_count > 0)}
						<div class="space-y-1">
							{#each stats.top_viewed as doc, i (doc.id)}
								{#if doc.view_count > 0}
									<button
										class="group flex w-full items-center gap-3 rounded-lg px-2 py-2 text-left transition-colors hover:bg-muted/50"
										onclick={() => goto(resolve(`/documents/${doc.id}`))}
									>
										<div
											class="flex size-8 shrink-0 items-center justify-center rounded-md bg-muted text-xs font-semibold text-muted-foreground"
										>
											{i + 1}
										</div>
										<div class="min-w-0 flex-1">
											<p
												class="truncate text-sm font-medium transition-colors group-hover:text-primary"
											>
												{doc.title ?? doc.original_file_name}
											</p>
											<p class="text-xs text-muted-foreground">
												<Eye class="mr-1 inline h-3 w-3" />{doc.view_count}
												<Heart class="mr-1 ml-2 inline h-3 w-3" />{doc.like_count}
											</p>
										</div>
										<TrendingUp
											class="size-4 shrink-0 text-muted-foreground transition-colors group-hover:text-primary"
										/>
									</button>
								{/if}
							{/each}
						</div>
					{:else}
						<p class="py-6 text-center text-sm text-muted-foreground">
							{$_('dashboard.charts.top_viewed_empty')}
						</p>
					{/if}
				</Card.Content>
			</Card.Root>

			<!-- Format Distribution -->
			<Card.Root class="transition-shadow hover:shadow-md">
				<Card.Header class="pb-2">
					<Card.Title class="text-base font-semibold">
						{$_('dashboard.charts.formats_title')}
					</Card.Title>
					<Card.Description class="text-xs">
						{$_('dashboard.charts.formats_desc')}
					</Card.Description>
				</Card.Header>
				<Card.Content>
					{#if isLoading}
						<Skeleton class="h-50 w-full" />
					{:else if stats}
						<FormatDistributionChart data={stats.format_distribution} />
					{/if}
				</Card.Content>
			</Card.Root>

			<!-- Enrichment Breakdown -->
			<Card.Root class="transition-shadow hover:shadow-md">
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
			<Card.Root class="transition-shadow hover:shadow-md">
				<Card.Header class="pb-2">
					<Card.Title class="text-base font-semibold">{$_('dashboard.actions.title')}</Card.Title>
				</Card.Header>
				<Card.Content class="grid gap-3">
					<Button
						variant="outline"
						class="group justify-start border-muted-foreground/20 transition-all hover:border-primary/50 hover:bg-primary/5"
						onclick={() => goto(resolve('/documents/upload'))}
					>
						<Upload
							class="mr-2 h-4 w-4 text-muted-foreground transition-colors group-hover:text-primary"
						/>
						{$_('dashboard.actions.upload')}
					</Button>
					<Button
						variant="outline"
						class="group justify-start border-muted-foreground/20 transition-all hover:border-primary/50 hover:bg-primary/5"
						onclick={() => goto(resolve('/documents/mine'))}
					>
						<FileText
							class="mr-2 h-4 w-4 text-muted-foreground transition-colors group-hover:text-primary"
						/>
						{$_('dashboard.actions.my_documents')}
					</Button>
				</Card.Content>
			</Card.Root>
		</div>
	</div>
</div>
