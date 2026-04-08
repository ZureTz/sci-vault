<script lang="ts">
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import {
		FileText,
		LoaderCircle,
		ArrowLeft,
		Clock,
		CircleAlert,
		CircleCheck,
		Hash,
		User,
		Calendar,
		BookOpen
	} from 'lucide-svelte';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import documentApi, { type DocumentResponse } from '$lib/api/document';
	import { showApiErrors } from '$lib/utils/api-error';

	let { data } = $props<{ data: { id: number } }>();

	let document = $state<DocumentResponse | null>(null);
	let isLoading = $state(true);

	async function loadDocument() {
		isLoading = true;
		try {
			document = await documentApi.getDocument(data.id);
		} catch (error: unknown) {
			showApiErrors(error, $_('document.detail.error'));
		} finally {
			isLoading = false;
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
		loadDocument();
	});
</script>

<svelte:head>
	<title>{document?.title || $_('document.detail.title')} | Sci-Vault</title>
</svelte:head>

<div class="container mx-auto max-w-4xl px-4 py-8">
	<!-- Actions Bar -->
	<div class="mb-6 flex items-center justify-between">
		<Button variant="ghost" size="sm" onclick={() => goto(resolve('/documents/mine'))}>
			<ArrowLeft class="mr-2 h-4 w-4" />
			{$_('document.detail.back')}
		</Button>

		{#if document && document.download_url}
			<Button href={document.download_url} target="_blank" rel="noreferrer">
				<BookOpen class="mr-2 h-4 w-4" />
				{$_('document.detail.download')}
			</Button>
		{/if}
	</div>

	<!-- Content -->
	{#if isLoading || !document}
		<div class="space-y-6">
			<Card.Root>
				<Card.Header class="gap-4">
					<Skeleton class="h-8 w-3/4" />
					<div class="flex gap-2">
						<Skeleton class="h-5 w-24" />
						<Skeleton class="h-5 w-32" />
					</div>
				</Card.Header>
				<Card.Content class="space-y-4">
					<Skeleton class="h-4 w-full" />
					<Skeleton class="h-4 w-5/6" />
					<Skeleton class="h-4 w-4/6" />
				</Card.Content>
			</Card.Root>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-6 md:grid-cols-3">
			<!-- Main Left Column -->
			<div class="space-y-6 md:col-span-2">
				<Card.Root>
					<Card.Header>
						<div class="flex items-start justify-between gap-4">
							<div class="space-y-2">
								<h1 class="text-2xl font-bold tracking-tight">
									{document.title || document.original_file_name}
								</h1>
								{#if document.title && document.original_file_name !== document.title}
									<p class="flex items-center text-sm text-muted-foreground">
										<FileText class="mr-2 h-4 w-4" />
										{document.original_file_name}
									</p>
								{/if}
							</div>
						</div>
					</Card.Header>

					<Card.Content class="space-y-6">
						{#if document.authors && document.authors.length > 0}
							<div class="flex items-start gap-2">
								<User class="mt-1 h-4 w-4 shrink-0 text-muted-foreground" />
								<div class="flex flex-wrap gap-1.5">
									{#each document.authors as author (author)}
										<Badge variant="secondary" class="font-normal">{author}</Badge>
									{/each}
								</div>
							</div>
						{/if}

						{#if document.summary}
							<div>
								<h3 class="mb-2 flex items-center gap-2 font-semibold">
									<BookOpen class="h-4 w-4 text-primary" />
									{$_('document.detail.summary')}
								</h3>
								<div
									class="rounded-lg border bg-muted/30 p-4 text-sm leading-relaxed text-foreground/90"
								>
									{document.summary}
								</div>
							</div>
						{/if}

						{#if document.tags && document.tags.length > 0}
							<div>
								<h3
									class="mb-2 flex items-center gap-2 border-b pb-1 text-sm font-medium text-muted-foreground"
								>
									<Hash class="h-3.5 w-3.5" />
									{$_('document.detail.tags')}
								</h3>
								<div class="mt-3 flex flex-wrap gap-2">
									{#each document.tags as tag (tag)}
										<Badge variant="outline" class="bg-primary/5">{tag}</Badge>
									{/each}
								</div>
							</div>
						{/if}
					</Card.Content>
				</Card.Root>
			</div>

			<!-- Sidebar Stats -->
			<div class="space-y-6">
				<Card.Root>
					<Card.Header>
						<Card.Title class="text-lg">{$_('document.detail.metadata')}</Card.Title>
					</Card.Header>
					<Card.Content>
						<dl class="space-y-4 text-sm">
							<div class="grid grid-cols-3 gap-2 border-b pb-3">
								<dt class="col-span-1 text-muted-foreground">{$_('document.detail.status')}</dt>
								<dd class="col-span-2 text-right">
									{#if document.enrich_status === 'done'}
										<span class="inline-flex items-center text-green-600 dark:text-green-400">
											<CircleCheck class="mr-1 h-3.5 w-3.5" />
											{$_('document.mine.status.done')}
										</span>
									{:else if document.enrich_status === 'failed'}
										<span class="inline-flex items-center text-red-600 dark:text-red-400">
											<CircleAlert class="mr-1 h-3.5 w-3.5" />
											{$_('document.mine.status.failed')}
										</span>
									{:else if document.enrich_status === 'processing'}
										<span class="inline-flex items-center text-blue-600 dark:text-blue-400">
											<LoaderCircle class="mr-1 h-3.5 w-3.5 animate-spin" />
											{$_('document.mine.status.processing')}
										</span>
									{:else}
										<span class="inline-flex items-center text-yellow-600 dark:text-yellow-400">
											<Clock class="mr-1 h-3.5 w-3.5" />
											{$_(`document.mine.status.${document.enrich_status}`)}
										</span>
									{/if}
								</dd>
							</div>

							{#if document.year}
								<div class="grid grid-cols-3 gap-2 border-b pb-3">
									<dt class="col-span-1 flex items-center text-muted-foreground">
										<Calendar class="mr-2 h-3.5 w-3.5" />
										{$_('document.detail.year')}
									</dt>
									<dd class="col-span-2 text-right font-medium">{document.year}</dd>
								</div>
							{/if}

							{#if document.doi}
								<div class="grid grid-cols-[1fr_2fr] gap-2 border-b pb-3">
									<dt class="mr-2 text-muted-foreground">DOI</dt>
									<dd class="text-right font-medium break-all">
										<a
											href={`https://doi.org/${document.doi}`}
											target="_blank"
											rel="noreferrer"
											class="text-primary hover:underline"
										>
											{document.doi}
										</a>
									</dd>
								</div>
							{/if}

							<div class="grid grid-cols-3 gap-2 border-b pb-3">
								<dt class="col-span-1 text-muted-foreground">{$_('document.detail.size')}</dt>
								<dd class="col-span-2 text-right font-medium">
									{formatFileSize(document.file_size)}
								</dd>
							</div>

							<div class="grid grid-cols-3 gap-2 border-b pb-3">
								<dt class="col-span-1 text-muted-foreground">{$_('document.detail.type')}</dt>
								<dd class="col-span-2 text-right font-medium">
									<Badge variant="outline"
										>{document.content_type.split('/').pop() || document.content_type}</Badge
									>
								</dd>
							</div>

							<div class="grid grid-cols-3 gap-2">
								<dt class="col-span-1 text-muted-foreground">{$_('document.detail.uploaded')}</dt>
								<dd class="col-span-2 text-right font-medium">{formatDate(document.created_at)}</dd>
							</div>
						</dl>
					</Card.Content>
				</Card.Root>
			</div>
		</div>
	{/if}
</div>
