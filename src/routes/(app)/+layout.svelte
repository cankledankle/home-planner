<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import { Toaster, toast } from 'svelte-sonner';
	import ThemeToggle from '$lib/components/ui/theme-toggle/ThemeToggle.svelte';
	import {
		LayoutDashboard,
		Home,
		History,
		Settings,
		LogOut,
		Menu,
		X,
		Upload
	} from '@lucide/svelte';

	let { children } = $props();
	let loading = $state(true);
	let mobileMenuOpen = $state(false);

	onMount(async () => {
		try {
			console.log('Checking auth...');
			const user = await api.me();
			console.log('Auth check result:', user);
			if (user) {
				auth.set(user);
			} else {
				console.log('No user, redirecting to login');
				goto('/login');
			}
		} catch (err) {
			console.error('Auth check failed:', err);
			goto('/login');
		} finally {
			loading = false;
		}
	});

	async function handleLogout() {
		try {
			await api.logout();
			auth.set(null);
			toast.success('Logged out successfully');
			goto('/login');
		} catch {
			toast.error('Failed to logout');
		}
	}

	const navItems = [
		{ href: '/', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/plans', label: 'Plans', icon: Home },
		{ href: '/import', label: 'Import', icon: Upload },
		{ href: '/activity', label: 'Activity', icon: History }
	];

	const adminNavItems = [{ href: '/settings', label: 'Settings', icon: Settings }];

	function isActive(href: string) {
		if (href === '/') {
			return $page.url.pathname === '/';
		}
		return $page.url.pathname.startsWith(href);
	}
</script>

<Toaster position="top-right" richColors />

{#if loading}
	<div class="flex min-h-screen items-center justify-center bg-background">
		<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-foreground"></div>
	</div>
{:else if $auth}
	<div class="flex min-h-screen bg-background">
		<!-- Desktop Sidebar -->
		<aside class="hidden w-64 flex-col border-r border-border bg-card lg:flex">
			<div class="border-b border-border p-6">
				<h1 class="text-xl font-bold text-card-foreground">Home Planner</h1>
			</div>

			<nav class="flex-1 space-y-1 p-4">
				{#each navItems as item}
					<a
						href={item.href}
						class="flex items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium transition-colors {isActive(
							item.href
						)
							? 'bg-accent text-accent-foreground'
							: 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'}"
					>
						<item.icon class="h-5 w-5" />
						{item.label}
					</a>
				{/each}

				{#if $auth.role === 'admin'}
					<div class="mt-4 border-t border-border pt-4">
						<p
							class="mb-2 px-4 text-xs font-semibold tracking-wider text-muted-foreground uppercase"
						>
							Admin
						</p>
						{#each adminNavItems as item}
							<a
								href={item.href}
								class="flex items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium transition-colors {isActive(
									item.href
								)
									? 'bg-accent text-accent-foreground'
									: 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'}"
							>
								<item.icon class="h-5 w-5" />
								{item.label}
							</a>
						{/each}
					</div>
				{/if}
			</nav>

			<div class="border-t border-border p-4">
				<div class="px-4 py-2">
					<ThemeToggle />
				</div>

				<div class="flex items-center gap-3 px-4 py-3">
					<div class="flex h-8 w-8 items-center justify-center rounded-full bg-muted">
						<span class="text-sm font-medium text-muted-foreground">
							{$auth.name.charAt(0).toUpperCase()}
						</span>
					</div>
					<div class="min-w-0 flex-1">
						<p class="truncate text-sm font-medium text-card-foreground">{$auth.name}</p>
						<p class="truncate text-xs text-muted-foreground">{$auth.email}</p>
					</div>
				</div>
				<Button variant="ghost" class="mt-2 w-full justify-start gap-3" onclick={handleLogout}>
					<LogOut class="h-5 w-5" />
					Logout
				</Button>
			</div>
		</aside>

		<!-- Mobile Header -->
		<div
			class="fixed top-0 right-0 left-0 z-50 flex h-16 items-center justify-between border-b border-border bg-card px-4 lg:hidden"
		>
			<h1 class="text-lg font-bold text-card-foreground">Home Planner</h1>
			<Button variant="ghost" size="icon" onclick={() => (mobileMenuOpen = !mobileMenuOpen)}>
				{#if mobileMenuOpen}
					<X class="h-5 w-5" />
				{:else}
					<Menu class="h-5 w-5" />
				{/if}
			</Button>
		</div>

		<!-- Mobile Menu -->
		{#if mobileMenuOpen}
			<div class="fixed inset-0 top-16 z-40 bg-background lg:hidden">
				<nav class="space-y-1 p-4">
					{#each navItems as item}
						<a
							href={item.href}
							class="flex items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium transition-colors {isActive(
								item.href
							)
								? 'bg-accent text-accent-foreground'
								: 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'}"
							onclick={() => (mobileMenuOpen = false)}
						>
							<item.icon class="h-5 w-5" />
							{item.label}
						</a>
					{/each}

					{#if $auth.role === 'admin'}
						<div class="mt-4 border-t border-border pt-4">
							<p
								class="mb-2 px-4 text-xs font-semibold tracking-wider text-muted-foreground uppercase"
							>
								Admin
							</p>
							{#each adminNavItems as item}
								<a
									href={item.href}
									class="flex items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium transition-colors {isActive(
										item.href
									)
										? 'bg-accent text-accent-foreground'
										: 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'}"
									onclick={() => (mobileMenuOpen = false)}
								>
									<item.icon class="h-5 w-5" />
									{item.label}
								</a>
							{/each}
						</div>
					{/if}

					<div class="mt-4 border-t border-border pt-4">
						<div class="px-4 py-2">
							<ThemeToggle />
						</div>
					</div>

					<div class="mt-4 border-t border-border pt-4">
						<div class="flex items-center gap-3 px-4 py-3">
							<div class="flex h-8 w-8 items-center justify-center rounded-full bg-muted">
								<span class="text-sm font-medium text-muted-foreground">
									{$auth.name.charAt(0).toUpperCase()}
								</span>
							</div>
							<div class="min-w-0 flex-1">
								<p class="truncate text-sm font-medium text-card-foreground">{$auth.name}</p>
								<p class="truncate text-xs text-muted-foreground">{$auth.email}</p>
							</div>
						</div>
						<Button variant="ghost" class="mt-2 w-full justify-start gap-3" onclick={handleLogout}>
							<LogOut class="h-5 w-5" />
							Logout
						</Button>
					</div>
				</nav>
			</div>
		{/if}

		<!-- Main Content -->
		<main class="min-w-0 flex-1 pt-16 lg:ml-0 lg:pt-0">
			<div class="overflow-x-hidden p-4 sm:p-6 lg:p-8">
				{@render children()}
			</div>
		</main>
	</div>
{/if}
