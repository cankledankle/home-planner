<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores';
	import { getDashboardStats, getRecentPlans, getActivities } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import { toast } from 'svelte-sonner';
	import {
		Home,
		CheckCircle2,
		AlertCircle,
		Flag,
		Plus,
		ArrowRight,
		Clock,
		Download
	} from '@lucide/svelte';
	import ExportModal from '$lib/components/plan/ExportModal.svelte';
	import type { Plan, Activity } from '$lib/types';
	import type { DashboardStats } from '$lib/api';

	let stats = $state<DashboardStats | null>(null);
	let recentPlans = $state<Plan[]>([]);
	let recentActivity = $state<Activity[]>([]);
	let loading = $state(true);
	let exportModalOpen = $state(false);

	onMount(async () => {
		try {
			const [statsData, plansData, activityData] = await Promise.all([
				getDashboardStats(),
				getRecentPlans(5),
				getActivities({ limit: 5 })
			]);
			stats = statsData;
			recentPlans = plansData ?? [];
			recentActivity = activityData?.data ?? [];
		} catch (err) {
			toast.error('Failed to load dashboard data');
		} finally {
			loading = false;
		}
	});

	const statCards = $derived([
		{
			label: 'Total Plans',
			value: stats?.total ?? 0,
			icon: Home,
			color: 'border-blue-200 dark:border-blue-800',
			iconBg: 'bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300'
		},
		{
			label: 'Complete',
			value: stats?.complete ?? 0,
			icon: CheckCircle2,
			color: 'border-emerald-200 dark:border-emerald-800',
			iconBg: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/50 dark:text-emerald-300'
		},
		{
			label: 'Incomplete',
			value: stats?.incomplete ?? 0,
			icon: AlertCircle,
			color: 'border-amber-200 dark:border-amber-800',
			iconBg: 'bg-amber-100 text-amber-700 dark:bg-amber-900/50 dark:text-amber-300'
		},
		{
			label: 'Flagged',
			value: stats?.flagged ?? 0,
			icon: Flag,
			color: 'border-red-200 dark:border-red-800',
			iconBg: 'bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-300'
		}
	]);

	function formatAction(action: string): string {
		return action
			.replace(/\./g, ' ')
			.replace(/_/g, ' ')
			.replace(/\b\w/g, (l) => l.toUpperCase());
	}

	function formatTime(dateString: string): string {
		const date = new Date(dateString);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(diff / 3600000);
		const days = Math.floor(diff / 86400000);

		if (minutes < 1) return 'Just now';
		if (minutes < 60) return `${minutes}m ago`;
		if (hours < 24) return `${hours}h ago`;
		if (days < 7) return `${days}d ago`;
		return date.toLocaleDateString();
	}
</script>

<div class="space-y-8">
	<!-- Header -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-3xl font-bold text-foreground">Dashboard</h1>
			<p class="mt-1 text-muted-foreground">Welcome back, {$auth?.name}!</p>
		</div>
		<Button onclick={() => (window.location.href = '/plans/new')}>
			<Plus class="mr-2 h-4 w-4" />
			New Plan
		</Button>
	</div>

	{#if loading}
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{#each Array(4) as _}
				<div class="animate-pulse rounded-xl border border-border bg-card p-6">
					<div class="mb-4 h-4 w-24 rounded bg-muted"></div>
					<div class="h-8 w-16 rounded bg-muted"></div>
				</div>
			{/each}
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{#each statCards as stat}
				<div
					class="rounded-xl border bg-card p-6 {stat.color} transition-all duration-200 hover:shadow-md"
				>
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm font-medium text-muted-foreground">{stat.label}</p>
							<p class="mt-2 text-3xl font-bold text-card-foreground">{stat.value}</p>
						</div>
						<div class="h-12 w-12 rounded-lg {stat.iconBg} flex items-center justify-center">
							<stat.icon class="h-6 w-6" />
						</div>
					</div>
				</div>
			{/each}
		</div>

		<!-- Two Column Layout -->
		<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
			<!-- Recent Plans -->
			<div class="overflow-hidden rounded-xl border border-border bg-card">
				<div class="flex items-center justify-between border-b border-border p-6">
					<h2 class="text-lg font-semibold text-card-foreground">Recent Plans</h2>
					<a
						href="/plans"
						class="flex items-center gap-1 text-sm font-medium text-primary transition-colors hover:text-primary/80"
					>
						View All
						<ArrowRight class="h-4 w-4" />
					</a>
				</div>
				<div class="divide-y divide-border">
					{#if !recentPlans?.length}
						<div class="p-8 text-center text-muted-foreground">
							No plans yet. Create your first plan to get started.
						</div>
					{:else}
						{#each recentPlans as plan}
							<a
								href="/plans/{plan.id}"
								class="group flex items-center gap-4 p-4 transition-colors hover:bg-accent"
							>
								<div class="min-w-0 flex-1">
									<div class="flex items-center gap-2">
										<h3
											class="truncate font-medium text-card-foreground transition-colors group-hover:text-primary"
										>
											{plan.name}
										</h3>
										<span
											class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium capitalize
											{plan.status === 'complete'
												? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/50 dark:text-emerald-300'
												: plan.status === 'flagged'
													? 'bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-300'
													: 'bg-amber-100 text-amber-700 dark:bg-amber-900/50 dark:text-amber-300'}"
										>
											{plan.status}
										</span>
									</div>
									<p class="mt-1 text-sm text-muted-foreground">
										{#if plan.beds || plan.baths}
											{plan.beds ?? 0} beds · {plan.baths ?? 0} baths
											{#if plan.heated_sf}· {plan.heated_sf.toLocaleString()} SF{/if}
										{:else}
											No specs added
										{/if}
									</p>
								</div>
								<ArrowRight
									class="h-4 w-4 text-muted-foreground transition-colors group-hover:text-primary"
								/>
							</a>
						{/each}
					{/if}
				</div>
			</div>

			<!-- Recent Activity -->
			<div class="overflow-hidden rounded-xl border border-border bg-card">
				<div class="flex items-center justify-between border-b border-border p-6">
					<h2 class="text-lg font-semibold text-card-foreground">Recent Activity</h2>
					<a
						href="/activity"
						class="flex items-center gap-1 text-sm font-medium text-primary transition-colors hover:text-primary/80"
					>
						View All
						<ArrowRight class="h-4 w-4" />
					</a>
				</div>
				<div class="divide-y divide-border">
					{#if !recentActivity?.length}
						<div class="p-8 text-center text-muted-foreground">No recent activity.</div>
					{:else}
						{#each recentActivity as activity}
							<div class="flex items-start gap-3 p-4">
								<div
									class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-muted"
								>
									<span class="text-sm font-medium text-muted-foreground">
										{activity.user?.name?.charAt(0).toUpperCase() ?? '?'}
									</span>
								</div>
								<div class="min-w-0 flex-1">
									<p class="text-sm">
										<span class="font-medium text-card-foreground"
											>{activity.user?.name ?? 'Unknown'}</span
										>
										<span class="text-muted-foreground">{formatAction(activity.action)}</span>
										{#if activity.plan}
											<span class="font-medium text-primary">{activity.plan.name}</span>
										{/if}
									</p>
									<div class="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
										<Clock class="h-3 w-3" />
										{formatTime(activity.created_at)}
									</div>
								</div>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="rounded-xl bg-gradient-to-r from-primary to-primary/80 p-6 text-primary-foreground">
			<h2 class="mb-2 text-lg font-semibold">Quick Actions</h2>
			<p class="mb-4 text-primary-foreground/80">Get started with these common tasks</p>
			<div class="flex flex-wrap gap-3">
				<a
					href="/plans"
					class="inline-flex items-center justify-center rounded-lg bg-background px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-accent"
				>
					View All Plans
				</a>
				<a
					href="/plans/new"
					class="inline-flex items-center justify-center rounded-lg border border-primary-foreground/30 bg-primary-foreground/10 px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary-foreground/20"
				>
					<Plus class="mr-2 h-4 w-4" />
					Add New Plan
				</a>
				<button
					onclick={() => (exportModalOpen = true)}
					class="inline-flex items-center justify-center rounded-lg border border-primary-foreground/30 bg-transparent px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary-foreground/10"
				>
					<Download class="mr-2 h-4 w-4" />
					Export Data
				</button>
			</div>
		</div>
	{/if}
</div>

<!-- Export Modal -->
<ExportModal bind:open={exportModalOpen} onClose={() => (exportModalOpen = false)} />
