<script lang="ts">
	import {
		Copy,
		Check,
		RefreshCw,
		ArrowRightLeft,
		Trash2,
		ShieldAlert,
		FlaskConical,
		Mail
	} from 'lucide-svelte';
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Separator } from '$lib/components/ui/separator';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import labApi, { type LabDetailResponse, type LabMemberInfo } from '$lib/api/lab';
	import { getActiveLab, setActiveLab, invalidateLabs } from '$lib/stores/lab.svelte';
	import { showApiErrors } from '$lib/utils/api-error';

	let activeLab = $derived(getActiveLab());
	let labDetail = $state<LabDetailResponse | null>(null);
	let members = $state<LabMemberInfo[]>([]);
	let isLoading = $state(true);

	// Invite code
	let inviteCodeCopied = $state(false);
	let resettingCode = $state(false);
	let resetDialogOpen = $state(false);

	// Transfer ownership
	let transferTargetId = $state<string>('');
	let transferring = $state(false);
	let transferDialogOpen = $state(false);

	// Delete lab
	let deleteStep = $state<1 | 2>(1);
	let requestingDelete = $state(false);
	let confirmName = $state('');
	let emailCode = $state('');
	let deleting = $state(false);

	$effect(() => {
		const lab = getActiveLab();
		if (lab) {
			isLoading = true;
			Promise.all([labApi.getLab(lab.id), labApi.getMembers(lab.id)])
				.then(([labRes, memberList]) => {
					labDetail = labRes;
					members = memberList.filter((m) => m.role !== 'owner');
					// Reset delete form state on lab switch
					deleteStep = 1;
					confirmName = '';
					emailCode = '';
					transferTargetId = '';
				})
				.catch((error: unknown) => {
					showApiErrors(error, $_('service.get_lab.failed'));
				})
				.finally(() => {
					isLoading = false;
				});
		} else {
			labDetail = null;
			members = [];
			isLoading = false;
		}
	});

	async function copyInviteCode() {
		if (!labDetail) return;
		try {
			await navigator.clipboard.writeText(labDetail.invite_code);
			inviteCodeCopied = true;
			toast.success($_('lab_dashboard.copied'));
			setTimeout(() => (inviteCodeCopied = false), 2000);
		} catch {
			// fallback
		}
	}

	async function handleResetInviteCode() {
		if (!activeLab || !labDetail) return;
		resetDialogOpen = false;
		resettingCode = true;
		try {
			const res = await labApi.resetInviteCode(activeLab.id);
			labDetail.invite_code = res.invite_code;
			toast.success($_('service.reset_invite_code.success'));
		} catch (error: unknown) {
			showApiErrors(error, $_('service.reset_invite_code.failed'));
		} finally {
			resettingCode = false;
		}
	}

	let transferTarget = $derived(members.find((m) => m.user_id === Number(transferTargetId)));

	async function handleTransfer() {
		if (!activeLab) return;
		const targetId = Number(transferTargetId);
		if (!targetId) return;

		transferDialogOpen = false;
		transferring = true;
		try {
			await labApi.transferOwnership(activeLab.id, { target_user_id: targetId });
			toast.success($_('service.transfer_ownership.success'));
			setActiveLab({ ...activeLab, role: 'member' });
			invalidateLabs();
			goto(resolve('/'));
		} catch (error: unknown) {
			showApiErrors(error, $_('service.transfer_ownership.failed'));
		} finally {
			transferring = false;
		}
	}

	async function handleRequestDelete() {
		if (!activeLab) return;
		requestingDelete = true;
		try {
			await labApi.requestDeleteLab(activeLab.id);
			toast.success($_('service.request_delete_lab.success'));
			deleteStep = 2;
		} catch (error: unknown) {
			showApiErrors(error, $_('service.request_delete_lab.failed'));
		} finally {
			requestingDelete = false;
		}
	}

	async function handleConfirmDelete() {
		if (!activeLab) return;
		deleting = true;
		try {
			await labApi.deleteLab(activeLab.id, {
				confirm_name: confirmName,
				email_code: emailCode
			});
			toast.success($_('service.delete_lab.success'));
			setActiveLab(null);
			invalidateLabs();
			goto(resolve('/'));
		} catch (error: unknown) {
			showApiErrors(error, $_('service.delete_lab.failed'));
		} finally {
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('lab.settings.title')} | Sci-Vault</title>
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
	{:else if activeLab.role !== 'owner'}
		<Card.Root class="shadow-sm">
			<Card.Content class="flex flex-col items-center gap-4 py-16 text-center">
				<div
					class="flex size-14 items-center justify-center rounded-2xl bg-yellow-500/10 ring-1 ring-yellow-500/30"
				>
					<ShieldAlert class="size-7 text-yellow-600 dark:text-yellow-400" />
				</div>
				<div>
					<p class="text-lg font-semibold">{$_('lab.settings.title')}</p>
					<p class="mt-1 text-sm text-muted-foreground">{$_('lab.settings.not_owner')}</p>
				</div>
			</Card.Content>
		</Card.Root>
	{:else if isLoading}
		<div class="space-y-6">
			<Skeleton class="h-8 w-48" />
			{#each Array.from({ length: 3 }, (__, i) => i) as i (i)}
				<Card.Root>
					<Card.Content class="space-y-3 p-6">
						<Skeleton class="h-5 w-40" />
						<Skeleton class="h-4 w-full" />
						<Skeleton class="h-10 w-32" />
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{:else if labDetail}
		<div class="space-y-6">
			<div>
				<h1 class="text-2xl font-bold tracking-tight">{$_('lab.settings.title')}</h1>
				<p class="mt-1 text-sm text-muted-foreground">{$_('lab.settings.description')}</p>
			</div>

			<!-- Invite Code Section -->
			<Card.Root>
				<Card.Header>
					<Card.Title class="flex items-center gap-2 text-base">
						<Copy class="size-4" />
						{$_('lab.settings.invite_code_section')}
					</Card.Title>
					<Card.Description>{$_('lab.settings.invite_code_desc')}</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					<div class="flex items-center gap-3">
						<code
							class="flex-1 rounded-md border bg-muted/50 px-4 py-2.5 font-mono text-lg tracking-widest"
						>
							{labDetail.invite_code}
						</code>
						<Button variant="outline" size="sm" class="gap-1.5" onclick={copyInviteCode}>
							{#if inviteCodeCopied}
								<Check class="size-4 text-green-600" />
							{:else}
								<Copy class="size-4" />
							{/if}
						</Button>
					</div>
					<Separator />
					<div>
						<p class="mb-3 text-sm text-muted-foreground">
							{$_('lab.settings.reset_invite_code_desc')}
						</p>
						<AlertDialog.Root bind:open={resetDialogOpen}>
							<AlertDialog.Trigger
								class="inline-flex h-9 items-center gap-2 rounded-md border border-input bg-background px-4 text-sm font-medium shadow-xs hover:bg-accent hover:text-accent-foreground"
								disabled={resettingCode}
							>
								<RefreshCw class="size-4 {resettingCode ? 'animate-spin' : ''}" />
								{$_('lab.settings.reset_invite_code')}
							</AlertDialog.Trigger>
							<AlertDialog.Content>
								<AlertDialog.Header>
									<AlertDialog.Title>{$_('lab.settings.reset_invite_code')}</AlertDialog.Title>
									<AlertDialog.Description>
										{$_('lab.settings.reset_confirm')}
									</AlertDialog.Description>
								</AlertDialog.Header>
								<AlertDialog.Footer>
									<AlertDialog.Cancel>{$_('profile.btn.cancel')}</AlertDialog.Cancel>
									<AlertDialog.Action
										onclick={(e: MouseEvent) => {
											e.preventDefault();
											handleResetInviteCode();
										}}
									>
										<RefreshCw class="size-3.5" />
										{$_('lab.settings.reset_invite_code')}
									</AlertDialog.Action>
								</AlertDialog.Footer>
							</AlertDialog.Content>
						</AlertDialog.Root>
					</div>
				</Card.Content>
			</Card.Root>

			<!-- Transfer Ownership Section -->
			<Card.Root>
				<Card.Header>
					<Card.Title class="flex items-center gap-2 text-base">
						<ArrowRightLeft class="size-4" />
						{$_('lab.settings.transfer_section')}
					</Card.Title>
					<Card.Description>{$_('lab.settings.transfer_desc')}</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					{#if members.length === 0}
						<p class="text-sm text-muted-foreground">{$_('lab.members.empty')}</p>
					{:else}
						<div class="flex items-end gap-3">
							<div class="flex-1 space-y-2">
								<Label for="transfer-target">{$_('lab.settings.select_member')}</Label>
								<select
									id="transfer-target"
									bind:value={transferTargetId}
									class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none"
								>
									<option value="">{$_('lab.settings.select_member')}</option>
									{#each members as member (member.user_id)}
										<option value={String(member.user_id)}>{member.username}</option>
									{/each}
								</select>
							</div>
							<AlertDialog.Root bind:open={transferDialogOpen}>
								<AlertDialog.Trigger
									class="inline-flex h-9 items-center gap-2 rounded-md border border-input bg-background px-4 text-sm font-medium shadow-xs hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50"
									disabled={!transferTargetId || transferring}
								>
									<ArrowRightLeft class="size-4" />
									{$_('lab.settings.transfer_btn')}
								</AlertDialog.Trigger>
								<AlertDialog.Content>
									<AlertDialog.Header>
										<AlertDialog.Title>{$_('lab.settings.transfer_section')}</AlertDialog.Title>
										<AlertDialog.Description>
											{$_('lab.settings.transfer_confirm', {
												values: { name: transferTarget?.username ?? '' }
											})}
										</AlertDialog.Description>
									</AlertDialog.Header>
									<AlertDialog.Footer>
										<AlertDialog.Cancel>{$_('profile.btn.cancel')}</AlertDialog.Cancel>
										<AlertDialog.Action
											variant="destructive"
											onclick={(e: MouseEvent) => {
												e.preventDefault();
												handleTransfer();
											}}
										>
											<ArrowRightLeft class="size-3.5" />
											{$_('lab.settings.transfer_btn')}
										</AlertDialog.Action>
									</AlertDialog.Footer>
								</AlertDialog.Content>
							</AlertDialog.Root>
						</div>
					{/if}
				</Card.Content>
			</Card.Root>

			<!-- Danger Zone -->
			<Card.Root class="border-destructive/30">
				<Card.Header>
					<Card.Title class="flex items-center gap-2 text-base text-destructive">
						<Trash2 class="size-4" />
						{$_('lab.settings.danger_zone')}
					</Card.Title>
					<Card.Description>{$_('lab.settings.delete_desc')}</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					{#if deleteStep === 1}
						<div class="flex items-center gap-3">
							<div class="flex-1">
								<p class="text-sm font-medium">{$_('lab.settings.delete_step1')}</p>
								<p class="mt-0.5 text-xs text-muted-foreground">
									{$_('lab.settings.delete_step1_desc')}
								</p>
							</div>
							<Button
								variant="destructive"
								class="gap-2"
								disabled={requestingDelete}
								onclick={handleRequestDelete}
							>
								<Mail class="size-4" />
								{$_('lab.settings.delete_step1')}
							</Button>
						</div>
					{:else}
						<div class="space-y-4 rounded-lg border border-destructive/20 bg-destructive/5 p-4">
							<p class="text-sm font-medium">{$_('lab.settings.delete_step2')}</p>
							<div class="space-y-3">
								<div class="space-y-1.5">
									<Label for="delete-name">{$_('lab.settings.delete_name_label')}</Label>
									<Input
										id="delete-name"
										bind:value={confirmName}
										placeholder={$_('lab.settings.delete_name_placeholder')}
									/>
								</div>
								<div class="space-y-1.5">
									<Label for="delete-code">{$_('lab.settings.delete_code_label')}</Label>
									<Input
										id="delete-code"
										bind:value={emailCode}
										placeholder={$_('lab.settings.delete_code_placeholder')}
										maxlength={6}
									/>
								</div>
							</div>
							<div class="flex gap-3">
								<Button
									variant="outline"
									onclick={() => {
										deleteStep = 1;
										confirmName = '';
										emailCode = '';
									}}
								>
									{$_('profile.btn.cancel')}
								</Button>
								<Button
									variant="destructive"
									class="gap-2"
									disabled={!confirmName || emailCode.length !== 6 || deleting}
									onclick={handleConfirmDelete}
								>
									<Trash2 class="size-4" />
									{deleting ? $_('lab.settings.deleting') : $_('lab.settings.delete_confirm_btn')}
								</Button>
							</div>
						</div>
					{/if}
				</Card.Content>
			</Card.Root>
		</div>
	{/if}
</div>
