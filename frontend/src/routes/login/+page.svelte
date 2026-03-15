<script lang="ts">
	import { Activity, Mail, Lock, User, KeyRound } from 'lucide-svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { _ } from 'svelte-i18n';

	let activeView = $state<'auth' | 'reset-password'>('auth');
	let activeTab = $state<'login' | 'register'>('login');

	// Form states
	let loginForm = $state({ identifier: '', password: '' });
	let registerForm = $state({
		username: '',
		email: '',
		password: '',
		confirmed_password: '',
		email_code: ''
	});
	let resetForm = $state({
		email: '',
		email_code: '',
		password: '',
		confirmed_password: ''
	});

	// Mocking request states
	let isSubmitting = $state(false);

	async function handleLogin() {
		isSubmitting = true;
		// TODO: Call login API mapping `identifier` to either email or username
		console.log('Login request:', loginForm);
		setTimeout(() => (isSubmitting = false), 1000);
	}

	async function handleRegister() {
		isSubmitting = true;
		console.log('Register request:', registerForm);
		setTimeout(() => (isSubmitting = false), 1000);
	}

	async function handleSendCode(email: string) {
		console.log('Send code to:', email);
	}

	async function handleResetPassword() {
		isSubmitting = true;
		console.log('Reset password request:', resetForm);
		setTimeout(() => {
			isSubmitting = false;
			activeView = 'auth'; // Go back
		}, 1000);
	}
</script>

<div class="flex min-h-screen items-center justify-center p-4">
	<!-- Background decoration -->
	<div class="absolute inset-0 z-[-1] overflow-hidden bg-background">
		<div
			class="absolute -top-[20%] -left-[10%] h-[50vh] w-[40vw] rounded-full bg-primary/5 blur-[100px]"
		></div>
		<div
			class="absolute top-[60%] left-[70%] h-[40vh] w-[30vw] rounded-full bg-primary/10 blur-[120px]"
		></div>
	</div>

	<div class="w-full max-w-md space-y-6">
		<div class="flex flex-col items-center justify-center space-y-2 text-center">
			<div
				class="flex aspect-square size-12 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow"
			>
				<Activity class="size-6" />
			</div>
			<div class="space-y-1">
				<h1 class="text-2xl font-bold tracking-tight">{$_('login.title.app')}</h1>
				<p class="text-sm text-muted-foreground">{$_('login.subtitle.app')}</p>
			</div>
		</div>

		{#if activeView === 'auth'}
			<Tabs.Root bind:value={activeTab} class="w-full">
				<Tabs.List class="grid w-full grid-cols-2">
					<Tabs.Trigger value="login">{$_('login.tab.login')}</Tabs.Trigger>
					<Tabs.Trigger value="register">{$_('login.tab.register')}</Tabs.Trigger>
				</Tabs.List>

				<!-- LOGIN TAB -->
				<Tabs.Content value="login">
					<Card.Root>
						<Card.Header>
							<Card.Title>{$_('login.welcome_back')}</Card.Title>
							<Card.Description>{$_('login.signin_desc')}</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-4">
							<div class="space-y-2">
								<Label for="identifier">{$_('login.identifier')}</Label>
								<div class="relative">
									<User class="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
									<Input
										id="identifier"
										placeholder={$_('login.identifier_placeholder')}
										class="pl-9"
										bind:value={loginForm.identifier}
									/>
								</div>
							</div>
							<div class="space-y-2">
								<div class="flex items-center justify-between">
									<Label for="password">{$_('login.password')}</Label>
									<button
										type="button"
										class="text-xs font-medium text-primary hover:underline"
										onclick={() => (activeView = 'reset-password')}
									>
										{$_('login.forgot_password')}
									</button>
								</div>
								<div class="relative">
									<Lock class="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
									<Input
										id="password"
										type="password"
										class="pl-9"
										bind:value={loginForm.password}
									/>
								</div>
							</div>
						</Card.Content>
						<Card.Footer>
							<Button class="w-full" onclick={handleLogin} disabled={isSubmitting}>
								{isSubmitting ? $_('login.btn.signing_in') : $_('login.btn.signin')}
							</Button>
						</Card.Footer>
					</Card.Root>
				</Tabs.Content>

				<!-- REGISTER TAB -->
				<Tabs.Content value="register">
					<Card.Root>
						<Card.Header>
							<Card.Title>{$_('login.create_account')}</Card.Title>
							<Card.Description>{$_('login.join_us')}</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-4">
							<div class="space-y-2">
								<Label for="reg-username">{$_('login.username')}</Label>
								<div class="relative">
									<User class="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
									<Input
										id="reg-username"
										placeholder={$_('login.username_placeholder')}
										class="pl-9"
										bind:value={registerForm.username}
									/>
								</div>
							</div>
							<div class="space-y-2">
								<Label for="reg-email">{$_('login.email')}</Label>
								<div class="relative">
									<Mail class="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
									<Input
										id="reg-email"
										type="email"
										placeholder={$_('login.email_placeholder')}
										class="pl-9"
										bind:value={registerForm.email}
									/>
								</div>
							</div>
							<div class="grid grid-cols-2 gap-4">
								<div class="col-span-2 space-y-2 sm:col-span-1">
									<Label for="reg-password">{$_('login.password')}</Label>
									<div class="relative">
										<Lock class="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
										<Input
											id="reg-password"
											type="password"
											class="pl-9"
											bind:value={registerForm.password}
										/>
									</div>
								</div>
								<div class="col-span-2 space-y-2 sm:col-span-1">
									<Label for="reg-confirm">{$_('login.confirm_password')}</Label>
									<div class="relative">
										<Lock class="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
										<Input
											id="reg-confirm"
											type="password"
											class="pl-9"
											bind:value={registerForm.confirmed_password}
										/>
									</div>
								</div>
							</div>
							<div class="space-y-2">
								<Label for="reg-code">{$_('login.verification_code')}</Label>
								<div class="flex gap-2">
									<div class="relative flex-1">
										<KeyRound class="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
										<Input
											id="reg-code"
											placeholder={$_('login.code_placeholder')}
											class="pl-9"
											maxlength={6}
											bind:value={registerForm.email_code}
										/>
									</div>
									<Button
										variant="outline"
										type="button"
										disabled={!registerForm.email}
										onclick={() => handleSendCode(registerForm.email)}
									>
										{$_('login.btn.send_code')}
									</Button>
								</div>
							</div>
						</Card.Content>
						<Card.Footer>
							<Button class="w-full" onclick={handleRegister} disabled={isSubmitting}>
								{isSubmitting ? $_('login.btn.creating_account') : $_('login.btn.create_account')}
							</Button>
						</Card.Footer>
					</Card.Root>
				</Tabs.Content>
			</Tabs.Root>
		{:else}
			<!-- RESET PASSWORD VIEW -->
			<Card.Root>
				<Card.Header>
					<Card.Title>{$_('login.reset_password_title')}</Card.Title>
					<Card.Description>{$_('login.reset_password_desc')}</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					<div class="space-y-2">
						<Label for="reset-email">{$_('login.email')}</Label>
						<div class="relative">
							<Mail class="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
							<Input
								id="reset-email"
								type="email"
								placeholder={$_('login.email_placeholder')}
								class="pl-9"
								bind:value={resetForm.email}
							/>
						</div>
					</div>
					<div class="space-y-2">
						<Label for="reset-code">{$_('login.verification_code')}</Label>
						<div class="flex gap-2">
							<div class="relative flex-1">
								<KeyRound class="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
								<Input
									id="reset-code"
									placeholder={$_('login.code_placeholder')}
									class="pl-9"
									maxlength={6}
									bind:value={resetForm.email_code}
								/>
							</div>
							<Button
								variant="outline"
								type="button"
								disabled={!resetForm.email}
								onclick={() => handleSendCode(resetForm.email)}
							>
								{$_('login.btn.send_code')}
							</Button>
						</div>
					</div>
					<div class="space-y-2">
						<Label for="reset-pass">{$_('login.new_password')}</Label>
						<div class="relative">
							<Lock class="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
							<Input id="reset-pass" type="password" class="pl-9" bind:value={resetForm.password} />
						</div>
					</div>
					<div class="space-y-2">
						<Label for="reset-confirm">{$_('login.confirm_password')}</Label>
						<div class="relative">
							<Lock class="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
							<Input
								id="reset-confirm"
								type="password"
								class="pl-9"
								bind:value={resetForm.confirmed_password}
							/>
						</div>
					</div>
				</Card.Content>
				<Card.Footer class="flex flex-col gap-2">
					<Button class="w-full" onclick={handleResetPassword} disabled={isSubmitting}>
						{isSubmitting ? $_('login.btn.resetting') : $_('login.btn.reset_password')}
					</Button>
					<Button
						variant="ghost"
						class="w-full text-muted-foreground"
						onclick={() => (activeView = 'auth')}
					>
						{$_('login.btn.back_to_login')}
					</Button>
				</Card.Footer>
			</Card.Root>
		{/if}
	</div>
</div>
