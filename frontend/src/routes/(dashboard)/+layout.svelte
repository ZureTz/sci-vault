<script lang="ts">
	import { onMount } from 'svelte';
	import { jwtDecode } from 'jwt-decode';
	import { toast } from 'svelte-sonner';
	import { _ } from 'svelte-i18n';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	import AppSidebar from '$lib/components/layout/AppSidebar.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as Separator from '$lib/components/ui/separator';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb';

	let { children } = $props();

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
				toast.error('Token expired, please login again.');
				goto(resolve('/login'));
			}
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
		} catch (error: unknown) {
			localStorage.removeItem('token');
			goto(resolve('/login'));
		}
	});
</script>

<Sidebar.Provider>
	<AppSidebar />
	<Sidebar.Inset>
		<header
			class="flex h-16 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12"
		>
			<div class="flex items-center gap-2 px-4">
				<Sidebar.Trigger class="-ml-1" />
				<Separator.Root orientation="vertical" class="mr-2 h-4" />
				<Breadcrumb.Root>
					<Breadcrumb.List>
						<Breadcrumb.Item>
							<Breadcrumb.Link href="/">{$_('breadcrumb.dashboard')}</Breadcrumb.Link>
						</Breadcrumb.Item>
					</Breadcrumb.List>
				</Breadcrumb.Root>
			</div>
		</header>
		<div class="mt-4 flex flex-1 flex-col gap-4 p-4 pt-0">
			{@render children()}
		</div>
	</Sidebar.Inset>
</Sidebar.Provider>
