<script lang="ts">
	import { onMount } from 'svelte';
	import { jwtDecode } from 'jwt-decode';
	import { toast } from 'svelte-sonner';
	import { _ } from 'svelte-i18n';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	import AppSidebar from '$lib/components/layout/AppSidebar.svelte';
	import ThemeToggle from '$lib/components/layout/ThemeToggle.svelte';
	import LanguageToggle from '$lib/components/layout/LanguageToggle.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as Separator from '$lib/components/ui/separator';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb';

	let { children } = $props();

	const crumbs = $derived.by((): { label: string; href?: string }[] => {
		const routeId = page.route.id ?? '';
		const base = { label: $_('breadcrumb.dashboard'), href: resolve('/') };
		if (routeId.includes('/settings')) {
			return [base, { label: $_('breadcrumb.settings') }];
		}
		if (routeId.includes('/profile')) {
			return [base, { label: $_('breadcrumb.profile') }];
		}
		if (routeId.includes('/documents/mine')) {
			return [base, { label: $_('breadcrumb.documents') }, { label: $_('breadcrumb.my_documents') }];
		}
		if (routeId.includes('/documents/upload')) {
			return [base, { label: $_('breadcrumb.documents') }, { label: $_('breadcrumb.upload') }];
		}
		return [base];
	});

	onMount(() => {
		const token = localStorage.getItem('token');
		if (!token) {
			goto(resolve('/login'));
			return;
		}

		try {
			const decoded = jwtDecode<{ exp?: number }>(token);
			const currentTime = Date.now() / 1000;
			if (decoded.exp && decoded.exp < currentTime) {
				localStorage.removeItem('token');
				localStorage.removeItem('user');
				toast.error('Token expired, please login again.');
				goto(resolve('/login'));
			}
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
		} catch (error: unknown) {
			localStorage.removeItem('token');
			localStorage.removeItem('user');
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
						{#each crumbs as crumb, i (crumb.label)}
							{#if i > 0}
								<Breadcrumb.Separator />
							{/if}
							<Breadcrumb.Item>
								{#if crumb.href}
									<Breadcrumb.Link href={crumb.href}>{crumb.label}</Breadcrumb.Link>
								{:else}
									<Breadcrumb.Page>{crumb.label}</Breadcrumb.Page>
								{/if}
							</Breadcrumb.Item>
						{/each}
					</Breadcrumb.List>
				</Breadcrumb.Root>
			</div>
			<div class="flex items-center gap-2 px-4">
				<ThemeToggle />
				<LanguageToggle />
			</div>
		</header>
		<div class="mt-4 flex flex-1 flex-col gap-4 p-4 pt-0">
			{@render children()}
		</div>
	</Sidebar.Inset>
</Sidebar.Provider>
