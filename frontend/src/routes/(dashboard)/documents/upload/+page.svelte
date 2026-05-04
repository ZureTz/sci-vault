<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';
	import {
		FileUp,
		LoaderCircle,
		CircleCheck,
		Clock,
		CircleAlert,
		Lock,
		FlaskConical
	} from 'lucide-svelte';

	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Select from '$lib/components/ui/select';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Progress } from '$lib/components/ui/progress';
	import documentApi, { type DocumentListItem, type DocumentVisibility } from '$lib/api/document';
	import { getActiveLab, getMyLabs } from '$lib/stores/lab.svelte';
	import { showApiErrors } from '$lib/utils/api-error';

	const MAX_BATCH_FILES = 20;
	const ALLOWED_UPLOAD_EXTENSIONS = new Set(['pdf', 'txt', 'md', 'docx', 'pptx', 'xlsx']);

	function extOf(filename: string): string {
		const dot = filename.lastIndexOf('.');
		return dot >= 0 ? filename.slice(dot + 1).toLowerCase() : '';
	}

	let fileInput = $state<HTMLInputElement | undefined>(undefined);
	let selectedFiles = $state<File[]>([]);
	let selectedFile = $derived(selectedFiles[0] ?? null);
	let isBatch = $derived(selectedFiles.length > 1);
	let title = $state('');
	let year = $state('');
	let doi = $state('');
	let visibility = $state<DocumentVisibility>('private');
	let selectedLabId = $state<string>('');
	// Labs come from the shared store (populated by AppSidebar's fetch).
	let myLabs = $derived(getMyLabs());
	let isSubmitting = $state(false);
	let uploadPercent = $state(0);
	let isDragging = $state(false);

	let pendingDocuments = $state<DocumentListItem[]>([]);
	let pollTimer = $state<ReturnType<typeof setInterval> | null>(null);

	async function fetchPendingDocuments() {
		try {
			const res = await documentApi.listPendingDocuments();
			pendingDocuments = res.documents;

			// Poll Enrich Status immediately after fetching pending documents to update any that might have changed since last poll
			pollEnrichStatus();
		} catch {
			// Silently fail on background poll
		}
	}

	async function pollEnrichStatus() {
		if (pendingDocuments.length === 0) return;
		for (const doc of pendingDocuments) {
			try {
				const res = await documentApi.getEnrichStatus(doc.id);
				const idx = pendingDocuments.findIndex((d) => d.id === doc.id);
				if (idx !== -1 && pendingDocuments[idx].enrich_status !== res.status) {
					pendingDocuments[idx].enrich_status = res.status;
					// If done or failed, we might want to refresh the entire pending list
					// to clear them out, or just refresh after a short delay
					if (['done', 'failed'].includes(res.status)) {
						setTimeout(fetchPendingDocuments, 1000);
					}
				}
			} catch {
				// skip
			}
		}
	}

	// Reactively follow the active lab from the store: switching the lab in the sidebar
	// auto-selects that lab as the upload target. Selecting "no lab" reverts to private.
	$effect(() => {
		const active = getActiveLab();
		if (active) {
			selectedLabId = String(active.id);
			visibility = 'lab';
		} else {
			selectedLabId = '';
			visibility = 'private';
		}
	});

	onMount(() => {
		fetchPendingDocuments();
		pollTimer = setInterval(pollEnrichStatus, 3000);
	});

	onDestroy(() => {
		if (pollTimer) clearInterval(pollTimer);
	});

	function processFiles(files: FileList | File[] | null) {
		if (!files || files.length === 0) {
			selectedFiles = [];
			return;
		}
		const list = Array.from(files);
		if (list.length > MAX_BATCH_FILES) {
			toast.error($_('document.upload.error.too_many_files', { values: { max: MAX_BATCH_FILES } }));
			if (fileInput) fileInput.value = '';
			return;
		}
		const invalid = list.find((f) => !ALLOWED_UPLOAD_EXTENSIONS.has(extOf(f.name)));
		if (invalid) {
			toast.error($_('document.upload.error.invalid_type'));
			if (fileInput) fileInput.value = '';
			return;
		}
		selectedFiles = list;
	}

	function handleFileChange(event: Event) {
		processFiles((event.target as HTMLInputElement).files);
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		isDragging = false;
		const dropped = event.dataTransfer?.files ?? null;
		if (fileInput && dropped) {
			fileInput.files = dropped;
		}
		processFiles(dropped);
	}

	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		isDragging = true;
	}

	function handleDragLeave(event: DragEvent) {
		event.preventDefault();
		isDragging = false;
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (selectedFiles.length === 0) {
			toast.error($_('document.upload.error.file_required'));
			return;
		}
		if (selectedFiles.length > MAX_BATCH_FILES) {
			toast.error($_('document.upload.error.too_many_files', { values: { max: MAX_BATCH_FILES } }));
			return;
		}
		if (visibility === 'lab' && !selectedLabId) {
			toast.error($_('document.upload.error.lab_required'));
			return;
		}

		isSubmitting = true;
		uploadPercent = 0;
		try {
			const labId = visibility === 'lab' ? Number(selectedLabId) : null;
			if (isBatch) {
				const res = await documentApi.batchUploadDocuments(
					{ files: selectedFiles, visibility, lab_id: labId },
					(pct) => (uploadPercent = pct)
				);
				const total = res.results.length;
				if (res.failed === 0) {
					toast.success($_('document.upload.batch_all_success', { values: { total } }));
				} else if (res.succeeded === 0) {
					toast.error($_('document.upload.batch_all_failed', { values: { total } }));
				} else {
					toast.warning(
						$_('document.upload.batch_success', {
							values: { succeeded: res.succeeded, failed: res.failed, total }
						})
					);
				}
				for (const r of res.results) {
					if (r.error) {
						toast.error(`${r.filename}: ${$_(r.error)}`);
					}
				}
			} else {
				const yearNum = year ? parseInt(year, 10) : null;
				await documentApi.uploadDocument(
					{
						file: selectedFiles[0],
						title: title || null,
						year: yearNum,
						doi: doi || null,
						visibility,
						lab_id: labId
					},
					(pct) => (uploadPercent = pct)
				);
				toast.success($_('document.upload.success'));
			}

			// Refresh the enrichment queue immediately
			fetchPendingDocuments();

			// Reset form (preserve visibility/lab choice for next upload)
			selectedFiles = [];
			title = '';
			year = '';
			doi = '';
			if (fileInput) fileInput.value = '';
		} catch (error: unknown) {
			showApiErrors(error, $_('document.upload.error.failed'));
		} finally {
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('document.upload.title')} | Sci-Vault</title>
</svelte:head>

<div class="mx-auto w-full max-w-2xl space-y-6">
	<!-- Header -->
	<div class="flex flex-col space-y-2">
		<h2 class="text-3xl font-bold tracking-tight">{$_('document.upload.title')}</h2>
		<p class="text-muted-foreground">{$_('document.upload.description')}</p>
	</div>

	<Card.Root class="shadow-sm">
		<Card.Content class="pt-6">
			<form onsubmit={handleSubmit} class="space-y-6">
				<!-- File picker -->
				<div class="space-y-1.5">
					<Label for="file">{$_('document.upload.file_label')}</Label>
					<div
						class={`flex cursor-pointer items-center gap-3 rounded-md border-2 border-dashed px-4 py-5 transition-colors ${
							isDragging
								? 'border-primary bg-primary/10'
								: 'border-input bg-muted/30 hover:bg-muted/50'
						}`}
						role="button"
						tabindex="0"
						onclick={() => fileInput?.click()}
						onkeydown={(e) => e.key === 'Enter' && fileInput?.click()}
						ondrop={handleDrop}
						ondragover={handleDragOver}
						ondragleave={handleDragLeave}
					>
						<FileUp class="h-5 w-5 shrink-0 text-muted-foreground" />
						<div class="flex min-w-0 flex-1 flex-col text-sm text-muted-foreground">
							{#if isBatch}
								<span class="truncate font-medium text-foreground">
									{$_('document.upload.selected_files', {
										values: { count: selectedFiles.length }
									})}
								</span>
								<span class="mt-0.5 truncate text-xs">
									{selectedFiles.map((f) => f.name).join(', ')}
								</span>
							{:else if selectedFile}
								<span class="truncate font-medium text-foreground">{selectedFile.name}</span>
								<span class="mt-0.5 text-xs">{(selectedFile.size / 1024 / 1024).toFixed(2)} MB</span
								>
							{:else}
								<span>{$_('document.upload.file_placeholder')}</span>
							{/if}
						</div>
					</div>
					<input
						id="file"
						type="file"
						accept=".pdf,.txt,.md,.docx,.pptx,.xlsx"
						multiple
						class="hidden"
						bind:this={fileInput}
						onchange={handleFileChange}
					/>
				</div>

				{#if !isBatch}
					<!-- Title -->
					<div class="space-y-1.5">
						<Label for="title">{$_('document.upload.title_label')}</Label>
						<Input
							id="title"
							bind:value={title}
							placeholder={$_('document.upload.title_placeholder')}
							maxlength={255}
						/>
					</div>

					<!-- Year & DOI in a row -->
					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-1.5">
							<Label for="year">{$_('document.upload.year_label')}</Label>
							<Input
								id="year"
								type="number"
								bind:value={year}
								placeholder={$_('document.upload.year_placeholder')}
								min={1000}
								max={9999}
							/>
						</div>
						<div class="space-y-1.5">
							<Label for="doi">{$_('document.upload.doi_label')}</Label>
							<Input
								id="doi"
								bind:value={doi}
								placeholder={$_('document.upload.doi_placeholder')}
								maxlength={255}
							/>
						</div>
					</div>
				{/if}

				<!-- Visibility selector -->
				<div class="space-y-2">
					<Label>{$_('document.upload.visibility_label')}</Label>
					<div class="grid grid-cols-2 gap-2">
						<button
							type="button"
							class={`flex items-start gap-3 rounded-md border p-3 text-left transition-colors ${
								visibility === 'private'
									? 'border-primary bg-primary/5 ring-1 ring-primary/30'
									: 'border-input hover:bg-muted/50'
							}`}
							onclick={() => (visibility = 'private')}
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
								visibility === 'lab'
									? 'border-primary bg-primary/5 ring-1 ring-primary/30'
									: 'border-input hover:bg-muted/50'
							}`}
							disabled={myLabs.length === 0}
							onclick={() => (visibility = 'lab')}
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
					{#if visibility === 'lab' && myLabs.length > 0}
						<Select.Root type="single" bind:value={selectedLabId}>
							<Select.Trigger class="w-full">
								{myLabs.find((l) => String(l.id) === selectedLabId)?.name ??
									$_('document.upload.select_lab')}
							</Select.Trigger>
							<Select.Content>
								{#each myLabs as lab (lab.id)}
									<Select.Item value={String(lab.id)} label={lab.name}>{lab.name}</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					{/if}
				</div>

				<!-- AI metadata enrichment hint -->
				<p class="text-sm text-muted-foreground">{$_('document.upload.metadata_hint')}</p>

				{#if isSubmitting}
					<div class="space-y-1.5">
						<div class="flex justify-between text-xs text-muted-foreground">
							<span>{$_('document.upload.submitting')}</span>
							<span>{uploadPercent}%</span>
						</div>
						<Progress value={uploadPercent} class="h-2" />
					</div>
				{/if}

				<Card.Footer class="px-0 pt-2 pb-0">
					<Button type="submit" class="w-full" disabled={isSubmitting}>
						{#if isSubmitting}
							{$_('document.upload.submitting')}
						{:else if isBatch}
							{$_('document.upload.submit_batch', { values: { count: selectedFiles.length } })}
						{:else}
							{$_('document.upload.submit')}
						{/if}
					</Button>
				</Card.Footer>
			</form>
		</Card.Content>
	</Card.Root>

	{#if pendingDocuments.length > 0}
		<div class="mt-8 space-y-3">
			<h3 class="px-1 text-sm font-medium text-muted-foreground">
				{$_('document.upload.pending_queue')}
			</h3>
			<div class="grid grid-cols-1 gap-3">
				{#each pendingDocuments as doc (doc.id)}
					<a
						href={resolve(`/documents/${doc.id}`)}
						class="block min-w-0 rounded-xl outline-none focus-visible:ring-2 focus-visible:ring-primary"
					>
						<Card.Root
							class="flex w-full items-center justify-between overflow-hidden p-4 shadow-sm transition-all hover:bg-muted/50 hover:shadow-md"
						>
							<div class="min-w-0 flex-1">
								<p class="truncate text-sm font-medium">{doc.title ?? doc.original_file_name}</p>
								{#if doc.title}
									<p class="truncate text-xs text-muted-foreground">{doc.original_file_name}</p>
								{/if}
							</div>
							<div class="ml-4 flex shrink-0 items-center justify-center">
								{#if doc.enrich_status === 'done'}
									<div
										class="flex items-center text-xs font-medium text-green-600 dark:text-green-500"
									>
										<CircleCheck class="mr-1.5 h-4 w-4" />
										{$_('document.mine.status.done')}
									</div>
								{:else if doc.enrich_status === 'failed'}
									<div class="flex items-center text-xs font-medium text-red-600 dark:text-red-500">
										<CircleAlert class="mr-1.5 h-4 w-4" />
										{$_('document.mine.status.failed')}
									</div>
								{:else if doc.enrich_status === 'processing'}
									<div class="flex items-center text-xs font-medium text-primary">
										<LoaderCircle class="mr-1.5 h-4 w-4 animate-spin" />
										{$_('document.mine.status.processing')}
									</div>
								{:else}
									<div class="flex items-center text-xs font-medium text-muted-foreground">
										<Clock class="mr-1.5 h-4 w-4" />
										{$_(`document.mine.status.${doc.enrich_status}`)}
									</div>
								{/if}
							</div>
						</Card.Root>
					</a>
				{/each}
			</div>
		</div>
	{/if}
</div>
