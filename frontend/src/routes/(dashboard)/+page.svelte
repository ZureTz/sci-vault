<script lang="ts">
	import { FlaskConical } from 'lucide-svelte';
	import { _ } from 'svelte-i18n';

	import * as Card from '$lib/components/ui/card';
</script>

<svelte:head>
	<title>{$_('lab_dashboard.title')} | Sci-Vault</title>
</svelte:head>

<div class="container mx-auto max-w-3xl px-4 py-8">
	<Card.Root class="shadow-sm">
		<Card.Content class="flex flex-col items-center gap-4 py-16 text-center">
			<div
				class="flex size-14 items-center justify-center rounded-2xl bg-primary/10 ring-1 ring-border/50"
			>
				<FlaskConical class="size-7 text-primary" />
			</div>
			<div>
				<p class="text-lg font-semibold">{$_('lab_dashboard.title')}</p>
				<p class="mt-1 text-sm text-muted-foreground">{$_('lab_dashboard.coming_soon')}</p>
			</div>
		</Card.Content>
	</Card.Root>
</div>
