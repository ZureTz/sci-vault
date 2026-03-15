<script lang="ts">
	import { _ } from 'svelte-i18n';
</script>

<div class="flex flex-col gap-4">
	<h1 class="text-3xl font-bold tracking-tight">{$_('dashboard.title')}</h1>
	<p class="text-muted-foreground">{$_('dashboard.description')}</p>
</div>
