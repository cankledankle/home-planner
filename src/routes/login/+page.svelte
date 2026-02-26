<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { auth } from '$lib/stores';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { toast } from 'svelte-sonner';

	let email = $state('');
	let password = $state('');
	let loading = $state(false);
	let error = $state('');

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		loading = true;

		try {
			const user = await api.login({ email, password });
			auth.set(user);
			toast.success('Welcome back!', {
				description: `Logged in as ${user.name}`
			});
			goto('/');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login failed';
			toast.error('Login failed', {
				description: error
			});
		} finally {
			loading = false;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center bg-slate-50">
	<div class="w-full max-w-md rounded-lg bg-white p-8 shadow-lg">
		<div class="mb-8 text-center">
			<h1 class="text-2xl font-bold text-slate-900">Home Planner</h1>
			<p class="mt-2 text-slate-600">Sign in to manage home plans</p>
		</div>

		<form onsubmit={handleSubmit} class="space-y-6">
			<div>
				<label for="email" class="mb-2 block text-sm font-medium text-slate-700"> Email </label>
				<Input
					id="email"
					type="email"
					bind:value={email}
					placeholder="you@example.com"
					required
					disabled={loading}
				/>
			</div>

			<div>
				<label for="password" class="mb-2 block text-sm font-medium text-slate-700">
					Password
				</label>
				<Input
					id="password"
					type="password"
					bind:value={password}
					placeholder="••••••••"
					required
					disabled={loading}
				/>
			</div>

			{#if error}
				<div class="rounded bg-red-50 p-3 text-sm text-red-600">
					{error}
				</div>
			{/if}

			<Button type="submit" class="w-full" disabled={loading}>
				{#if loading}
					Signing in...
				{:else}
					Sign In
				{/if}
			</Button>
		</form>
	</div>
</div>
