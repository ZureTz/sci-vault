<script lang="ts">
	import SunIcon from 'lucide-svelte/icons/sun';
	import MoonIcon from 'lucide-svelte/icons/moon';

	import { resetMode, setMode } from 'mode-watcher';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import { _ } from 'svelte-i18n';
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger
		class={buttonVariants({ variant: 'outline', size: 'icon' }) + ' bg-background/80 backdrop-blur'}
	>
		<SunIcon
			class="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 transition-all! dark:scale-0 dark:-rotate-90"
		/>
		<MoonIcon
			class="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 transition-all! dark:scale-100 dark:rotate-0"
		/>
		<span class="sr-only">{$_('app.toggle_theme')}</span>
	</DropdownMenu.Trigger>
	<DropdownMenu.Content align="end">
		<DropdownMenu.Item onclick={() => setMode('light')}>{$_('theme.light')}</DropdownMenu.Item>
		<DropdownMenu.Item onclick={() => setMode('dark')}>{$_('theme.dark')}</DropdownMenu.Item>
		<DropdownMenu.Item onclick={() => resetMode()}>{$_('theme.system')}</DropdownMenu.Item>
	</DropdownMenu.Content>
</DropdownMenu.Root>
