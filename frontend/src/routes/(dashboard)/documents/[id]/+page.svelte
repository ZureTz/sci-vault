<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { _, locale } from 'svelte-i18n';
	import { goto, afterNavigate } from '$app/navigation';
	import { resolve } from '$app/paths';
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
		BookOpen,
		RefreshCw,
		Languages,
		Lock,
		FlaskConical,
		Pencil,
		Sparkles,
		Heart,
		Eye
	} from 'lucide-svelte';

	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Card from '$lib/components/ui/card';
	import * as Select from '$lib/components/ui/select';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from 'svelte-sonner';
	import documentApi, { type DocumentResponse, type DocumentVisibility } from '$lib/api/document';
	import interactionApi from '$lib/api/interaction';
	import recommendApi, { type SimilarDocumentItem } from '$lib/api/recommend';
	import { getActiveLab, getMyLabs } from '$lib/stores/lab.svelte';
	import { getUser } from '$lib/stores/user.svelte';
	import { translateSummary } from '$lib/api/translate';
	import { showApiErrors } from '$lib/utils/api-error';

	let { data } = $props<{ data: { id: number } }>();

	let document = $state<DocumentResponse | null>(null);
	let isLoading = $state(true);
	let isRestarting = $state(false);
	let pollTimer = $state<ReturnType<typeof setInterval> | null>(null);

	// Like toggle. Optimistic: flip state immediately, roll back on failure.
	let isLikePending = $state(false);

	async function toggleLike() {
		if (!document || isLikePending) return;
		isLikePending = true;
		const wasLiked = document.liked_by_me;
		document.liked_by_me = !wasLiked;
		document.like_count += wasLiked ? -1 : 1;
		try {
			const res = wasLiked
				? await interactionApi.unlike(document.id)
				: await interactionApi.like(document.id);
			document.liked_by_me = res.liked;
			document.like_count = res.like_count;
		} catch (error: unknown) {
			document.liked_by_me = wasLiked;
			document.like_count += wasLiked ? 1 : -1;
			showApiErrors(error, $_('document.detail.like.failed'));
		} finally {
			isLikePending = false;
		}
	}

	let isTranslating = $state(false);
	let translatedSummary = $state('');
	let showOriginal = $state(false);

	// Similar-documents recommendations
	let similarDocs = $state<SimilarDocumentItem[]>([]);
	let isLoadingSimilar = $state(false);
	let similarLoaded = $state(false);

	let isEnglishLocale = $derived(($locale ?? 'en').startsWith('en'));
	let currentUser = $derived(getUser());
	let isOwner = $derived(document != null && document.uploaded_by === Number(currentUser.id));

	// Back navigation: if the user arrived from inside the app, pop history;
	// otherwise (direct link, refresh, external) fall back to a sensible default
	// so the button is never a no-op. document.referrer is unreliable for this
	// because SvelteKit's client-side navigation doesn't update it, so we track
	// arrival via afterNavigate instead.
	let arrivedInternally = $state(false);
	afterNavigate((nav) => {
		if (nav.from && nav.type !== 'enter') {
			arrivedInternally = true;
		}
	});

	function goBack() {
		if (arrivedInternally) {
			window.history.back();
		} else {
			goto(resolve('/documents/mine'));
		}
	}

	// Visibility edit state. Read labs from the shared store — the sidebar
	// owns the getMyLabs fetch, so this page doesn't fire a duplicate request.
	let visDialogOpen = $state(false);
	let visEditValue = $state<DocumentVisibility>('private');
	let visEditLabId = $state<string>('');
	let visSubmitting = $state(false);
	let myLabs = $derived(getMyLabs());

	function openVisDialog() {
		if (!document) return;
		visEditValue = document.visibility;
		// Prefer the document's current lab; fall back to the lab active in the sidebar.
		if (document.lab_id) {
			visEditLabId = String(document.lab_id);
		} else {
			const active = getActiveLab();
			visEditLabId = active ? String(active.id) : '';
		}
		visDialogOpen = true;
	}

	async function handleVisSubmit() {
		if (!document) return;
		if (visEditValue === 'lab' && !visEditLabId) {
			toast.error($_('document.detail.visibility.lab_required'));
			return;
		}
		visSubmitting = true;
		try {
			await documentApi.updateVisibility(document.id, {
				visibility: visEditValue,
				lab_id: visEditValue === 'lab' ? Number(visEditLabId) : null
			});
			toast.success($_('document.detail.visibility.success'));
			visDialogOpen = false;
			await loadDocument(false);
		} catch (error: unknown) {
			showApiErrors(error, $_('document.detail.visibility.failed'));
		} finally {
			visSubmitting = false;
		}
	}

	async function loadDocument(showSpinner = true) {
		if (showSpinner) isLoading = true;
		try {
			document = await documentApi.getDocument(data.id);

			// Fetch real-time status immediately instead of waiting for the first poll tick
			pollEnrichStatus();
			// Recommendations depend on the source doc's embedding being ready.
			if (document.enrich_status === 'done') {
				loadSimilar();
			}
		} catch (error: unknown) {
			showApiErrors(error, $_('document.detail.error'));
		} finally {
			if (showSpinner) isLoading = false;
		}
	}

	async function loadSimilar() {
		if (!document || document.enrich_status !== 'done') return;
		isLoadingSimilar = true;
		try {
			const active = getActiveLab();
			const res = await recommendApi.getSimilar(document.id, {
				lab_id: active ? active.id : undefined,
				limit: 10
			});
			similarDocs = res.results;
		} catch {
			// Fail silently — recommendations are non-critical UI.
			similarDocs = [];
		} finally {
			isLoadingSimilar = false;
			similarLoaded = true;
		}
	}

	async function pollEnrichStatus() {
		if (!document) return;
		const status = document.enrich_status;
		if (status === 'not_started' || status === 'pending' || status === 'processing') {
			try {
				const res = await documentApi.getEnrichStatus(document.id);
				if (document.enrich_status !== res.status) {
					document.enrich_status = res.status;
					if (res.status === 'done') {
						await loadDocument(false);
					}
				}
			} catch {
				// silently ignore polling errors
			}
		}
	}

	async function restartEnrichment() {
		if (!document) return;
		isRestarting = true;
		try {
			await documentApi.restartEnrichment(document.id);
			toast.success($_('service.restart_enrichment.success'));
			document.enrich_status = 'pending';
		} catch (error: unknown) {
			showApiErrors(error, $_('service.restart_enrichment.failed'));
		} finally {
			isRestarting = false;
		}
	}

	async function handleTranslate() {
		if (!document?.summary || !$locale) return;
		isTranslating = true;
		translatedSummary = '';
		showOriginal = false;
		try {
			await translateSummary(
				document.summary,
				$locale,
				(chunk) => {
					translatedSummary += chunk;
				},
				() => {
					isTranslating = false;
				},
				(error) => {
					isTranslating = false;
					toast.error($_('document.detail.translate_error'));
					console.error('Translation error:', error);
				}
			);
		} catch {
			isTranslating = false;
			toast.error($_('document.detail.translate_error'));
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

	// Reactively (re)load whenever the route's :id changes — without this,
	// clicking a similar-doc link inside this page just updates the URL but
	// reuses the same component instance, so data.id changes but the body
	// never refetches.
	$effect(() => {
		const id = data.id;
		// Reset stale per-doc state so skeletons show during the refetch.
		similarDocs = [];
		similarLoaded = false;
		translatedSummary = '';
		showOriginal = false;
		isTranslating = false;
		loadDocument();
		// Note: `id` is read above purely to register the dependency.
		void id;
	});

	onMount(() => {
		pollTimer = setInterval(pollEnrichStatus, 3000);
	});

	onDestroy(() => {
		if (pollTimer) clearInterval(pollTimer);
	});
</script>

<svelte:head>
	<title>{document?.title || $_('document.detail.title')} | Sci-Vault</title>
</svelte:head>

<div class="mx-auto w-full max-w-6xl space-y-6">
	<!-- Actions Bar -->
	<div class="flex items-center justify-between">
		<Button variant="ghost" size="sm" onclick={goBack}>
			<ArrowLeft class="mr-2 h-4 w-4" />
			{$_('document.detail.back')}
		</Button>

		{#if document && document.download_url}
			<Button variant="outline" href={document.download_url} target="_blank" rel="noreferrer">
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
							<div class="w-full space-y-2">
								<div class="flex flex-wrap items-center gap-3">
									<h1 class="text-2xl font-bold tracking-tight">
										{document.title || document.original_file_name}
									</h1>
									<div class="flex items-center gap-2">
										{#if document.enrich_status === 'done'}
											<Badge
												variant="outline"
												class="border-green-500/30 bg-green-500/10 whitespace-nowrap text-green-700 dark:text-green-400"
											>
												<CircleCheck class="mr-1 h-3.5 w-3.5" />
												{$_('document.mine.status.done')}
											</Badge>
										{:else if document.enrich_status === 'failed'}
											<Badge variant="destructive" class="whitespace-nowrap">
												<CircleAlert class="mr-1 h-3.5 w-3.5" />
												{$_('document.mine.status.failed')}
											</Badge>
										{:else if document.enrich_status === 'processing'}
											<Badge
												variant="outline"
												class="border-blue-500/30 bg-blue-500/10 whitespace-nowrap text-blue-700 dark:text-blue-400"
											>
												<LoaderCircle class="mr-1 h-3.5 w-3.5 animate-spin" />
												{$_('document.mine.status.processing')}
											</Badge>
										{:else}
											<Badge
												variant="outline"
												class="border-yellow-500/30 bg-yellow-500/10 whitespace-nowrap text-yellow-700 dark:text-yellow-400"
											>
												<Clock class="mr-1 h-3.5 w-3.5" />
												{$_(`document.mine.status.${document.enrich_status}`)}
											</Badge>
										{/if}

										{#if document.enrich_status === 'failed' || document.enrich_status === 'not_started'}
											<Button
												variant="outline"
												size="sm"
												class="h-6 px-2 text-xs"
												onclick={restartEnrichment}
												disabled={isRestarting}
											>
												<RefreshCw class="mr-1 h-3 w-3 {isRestarting ? 'animate-spin' : ''}" />
												{$_('document.detail.restart')}
											</Button>
										{/if}
									</div>
								</div>
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
								<div class="mb-2 flex items-center justify-between">
									<h3 class="flex items-center gap-2 font-semibold">
										<BookOpen class="h-4 w-4 text-primary" />
										{$_('document.detail.summary')}
									</h3>
									{#if !isEnglishLocale}
										<Button
											variant="ghost"
											size="sm"
											class="h-7 gap-1.5 text-xs text-muted-foreground"
											onclick={handleTranslate}
											disabled={isTranslating}
										>
											{#if isTranslating}
												<LoaderCircle class="h-3.5 w-3.5 animate-spin" />
											{:else}
												<Languages class="h-3.5 w-3.5" />
											{/if}
											{$_('document.detail.translate')}
										</Button>
									{/if}
								</div>
								<div
									class="rounded-lg border bg-muted/30 p-4 text-sm leading-relaxed text-foreground/90"
								>
									{#if translatedSummary && !showOriginal}
										{translatedSummary}{#if isTranslating}<span
												class="inline-block h-4 w-0.5 animate-pulse bg-foreground/60"
											></span>{/if}
									{:else}
										{document.summary}
									{/if}
								</div>
								{#if translatedSummary || isTranslating}
									<button
										class="mt-1.5 flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
										onclick={() => (showOriginal = !showOriginal)}
									>
										<Languages class="h-3 w-3" />
										{showOriginal
											? $_('document.detail.show_translation')
											: $_('document.detail.show_original')}
									</button>
								{/if}
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

					<!-- Engagement footer (Twitter-style): separator above, counts on the
						 left, like toggle on the right. -->
					<Card.Footer class="flex flex-wrap items-center justify-between gap-3 border-t pt-4">
						<div class="flex items-center gap-5 text-sm text-muted-foreground">
							<span class="flex items-center gap-1.5">
								<Eye class="h-4 w-4" />
								<span class="font-semibold text-foreground">{document.view_count}</span>
								<span>{$_('document.detail.engagement.views')}</span>
							</span>
							<span class="flex items-center gap-1.5">
								<Heart class="h-4 w-4 {document.liked_by_me ? 'fill-current text-rose-500' : ''}" />
								<span class="font-semibold text-foreground">{document.like_count}</span>
								<span>{$_('document.detail.engagement.likes')}</span>
							</span>
						</div>
						<Button
							variant={document.liked_by_me ? 'default' : 'outline'}
							size="sm"
							onclick={toggleLike}
							disabled={isLikePending}
							aria-pressed={document.liked_by_me}
						>
							<Heart class="mr-1.5 h-4 w-4 {document.liked_by_me ? 'fill-current' : ''}" />
							{document.liked_by_me
								? $_('document.detail.like.liked')
								: $_('document.detail.like.like')}
						</Button>
					</Card.Footer>
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

							<div class="grid grid-cols-3 gap-2 border-b pb-3">
								<dt class="col-span-1 text-muted-foreground">{$_('document.detail.uploaded')}</dt>
								<dd class="col-span-2 text-right font-medium">{formatDate(document.created_at)}</dd>
							</div>

							<div class="grid grid-cols-3 items-center gap-2 border-b pb-3">
								<dt class="col-span-1 flex items-center text-muted-foreground">
									<User class="mr-2 h-3.5 w-3.5" />
									{$_('document.detail.uploader')}
								</dt>
								<dd class="col-span-2 text-right">
									<a
										href={resolve(`/profile/${document.uploaded_by}`)}
										class="font-medium text-primary hover:underline"
									>
										{document.uploaded_by_username ?? $_('document.detail.unknown_uploader')}
									</a>
								</dd>
							</div>

							<div class="grid grid-cols-3 items-center gap-2">
								<dt class="col-span-1 text-muted-foreground">
									{$_('document.detail.visibility.label')}
								</dt>
								<dd class="col-span-2 flex items-center justify-end gap-1.5">
									{#if document.visibility === 'lab' && document.lab_name}
										<Badge
											variant="outline"
											class="max-w-40 gap-1 border-blue-500/30 bg-blue-500/10 text-blue-700 dark:text-blue-400"
										>
											<FlaskConical class="size-3 shrink-0" />
											<span class="truncate" title={document.lab_name}>{document.lab_name}</span>
										</Badge>
									{:else}
										<Badge variant="secondary" class="gap-1">
											<Lock class="size-3" />
											{$_('document.mine.visibility.private')}
										</Badge>
									{/if}
									{#if isOwner}
										<Button
											variant="ghost"
											size="icon"
											class="size-7 text-muted-foreground hover:text-primary"
											onclick={openVisDialog}
										>
											<Pencil class="size-3.5" />
										</Button>
									{/if}
								</dd>
							</div>
						</dl>
					</Card.Content>
				</Card.Root>
			</div>
		</div>

		<!-- Similar documents -->
		{#if document.enrich_status === 'done'}
			<Card.Root>
				<Card.Header class="pb-3">
					<Card.Title class="flex items-center gap-2 text-lg">
						<Sparkles class="h-4 w-4 text-primary" />
						{$_('document.detail.similar.title')}
					</Card.Title>
					<p class="text-sm text-muted-foreground">
						{$_('document.detail.similar.description')}
					</p>
				</Card.Header>
				<Card.Content>
					{#if isLoadingSimilar}
						<div class="grid gap-3 sm:grid-cols-2">
							{#each Array.from({ length: 4 }, (_, i) => i) as i (i)}
								<div class="space-y-2 rounded-lg border p-4">
									<Skeleton class="h-4 w-3/4" />
									<Skeleton class="h-3 w-full" />
									<Skeleton class="h-3 w-5/6" />
								</div>
							{/each}
						</div>
					{:else if similarDocs.length === 0}
						<p class="py-6 text-center text-sm text-muted-foreground">
							{similarLoaded
								? $_('document.detail.similar.empty')
								: $_('document.detail.similar.unavailable')}
						</p>
					{:else}
						<div class="grid gap-3 sm:grid-cols-2">
							{#each similarDocs as item (item.doc_id)}
								<a
									href={resolve(`/documents/${item.doc_id}`)}
									class="group flex flex-col gap-2 rounded-lg border bg-card p-4 transition-colors hover:border-primary/50 hover:bg-muted/40"
								>
									<div class="flex items-start justify-between gap-2">
										<h4
											class="line-clamp-2 text-sm leading-snug font-semibold group-hover:text-primary"
										>
											{item.title || item.original_file_name}
										</h4>
										<Badge variant="outline" class="shrink-0 gap-1 font-mono text-[10px]">
											{Math.round(item.similarity * 100)}%
										</Badge>
									</div>
									{#if item.summary}
										<p class="line-clamp-3 text-xs text-muted-foreground">
											{item.summary}
										</p>
									{/if}
									{#if item.tags && item.tags.length > 0}
										<div class="flex flex-wrap gap-1">
											{#each item.tags.slice(0, 4) as tag (tag)}
												<Badge variant="secondary" class="text-[10px] font-normal">{tag}</Badge>
											{/each}
										</div>
									{/if}
								</a>
							{/each}
						</div>
					{/if}
				</Card.Content>
			</Card.Root>
		{/if}
	{/if}
</div>

<!-- Visibility edit dialog -->
{#if document}
	<AlertDialog.Root bind:open={visDialogOpen}>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>{$_('document.detail.visibility.dialog_title')}</AlertDialog.Title>
				<AlertDialog.Description>
					{$_('document.detail.visibility.dialog_desc')}
				</AlertDialog.Description>
			</AlertDialog.Header>

			<div class="space-y-4 px-6">
				<div class="grid grid-cols-2 gap-2">
					<button
						type="button"
						class={`flex items-start gap-3 rounded-md border p-3 text-left transition-colors ${
							visEditValue === 'private'
								? 'border-primary bg-primary/5 ring-1 ring-primary/30'
								: 'border-input hover:bg-muted/50'
						}`}
						onclick={() => (visEditValue = 'private')}
					>
						<Lock class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
						<div class="min-w-0 flex-1">
							<div class="text-sm font-medium">{$_('document.upload.visibility_private')}</div>
							<div class="text-xs text-muted-foreground">
								{$_('document.upload.visibility_private_hint')}
							</div>
						</div>
					</button>
					<button
						type="button"
						class={`flex items-start gap-3 rounded-md border p-3 text-left transition-colors ${
							visEditValue === 'lab'
								? 'border-primary bg-primary/5 ring-1 ring-primary/30'
								: 'border-input hover:bg-muted/50'
						}`}
						disabled={myLabs.length === 0}
						onclick={() => (visEditValue = 'lab')}
					>
						<FlaskConical class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
						<div class="min-w-0 flex-1">
							<div class="text-sm font-medium">{$_('document.upload.visibility_lab')}</div>
							<div class="text-xs text-muted-foreground">
								{myLabs.length === 0
									? $_('document.upload.visibility_lab_no_labs')
									: $_('document.upload.visibility_lab_hint')}
							</div>
						</div>
					</button>
				</div>

				{#if visEditValue === 'lab' && myLabs.length > 0}
					<div class="space-y-1.5">
						<Label for="vis-lab-select">{$_('document.upload.select_lab')}</Label>
						<Select.Root type="single" bind:value={visEditLabId}>
							<Select.Trigger id="vis-lab-select" class="w-full">
								{myLabs.find((l) => String(l.id) === visEditLabId)?.name ??
									$_('document.upload.select_lab')}
							</Select.Trigger>
							<Select.Content>
								{#each myLabs as lab (lab.id)}
									<Select.Item value={String(lab.id)} label={lab.name}>{lab.name}</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
				{/if}
			</div>

			<AlertDialog.Footer>
				<AlertDialog.Cancel>{$_('profile.btn.cancel')}</AlertDialog.Cancel>
				<AlertDialog.Action
					disabled={visSubmitting || (visEditValue === 'lab' && !visEditLabId)}
					onclick={(e: MouseEvent) => {
						e.preventDefault();
						handleVisSubmit();
					}}
				>
					<Pencil class="size-3.5" />
					{$_('document.detail.visibility.apply')}
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}
