<script lang="ts">
	import { Activity, Compass, LogOut, Settings, User } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Avatar from '$lib/components/ui/avatar';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { Skeleton } from '$lib/components/ui/skeleton';

	let { ref = $bindable(null), ...restProps } = $props();

	let currentUser = $state({ id: '', username: 'User', email: '' });
	let userInitials = $derived(
		currentUser.username ? currentUser.username.substring(0, 2).toUpperCase() : 'US'
	);

	let initDone = $state(false);

	onMount(() => {
		const userStr = localStorage.getItem('user');
		if (userStr) {
			try {
				currentUser = JSON.parse(userStr);
			} catch (e) {
				console.error('Failed to parse user info', e);
			}
		}
		initDone = true;
	});

	function handleLogout() {
		localStorage.removeItem('token');
		localStorage.removeItem('user');
		goto(resolve('/login'));
	}
</script>

<Sidebar.Root bind:ref {...restProps}>
	<Sidebar.Header class="h-16 justify-center border-b p-0 transition-[height] ease-linear">
		<div class="flex items-center gap-2 px-4">
			<div
				class="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"
			>
				<Activity class="size-4" />
			</div>
			<div class="flex flex-col gap-0.5 leading-none">
				<span class="font-semibold">{$_('app.title')}</span>
				<span class="">{$_('app.version')}</span>
			</div>
		</div>
	</Sidebar.Header>
	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupLabel>{$_('sidebar.navigation')}</Sidebar.GroupLabel>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					<Sidebar.MenuItem>
						<Sidebar.MenuButton>
							{#snippet child({ props })}
								<a href={resolve('/')} {...props}>
									<Compass />
									<span>{$_('sidebar.dashboard')}</span>
								</a>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
					<Sidebar.MenuItem>
						<Sidebar.MenuButton>
							{#snippet child({ props })}
								<a href={resolve('/settings')} {...props}>
									<Settings />
									<span>{$_('sidebar.settings')}</span>
								</a>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>
	<Sidebar.Footer>
		<Sidebar.Menu>
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
										<Avatar.Fallback class="rounded-lg">{userInitials}</Avatar.Fallback>
									</Avatar.Root>
									<div class="grid flex-1 text-left text-sm leading-tight">
										<span class="truncate font-semibold">{currentUser.username}</span>
										<span class="truncate text-xs">{currentUser.email}</span>
									</div>
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
									<Avatar.Fallback class="rounded-lg">{userInitials}</Avatar.Fallback>
								</Avatar.Root>
								<div class="grid flex-1 text-left text-sm leading-tight">
									<span class="truncate font-semibold">{currentUser.username}</span>
									<span class="truncate text-xs">{currentUser.email}</span>
								</div>
							</div>
						</DropdownMenu.Label>
						<DropdownMenu.Separator />
						<DropdownMenu.Group>
							<DropdownMenu.Item>
								<User />
								{$_('sidebar.profile')}
							</DropdownMenu.Item>
						</DropdownMenu.Group>
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
