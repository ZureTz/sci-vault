<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { _ } from 'svelte-i18n';
	import {
		FileText,
		Upload,
		LoaderCircle,
		CircleCheck,
		Clock,
		CircleAlert,
		Eye,
		RefreshCw
	} from 'lucide-svelte';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { toast } from 'svelte-sonner';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import documentApi, { type DocumentListItem } from '$lib/api/document';
	import { showApiErrors } from '$lib/utils/api-error';

	const PAGE_SIZE = 10;

	let documents = $state<DocumentListItem[]>([]);
	let total = $state(0);
	let currentPage = $state(1);
	let isLoading = $state(true);
	let pollTimer = $state<ReturnType<typeof setInterval> | null>(null);

	let totalPages = $derived(Math.max(1, Math.ceil(total / PAGE_SIZE)));

	async function loadDocuments() {
		isLoading = true;
		try {
			const res = await documentApi.listMyDocuments(currentPage, PAGE_SIZE);
			documents = res.documents;
			total = res.total;

			// Fetch real-time status immediately instead of waiting for the first poll tick
			pollEnrichStatus();
		} catch (error: unknown) {
			showApiErrors(error, $_('document.mine.error'));
		} finally {
			isLoading = false;
		}
	}

	async function restartEnrichment(docId: number) {
		try {
			await documentApi.restartEnrichment(docId);
			toast.success($_('service.restart_enrichment.success'));
			const idx = documents.findIndex((d) => d.id === docId);
			if (idx !== -1) {
				documents[idx].enrich_status = 'pending';
			}
		} catch (error: unknown) {
			showApiErrors(error, $_('service.restart_enrichment.failed'));
		}
	}

	async function pollEnrichStatus() {
		const pending = documents.filter(
			(d) =>
				d.enrich_status === 'not_started' ||
				d.enrich_status === 'pending' ||
				d.enrich_status === 'processing'
		);
		if (pending.length === 0) return;

		for (const doc of pending) {
			try {
				const res = await documentApi.getEnrichStatus(doc.id);
				const idx = documents.findIndex((d) => d.id === doc.id);
				if (idx !== -1 && documents[idx].enrich_status !== res.status) {
					documents[idx].enrich_status = res.status;
				}
			} catch {
				// silently ignore polling errors
			}
		}
	}

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	onMount(() => {
		loadDocuments();
		pollTimer = setInterval(pollEnrichStatus, 3000);
	});

	onDestroy(() => {
		if (pollTimer) clearInterval(pollTimer);
	});
</script>

<svelte:head>
	<title>{$_('document.mine.title')} | Sci-Vault</title>
</svelte:head>

<div class="flex-1 space-y-6">
	<!-- Header -->
	<div class="flex flex-col justify-between space-y-4 sm:flex-row sm:items-center sm:space-y-0">
		<div class="flex items-center gap-3">
			<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
				<FileText class="h-5 w-5" />
			</div>
			<div class="space-y-1">
				<h2 class="text-3xl font-bold tracking-tight">{$_('document.mine.title')}</h2>
				<p class="text-sm text-muted-foreground">{$_('document.mine.description')}</p>
			</div>
		</div>
		<div class="flex items-center space-x-2">
			<Button variant="outline" onclick={() => goto(resolve('/documents/upload'))}>
				<Upload class="mr-2 h-4 w-4" />
				{$_('document.mine.go_upload')}
			</Button>
		</div>
	</div>

	<Card.Root class="shadow-sm">
		<Card.Content class="p-0">
			{#if isLoading}
				<div class="divide-y">
					{#each Array.from({ length: 4 }, (_, i) => i) as i (i)}
						<div class="flex items-center gap-4 px-6 py-4">
							<div class="flex flex-1 flex-col gap-1.5">
								<Skeleton class="h-4 w-48" />
								<Skeleton class="h-3 w-32" />
							</div>
							<Skeleton class="h-4 w-12" />
							<Skeleton class="h-5 w-20 rounded-full" />
							<Skeleton class="h-4 w-20" />
						</div>
					{/each}
				</div>
			{:else if documents.length === 0}
				<div class="flex flex-col items-center gap-4 py-16">
					<div class="flex h-16 w-16 items-center justify-center rounded-full bg-muted">
						<FileText class="h-8 w-8 text-muted-foreground" />
					</div>
					<div class="text-center">
						<p class="font-medium">{$_('document.mine.empty')}</p>
						<p class="mt-1 text-sm text-muted-foreground">{$_('document.mine.empty_hint')}</p>
					</div>
					<Button onclick={() => goto(resolve('/documents/upload'))}>
						<Upload class="h-4 w-4" />
						{$_('document.mine.go_upload')}
					</Button>
				</div>
			{:else}
				<div class="overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>{$_('document.mine.table.title')}</Table.Head>
								<Table.Head class="hidden w-24 text-right md:table-cell"
									>{$_('document.mine.table.file_size')}</Table.Head
								>
								<Table.Head class="w-28 sm:w-32">{$_('document.mine.table.status')}</Table.Head>
								<Table.Head class="hidden w-32 sm:table-cell"
									>{$_('document.mine.table.created_at')}</Table.Head
								>
								<Table.Head class="w-20 text-center">{$_('document.mine.table.actions')}</Table.Head
								>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each documents as doc (doc.id)}
								<Table.Row class="group transition-colors hover:bg-muted/50 hover:shadow-sm">
									<Table.Cell class="max-w-48 font-medium sm:max-w-[16rem] md:max-w-[24rem]">
										<a
											href={resolve(`/documents/${doc.id}`)}
											class="flex items-center gap-3 rounded-sm outline-none focus-visible:ring-1 focus-visible:ring-primary"
										>
											<FileText
												class="h-4 w-4 shrink-0 text-muted-foreground/70 transition-colors group-hover:text-primary"
											/>
											<div class="min-w-0 flex-1">
												<span
													class="block truncate font-medium transition-colors group-hover:text-primary"
													title={doc.title ?? doc.original_file_name}
													>{doc.title ?? doc.original_file_name}</span
												>
												{#if doc.title}
													<span
														class="mt-0.5 block truncate text-xs font-normal text-muted-foreground/80 group-hover:text-muted-foreground"
														title={doc.original_file_name}>{doc.original_file_name}</span
													>
												{/if}
											</div>
										</a>
									</Table.Cell>
									<Table.Cell class="hidden text-right text-xs text-muted-foreground md:table-cell">
										{formatFileSize(doc.file_size)}
									</Table.Cell>
									<Table.Cell>
										{#if doc.enrich_status === 'done'}
											<Badge
												variant="outline"
												class="border-green-500/30 bg-green-500/10 text-green-700 dark:text-green-400"
											>
												<CircleCheck />
												{$_('document.mine.status.done')}
											</Badge>
										{:else if doc.enrich_status === 'failed'}
											<Badge variant="destructive">
												<CircleAlert />
												{$_('document.mine.status.failed')}
											</Badge>
										{:else if doc.enrich_status === 'processing'}
											<Badge
												variant="outline"
												class="border-blue-500/30 bg-blue-500/10 text-blue-700 dark:text-blue-400"
											>
												<LoaderCircle class="animate-spin" />
												{$_('document.mine.status.processing')}
											</Badge>
										{:else}
											<Badge
												variant="outline"
												class="border-yellow-500/30 bg-yellow-500/10 text-yellow-700 dark:text-yellow-400"
											>
												<Clock />
												{$_(`document.mine.status.${doc.enrich_status}`)}
											</Badge>
										{/if}
									</Table.Cell>
									<Table.Cell class="hidden text-xs text-muted-foreground sm:table-cell">
										{formatDate(doc.created_at)}
									</Table.Cell>
									<Table.Cell class="text-center">
										<div class="flex justify-center gap-1">
											{#if doc.enrich_status === 'failed' || doc.enrich_status === 'not_started'}
												<Button
													variant="ghost"
													size="icon"
													class="h-8 w-8 text-muted-foreground transition-colors hover:bg-primary/10 hover:text-primary"
													onclick={() => restartEnrichment(doc.id)}
												>
													<RefreshCw strokeWidth={2.5} class="h-4 w-4" />
												</Button>
											{/if}
											<Button
												variant="ghost"
												size="icon"
												href={resolve(`/documents/${doc.id}`)}
												class="h-8 w-8 text-muted-foreground transition-colors hover:bg-primary/10 hover:text-primary"
											>
												<Eye strokeWidth={2.5} class="h-4 w-4" />
											</Button>
										</div>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>

				<!-- Pagination -->
				{#if totalPages > 1}
					<div class="flex items-center justify-between border-t px-6 py-3">
						<p class="text-sm text-muted-foreground">
							{(currentPage - 1) * PAGE_SIZE + 1}–{Math.min(currentPage * PAGE_SIZE, total)} / {total}
						</p>
						<div class="flex items-center gap-2">
							<Button
								variant="outline"
								size="sm"
								disabled={currentPage <= 1}
								onclick={() => {
									currentPage -= 1;
									loadDocuments();
								}}
							>
								{$_('document.mine.pagination.prev')}
							</Button>
							<span class="text-sm text-muted-foreground">{currentPage} / {totalPages}</span>
							<Button
								variant="outline"
								size="sm"
								disabled={currentPage >= totalPages}
								onclick={() => {
									currentPage += 1;
									loadDocuments();
								}}
							>
								{$_('document.mine.pagination.next')}
							</Button>
						</div>
					</div>
				{/if}
			{/if}
		</Card.Content>
	</Card.Root>
</div>
