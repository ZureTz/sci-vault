<script lang="ts">
	import { FlaskConical, Loader2 } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { _ } from 'svelte-i18n';

	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import labApi from '$lib/api/lab';
	import { invalidateLabs } from '$lib/stores/lab.svelte';

	let inviteCode = $state('');
	let submitting = $state(false);
	let joinedLab = $state<{ name: string; member_count: number } | null>(null);

	async function handleJoin() {
		const code = inviteCode.trim();
		if (!code) return;

		submitting = true;
		try {
			const lab = await labApi.joinLabByCode({ invite_code: code });
			joinedLab = { name: lab.name, member_count: lab.member_count };
			inviteCode = '';
			invalidateLabs();
			toast.success($_('lab.join.success', { values: { name: lab.name } }));
		} catch (err: unknown) {
			const status = (err as { response?: { status?: number } })?.response?.status;
			if (status === 404) {
				toast.error($_('service.join_lab.invalid_code'));
			} else if (status === 409) {
				toast.error($_('service.join_lab.already_member'));
			} else {
				toast.error($_('service.join_lab.failed'));
			}
		} finally {
			submitting = false;
		}
	}
</script>

<div class="mx-auto max-w-md py-8">
	<Card.Root>
		<Card.Header>
			<div class="flex items-center gap-3">
				<div
					class="flex size-10 items-center justify-center rounded-lg bg-primary text-primary-foreground"
				>
					<FlaskConical class="size-5" />
				</div>
				<div>
					<Card.Title>{$_('lab.join.title')}</Card.Title>
					<Card.Description>{$_('lab.join.description')}</Card.Description>
				</div>
			</div>
		</Card.Header>
		<Card.Content>
			<form
				onsubmit={(e) => {
					e.preventDefault();
					handleJoin();
				}}
				class="flex flex-col gap-4"
			>
				<div class="flex flex-col gap-2">
					<Label for="invite-code">{$_('lab.join.code_label')}</Label>
					<Input
						id="invite-code"
						bind:value={inviteCode}
						placeholder={$_('lab.join.code_placeholder')}
						disabled={submitting}
						class="font-mono tracking-widest uppercase"
					/>
				</div>
				<Button type="submit" disabled={submitting || !inviteCode.trim()} class="w-full">
					{#if submitting}
						<Loader2 class="mr-2 size-4 animate-spin" />
					{/if}
					{submitting ? $_('lab.join.joining') : $_('lab.join.submit')}
				</Button>
			</form>

			{#if joinedLab}
				<div
					class="mt-4 rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-950"
				>
					<p class="text-sm font-medium text-green-800 dark:text-green-200">
						{$_('lab.join.welcome', { values: { name: joinedLab.name } })}
					</p>
					<p class="mt-1 text-xs text-green-600 dark:text-green-400">
						{$_('lab.join.member_count', { values: { count: joinedLab.member_count } })}
					</p>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
