<script lang="ts">
	import { Crown, UserMinus, FlaskConical } from 'lucide-svelte';
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Card from '$lib/components/ui/card';
	import * as Avatar from '$lib/components/ui/avatar';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import labApi, { type LabMemberInfo } from '$lib/api/lab';
	import { getActiveLab } from '$lib/stores/lab.svelte';
	import { getUser } from '$lib/stores/user.svelte';
	import { showApiErrors } from '$lib/utils/api-error';

	let activeLab = $derived(getActiveLab());
	let currentUser = $derived(getUser());
	let members = $state<LabMemberInfo[]>([]);
	let isLoading = $state(true);
	let kickingUserId = $state<number | null>(null);
	let kickDialogOpen = $state(false);

	$effect(() => {
		const lab = getActiveLab();
		if (lab) {
			isLoading = true;
			labApi
				.getMembers(lab.id)
				.then((result) => {
					members = result;
				})
				.catch((error: unknown) => {
					showApiErrors(error, $_('service.get_members.failed'));
				})
				.finally(() => {
					isLoading = false;
				});
		} else {
			members = [];
			isLoading = false;
		}
	});

	async function handleKick(member: LabMemberInfo) {
		if (!activeLab) return;
		kickDialogOpen = false;
		kickingUserId = member.user_id;
		try {
			await labApi.kickMember(activeLab.id, member.user_id);
			toast.success($_('service.kick_member.success'));
			members = members.filter((m) => m.user_id !== member.user_id);
		} catch (error: unknown) {
			showApiErrors(error, $_('service.kick_member.failed'));
		} finally {
			kickingUserId = null;
		}
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}
</script>

<svelte:head>
	<title>{$_('lab.members.title')} | Sci-Vault</title>
</svelte:head>

<div class="container mx-auto max-w-3xl px-4 py-8">
	{#if !activeLab}
		<Card.Root class="shadow-sm">
			<Card.Content class="flex flex-col items-center gap-4 py-16 text-center">
				<div
					class="flex size-14 items-center justify-center rounded-2xl bg-primary/10 ring-1 ring-border/50"
				>
					<FlaskConical class="size-7 text-primary" />
				</div>
				<div>
					<p class="text-lg font-semibold">{$_('lab_dashboard.no_lab_selected')}</p>
					<p class="mt-1 text-sm text-muted-foreground">
						{$_('lab_dashboard.no_lab_selected_desc')}
					</p>
				</div>
			</Card.Content>
		</Card.Root>
	{:else}
		<div class="mb-6">
			<h1 class="text-2xl font-bold tracking-tight">{$_('lab.members.title')}</h1>
			<p class="mt-1 text-sm text-muted-foreground">{$_('lab.members.description')}</p>
		</div>

		<Card.Root>
			<Card.Content class="p-0">
				{#if isLoading}
					<div class="divide-y">
						{#each Array.from({ length: 4 }, (__, i) => i) as i (i)}
							<div class="flex items-center gap-4 p-4">
								<Skeleton class="size-10 rounded-full" />
								<div class="flex-1 space-y-2">
									<Skeleton class="h-4 w-32" />
									<Skeleton class="h-3 w-20" />
								</div>
								<Skeleton class="h-6 w-16 rounded-full" />
							</div>
						{/each}
					</div>
				{:else if members.length === 0}
					<div class="py-12 text-center">
						<p class="text-sm text-muted-foreground">{$_('lab.members.empty')}</p>
					</div>
				{:else}
					<div class="divide-y">
						{#each members as member (member.user_id)}
							<div class="flex items-center gap-4 p-4 transition-colors hover:bg-muted/30">
								<button
									class="shrink-0 cursor-pointer"
									onclick={() => goto(resolve(`/profile/${member.user_id}`))}
								>
									<Avatar.Root class="size-10">
										<Avatar.Fallback>
											{member.username.substring(0, 2).toUpperCase()}
										</Avatar.Fallback>
									</Avatar.Root>
								</button>
								<div class="min-w-0 flex-1">
									<div class="flex items-center gap-2">
										<button
											class="cursor-pointer truncate text-sm font-medium hover:underline"
											onclick={() => goto(resolve(`/profile/${member.user_id}`))}
										>
											{member.username}
										</button>
										{#if member.user_id === Number(currentUser.id)}
											<span class="text-xs text-muted-foreground">{$_('lab.members.you')}</span>
										{/if}
									</div>
									<p class="text-xs text-muted-foreground">
										{$_('lab.members.joined')}
										{formatDate(member.joined_at)}
									</p>
								</div>
								<div class="flex items-center gap-2">
									{#if member.role === 'owner'}
										<Badge
											variant="outline"
											class="gap-1 border-yellow-500/30 bg-yellow-500/10 text-yellow-700 dark:text-yellow-400"
										>
											<Crown class="size-3" />
											{$_('lab.members.owner')}
										</Badge>
									{:else}
										<Badge variant="secondary">{$_('lab.members.member')}</Badge>
									{/if}

									{#if activeLab.role === 'owner' && member.role !== 'owner' && member.user_id !== Number(currentUser.id)}
										<AlertDialog.Root bind:open={kickDialogOpen}>
											<AlertDialog.Trigger
												class="inline-flex h-8 items-center gap-1 rounded-md px-3 text-sm text-destructive hover:bg-destructive/10 hover:text-destructive"
											>
												<UserMinus class="size-3.5" />
												{$_('lab.members.remove')}
											</AlertDialog.Trigger>
											<AlertDialog.Content>
												<AlertDialog.Header>
													<AlertDialog.Title>{$_('lab.members.remove')}</AlertDialog.Title>
													<AlertDialog.Description>
														{$_('lab.members.remove_confirm', {
															values: { name: member.username }
														})}
													</AlertDialog.Description>
												</AlertDialog.Header>
												<AlertDialog.Footer>
													<AlertDialog.Cancel>{$_('profile.btn.cancel')}</AlertDialog.Cancel>
													<AlertDialog.Action
														variant="destructive"
														disabled={kickingUserId === member.user_id}
														onclick={(e: MouseEvent) => {
															e.preventDefault();
															handleKick(member);
														}}
													>
														<UserMinus class="size-3.5" />
														{$_('lab.members.remove')}
													</AlertDialog.Action>
												</AlertDialog.Footer>
											</AlertDialog.Content>
										</AlertDialog.Root>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</Card.Content>
		</Card.Root>
	{/if}
</div>
