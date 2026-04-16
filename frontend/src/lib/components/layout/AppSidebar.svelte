<script lang="ts">
	import {
		Activity,
		Check,
		ChevronRight,
		ChevronsUpDown,
		Compass,
		FileText,
		FlaskConical,
		LogOut,
		Plus,
		Search,
		Settings,
		Upload,
		User,
		Users
	} from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';

	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Collapsible } from 'bits-ui';
	import * as Avatar from '$lib/components/ui/avatar';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import userApi from '$lib/api/user';
	import labApi, { type LabListItem } from '$lib/api/lab';
	import { getLabsVersion, getActiveLab, setActiveLab } from '$lib/stores/lab.svelte';
	import { getUser, getAvatarUrl, setAvatarUrl, clearUser } from '$lib/stores/user.svelte';

	let { ref = $bindable(null), ...restProps } = $props();

	let currentUser = $derived(getUser());
	let userInitials = $derived(
		currentUser.username ? currentUser.username.substring(0, 2).toUpperCase() : 'US'
	);

	let initDone = $state(false);
	let avatarUrl = $derived(getAvatarUrl());

	// Lab selector state
	let myLabs = $state<LabListItem[]>([]);
	let labsLoaded = $state(false);
	let selectedLabId = $state<number | null>(null);
	let selectedLab = $derived(myLabs.find((l) => l.id === selectedLabId) ?? null);

	// Lab / workspace-level items
	const topNavItems = [
		{ title: 'sidebar.lab_dashboard', url: '/' as const, icon: FlaskConical },
		{ title: 'sidebar.lab_members', url: '/members' as const, icon: Users },
		{ title: 'sidebar.lab_settings', url: '/lab-settings' as const, icon: Settings }
	];

	// Personal items
	let bottomNavItems = $derived([
		{ title: 'sidebar.dashboard', url: '/mine/dashboard' as const, icon: Compass },
		{
			title: 'sidebar.documents',
			icon: FileText,
			items: [
				{ title: 'sidebar.my_documents', url: '/documents/mine' as const, icon: FileText },
				{ title: 'sidebar.upload', url: '/documents/upload' as const, icon: Upload }
			]
		},
		{
			title: 'sidebar.account',
			icon: User,
			items: [
				{ title: 'sidebar.profile', url: `/profile/${currentUser.id}` as const, icon: User },
				{ title: 'sidebar.settings', url: '/settings' as const, icon: Settings }
			]
		}
	]);

	let isDocGroupActive = $derived(page.route.id?.startsWith('/(dashboard)/documents') ?? false);
	let isAccountGroupActive = $derived(
		page.route.id?.startsWith('/(dashboard)/profile') || page.route.id === '/(dashboard)/settings'
	);

	const groupActiveMap: Record<string, boolean> = $derived({
		'sidebar.documents': isDocGroupActive,
		'sidebar.account': isAccountGroupActive
	});

	async function reloadLabs() {
		const result = await labApi.getMyLabs().catch(() => null);
		if (result === null) return;
		myLabs = result;
		if (selectedLabId !== null && !myLabs.some((l) => l.id === selectedLabId)) {
			selectLab(null);
		}
		if (selectedLabId === null && myLabs.length === 1) {
			selectLab(myLabs[0].id);
		}
		// Keep active lab store in sync (e.g. if name/role changed)
		if (selectedLabId !== null) {
			const current = myLabs.find((l) => l.id === selectedLabId);
			if (current) {
				setActiveLab({ id: current.id, name: current.name, role: current.role });
			}
		}
		labsLoaded = true;
	}

	// Re-fetch labs whenever a page signals a change (join / create).
	$effect(() => {
		getLabsVersion(); // tracked dependency
		reloadLabs();
	});

	onMount(async () => {
		// Restore persisted lab selection from store
		const storedLab = getActiveLab();
		if (storedLab) selectedLabId = storedLab.id;

		await Promise.allSettled([
			userApi
				.getAvatar(currentUser.id)
				.then((a) => setAvatarUrl(a.avatar_url))
				.catch(() => {}),
			reloadLabs()
		]);

		initDone = true;
	});

	function selectLab(id: number | null) {
		selectedLabId = id;
		if (id === null) {
			setActiveLab(null);
		} else {
			const lab = myLabs.find((l) => l.id === id);
			if (lab) {
				setActiveLab({ id: lab.id, name: lab.name, role: lab.role });
			}
		}
	}

	function handleLogout() {
		localStorage.removeItem('token');
		setActiveLab(null);
		clearUser();
		goto(resolve('/login'));
	}
</script>

<Sidebar.Root collapsible="offcanvas" bind:ref {...restProps}>
	<Sidebar.Header class="border-b p-0">
		<!-- App logo -->
		<a href={resolve('/welcome')} class="flex h-16 items-center gap-2 px-4">
			<div
				class="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"
			>
				<Activity class="size-4" />
			</div>
			<div class="flex flex-col gap-0.5 leading-none">
				<span class="font-semibold">{$_('app.title')}</span>
				<span class="">{$_('app.version')}</span>
			</div>
		</a>

		<!-- Lab selector -->
		<div class="px-2 pb-2">
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Sidebar.MenuButton
							size="lg"
							class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
							{...props}
						>
							{#if !labsLoaded}
								<Skeleton class="size-8 rounded-lg" />
								<div class="grid flex-1 gap-1">
									<Skeleton class="h-3.5 w-28" />
									<Skeleton class="h-3 w-16" />
								</div>
							{:else if selectedLab}
								<div
									class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10"
								>
									<FlaskConical class="size-4 text-primary" />
								</div>
								<div class="grid flex-1 text-left text-sm leading-tight">
									<span class="truncate font-semibold">{selectedLab.name}</span>
									<span class="truncate text-xs text-muted-foreground capitalize">
										{$_(`profile.labs.role.${selectedLab.role}`)}
									</span>
								</div>
							{:else}
								<div class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted">
									<FlaskConical class="size-4 text-muted-foreground" />
								</div>
								<span class="flex-1 text-left text-sm text-muted-foreground">
									{myLabs.length === 0 ? $_('sidebar.no_labs_hint') : $_('sidebar.select_lab')}
								</span>
							{/if}
							<ChevronsUpDown class="ml-auto size-4 shrink-0 text-muted-foreground" />
						</Sidebar.MenuButton>
					{/snippet}
				</DropdownMenu.Trigger>

				<DropdownMenu.Content
					class="w-[--bits-dropdown-menu-anchor-width] min-w-56 rounded-lg"
					side="bottom"
					align="start"
					sideOffset={4}
				>
					{#if myLabs.length > 0}
						<DropdownMenu.Label class="text-xs text-muted-foreground">
							{$_('sidebar.your_labs')}
						</DropdownMenu.Label>
						{#each myLabs as lab (lab.id)}
							<DropdownMenu.Item onclick={() => selectLab(lab.id)} class="gap-2">
								<div class="flex size-6 shrink-0 items-center justify-center rounded bg-primary/10">
									<FlaskConical class="size-3.5 text-primary" />
								</div>
								<div class="grid flex-1 leading-tight">
									<span class="truncate text-sm font-medium">{lab.name}</span>
									<span class="text-xs text-muted-foreground capitalize">
										{$_(`profile.labs.role.${lab.role}`)}
										· {lab.member_count}
										{$_('sidebar.members')}
									</span>
								</div>
								{#if selectedLabId === lab.id}
									<Check class="ml-auto size-4 text-primary" />
								{/if}
							</DropdownMenu.Item>
						{/each}
						<DropdownMenu.Separator />
					{:else}
						<div class="px-2 py-3 text-center">
							<p class="text-sm font-medium">{$_('sidebar.no_labs')}</p>
							<p class="mt-0.5 text-xs text-muted-foreground">{$_('sidebar.no_labs_desc')}</p>
						</div>
						<DropdownMenu.Separator />
					{/if}
					<DropdownMenu.Group>
						<DropdownMenu.Item onclick={() => goto(resolve('/labs/create'))} class="gap-2">
							<Plus class="size-4" />
							{$_('sidebar.create_lab')}
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => goto(resolve('/labs/join'))} class="gap-2">
							<FlaskConical class="size-4" />
							{$_('sidebar.join_lab')}
						</DropdownMenu.Item>
					</DropdownMenu.Group>
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		</div>
	</Sidebar.Header>

	<Sidebar.Content>
		<!-- Search — standalone, spans all scopes -->
		<Sidebar.Group>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					<Sidebar.MenuItem>
						<Sidebar.MenuButton isActive={page.url.pathname === resolve('/search')}>
							{#snippet child({ props })}
								<a href={resolve('/search')} {...props}>
									<Search />
									<span>{$_('sidebar.search')}</span>
								</a>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Sidebar.Separator />

		<!-- Lab / workspace-level navigation -->
		<Sidebar.Group>
			<Sidebar.GroupLabel>{$_('sidebar.workspace')}</Sidebar.GroupLabel>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					{#each topNavItems as item (item.title)}
						<Sidebar.MenuItem>
							<Sidebar.MenuButton isActive={page.url.pathname === resolve(item.url)}>
								{#snippet child({ props })}
									<a href={resolve(item.url)} {...props}>
										<item.icon />
										<span>{$_(item.title)}</span>
									</a>
								{/snippet}
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					{/each}
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Sidebar.Separator />

		<!-- Personal navigation -->
		<Sidebar.Group>
			<Sidebar.GroupLabel>{$_('sidebar.personal')}</Sidebar.GroupLabel>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					{#each bottomNavItems as item (item.title)}
						{#if item.items}
							<Collapsible.Root
								open={groupActiveMap[item.title] ?? false}
								class="group/collapsible"
							>
								<Sidebar.MenuItem>
									<Collapsible.Trigger>
										{#snippet child({ props })}
											<Sidebar.MenuButton {...props}>
												<item.icon />
												<span>{$_(item.title)}</span>
												<ChevronRight
													class="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90"
												/>
											</Sidebar.MenuButton>
										{/snippet}
									</Collapsible.Trigger>
									<Collapsible.Content>
										<Sidebar.MenuSub>
											{#each item.items as subItem (subItem.title)}
												<Sidebar.MenuSubItem>
													<Sidebar.MenuSubButton
														isActive={page.url.pathname === resolve(subItem.url) ||
															(subItem.url === '/documents/mine' &&
																page.route.id === '/(dashboard)/documents/[id]')}
													>
														{#snippet child({ props })}
															<a href={resolve(subItem.url)} {...props}>
																<subItem.icon class="size-3.5" />
																<span>{$_(subItem.title)}</span>
															</a>
														{/snippet}
													</Sidebar.MenuSubButton>
												</Sidebar.MenuSubItem>
											{/each}
										</Sidebar.MenuSub>
									</Collapsible.Content>
								</Sidebar.MenuItem>
							</Collapsible.Root>
						{:else}
							<Sidebar.MenuItem>
								<Sidebar.MenuButton isActive={page.url.pathname === resolve(item.url)}>
									{#snippet child({ props })}
										<a href={resolve(item.url)} {...props}>
											<item.icon />
											<span>{$_(item.title)}</span>
										</a>
									{/snippet}
								</Sidebar.MenuButton>
							</Sidebar.MenuItem>
						{/if}
					{/each}
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>

	<Sidebar.Footer>
		<Sidebar.Menu>
			<!-- User / logout -->
			<Sidebar.MenuItem>
				<DropdownMenu.Root>
					<DropdownMenu.Trigger>
						{#snippet child({ props })}
							<Sidebar.MenuButton
								size="lg"
								class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
								{...props}
							>
								{#if !initDone}
									<div class="flex items-center gap-2">
										<Skeleton class="h-8 w-8 rounded-lg" />
										<div class="grid flex-1 gap-1">
											<Skeleton class="h-4 w-20" />
											<Skeleton class="h-3 w-24" />
										</div>
									</div>
								{:else}
									<Avatar.Root class="h-8 w-8 rounded-lg">
										{#if avatarUrl}
											<Avatar.Image
												src={avatarUrl}
												alt={currentUser.username}
												class="object-cover"
											/>
										{/if}
										<Avatar.Fallback class="rounded-lg">{userInitials}</Avatar.Fallback>
									</Avatar.Root>
									<div class="grid flex-1 text-left text-sm leading-tight">
										<span class="truncate font-semibold">{currentUser.username}</span>
										<span class="truncate text-xs">{currentUser.email}</span>
									</div>
									<ChevronsUpDown class="ml-auto size-4" />
								{/if}
							</Sidebar.MenuButton>
						{/snippet}
					</DropdownMenu.Trigger>
					<DropdownMenu.Content
						class="w-[--bits-dropdown-menu-anchor-width] min-w-56 rounded-lg"
						side="top"
						align="end"
						sideOffset={4}
					>
						<DropdownMenu.Label class="p-0 font-normal">
							<div class="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
								<Avatar.Root class="h-8 w-8 rounded-lg">
									{#if avatarUrl}
										<Avatar.Image src={avatarUrl} alt={currentUser.username} class="object-cover" />
									{/if}
									<Avatar.Fallback class="rounded-lg">{userInitials}</Avatar.Fallback>
								</Avatar.Root>
								<div class="grid flex-1 text-left text-sm leading-tight">
									<span class="truncate font-semibold">{currentUser.username}</span>
									<span class="truncate text-xs">{currentUser.email}</span>
								</div>
							</div>
						</DropdownMenu.Label>
						<DropdownMenu.Separator />
						<DropdownMenu.Item onclick={handleLogout}>
							<LogOut />
							{$_('sidebar.logout')}
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Sidebar.MenuItem>
		</Sidebar.Menu>
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
