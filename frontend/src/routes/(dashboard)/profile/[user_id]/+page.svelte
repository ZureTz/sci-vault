<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { onMount, untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { MapPin, Globe, ExternalLink, UserRound, Upload, SquarePen } from 'lucide-svelte';
	import { jwtDecode } from 'jwt-decode';

	import { page } from '$app/state';
	import type { PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Avatar from '$lib/components/ui/avatar';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { Separator } from '$lib/components/ui/separator';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import userApi from '$lib/api/user';

	let { data }: { data: PageData } = $props();

	let profilePromise = $state(untrack(() => data.profile));
	let currentUserId = $state<number | null>(null);
	let isEditing = $state(false);
	let editForm = $state({ nickname: '', bio: '', website: '', location: '' });
	let fileInput = $state<HTMLInputElement | undefined>(undefined);

	// Re-sync when SvelteKit updates data.profile on param change (e.g. /profile/9 → /profile/11)
	$effect(() => {
		profilePromise = data.profile;
		untrack(() => {
			isEditing = false;
		});
	});

	onMount(() => {
		const token = localStorage.getItem('token');
		if (token) {
			try {
				const decoded = jwtDecode<{ user_id?: number }>(token);
				currentUserId = Number(decoded.user_id);
			} catch {
				// ignore invalid token
			}
		}
	});

	function startEdit(profile: {
		nickname?: string;
		bio?: string;
		website?: string;
		location?: string;
	}) {
		editForm = {
			nickname: profile.nickname ?? '',
			bio: profile.bio ?? '',
			website: profile.website ?? '',
			location: profile.location ?? ''
		};
		isEditing = true;
	}

	async function saveEdit(userId: number) {
		try {
			if (editForm.website && !/^https?:\/\//.test(editForm.website)) {
				editForm.website = 'https://' + editForm.website;
			}
			await userApi.updateProfile(editForm);
			isEditing = false;
			profilePromise = userApi.getProfile(userId);
			toast.success($_('profile.success.profile_updated'));
		} catch {
			toast.error($_('profile.error.update_failed'));
		}
	}

	async function handleAvatarUpload(event: Event, userId: number) {
		const file = (event.target as HTMLInputElement).files?.[0];
		if (!file) return;
		if (file.size > 5 * 1024 * 1024) {
			toast.error($_('profile.error.avatar_too_large'));
			return;
		}
		try {
			await userApi.uploadAvatar(file);
			profilePromise = userApi.getProfile(userId);
			toast.success($_('profile.success.avatar_updated'));
		} catch {
			toast.error($_('profile.error.upload_failed'));
		}
	}
</script>

<svelte:head>
	<title>{$_('profile.title')} | Sci-Vault</title>
</svelte:head>

<div class="container mx-auto max-w-3xl px-4 py-8">
	<Card.Root class="overflow-hidden shadow-sm">
		<!-- Cover banner -->
		<div class="h-36 bg-linear-to-br from-primary/30 via-primary/10 to-muted"></div>

		{#await profilePromise}
			<!-- Skeleton state -->
			<Card.Header class="pt-0 pb-4">
				<div class="-mt-14 flex items-end gap-4">
					<Skeleton class="h-24 w-24 shrink-0 rounded-full border-4 border-background" />
				</div>
				<div class="mt-3 space-y-2">
					<Skeleton class="h-7 w-48" />
					<Skeleton class="h-4 w-28" />
				</div>
			</Card.Header>

			<Separator />

			<Card.Content class="space-y-3 pt-5">
				<Skeleton class="h-4 w-full" />
				<Skeleton class="h-4 w-4/5" />
				<Skeleton class="h-4 w-36" />
			</Card.Content>
		{:then profile}
			{@const urlUserId = Number(page.params.user_id)}
			{@const isOwner = currentUserId === urlUserId}

			{#if !profile}
				<!-- Profile doesn't exist -->
				{#if isOwner}
					<!-- Own profile, not yet created — show empty editable form -->
					<Card.Header class="pt-0 pb-4">
						<div class="-mt-14 flex items-end justify-between gap-4">
							<Avatar.Root class="h-24 w-24 shrink-0 border-4 border-background shadow-md">
								<Avatar.Fallback class="bg-primary/10 text-3xl font-semibold text-primary"
									>?</Avatar.Fallback
								>
							</Avatar.Root>
							<div class="mt-4 flex gap-2">
								<Button size="sm" onclick={() => saveEdit(urlUserId)}
									>{$_('profile.btn.save')}</Button
								>
							</div>
						</div>
						<div class="mt-3 space-y-4 pt-2">
							<p class="text-sm text-muted-foreground">{$_('profile.empty_own_profile')}</p>
							<div class="space-y-1.5">
								<Label for="nickname">{$_('profile.form.nickname')}</Label>
								<Input id="nickname" bind:value={editForm.nickname} />
							</div>
							<div class="space-y-1.5">
								<Label for="location">{$_('profile.form.location')}</Label>
								<Input id="location" bind:value={editForm.location} />
							</div>
							<div class="space-y-1.5">
								<Label for="website">{$_('profile.form.website')}</Label>
								<Input id="website" bind:value={editForm.website} />
							</div>
							<div class="space-y-1.5">
								<Label for="bio">{$_('profile.form.bio')}</Label>
								<textarea
									id="bio"
									bind:value={editForm.bio}
									rows={3}
									class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
								></textarea>
							</div>
						</div>
					</Card.Header>
				{:else}
					<!-- Someone else's profile doesn't exist -->
					<Card.Content class="flex flex-col items-center gap-3 py-16 text-center">
						<p class="text-lg font-medium">{$_('profile.not_found')}</p>
						<p class="text-sm text-muted-foreground">{$_('profile.not_found_desc')}</p>
					</Card.Content>
				{/if}
			{:else}
				{@const initials = profile.nickname ? profile.nickname.charAt(0).toUpperCase() : '?'}
				{@const displayName = profile.nickname || `User #${profile.user_id}`}

				<Card.Header class="pt-0 pb-4">
					<div class="flex items-center justify-between">
						<!-- Avatar overlapping cover -->
						<div class="group relative -mt-14 flex items-end">
							<Avatar.Root class="h-24 w-24 shrink-0 border-4 border-background shadow-md">
								{#if profile.avatar_url}
									<Avatar.Image src={profile.avatar_url} alt={displayName} class="object-cover" />
								{/if}
								<Avatar.Fallback class="bg-primary/10 text-3xl font-semibold text-primary">
									{initials}
								</Avatar.Fallback>
							</Avatar.Root>

							{#if isOwner && !isEditing}
								<button
									onclick={() => fileInput?.click()}
									class="absolute -right-2 -bottom-2 z-10 rounded-full border bg-background p-1.5 shadow-sm hover:bg-muted"
									title={$_('profile.tooltip.upload_avatar')}
								>
									<Upload class="h-4 w-4 text-muted-foreground" />
								</button>
								<input
									type="file"
									accept="image/*"
									class="hidden"
									bind:this={fileInput}
									onchange={(e) => handleAvatarUpload(e, profile.user_id)}
								/>
							{/if}
						</div>

						<!-- Action buttons -->
						<div class="mt-4 flex gap-2">
							{#if isOwner}
								{#if isEditing}
									<Button variant="outline" size="sm" onclick={() => (isEditing = false)}>
										{$_('profile.btn.cancel')}
									</Button>
									<Button size="sm" onclick={() => saveEdit(profile.user_id)}
										>{$_('profile.btn.save')}</Button
									>
								{:else}
									<Button
										variant="outline"
										size="sm"
										class="gap-1.5"
										onclick={() => startEdit(profile)}
									>
										<SquarePen class="h-4 w-4" />
										{$_('profile.btn.edit')}
									</Button>
								{/if}
							{/if}
						</div>
					</div>

					<!-- Name & location / edit form -->
					<div class="mt-3">
						{#if isEditing}
							<div class="space-y-4 pt-2">
								<div class="space-y-1.5">
									<Label for="nickname">{$_('profile.form.nickname')}</Label>
									<Input id="nickname" bind:value={editForm.nickname} />
								</div>
								<div class="space-y-1.5">
									<Label for="location">{$_('profile.form.location')}</Label>
									<Input id="location" bind:value={editForm.location} />
								</div>
								<div class="space-y-1.5">
									<Label for="website">{$_('profile.form.website')}</Label>
									<Input id="website" bind:value={editForm.website} />
								</div>
								<div class="space-y-1.5">
									<Label for="bio">{$_('profile.form.bio')}</Label>
									<textarea
										id="bio"
										bind:value={editForm.bio}
										rows={3}
										class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
									></textarea>
								</div>
							</div>
						{:else}
							<div class="space-y-1.5">
								<Card.Title class="text-2xl font-bold">{displayName}</Card.Title>
								{#if profile.location}
									<div class="flex items-center gap-1.5 text-sm text-muted-foreground">
										<MapPin class="h-3.5 w-3.5 shrink-0" />
										<span>{profile.location}</span>
									</div>
								{/if}
							</div>
						{/if}
					</div>
				</Card.Header>

				{#if !isEditing && (profile.bio || profile.website)}
					<Separator />

					<Card.Content class="space-y-4 pt-5">
						{#if profile.bio}
							<div class="flex gap-3">
								<UserRound class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
								<p class="text-sm leading-relaxed text-muted-foreground">{profile.bio}</p>
							</div>
						{/if}

						{#if profile.website}
							<Tooltip.Provider>
								<Tooltip.Root>
									<Tooltip.Trigger>
										<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
										<a
											href={profile.website}
											target="_blank"
											rel="external noopener noreferrer"
											class="inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:underline"
										>
											<Globe class="h-4 w-4 shrink-0" />
											<span class="max-w-xs truncate"
												>{profile.website.replace(/^https?:\/\//, '')}</span
											>
											<ExternalLink class="h-3 w-3 shrink-0 opacity-60" />
										</a>
									</Tooltip.Trigger>
									<Tooltip.Content>
										<p>{$_('profile.tooltip.open_website')}</p>
									</Tooltip.Content>
								</Tooltip.Root>
							</Tooltip.Provider>
						{/if}
					</Card.Content>
				{/if}
			{/if}
		{/await}
	</Card.Root>
</div>
