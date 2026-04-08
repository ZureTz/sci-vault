<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { _ } from 'svelte-i18n';
	import {
		Activity,
		Brain,
		Search,
		Shield,
		ArrowRight,
		Upload,
		Sparkles,
		ChevronRight
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import ThemeToggle from '$lib/components/layout/ThemeToggle.svelte';
	import LanguageToggle from '$lib/components/layout/LanguageToggle.svelte';
	import { intersect } from '$lib/actions/intersect';

	let isLoggedIn = $state(false);

	onMount(() => {
		isLoggedIn = !!localStorage.getItem('token');
	});

	const features = [
		{
			icon: Brain,
			titleKey: 'welcome.features.ai_enrichment.title',
			descKey: 'welcome.features.ai_enrichment.description',
			gradient: 'from-violet-500/20 to-purple-500/10',
			iconColor: 'text-violet-500'
		},
		{
			icon: Search,
			titleKey: 'welcome.features.vector_search.title',
			descKey: 'welcome.features.vector_search.description',
			gradient: 'from-blue-500/20 to-cyan-500/10',
			iconColor: 'text-blue-500'
		},
		{
			icon: Shield,
			titleKey: 'welcome.features.secure_storage.title',
			descKey: 'welcome.features.secure_storage.description',
			gradient: 'from-emerald-500/20 to-teal-500/10',
			iconColor: 'text-emerald-500'
		}
	];

	const steps = [
		{
			labelKey: 'welcome.workflow.step1.label',
			titleKey: 'welcome.workflow.step1.title',
			descKey: 'welcome.workflow.step1.description',
			icon: Upload
		},
		{
			labelKey: 'welcome.workflow.step2.label',
			titleKey: 'welcome.workflow.step2.title',
			descKey: 'welcome.workflow.step2.description',
			icon: Sparkles
		},
		{
			labelKey: 'welcome.workflow.step3.label',
			titleKey: 'welcome.workflow.step3.title',
			descKey: 'welcome.workflow.step3.description',
			icon: Search
		}
	];
</script>

<svelte:head>
	<title>{$_('app.title')} — {$_('login.subtitle.app')}</title>
</svelte:head>

<div class="flex min-h-screen flex-col bg-background text-foreground">
	<!-- Header -->
	<header class="sticky top-0 z-50 border-b bg-background/80 backdrop-blur-md">
		<div class="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
			<div class="flex items-center gap-2.5">
				<div
					class="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"
				>
					<Activity class="size-4" />
				</div>
				<span class="font-semibold">{$_('app.title')}</span>
				<span class="text-xs text-muted-foreground">{$_('app.version')}</span>
			</div>
			<div class="flex items-center gap-2">
				<ThemeToggle />
				<LanguageToggle />
				{#if isLoggedIn}
					<Button href={resolve('/')} size="sm" class="ml-2">
						{$_('welcome.footer.go_to_dashboard')}
						<ChevronRight class="ml-1 size-3.5" />
					</Button>
				{:else}
					<Button href={resolve('/login')} size="sm" class="ml-2">
						{$_('welcome.footer.sign_in')}
						<ChevronRight class="ml-1 size-3.5" />
					</Button>
				{/if}
			</div>
		</div>
	</header>

	<main class="flex-1">
		<!-- Hero -->
		<section class="relative overflow-hidden px-6 py-24 sm:py-32">
			<!-- Decorative background blobs -->
			<div class="pointer-events-none absolute inset-0 -z-10 overflow-hidden" aria-hidden="true">
				<div
					class="absolute -top-40 left-1/2 size-150 -translate-x-1/2 rounded-full bg-primary/10 blur-3xl"
				></div>
				<div
					class="absolute -bottom-20 left-1/4 size-100 rounded-full bg-violet-500/10 blur-3xl"
				></div>
				<div
					class="absolute right-1/4 -bottom-20 size-100 rounded-full bg-blue-500/10 blur-3xl"
				></div>
			</div>

			<div use:intersect class="animate-on-scroll mx-auto max-w-4xl text-center delay-100">
				<!-- Badge -->
				<div
					class="mb-6 inline-flex items-center gap-1.5 rounded-full border bg-muted/60 px-3 py-1 text-xs font-medium text-muted-foreground"
				>
					<Sparkles class="size-3 text-primary" />
					{$_('welcome.hero.badge')}
				</div>

				<!-- Heading -->
				<h1 class="text-5xl font-bold tracking-tight sm:text-6xl lg:text-7xl">
					{$_('welcome.hero.title')}
					<span
						class="bg-linear-to-r from-primary via-violet-500 to-blue-500 bg-clip-text text-transparent"
					>
						{$_('welcome.hero.title_highlight')}
					</span>
				</h1>

				<!-- Subtitle -->
				<p class="mx-auto mt-6 max-w-2xl text-lg leading-relaxed text-muted-foreground">
					{$_('welcome.hero.subtitle')}
				</p>

				<!-- CTAs -->
				<div class="mt-10 flex flex-wrap items-center justify-center gap-4">
					{#if isLoggedIn}
						<Button href={resolve('/')} size="lg" class="gap-2 px-8 text-base">
							{$_('welcome.footer.go_to_dashboard')}
							<ArrowRight class="size-4" />
						</Button>
					{:else}
						<Button href={resolve('/login')} size="lg" class="gap-2 px-8 text-base">
							{$_('welcome.hero.cta_primary')}
							<ArrowRight class="size-4" />
						</Button>
						<Button href={resolve('/login')} variant="outline" size="lg" class="px-8 text-base">
							{$_('welcome.hero.cta_secondary')}
						</Button>
					{/if}
				</div>
			</div>

			<!-- Floating mockup card -->
			<div use:intersect class="animate-on-scroll mx-auto mt-20 max-w-3xl delay-200">
				<div
					class="overflow-hidden rounded-2xl border bg-card shadow-2xl ring-1 shadow-primary/10 ring-border"
				>
					<!-- Fake browser bar -->
					<div class="flex items-center gap-2 border-b bg-muted/50 px-4 py-3">
						<div class="flex gap-1.5">
							<div class="size-3 rounded-full bg-red-400/70"></div>
							<div class="size-3 rounded-full bg-yellow-400/70"></div>
							<div class="size-3 rounded-full bg-green-400/70"></div>
						</div>
						<div
							class="mx-auto max-w-xs flex-1 rounded-md bg-background/60 px-3 py-1 text-center text-xs text-muted-foreground"
						>
							sci-vault.app / documents
						</div>
					</div>
					<!-- Content preview -->
					<div class="space-y-4 p-6">
						<div class="flex items-center justify-between">
							<div class="h-5 w-32 animate-pulse rounded bg-muted"></div>
							<div class="h-8 w-24 animate-pulse rounded-md bg-primary/20"></div>
						</div>
						<div class="space-y-2">
							{#each [90, 75, 60, 85] as w (w)}
								<div class="flex items-center gap-4">
									<div class="h-4 animate-pulse rounded bg-muted" style="width: {w}%"></div>
									<div class="ml-auto h-5 w-14 animate-pulse rounded-full bg-emerald-500/20"></div>
								</div>
							{/each}
						</div>
						<div class="mt-4 space-y-2 rounded-xl border bg-muted/30 p-4">
							<div class="flex items-center gap-2">
								<Brain class="size-4 text-violet-500" />
								<div class="h-4 w-20 animate-pulse rounded bg-muted"></div>
							</div>
							<div class="space-y-1.5">
								<div class="h-3 w-full animate-pulse rounded bg-muted/60"></div>
								<div class="h-3 w-5/6 animate-pulse rounded bg-muted/60"></div>
								<div class="h-3 w-4/6 animate-pulse rounded bg-muted/60"></div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</section>

		<!-- Features -->
		<section class="bg-muted/30 px-6 py-24">
			<div use:intersect class="animate-on-scroll mx-auto max-w-6xl delay-300">
				<div class="mb-16 text-center">
					<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">
						{$_('welcome.features.title')}
					</h2>
					<p class="mx-auto mt-4 max-w-xl text-muted-foreground">
						{$_('welcome.features.subtitle')}
					</p>
				</div>

				<div class="grid gap-6 sm:grid-cols-3">
					{#each features as feat (feat.titleKey)}
						{@const Icon = feat.icon}
						<div
							class="group relative overflow-hidden rounded-2xl border bg-card p-6 transition-all duration-300 hover:-translate-y-1 hover:shadow-lg hover:shadow-primary/5"
						>
							<div
								class="absolute inset-0 -z-10 bg-linear-to-br {feat.gradient} opacity-0 transition-opacity duration-300 group-hover:opacity-100"
							></div>
							<div class="mb-4 inline-flex rounded-xl border bg-background p-3 shadow-sm">
								<Icon class="size-5 {feat.iconColor}" />
							</div>
							<h3 class="mb-2 font-semibold">{$_(feat.titleKey)}</h3>
							<p class="text-sm leading-relaxed text-muted-foreground">{$_(feat.descKey)}</p>
						</div>
					{/each}
				</div>
			</div>
		</section>

		<!-- How it works -->
		<section class="px-6 py-24">
			<div use:intersect class="animate-on-scroll mx-auto max-w-4xl delay-400">
				<div class="mb-16 text-center">
					<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">
						{$_('welcome.workflow.title')}
					</h2>
				</div>

				<div class="relative">
					<!-- Connecting line (desktop) -->
					<div
						class="absolute top-10 right-0 left-0 hidden h-px bg-linear-to-r from-transparent via-border to-transparent sm:block"
						style="top: 2.5rem;"
					></div>

					<div class="grid gap-10 sm:grid-cols-3">
						{#each steps as step, i (step.titleKey)}
							{@const Icon = step.icon}
							<div class="relative flex flex-col items-center text-center">
								<div
									class="relative mb-6 flex size-20 items-center justify-center rounded-2xl border bg-card shadow-md ring-4 ring-background"
								>
									<Icon class="size-8 text-primary" />
									<div
										class="absolute -top-3 -right-3 flex size-6 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground"
									>
										{i + 1}
									</div>
								</div>
								<h3 class="mb-2 text-lg font-semibold">{$_(step.titleKey)}</h3>
								<p class="max-w-45 text-sm leading-relaxed text-muted-foreground">
									{$_(step.descKey)}
								</p>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</section>

		<!-- CTA Banner -->
		<section class="px-6 py-20">
			<div use:intersect class="animate-on-scroll mx-auto max-w-3xl delay-500">
				<div
					class="relative overflow-hidden rounded-3xl bg-primary px-8 py-14 text-center text-primary-foreground shadow-2xl shadow-primary/30"
				>
					<!-- Decorative blobs inside banner -->
					<div
						class="pointer-events-none absolute -top-10 -left-10 size-48 rounded-full bg-white/10 blur-2xl"
						aria-hidden="true"
					></div>
					<div
						class="pointer-events-none absolute -right-10 -bottom-10 size-48 rounded-full bg-white/10 blur-2xl"
						aria-hidden="true"
					></div>

					<Activity class="mx-auto mb-4 size-10 opacity-80" />
					<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">
						{$_('app.title')}
					</h2>
					<p class="mx-auto mt-3 max-w-md text-primary-foreground/80">
						{$_('login.subtitle.app')}
					</p>
					<div class="mt-8 flex flex-wrap justify-center gap-4">
						{#if isLoggedIn}
							<Button
								href={resolve('/')}
								variant="secondary"
								size="lg"
								class="gap-2 px-8 text-base font-semibold"
							>
								{$_('welcome.footer.go_to_dashboard')}
								<ArrowRight class="size-4" />
							</Button>
						{:else}
							<Button
								href={resolve('/login')}
								variant="secondary"
								size="lg"
								class="gap-2 px-8 text-base font-semibold"
							>
								{$_('welcome.hero.cta_primary')}
								<ArrowRight class="size-4" />
							</Button>
						{/if}
					</div>
				</div>
			</div>
		</section>
	</main>

	<!-- Footer -->
	<footer class="border-t px-6 py-8">
		<div class="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 sm:flex-row">
			<div class="flex items-center gap-2 text-sm text-muted-foreground">
				<div
					class="flex aspect-square size-5 items-center justify-center rounded bg-primary text-primary-foreground"
				>
					<Activity class="size-3" />
				</div>
				{$_('welcome.footer.copyright')}
			</div>
			{#if isLoggedIn}
				<Button href={resolve('/')} variant="ghost" size="sm">
					{$_('welcome.footer.go_to_dashboard')}
					<ChevronRight class="ml-1 size-3.5" />
				</Button>
			{:else}
				<Button href={resolve('/login')} variant="ghost" size="sm">
					{$_('welcome.footer.sign_in')}
					<ChevronRight class="ml-1 size-3.5" />
				</Button>
			{/if}
		</div>
	</footer>
</div>

<style>
	/* 页面滚动/加载入场动画 */
	.animate-on-scroll {
		opacity: 0;
		transform: translateY(24px);
		transition:
			opacity 0.8s cubic-bezier(0.16, 1, 0.3, 1),
			transform 0.8s cubic-bezier(0.16, 1, 0.3, 1);
	}

	:global(.animate-on-scroll.intersected) {
		opacity: 1;
		transform: translateY(0);
	}

	.delay-100 {
		transition-delay: 100ms;
	}
	.delay-200 {
		transition-delay: 200ms;
	}
	.delay-300 {
		transition-delay: 300ms;
	}
	.delay-400 {
		transition-delay: 400ms;
	}
	.delay-500 {
		transition-delay: 500ms;
	}
</style>
