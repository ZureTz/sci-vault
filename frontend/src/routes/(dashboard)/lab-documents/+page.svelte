<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { _ } from 'svelte-i18n';
	import {
		type ColumnDef,
		type SortingState,
		type VisibilityState,
		getCoreRowModel
	} from '@tanstack/table-core';
	import {
		FileText,
		LoaderCircle,
		CircleCheck,
		Clock,
		CircleAlert,
		Eye,
		RefreshCw,
		FlaskConical,
		Pencil,
		X,
		Search,
		ChevronDown,
		ChevronUp,
		ChevronsUpDown,
		ChevronLeft,
		ChevronRight,
		ChevronsLeft,
		ChevronsRight,
		Ellipsis,
		SlidersHorizontal,
		ShieldOff,
		Trash2,
		User as UserIcon
	} from 'lucide-svelte';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { toast } from 'svelte-sonner';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Card from '$lib/components/ui/card';
	import * as Select from '$lib/components/ui/select';
	import * as Table from '$lib/components/ui/table';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { createSvelteTable, FlexRender } from '$lib/components/ui/data-table';
	import documentApi, {
		type DocumentListItem,
		type ListLabDocumentsParams
	} from '$lib/api/document';
	import { getActiveLab } from '$lib/stores/lab.svelte';
	import { showApiErrors } from '$lib/utils/api-error';

	type StatusFilter = 'all' | 'not_started' | 'pending' | 'processing' | 'done' | 'failed';
	type SortKey = 'created_at' | 'title' | 'file_size' | 'view_count';

	// ===== Active-lab gating =====
	// Depend on the lab ID (a primitive) so the fetch effect refires on lab swap
	// even if the ActiveLab object reference changes.
	let activeLabId = $derived(getActiveLab()?.id ?? null);
	let activeLabRole = $derived(getActiveLab()?.role ?? null);
	let activeLabName = $derived(getActiveLab()?.name ?? '');
	let canViewPage = $derived(activeLabId !== null && activeLabRole === 'owner');

	// ===== Data + server state =====
	let documents = $state<DocumentListItem[]>([]);
	let total = $state(0);
	let isLoading = $state(true);
	let pollTimer = $state<ReturnType<typeof setInterval> | null>(null);

	// ===== Query controls (server-side) =====
	let pageIndex = $state(0);
	let pageSize = $state(10);
	let search = $state('');
	let debouncedSearch = $state('');
	let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null;
	let statusFilter = $state<StatusFilter>('all');

	let sorting = $state<SortingState>([{ id: 'created_at', desc: true }]);

	// ===== TanStack table UI-only state =====
	let columnVisibility = $state<VisibilityState>({});

	// ===== Edit metadata dialog state =====
	let editDialogOpen = $state(false);
	let editTarget = $state<DocumentListItem | null>(null);
	let editTitle = $state('');
	let editYear = $state<string>('');
	let editDoi = $state('');
	let editSubmitting = $state(false);

	// ===== Delete confirmation state =====
	let deleteDialogOpen = $state(false);
	let deleteTarget = $state<DocumentListItem | null>(null);
	let deleteSubmitting = $state(false);

	const pageCount = $derived(Math.max(1, Math.ceil(total / pageSize)));

	// ===== Debounced search =====
	$effect(() => {
		const s = search;
		if (searchDebounceTimer) clearTimeout(searchDebounceTimer);
		searchDebounceTimer = setTimeout(() => {
			debouncedSearch = s;
			pageIndex = 0;
		}, 300);
	});

	// ===== Fetching =====
	async function loadDocuments() {
		const labId = activeLabId;
		if (labId === null || activeLabRole !== 'owner') {
			documents = [];
			total = 0;
			isLoading = false;
			return;
		}
		isLoading = true;
		try {
			const sortEntry = sorting[0];
			const sortBy = (sortEntry?.id as SortKey | undefined) ?? 'created_at';
			const sortOrder = sortEntry?.desc ? 'desc' : 'asc';

			const params: ListLabDocumentsParams = {
				page: pageIndex + 1,
				page_size: pageSize,
				sort_by: sortBy,
				sort_order: sortOrder
			};
			if (debouncedSearch.trim()) params.search = debouncedSearch.trim();
			if (statusFilter !== 'all') params.status = statusFilter;

			const res = await documentApi.listLabDocuments(labId, params);
			documents = res.documents;
			total = res.total;
			pollEnrichStatus();
		} catch (error: unknown) {
			showApiErrors(error, $_('document.lab.error'));
		} finally {
			isLoading = false;
		}
	}

	// Refetch whenever any server-affecting control changes — including the
	// active lab ID, so a lab swap reloads the table.
	$effect(() => {
		void debouncedSearch;
		void statusFilter;
		void pageIndex;
		void pageSize;
		void sorting;
		void activeLabId;
		void activeLabRole;
		loadDocuments();
	});

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
				// silently ignore
			}
		}
	}

	async function restartEnrichment(docId: number) {
		try {
			await documentApi.restartEnrichment(docId);
			toast.success($_('service.restart_enrichment.success'));
			const idx = documents.findIndex((d) => d.id === docId);
			if (idx !== -1) documents[idx].enrich_status = 'pending';
		} catch (error: unknown) {
			showApiErrors(error, $_('service.restart_enrichment.failed'));
		}
	}

	function openEditDialog(doc: DocumentListItem) {
		editTarget = doc;
		editTitle = doc.title ?? '';
		editYear = '';
		editDoi = '';
		(async () => {
			try {
				const full = await documentApi.getDocument(doc.id);
				editTitle = full.title ?? '';
				editYear = full.year != null ? String(full.year) : '';
				editDoi = full.doi ?? '';
			} catch {
				// best effort — fall back to list-level fields
			}
		})();
		editDialogOpen = true;
	}

	async function handleEditSubmit() {
		if (!editTarget) return;
		editSubmitting = true;
		try {
			const yearNum = editYear.trim() ? Number(editYear.trim()) : null;
			if (yearNum !== null && (Number.isNaN(yearNum) || yearNum < 1000 || yearNum > 9999)) {
				toast.error($_('document.lab.edit_dialog.year_invalid'));
				editSubmitting = false;
				return;
			}
			await documentApi.updateMetadata(editTarget.id, {
				title: editTitle.trim() ? editTitle.trim() : '',
				year: yearNum,
				doi: editDoi.trim() ? editDoi.trim() : ''
			});
			toast.success($_('document.lab.edit_dialog.success'));
			editDialogOpen = false;
			loadDocuments();
		} catch (error: unknown) {
			showApiErrors(error, $_('document.lab.edit_dialog.failed'));
		} finally {
			editSubmitting = false;
		}
	}

	function openDeleteDialog(doc: DocumentListItem) {
		deleteTarget = doc;
		deleteDialogOpen = true;
	}

	async function handleDeleteSubmit() {
		if (!deleteTarget) return;
		deleteSubmitting = true;
		try {
			await documentApi.deleteDocument(deleteTarget.id);
			toast.success($_('document.lab.delete.success'));
			deleteDialogOpen = false;
			loadDocuments();
		} catch (error: unknown) {
			showApiErrors(error, $_('document.lab.delete.failed'));
		} finally {
			deleteSubmitting = false;
		}
	}

	// ===== Formatting helpers =====
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

	const columns: ColumnDef<DocumentListItem>[] = [
		{
			id: 'title',
			accessorKey: 'title',
			header: () => $_('document.lab.table.title'),
			enableSorting: true,
			enableHiding: false
		},
		{
			id: 'uploader',
			accessorKey: 'uploaded_by_username',
			header: () => $_('document.lab.table.uploader'),
			enableSorting: false
		},
		{
			id: 'file_size',
			accessorKey: 'file_size',
			header: () => $_('document.lab.table.file_size'),
			enableSorting: true
		},
		{
			id: 'enrich_status',
			accessorKey: 'enrich_status',
			header: () => $_('document.lab.table.status'),
			enableSorting: false
		},
		{
			id: 'created_at',
			accessorKey: 'created_at',
			header: () => $_('document.lab.table.created_at'),
			enableSorting: true
		},
		{
			id: 'actions',
			header: () => $_('document.lab.table.actions'),
			enableSorting: false,
			enableHiding: false
		}
	];

	const table = createSvelteTable<DocumentListItem>({
		get data() {
			return documents;
		},
		columns,
		manualPagination: true,
		manualSorting: true,
		manualFiltering: true,
		get pageCount() {
			return pageCount;
		},
		getRowId: (row) => String(row.id),
		state: {
			get sorting() {
				return sorting;
			},
			get columnVisibility() {
				return columnVisibility;
			},
			get pagination() {
				return { pageIndex, pageSize };
			}
		},
		onSortingChange: (updater) => {
			sorting = typeof updater === 'function' ? updater(sorting) : updater;
		},
		onColumnVisibilityChange: (updater) => {
			columnVisibility = typeof updater === 'function' ? updater(columnVisibility) : updater;
		},
		onPaginationChange: (updater) => {
			const prev = { pageIndex, pageSize };
			const next = typeof updater === 'function' ? updater(prev) : updater;
			pageIndex = next.pageIndex;
			pageSize = next.pageSize;
		},
		getCoreRowModel: getCoreRowModel()
	});

	function cycleSort(columnId: string) {
		const current = sorting[0];
		if (!current || current.id !== columnId) {
			sorting = [{ id: columnId, desc: true }];
			return;
		}
		if (current.desc) {
			sorting = [{ id: columnId, desc: false }];
		} else {
			sorting = [{ id: 'created_at', desc: true }];
		}
	}

	const columnLabels: Record<string, string> = $derived({
		title: $_('document.lab.table.title'),
		uploader: $_('document.lab.table.uploader'),
		file_size: $_('document.lab.table.file_size'),
		enrich_status: $_('document.lab.table.status'),
		created_at: $_('document.lab.table.created_at')
	});

	onMount(() => {
		pollTimer = setInterval(pollEnrichStatus, 3000);
	});

	onDestroy(() => {
		if (pollTimer) clearInterval(pollTimer);
		if (searchDebounceTimer) clearTimeout(searchDebounceTimer);
	});
</script>

<svelte:head>
	<title>{$_('document.lab.title')} | Sci-Vault</title>
</svelte:head>

<div class="flex-1 space-y-6">
	<!-- Header -->
	<div class="flex flex-col justify-between space-y-4 sm:flex-row sm:items-center sm:space-y-0">
		<div class="flex items-center gap-3">
			<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
				<FlaskConical class="h-5 w-5" />
			</div>
			<div class="space-y-1">
				<h2 class="text-3xl font-bold tracking-tight">
					{$_('document.lab.title')}
					{#if activeLabName}
						<span class="text-muted-foreground">— {activeLabName}</span>
					{/if}
				</h2>
				<p class="text-sm text-muted-foreground">{$_('document.lab.description')}</p>
			</div>
		</div>
		{#if canViewPage}
			<div class="flex items-center space-x-2">
				<Button
					variant="outline"
					onclick={loadDocuments}
					disabled={isLoading}
					aria-label={$_('document.lab.refresh')}
				>
					<RefreshCw class={`mr-2 h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
					{$_('document.lab.refresh')}
				</Button>
			</div>
		{/if}
	</div>

	{#if activeLabId === null}
		<!-- No active lab selected -->
		<Card.Root class="shadow-sm">
			<Card.Content class="flex flex-col items-center gap-4 py-16">
				<div class="flex h-16 w-16 items-center justify-center rounded-full bg-muted">
					<FlaskConical class="h-8 w-8 text-muted-foreground" />
				</div>
				<div class="text-center">
					<p class="font-medium">{$_('document.lab.no_lab')}</p>
					<p class="mt-1 text-sm text-muted-foreground">{$_('document.lab.no_lab_hint')}</p>
				</div>
			</Card.Content>
		</Card.Root>
	{:else if activeLabRole !== 'owner'}
		<!-- Active lab but caller is not the owner -->
		<Card.Root class="shadow-sm">
			<Card.Content class="flex flex-col items-center gap-4 py-16">
				<div class="flex h-16 w-16 items-center justify-center rounded-full bg-muted">
					<ShieldOff class="h-8 w-8 text-muted-foreground" />
				</div>
				<div class="text-center">
					<p class="font-medium">{$_('document.lab.not_owner')}</p>
					<p class="mt-1 text-sm text-muted-foreground">{$_('document.lab.not_owner_hint')}</p>
				</div>
			</Card.Content>
		</Card.Root>
	{:else}
		<!-- Toolbar: search + status filter + column visibility -->
		<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
			<div class="flex flex-1 flex-col gap-2 sm:flex-row sm:items-center">
				<div class="relative w-full sm:max-w-xs">
					<Search
						class="pointer-events-none absolute top-1/2 left-2.5 h-4 w-4 -translate-y-1/2 text-muted-foreground"
					/>
					<Input
						type="search"
						placeholder={$_('document.lab.search_placeholder')}
						bind:value={search}
						class="pl-8"
					/>
				</div>

				<Select.Root type="single" bind:value={statusFilter}>
					<Select.Trigger class="w-full sm:w-45">
						{statusFilter === 'all'
							? $_('document.lab.filter.status_all')
							: $_(`document.lab.status.${statusFilter}`)}
					</Select.Trigger>
					<Select.Content>
						<Select.Item value="all">{$_('document.lab.filter.status_all')}</Select.Item>
						<Select.Item value="not_started">{$_('document.lab.status.not_started')}</Select.Item>
						<Select.Item value="pending">{$_('document.lab.status.pending')}</Select.Item>
						<Select.Item value="processing">{$_('document.lab.status.processing')}</Select.Item>
						<Select.Item value="done">{$_('document.lab.status.done')}</Select.Item>
						<Select.Item value="failed">{$_('document.lab.status.failed')}</Select.Item>
					</Select.Content>
				</Select.Root>

				{#if statusFilter !== 'all' || debouncedSearch}
					<Button
						variant="ghost"
						size="sm"
						onclick={() => {
							statusFilter = 'all';
							search = '';
						}}
					>
						<X class="mr-1 h-4 w-4" />
						{$_('document.lab.filter.reset')}
					</Button>
				{/if}
			</div>

			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button {...props} variant="outline" size="sm">
							<SlidersHorizontal class="mr-2 h-4 w-4" />
							{$_('document.lab.columns')}
							<ChevronDown class="ml-2 h-4 w-4" />
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content align="end" class="w-44">
					<DropdownMenu.Label>{$_('document.lab.columns')}</DropdownMenu.Label>
					<DropdownMenu.Separator />
					{#each table
						.getAllColumns()
						.filter((c) => c.getCanHide() && columnLabels[c.id]) as column (column.id)}
						<DropdownMenu.CheckboxItem
							checked={column.getIsVisible()}
							onCheckedChange={(v) => column.toggleVisibility(!!v)}
						>
							{columnLabels[column.id]}
						</DropdownMenu.CheckboxItem>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		</div>

		<Card.Root class="shadow-sm">
			<Card.Content class="p-0">
				<div class="overflow-x-auto">
					<Table.Root>
						<Table.Header>
							{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
								<Table.Row>
									{#each headerGroup.headers as header (header.id)}
										{@const col = header.column}
										<Table.Head
											class={col.id === 'actions'
												? 'w-20 text-center'
												: col.id === 'file_size'
													? 'w-28 text-right'
													: col.id === 'created_at'
														? 'w-32'
														: col.id === 'uploader'
															? 'w-40'
															: col.id === 'enrich_status'
																? 'w-32'
																: ''}
										>
											{#if col.getCanSort()}
												<button
													type="button"
													class="-ml-2 inline-flex h-8 items-center gap-1 rounded px-2 font-medium text-muted-foreground hover:bg-muted/60 hover:text-foreground"
													onclick={() => cycleSort(col.id)}
												>
													<FlexRender
														content={col.columnDef.header}
														context={header.getContext()}
													/>
													{#if sorting[0]?.id === col.id && !sorting[0].desc}
														<ChevronUp class="h-3.5 w-3.5" />
													{:else if sorting[0]?.id === col.id && sorting[0].desc}
														<ChevronDown class="h-3.5 w-3.5" />
													{:else}
														<ChevronsUpDown class="h-3.5 w-3.5 opacity-50" />
													{/if}
												</button>
											{:else}
												<FlexRender content={col.columnDef.header} context={header.getContext()} />
											{/if}
										</Table.Head>
									{/each}
								</Table.Row>
							{/each}
						</Table.Header>
						<Table.Body>
							{#if isLoading}
								{#each Array.from({ length: Math.min(pageSize, 10) }, (_, i) => i) as i (i)}
									<Table.Row>
										{#each table.getVisibleLeafColumns() as col (col.id)}
											<Table.Cell>
												<Skeleton class="h-4 w-full max-w-50" />
											</Table.Cell>
										{/each}
									</Table.Row>
								{/each}
							{:else if documents.length === 0}
								<Table.Row>
									<Table.Cell
										colspan={table.getVisibleLeafColumns().length}
										class="h-60 text-center"
									>
										<div class="flex flex-col items-center gap-4 py-8">
											<div class="flex h-16 w-16 items-center justify-center rounded-full bg-muted">
												<FileText class="h-8 w-8 text-muted-foreground" />
											</div>
											<div class="text-center">
												<p class="font-medium">
													{debouncedSearch || statusFilter !== 'all'
														? $_('document.lab.empty_filtered')
														: $_('document.lab.empty')}
												</p>
												<p class="mt-1 text-sm text-muted-foreground">
													{debouncedSearch || statusFilter !== 'all'
														? $_('document.lab.empty_filtered_hint')
														: $_('document.lab.empty_hint')}
												</p>
											</div>
										</div>
									</Table.Cell>
								</Table.Row>
							{:else}
								{#each table.getRowModel().rows as row (row.id)}
									{@const doc = row.original}
									<Table.Row class="group transition-colors hover:bg-muted/50">
										{#each row.getVisibleCells() as cell (cell.id)}
											{@const colId = cell.column.id}
											<Table.Cell
												class={colId === 'actions'
													? 'text-center'
													: colId === 'file_size'
														? 'text-right text-xs text-muted-foreground'
														: colId === 'created_at'
															? 'text-xs text-muted-foreground'
															: colId === 'title'
																? 'max-w-48 font-medium sm:max-w-[16rem] md:max-w-[24rem]'
																: ''}
											>
												{#if colId === 'title'}
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
												{:else if colId === 'uploader'}
													<div class="flex items-center gap-1.5 text-sm text-muted-foreground">
														<UserIcon class="h-3.5 w-3.5 shrink-0" />
														<span class="truncate" title={doc.uploaded_by_username ?? ''}>
															{doc.uploaded_by_username ?? $_('document.lab.unknown_uploader')}
														</span>
													</div>
												{:else if colId === 'file_size'}
													{formatFileSize(doc.file_size)}
												{:else if colId === 'enrich_status'}
													{#if doc.enrich_status === 'done'}
														<Badge
															variant="outline"
															class="border-green-500/30 bg-green-500/10 text-green-700 dark:text-green-400"
														>
															<CircleCheck />
															{$_('document.lab.status.done')}
														</Badge>
													{:else if doc.enrich_status === 'failed'}
														<Badge variant="destructive">
															<CircleAlert />
															{$_('document.lab.status.failed')}
														</Badge>
													{:else if doc.enrich_status === 'processing'}
														<Badge
															variant="outline"
															class="border-blue-500/30 bg-blue-500/10 text-blue-700 dark:text-blue-400"
														>
															<LoaderCircle class="animate-spin" />
															{$_('document.lab.status.processing')}
														</Badge>
													{:else}
														<Badge
															variant="outline"
															class="border-yellow-500/30 bg-yellow-500/10 text-yellow-700 dark:text-yellow-400"
														>
															<Clock />
															{$_(`document.lab.status.${doc.enrich_status}`)}
														</Badge>
													{/if}
												{:else if colId === 'created_at'}
													{formatDate(doc.created_at)}
												{:else if colId === 'actions'}
													<DropdownMenu.Root>
														<DropdownMenu.Trigger>
															{#snippet child({ props })}
																<Button
																	{...props}
																	variant="ghost"
																	size="icon"
																	class="h-8 w-8 data-[state=open]:bg-muted"
																	aria-label={$_('document.lab.actions.open_menu')}
																>
																	<Ellipsis class="h-4 w-4" />
																</Button>
															{/snippet}
														</DropdownMenu.Trigger>
														<DropdownMenu.Content align="end" class="w-48">
															<DropdownMenu.Item
																onSelect={() => goto(resolve(`/documents/${doc.id}`))}
															>
																<Eye class="mr-2 h-4 w-4" />
																{$_('document.lab.actions.view')}
															</DropdownMenu.Item>
															<DropdownMenu.Item onSelect={() => openEditDialog(doc)}>
																<Pencil class="mr-2 h-4 w-4" />
																{$_('document.lab.actions.edit')}
															</DropdownMenu.Item>
															{#if doc.enrich_status === 'failed' || doc.enrich_status === 'not_started'}
																<DropdownMenu.Item onSelect={() => restartEnrichment(doc.id)}>
																	<RefreshCw class="mr-2 h-4 w-4" />
																	{$_('document.lab.actions.restart')}
																</DropdownMenu.Item>
															{/if}
															<DropdownMenu.Separator />
															<DropdownMenu.Item
																class="text-destructive focus:text-destructive"
																onSelect={() => openDeleteDialog(doc)}
															>
																<Trash2 class="mr-2 h-4 w-4" />
																{$_('document.lab.actions.delete')}
															</DropdownMenu.Item>
														</DropdownMenu.Content>
													</DropdownMenu.Root>
												{/if}
											</Table.Cell>
										{/each}
									</Table.Row>
								{/each}
							{/if}
						</Table.Body>
					</Table.Root>
				</div>

				<!-- Pagination -->
				<div
					class="flex flex-col items-center justify-between gap-3 border-t px-4 py-3 sm:flex-row"
				>
					<div class="text-sm text-muted-foreground">
						{$_('document.lab.pagination.info', { values: { total } })}
					</div>
					<div class="flex flex-col items-center gap-3 sm:flex-row sm:gap-6">
						<div class="flex items-center gap-2">
							<Label for="page-size" class="text-sm text-muted-foreground">
								{$_('document.lab.pagination.per_page')}
							</Label>
							<Select.Root
								type="single"
								value={String(pageSize)}
								onValueChange={(v) => {
									pageSize = Number(v);
									pageIndex = 0;
								}}
							>
								<Select.Trigger id="page-size" class="h-8 w-18">
									{pageSize}
								</Select.Trigger>
								<Select.Content>
									{#each [10, 20, 30, 50, 100] as n (n)}
										<Select.Item value={String(n)}>{n}</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>
						<div class="text-sm text-muted-foreground">
							{$_('document.lab.pagination.page_info', {
								values: { current: pageIndex + 1, total: pageCount }
							})}
						</div>
						<div class="flex items-center gap-1">
							<Button
								variant="outline"
								size="icon"
								class="h-8 w-8"
								disabled={pageIndex === 0}
								onclick={() => (pageIndex = 0)}
								aria-label={$_('document.lab.pagination.first')}
							>
								<ChevronsLeft class="h-4 w-4" />
							</Button>
							<Button
								variant="outline"
								size="icon"
								class="h-8 w-8"
								disabled={pageIndex === 0}
								onclick={() => (pageIndex = Math.max(0, pageIndex - 1))}
								aria-label={$_('document.lab.pagination.prev')}
							>
								<ChevronLeft class="h-4 w-4" />
							</Button>
							<Button
								variant="outline"
								size="icon"
								class="h-8 w-8"
								disabled={pageIndex >= pageCount - 1}
								onclick={() => (pageIndex = Math.min(pageCount - 1, pageIndex + 1))}
								aria-label={$_('document.lab.pagination.next')}
							>
								<ChevronRight class="h-4 w-4" />
							</Button>
							<Button
								variant="outline"
								size="icon"
								class="h-8 w-8"
								disabled={pageIndex >= pageCount - 1}
								onclick={() => (pageIndex = pageCount - 1)}
								aria-label={$_('document.lab.pagination.last')}
							>
								<ChevronsRight class="h-4 w-4" />
							</Button>
						</div>
					</div>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}
</div>

<!-- Edit metadata dialog -->
<AlertDialog.Root bind:open={editDialogOpen}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>{$_('document.lab.edit_dialog.title')}</AlertDialog.Title>
			<AlertDialog.Description>
				{$_('document.lab.edit_dialog.description')}
			</AlertDialog.Description>
		</AlertDialog.Header>

		<div class="space-y-4 px-6">
			<div class="space-y-1.5">
				<Label for="edit-title">{$_('document.lab.edit_dialog.title_label')}</Label>
				<Input
					id="edit-title"
					bind:value={editTitle}
					placeholder={editTarget?.original_file_name ?? ''}
					maxlength={255}
				/>
			</div>
			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1.5">
					<Label for="edit-year">{$_('document.lab.edit_dialog.year_label')}</Label>
					<Input
						id="edit-year"
						type="number"
						inputmode="numeric"
						min="1000"
						max="9999"
						bind:value={editYear}
						placeholder="2024"
					/>
				</div>
				<div class="space-y-1.5">
					<Label for="edit-doi">{$_('document.lab.edit_dialog.doi_label')}</Label>
					<Input id="edit-doi" bind:value={editDoi} placeholder="10.xxxx/..." maxlength={255} />
				</div>
			</div>
		</div>

		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={editSubmitting}>
				{$_('profile.btn.cancel')}
			</AlertDialog.Cancel>
			<AlertDialog.Action
				disabled={editSubmitting}
				onclick={(e: MouseEvent) => {
					e.preventDefault();
					handleEditSubmit();
				}}
			>
				<Pencil class="size-3.5" />
				{$_('document.lab.edit_dialog.save')}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>

<!-- Delete confirmation -->
<AlertDialog.Root bind:open={deleteDialogOpen}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>{$_('document.lab.delete.title')}</AlertDialog.Title>
			<AlertDialog.Description>
				{$_('document.lab.delete.description', {
					values: { name: deleteTarget?.title ?? deleteTarget?.original_file_name ?? '' }
				})}
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={deleteSubmitting}>
				{$_('profile.btn.cancel')}
			</AlertDialog.Cancel>
			<AlertDialog.Action
				disabled={deleteSubmitting}
				class="text-destructive-foreground bg-destructive hover:bg-destructive/90"
				onclick={(e: MouseEvent) => {
					e.preventDefault();
					handleDeleteSubmit();
				}}
			>
				<Trash2 class="size-3.5" />
				{$_('document.lab.delete.confirm')}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
