<script lang="ts">
	import { onMount } from 'svelte';
	import { jwtDecode } from 'jwt-decode';
	import { MediaQuery } from 'svelte/reactivity';
	import { toast } from 'svelte-sonner';
	import { _ } from 'svelte-i18n';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	import AppSidebar from '$lib/components/layout/AppSidebar.svelte';
	import { getDocDetailOrigin } from '$lib/stores/nav.svelte';
	import { clearUser } from '$lib/stores/user.svelte';
	import ThemeToggle from '$lib/components/layout/ThemeToggle.svelte';
	import LanguageToggle from '$lib/components/layout/LanguageToggle.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as Separator from '$lib/components/ui/separator';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb';
	import * as Drawer from '$lib/components/ui/drawer';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { buttonVariants } from '$lib/components/ui/button';

	let { children } = $props();

	let collapseOpen = $state(false);
	const isDesktop = new MediaQuery('(min-width: 768px)');
	// Desktop: show full trail up to 3 items; collapse middle when deeper.
	// Mobile: show only the current page; collapse ALL ancestors into the ellipsis.
	const DESKTOP_MAX = 3;

	type CrumbHref =
		| '/'
		| '/mine/dashboard'
		| '/search'
		| '/mine/history'
		| '/documents/mine'
		| '/documents/upload';
	type Crumb = { label: string; href?: CrumbHref };

	const crumbs = $derived.by((): Crumb[] => {
		const routeId = page.route.id;

		// Group roots — clickable links (href is the unresolved route id; resolved at the link site)
		const labBase: Crumb = { label: $_('breadcrumb.lab_dashboard'), href: '/' };
		const personalBase: Crumb = {
			label: $_('breadcrumb.dashboard'),
			href: '/mine/dashboard'
		};
		// Intermediate group labels — not real pages, so not clickable
		const documentsCrumb = { label: $_('breadcrumb.documents') };
		const accountCrumb = { label: $_('sidebar.account') };

		if (!routeId) return [labBase];

		switch (routeId) {
			// ── Workspace group ───────────────────────────────────────────
			case '/(dashboard)':
				return [{ label: $_('breadcrumb.lab_dashboard') }];
			case '/(dashboard)/labs/create':
				return [labBase, { label: $_('breadcrumb.create_lab') }];
			case '/(dashboard)/labs/join':
				return [labBase, { label: $_('breadcrumb.join_lab') }];
			case '/(dashboard)/search':
				return [{ label: $_('breadcrumb.search') }];
			case '/(dashboard)/members':
				return [labBase, { label: $_('breadcrumb.lab_members') }];
			case '/(dashboard)/lab-documents':
				return [labBase, { label: $_('breadcrumb.lab_documents') }];
			case '/(dashboard)/lab-settings':
				return [labBase, { label: $_('breadcrumb.lab_settings') }];

			// ── Personal group ────────────────────────────────────────────
			case '/(dashboard)/mine/dashboard':
				return [{ label: $_('breadcrumb.dashboard') }];
			case '/(dashboard)/documents':
				return [personalBase, { label: $_('breadcrumb.documents') }];
			case '/(dashboard)/documents/mine':
				return [personalBase, documentsCrumb, { label: $_('breadcrumb.my_documents') }];
			case '/(dashboard)/documents/upload':
				return [personalBase, documentsCrumb, { label: $_('breadcrumb.upload') }];
			case '/(dashboard)/mine/history':
				return [personalBase, { label: $_('breadcrumb.history') }];
			case '/(dashboard)/documents/[id]': {
				// The detail page is reachable from My Documents, Upload, Search,
				// History, or directly. The trail leans on the captured origin
				// (set by the sidebar's afterNavigate hook) so the breadcrumb
				// reflects how the user actually got here.
				const detail = { label: $_('breadcrumb.document_detail') };
				switch (getDocDetailOrigin()) {
					case '/search':
						return [{ label: $_('breadcrumb.search'), href: '/search' }, detail];
					case '/mine/history':
						return [
							personalBase,
							{ label: $_('breadcrumb.history'), href: '/mine/history' },
							detail
						];
					case '/documents/upload':
						return [
							personalBase,
							documentsCrumb,
							{ label: $_('breadcrumb.upload'), href: '/documents/upload' },
							detail
						];
					case '/documents/mine':
						return [
							personalBase,
							documentsCrumb,
							{ label: $_('breadcrumb.my_documents'), href: '/documents/mine' },
							detail
						];
					default:
						return [personalBase, documentsCrumb, detail];
				}
			}
			case '/(dashboard)/profile':
			case '/(dashboard)/profile/[user_id]':
				return [personalBase, accountCrumb, { label: $_('breadcrumb.profile') }];
			case '/(dashboard)/settings':
				return [personalBase, accountCrumb, { label: $_('breadcrumb.settings') }];

			default:
				return [labBase];
		}
	});

	// Split crumbs into (head, middle, tail) so rendering is branchless in the template.
	// - No collapse: head = all crumbs.
	// - Desktop collapse (>DESKTOP_MAX): head = [first], middle = ...inner, tail = last 2.
	// - Mobile collapse (>1): head = [], middle = all but last, tail = [last].
	const crumbRender = $derived.by(() => {
		const empty: Crumb[] = [];
		if (isDesktop.current) {
			if (crumbs.length <= DESKTOP_MAX) {
				return { head: crumbs, middle: empty, tail: empty };
			}
			return {
				head: crumbs.slice(0, 1),
				middle: crumbs.slice(1, -2),
				tail: crumbs.slice(-2)
			};
		}
		if (crumbs.length <= 1) {
			return { head: crumbs, middle: empty, tail: empty };
		}
		return {
			head: empty,
			middle: crumbs.slice(0, -1),
			tail: crumbs.slice(-1)
		};
	});

	onMount(() => {
		const token = localStorage.getItem('token');
		if (!token) {
			goto(resolve('/welcome'));
			return;
		}

		try {
			const decoded = jwtDecode<{ exp?: number }>(token);
			const currentTime = Date.now() / 1000;
			if (decoded.exp && decoded.exp < currentTime) {
				localStorage.removeItem('token');
				clearUser();
				toast.error('Token expired, please login again.');
				goto(resolve('/login'));
			}
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
		} catch (error: unknown) {
			localStorage.removeItem('token');
			clearUser();
			goto(resolve('/login'));
		}
	});
</script>

<Sidebar.Provider>
	<AppSidebar />
	<Sidebar.Inset class="min-w-0 overflow-x-clip">
		<header
			class="sticky top-0 z-10 flex h-16 shrink-0 items-center gap-2 border-b bg-background transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12"
		>
			<div class="flex flex-1 items-center gap-2 px-4">
				<Sidebar.Trigger class="-ml-1" />
				<Separator.Root orientation="vertical" class="mr-2 h-4" />
				<Breadcrumb.Root>
					<Breadcrumb.List>
						{@const { head, middle, tail } = crumbRender}

						{#each head as crumb, i (crumb.label)}
							{#if i > 0}
								<Breadcrumb.Separator />
							{/if}
							{@render crumbItem(crumb)}
						{/each}

						{#if middle.length > 0}
							{#if head.length > 0}
								<Breadcrumb.Separator />
							{/if}
							<Breadcrumb.Item>
								{#if isDesktop.current}
									<DropdownMenu.Root bind:open={collapseOpen}>
										<DropdownMenu.Trigger
											class="flex items-center gap-1"
											aria-label={$_('breadcrumb.toggle_menu')}
										>
											<Breadcrumb.Ellipsis class="size-4" />
										</DropdownMenu.Trigger>
										<DropdownMenu.Content align="start">
											{#each middle as crumb, i (i)}
												<DropdownMenu.Item>
													{#if crumb.href}
														<a href={resolve(crumb.href)}>{crumb.label}</a>
													{:else}
														<span>{crumb.label}</span>
													{/if}
												</DropdownMenu.Item>
											{/each}
										</DropdownMenu.Content>
									</DropdownMenu.Root>
								{:else}
									<Drawer.Root bind:open={collapseOpen}>
										<Drawer.Trigger aria-label={$_('breadcrumb.toggle_menu')}>
											<Breadcrumb.Ellipsis class="size-4" />
										</Drawer.Trigger>
										<Drawer.Content>
											<Drawer.Header class="text-start">
												<Drawer.Title>{$_('breadcrumb.navigate_to')}</Drawer.Title>
												<Drawer.Description>
													{$_('breadcrumb.navigate_to_desc')}
												</Drawer.Description>
											</Drawer.Header>
											<div class="grid gap-1 px-4">
												{#each middle as crumb, i (i)}
													{#if crumb.href}
														<a href={resolve(crumb.href)} class="py-2 text-sm">{crumb.label}</a>
													{:else}
														<span class="py-2 text-sm font-medium">{crumb.label}</span>
													{/if}
												{/each}
											</div>
											<Drawer.Footer class="pt-4">
												<Drawer.Close class={buttonVariants({ variant: 'outline' })}>
													{$_('common.close')}
												</Drawer.Close>
											</Drawer.Footer>
										</Drawer.Content>
									</Drawer.Root>
								{/if}
							</Breadcrumb.Item>
						{/if}

						{#each tail as crumb, i (crumb.label)}
							{#if i > 0 || middle.length > 0 || head.length > 0}
								<Breadcrumb.Separator />
							{/if}
							{@render crumbItem(crumb)}
						{/each}
					</Breadcrumb.List>
				</Breadcrumb.Root>

				{#snippet crumbItem(crumb: Crumb)}
					<Breadcrumb.Item>
						{#if crumb.href}
							<Breadcrumb.Link href={resolve(crumb.href)} class="max-w-32 truncate md:max-w-none">
								{crumb.label}
							</Breadcrumb.Link>
						{:else}
							<Breadcrumb.Page class="max-w-32 truncate md:max-w-none">
								{crumb.label}
							</Breadcrumb.Page>
						{/if}
					</Breadcrumb.Item>
				{/snippet}
			</div>
			<div class="flex items-center gap-2 px-4">
				<ThemeToggle />
				<LanguageToggle />
			</div>
		</header>
		<main class="container mx-auto mt-4 flex flex-1 flex-col gap-4 p-4 pt-0">
			{@render children()}
		</main>
	</Sidebar.Inset>
</Sidebar.Provider>
