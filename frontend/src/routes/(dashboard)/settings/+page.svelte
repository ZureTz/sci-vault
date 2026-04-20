<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';
	import { Lock, KeyRound, ShieldCheck, Eye, EyeOff } from 'lucide-svelte';

	import * as Card from '$lib/components/ui/card';
	import * as InputGroup from '$lib/components/ui/input-group';
	import { Label } from '$lib/components/ui/label';
	import { Button } from '$lib/components/ui/button';

	import userApi from '$lib/api/user';
	import { showApiErrors } from '$lib/utils/api-error';
	import { validateChangePasswordForm, type ChangePasswordFormErrors } from '$lib/utils/validation';

	let form = $state({
		current_password: '',
		new_password: '',
		confirmed_password: ''
	});
	let errors = $state<ChangePasswordFormErrors>({});
	let isSubmitting = $state(false);
	let showCurrent = $state(false);
	let showNew = $state(false);
	let showConfirm = $state(false);

	async function handleSubmit() {
		const v = validateChangePasswordForm(form);
		if (v) {
			errors = v;
			return;
		}
		errors = {};

		isSubmitting = true;
		try {
			await userApi.changePassword(form);
			toast.success($_('settings.change_password.success'));
			form = { current_password: '', new_password: '', confirmed_password: '' };
		} catch (error: unknown) {
			showApiErrors(error, $_('settings.change_password.failed'));
		} finally {
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('settings.title')} | Sci-Vault</title>
</svelte:head>

<div class="mx-auto w-full max-w-2xl space-y-6">
	<div class="space-y-1">
		<h1 class="text-2xl font-bold tracking-tight">{$_('settings.title')}</h1>
		<p class="text-sm text-muted-foreground">{$_('settings.description')}</p>
	</div>

	<Card.Root>
		<Card.Header>
			<div class="flex items-center gap-2">
				<ShieldCheck class="h-4 w-4 text-muted-foreground" />
				<Card.Title class="text-base">{$_('settings.change_password.title')}</Card.Title>
			</div>
			<Card.Description>{$_('settings.change_password.description')}</Card.Description>
		</Card.Header>

		<form
			onsubmit={(e) => {
				e.preventDefault();
				handleSubmit();
			}}
		>
			<Card.Content class="space-y-4">
				<div class="space-y-2">
					<Label for="current-password">{$_('settings.change_password.current')}</Label>
					<InputGroup.Root>
						<InputGroup.Addon>
							<Lock />
						</InputGroup.Addon>
						<InputGroup.Input
							id="current-password"
							type={showCurrent ? 'text' : 'password'}
							autocomplete="current-password"
							bind:value={form.current_password}
						/>
						<InputGroup.Addon align="inline-end">
							<InputGroup.Button
								size="icon-xs"
								onclick={() => (showCurrent = !showCurrent)}
								aria-label={$_(
									showCurrent
										? 'settings.change_password.hide_password'
										: 'settings.change_password.show_password'
								)}
							>
								{#if showCurrent}
									<EyeOff />
								{:else}
									<Eye />
								{/if}
							</InputGroup.Button>
						</InputGroup.Addon>
					</InputGroup.Root>
					{#if errors.current_password}
						<p class="text-sm text-destructive">{$_(errors.current_password)}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<Label for="new-password">{$_('settings.change_password.new')}</Label>
					<InputGroup.Root>
						<InputGroup.Addon>
							<KeyRound />
						</InputGroup.Addon>
						<InputGroup.Input
							id="new-password"
							type={showNew ? 'text' : 'password'}
							autocomplete="new-password"
							bind:value={form.new_password}
						/>
						<InputGroup.Addon align="inline-end">
							<InputGroup.Button
								size="icon-xs"
								onclick={() => (showNew = !showNew)}
								aria-label={$_(
									showNew
										? 'settings.change_password.hide_password'
										: 'settings.change_password.show_password'
								)}
							>
								{#if showNew}
									<EyeOff />
								{:else}
									<Eye />
								{/if}
							</InputGroup.Button>
						</InputGroup.Addon>
					</InputGroup.Root>
					{#if errors.new_password}
						<p class="text-sm text-destructive">{$_(errors.new_password)}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<Label for="confirm-password">{$_('settings.change_password.confirm')}</Label>
					<InputGroup.Root>
						<InputGroup.Addon>
							<KeyRound />
						</InputGroup.Addon>
						<InputGroup.Input
							id="confirm-password"
							type={showConfirm ? 'text' : 'password'}
							autocomplete="new-password"
							bind:value={form.confirmed_password}
						/>
						<InputGroup.Addon align="inline-end">
							<InputGroup.Button
								size="icon-xs"
								onclick={() => (showConfirm = !showConfirm)}
								aria-label={$_(
									showConfirm
										? 'settings.change_password.hide_password'
										: 'settings.change_password.show_password'
								)}
							>
								{#if showConfirm}
									<EyeOff />
								{:else}
									<Eye />
								{/if}
							</InputGroup.Button>
						</InputGroup.Addon>
					</InputGroup.Root>
					{#if errors.confirmed_password}
						<p class="text-sm text-destructive">{$_(errors.confirmed_password)}</p>
					{/if}
				</div>
			</Card.Content>

			<Card.Footer class="mt-6 justify-end border-t pt-6">
				<Button type="submit" disabled={isSubmitting}>
					{isSubmitting
						? $_('settings.change_password.submitting')
						: $_('settings.change_password.submit')}
				</Button>
			</Card.Footer>
		</form>
	</Card.Root>
</div>
