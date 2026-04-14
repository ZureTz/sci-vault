<script lang="ts">
	import {
		FlaskConical,
		Copy,
		Check,
		Users,
		Settings,
		Crown,
		ArrowRight,
		LogOut,
		Mail
	} from 'lucide-svelte';
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Separator } from '$lib/components/ui/separator';
	import labApi, { type LabDetailResponse } from '$lib/api/lab';
	import { getActiveLab, setActiveLab, invalidateLabs } from '$lib/stores/lab.svelte';
	import { showApiErrors } from '$lib/utils/api-error';

	let activeLab = $derived(getActiveLab());
	let labDetail = $state<LabDetailResponse | null>(null);
	let isLoading = $state(true);
	let copied = $state(false);

	// Leave lab state
	let leaveDialogOpen = $state(false);
	let leaveCodeSent = $state(false);
	let requestingLeave = $state(false);
	let leaveEmailCode = $state('');
	let confirmingLeave = $state(false);

	function resetLeaveState() {
		leaveDialogOpen = false;
		leaveCodeSent = false;
		leaveEmailCode = '';
	}

	$effect(() => {
		const lab = getActiveLab();
		// Reset leave flow on lab switch
		resetLeaveState();
		if (lab) {
			isLoading = true;
			labApi
				.getLab(lab.id)
				.then((detail) => {
					labDetail = detail;
				})
				.catch((error: unknown) => {
					showApiErrors(error, $_('service.get_lab.failed'));
				})
				.finally(() => {
					isLoading = false;
				});
		} else {
			labDetail = null;
			isLoading = false;
		}
	});

	async function copyInviteCode() {
		if (!labDetail) return;
		try {
			await navigator.clipboard.writeText(labDetail.invite_code);
			copied = true;
			toast.success($_('lab_dashboard.copied'));
			setTimeout(() => (copied = false), 2000);
		} catch {
			// fallback
		}
	}

	async function handleRequestLeave() {
		if (!activeLab) return;
		requestingLeave = true;
		try {
			await labApi.requestLeaveLab(activeLab.id);
			toast.success($_('service.request_leave_lab.success'));
			leaveCodeSent = true;
		} catch (error: unknown) {
			showApiErrors(error, $_('service.request_leave_lab.failed'));
		} finally {
			requestingLeave = false;
		}
	}

	async function handleConfirmLeave() {
		if (!activeLab) return;
		confirmingLeave = true;
		try {
			await labApi.leaveLab(activeLab.id, { email_code: leaveEmailCode });
			toast.success($_('service.leave_lab.success'));
			setActiveLab(null);
			invalidateLabs();
		} catch (error: unknown) {
			showApiErrors(error, $_('service.leave_lab.failed'));
		} finally {
			confirmingLeave = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('lab_dashboard.title')} | Sci-Vault</title>
</svelte:head>

<div class="space-y-6">
	{#if !activeLab}
		<!-- No lab selected state -->
		<div class="flex h-[80vh] flex-col items-center justify-center space-y-6">
			<Card.Root class="w-full max-w-lg shadow-sm">
				<Card.Content
					class="flex flex-col items-center justify-center p-12 text-center text-muted-foreground"
				>
					<div
						class="mb-6 flex size-14 items-center justify-center rounded-full bg-primary/10 ring-1 ring-border/50"
					>
						<FlaskConical class="size-7 text-primary" />
					</div>
					<h3 class="mb-2 text-2xl font-bold tracking-tight text-foreground">
						{$_('lab_dashboard.no_lab_selected')}
					</h3>
					<p class="mb-8">
						{$_('lab_dashboard.no_lab_selected_desc')}
					</p>
					<div class="flex w-full flex-col gap-3 sm:flex-row sm:justify-center">
						<Button
							variant="outline"
							onclick={() => goto(resolve('/labs/join'))}
							class="w-full sm:w-auto"
						>
							{$_('sidebar.join_lab')}
						</Button>
						<Button onclick={() => goto(resolve('/labs/create'))} class="w-full sm:w-auto">
							{$_('sidebar.create_lab')}
						</Button>
					</div>
				</Card.Content>
			</Card.Root>
		</div>
	{:else if isLoading}
		<!-- Loading state -->
		<div class="space-y-6">
			<div class="flex items-center gap-4">
				<Skeleton class="h-12 w-12 rounded-xl" />
				<div class="space-y-2">
					<Skeleton class="h-6 w-48" />
					<Skeleton class="h-4 w-32" />
				</div>
			</div>
			<div class="grid gap-4 sm:grid-cols-3">
				{#each Array.from({ length: 3 }, (__, i) => i) as i (i)}
					<Card.Root>
						<Card.Content class="p-4">
							<Skeleton class="h-4 w-20" />
							<Skeleton class="mt-2 h-7 w-16" />
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
		</div>
	{:else if labDetail}
		<!-- Lab Overview -->
		<div class="space-y-6">
			<!-- Header -->
			<div class="flex flex-col justify-between space-y-4 sm:flex-row sm:items-center sm:space-y-0">
				<div class="flex items-center gap-4">
					<div
						class="flex size-12 shrink-0 items-center justify-center rounded-xl bg-primary/10 ring-1 ring-border/50"
					>
						<FlaskConical class="size-6 text-primary" />
					</div>
					<div class="space-y-1">
						<div class="flex items-center gap-2">
							<h2 class="text-3xl font-bold tracking-tight">{labDetail.name}</h2>
							{#if labDetail.my_role === 'owner'}
								<Badge
									variant="outline"
									class="gap-1 border-yellow-500/30 bg-yellow-500/10 text-yellow-700 dark:text-yellow-400"
								>
									<Crown class="size-3" />
									{$_('lab_dashboard.owner_badge')}
								</Badge>
							{/if}
						</div>
						<p class="text-sm text-muted-foreground capitalize">
							{$_('lab_dashboard.your_role')}: {$_(`profile.labs.role.${labDetail.my_role}`)}
						</p>
					</div>
				</div>
			</div>

			<!-- Stat Cards -->
			<div class="grid gap-4 sm:grid-cols-3">
				<!-- Members -->
				<Card.Root class="transition-shadow hover:shadow-md">
					<Card.Content class="p-4">
						<div class="flex items-center justify-between">
							<span class="text-sm font-medium text-muted-foreground">
								{$_('sidebar.lab_members')}
							</span>
							<div
								class="flex size-8 items-center justify-center rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400"
							>
								<Users class="size-4" />
							</div>
						</div>
						<div class="mt-2 text-2xl font-bold tracking-tight">{labDetail.member_count}</div>
					</Card.Content>
				</Card.Root>

				<!-- Invite Code -->
				<Card.Root class="transition-shadow hover:shadow-md sm:col-span-2">
					<Card.Content class="p-4">
						<div class="flex items-center justify-between">
							<span class="text-sm font-medium text-muted-foreground">
								{$_('lab_dashboard.invite_code')}
							</span>
							<Button
								variant="ghost"
								size="sm"
								class="h-8 gap-1.5 text-xs"
								onclick={copyInviteCode}
							>
								{#if copied}
									<Check class="size-3.5 text-green-600" />
								{:else}
									<Copy class="size-3.5" />
								{/if}
								{$_('lab_dashboard.copy_code')}
							</Button>
						</div>
						<div class="mt-2 font-mono text-2xl font-bold tracking-widest">
							{labDetail.invite_code}
						</div>
					</Card.Content>
				</Card.Root>
			</div>

			<!-- Description -->
			{#if labDetail.description}
				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Title class="text-base font-semibold">
							{$_('lab_dashboard.description')}
						</Card.Title>
					</Card.Header>
					<Card.Content>
						<p class="text-sm text-muted-foreground">{labDetail.description}</p>
					</Card.Content>
				</Card.Root>
			{/if}

			<Separator />

			<!-- Quick Actions -->
			<div>
				<h2 class="mb-3 text-base font-semibold">{$_('lab_dashboard.quick_actions')}</h2>
				<div class="grid gap-3 sm:grid-cols-2">
					<Button
						variant="outline"
						class="group h-auto justify-start gap-3 border-muted-foreground/20 p-4 transition-all hover:border-primary/50 hover:bg-primary/5"
						onclick={() => goto(resolve('/members'))}
					>
						<div
							class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-blue-500/10 text-blue-600 transition-colors group-hover:bg-blue-500/20 dark:text-blue-400"
						>
							<Users class="size-4" />
						</div>
						<div class="text-left">
							<div class="text-sm font-medium">{$_('lab_dashboard.view_members')}</div>
							<div class="text-xs text-muted-foreground">
								{$_('lab_dashboard.member_count', { values: { count: labDetail.member_count } })}
							</div>
						</div>
						<ArrowRight class="ml-auto size-4 text-muted-foreground" />
					</Button>

					{#if labDetail.my_role === 'owner'}
						<Button
							variant="outline"
							class="group h-auto justify-start gap-3 border-muted-foreground/20 p-4 transition-all hover:border-primary/50 hover:bg-primary/5"
							onclick={() => goto(resolve('/lab-settings'))}
						>
							<div
								class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-yellow-500/10 text-yellow-600 transition-colors group-hover:bg-yellow-500/20 dark:text-yellow-400"
							>
								<Settings class="size-4" />
							</div>
							<div class="text-left">
								<div class="text-sm font-medium">{$_('lab_dashboard.lab_settings')}</div>
								<div class="text-xs text-muted-foreground">
									{$_('lab_dashboard.lab_settings_hint')}
								</div>
							</div>
							<ArrowRight class="ml-auto size-4 text-muted-foreground" />
						</Button>
					{:else}
						<AlertDialog.Root
							bind:open={leaveDialogOpen}
							onOpenChange={(open) => {
								if (!open) resetLeaveState();
							}}
						>
							<AlertDialog.Trigger
								class="group flex h-auto w-full items-center justify-start gap-3 rounded-md border border-destructive/20 p-4 text-left transition-all hover:border-destructive/50 hover:bg-destructive/5"
							>
								<div
									class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-destructive/10 text-destructive transition-colors group-hover:bg-destructive/20"
								>
									<LogOut class="size-4" />
								</div>
								<div class="text-left">
									<div class="text-sm font-medium text-destructive">
										{$_('lab_dashboard.leave_lab')}
									</div>
									<div class="text-xs text-muted-foreground">
										{$_('lab_dashboard.leave_lab_hint')}
									</div>
								</div>
							</AlertDialog.Trigger>
							<AlertDialog.Content>
								<AlertDialog.Header>
									<AlertDialog.Title>
										{$_('lab_dashboard.leave_confirm', { values: { name: activeLab.name } })}
									</AlertDialog.Title>
									<AlertDialog.Description>
										{#if !leaveCodeSent}
											{$_('lab_dashboard.leave_step1_desc')}
										{:else}
											{$_('lab_dashboard.leave_step2')}
										{/if}
									</AlertDialog.Description>
								</AlertDialog.Header>
								{#if leaveCodeSent}
									<div class="space-y-1.5 px-6">
										<Label for="leave-code">{$_('lab_dashboard.leave_code_label')}</Label>
										<Input
											id="leave-code"
											bind:value={leaveEmailCode}
											placeholder={$_('lab.settings.delete_code_placeholder')}
											maxlength={6}
										/>
									</div>
								{/if}
								<AlertDialog.Footer>
									<AlertDialog.Cancel onclick={resetLeaveState}>
										{$_('profile.btn.cancel')}
									</AlertDialog.Cancel>
									{#if !leaveCodeSent}
										<AlertDialog.Action
											variant="destructive"
											disabled={requestingLeave}
											onclick={(e: MouseEvent) => {
												e.preventDefault();
												handleRequestLeave();
											}}
										>
											<Mail class="size-3.5" />
											{$_('lab_dashboard.leave_send_code')}
										</AlertDialog.Action>
									{:else}
										<AlertDialog.Action
											variant="destructive"
											disabled={leaveEmailCode.length !== 6 || confirmingLeave}
											onclick={handleConfirmLeave}
										>
											<LogOut class="size-3.5" />
											{$_('lab_dashboard.leave_lab')}
										</AlertDialog.Action>
									{/if}
								</AlertDialog.Footer>
							</AlertDialog.Content>
						</AlertDialog.Root>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>
