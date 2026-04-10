<script lang="ts">
	import { FlaskConical, Loader2 } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { _ } from 'svelte-i18n';

	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import labApi from '$lib/api/lab';
	import { invalidateLabs } from '$lib/stores/lab.svelte';

	let name = $state('');
	let description = $state('');
	let submitting = $state(false);

	async function handleCreate() {
		if (!name.trim()) return;

		submitting = true;
		try {
			const lab = await labApi.createLab({
				name: name.trim(),
				description: description.trim() || undefined
			});
			invalidateLabs();
			toast.success($_('lab.create.success', { values: { name: lab.name } }));
			goto(resolve('/labs/join'));
		} catch {
			toast.error($_('service.create_lab.failed'));
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
					<Card.Title>{$_('lab.create.title')}</Card.Title>
					<Card.Description>{$_('lab.create.description')}</Card.Description>
				</div>
			</div>
		</Card.Header>
		<Card.Content>
			<form
				onsubmit={(e) => {
					e.preventDefault();
					handleCreate();
				}}
				class="flex flex-col gap-4"
			>
				<div class="flex flex-col gap-2">
					<Label for="lab-name">{$_('lab.create.name_label')}</Label>
					<Input
						id="lab-name"
						bind:value={name}
						placeholder={$_('lab.create.name_placeholder')}
						disabled={submitting}
						maxlength={100}
					/>
				</div>
				<div class="flex flex-col gap-2">
					<Label for="lab-desc">
						{$_('lab.create.desc_label')}
						<span class="ml-1 text-xs text-muted-foreground">({$_('lab.create.optional')})</span>
					</Label>
					<textarea
						id="lab-desc"
						bind:value={description}
						placeholder={$_('lab.create.desc_placeholder')}
						disabled={submitting}
						maxlength={500}
						rows={3}
						class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
					></textarea>
				</div>
				<Button type="submit" disabled={submitting || !name.trim()} class="w-full">
					{#if submitting}
						<Loader2 class="mr-2 size-4 animate-spin" />
					{/if}
					{submitting ? $_('lab.create.creating') : $_('lab.create.submit')}
				</Button>
			</form>
		</Card.Content>
	</Card.Root>
</div>
