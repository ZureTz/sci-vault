<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';
	import { FileUp } from 'lucide-svelte';

	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import documentApi from '$lib/api/document';
	import { showApiErrors } from '$lib/utils/api-error';

	let fileInput = $state<HTMLInputElement | undefined>(undefined);
	let selectedFile = $state<File | null>(null);
	let title = $state('');
	let year = $state('');
	let doi = $state('');
	let isSubmitting = $state(false);

	function handleFileChange(event: Event) {
		const file = (event.target as HTMLInputElement).files?.[0] ?? null;
		selectedFile = file;
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!selectedFile) {
			toast.error($_('document.upload.error.file_required'));
			return;
		}

		isSubmitting = true;
		try {
			const yearNum = year ? parseInt(year, 10) : null;
			await documentApi.uploadDocument({
				file: selectedFile,
				title: title || null,
				year: yearNum,
				doi: doi || null
			});
			toast.success($_('document.upload.success'));
			// Reset form
			selectedFile = null;
			title = '';
			year = '';
			doi = '';
			if (fileInput) fileInput.value = '';
		} catch (error: unknown) {
			showApiErrors(error, $_('document.upload.error.failed'));
		} finally {
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('document.upload.title')} | Sci-Vault</title>
</svelte:head>

<div class="container mx-auto max-w-2xl px-4 py-8">
	<Card.Root class="shadow-sm">
		<Card.Header>
			<div class="flex items-center gap-3">
				<div
					class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary"
				>
					<FileUp class="h-5 w-5" />
				</div>
				<div>
					<Card.Title class="text-xl">{$_('document.upload.title')}</Card.Title>
					<Card.Description>{$_('document.upload.description')}</Card.Description>
				</div>
			</div>
		</Card.Header>

		<Card.Content>
			<form onsubmit={handleSubmit} class="space-y-5">
				<!-- File picker -->
				<div class="space-y-1.5">
					<Label for="file">{$_('document.upload.file_label')}</Label>
					<div
						class="flex cursor-pointer items-center gap-3 rounded-md border border-dashed border-input bg-muted/30 px-4 py-5 transition-colors hover:bg-muted/50"
						role="button"
						tabindex="0"
						onclick={() => fileInput?.click()}
						onkeydown={(e) => e.key === 'Enter' && fileInput?.click()}
					>
						<FileUp class="h-5 w-5 shrink-0 text-muted-foreground" />
						<span class="text-sm text-muted-foreground">
							{#if selectedFile}
								<span class="font-medium text-foreground">{selectedFile.name}</span>
								<span class="ml-2 text-xs">({(selectedFile.size / 1024 / 1024).toFixed(2)} MB)</span
								>
							{:else}
								{$_('document.upload.file_placeholder')}
							{/if}
						</span>
					</div>
					<input
						id="file"
						type="file"
						accept=".pdf,.doc,.docx,.txt"
						class="hidden"
						bind:this={fileInput}
						onchange={handleFileChange}
					/>
				</div>

				<!-- Title -->
				<div class="space-y-1.5">
					<Label for="title">{$_('document.upload.title_label')}</Label>
					<Input
						id="title"
						bind:value={title}
						placeholder={$_('document.upload.title_placeholder')}
						maxlength={255}
					/>
				</div>

				<!-- Year & DOI in a row -->
				<div class="grid grid-cols-2 gap-4">
					<div class="space-y-1.5">
						<Label for="year">{$_('document.upload.year_label')}</Label>
						<Input
							id="year"
							type="number"
							bind:value={year}
							placeholder={$_('document.upload.year_placeholder')}
							min={1000}
							max={9999}
						/>
					</div>
					<div class="space-y-1.5">
						<Label for="doi">{$_('document.upload.doi_label')}</Label>
						<Input
							id="doi"
							bind:value={doi}
							placeholder={$_('document.upload.doi_placeholder')}
							maxlength={255}
						/>
					</div>
				</div>

				<!-- AI metadata enrichment hint -->
				<p class="text-sm text-muted-foreground">{$_('document.upload.metadata_hint')}</p>

				<Card.Footer class="px-0 pt-2 pb-0">
					<Button type="submit" class="w-full" disabled={isSubmitting}>
						{#if isSubmitting}
							{$_('document.upload.submitting')}
						{:else}
							{$_('document.upload.submit')}
						{/if}
					</Button>
				</Card.Footer>
			</form>
		</Card.Content>
	</Card.Root>
</div>
