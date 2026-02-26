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
			color: 'bg-blue-50 text-blue-700 border-blue-200',
			iconBg: 'bg-blue-100'
		},
		{
			label: 'Complete',
			value: stats?.complete ?? 0,
			icon: CheckCircle2,
			color: 'bg-emerald-50 text-emerald-700 border-emerald-200',
			iconBg: 'bg-emerald-100'
		},
		{
			label: 'Incomplete',
			value: stats?.incomplete ?? 0,
			icon: AlertCircle,
			color: 'bg-amber-50 text-amber-700 border-amber-200',
			iconBg: 'bg-amber-100'
		},
		{
			label: 'Flagged',
			value: stats?.flagged ?? 0,
			icon: Flag,
			color: 'bg-red-50 text-red-700 border-red-200',
			iconBg: 'bg-red-100'
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
			<h1 class="text-3xl font-bold text-slate-900">Dashboard</h1>
			<p class="mt-1 text-slate-600">Welcome back, {$auth?.name}!</p>
		</div>
		<Button onclick={() => (window.location.href = '/plans/new')}>
			<Plus class="mr-2 h-4 w-4" />
			New Plan
		</Button>
	</div>

	{#if loading}
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{#each Array(4) as _}
				<div class="animate-pulse rounded-xl border border-slate-200 bg-white p-6">
					<div class="mb-4 h-4 w-24 rounded bg-slate-200"></div>
					<div class="h-8 w-16 rounded bg-slate-200"></div>
				</div>
			{/each}
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{#each statCards as stat}
				<div
					class="rounded-xl border bg-white p-6 {stat.color} transition-all duration-200 hover:shadow-md"
				>
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm font-medium opacity-80">{stat.label}</p>
							<p class="mt-2 text-3xl font-bold">{stat.value}</p>
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
			<div class="overflow-hidden rounded-xl border border-slate-200 bg-white">
				<div class="flex items-center justify-between border-b border-slate-200 p-6">
					<h2 class="text-lg font-semibold text-slate-900">Recent Plans</h2>
					<a
						href="/plans"
						class="flex items-center gap-1 text-sm font-medium text-blue-600 transition-colors hover:text-blue-700"
					>
						View All
						<ArrowRight class="h-4 w-4" />
					</a>
				</div>
				<div class="divide-y divide-slate-100">
					{#if !recentPlans?.length}
						<div class="p-8 text-center text-slate-500">
							No plans yet. Create your first plan to get started.
						</div>
					{:else}
						{#each recentPlans as plan}
							<a
								href="/plans/{plan.id}"
								class="group flex items-center gap-4 p-4 transition-colors hover:bg-slate-50"
							>
								<div class="min-w-0 flex-1">
									<div class="flex items-center gap-2">
										<h3
											class="truncate font-medium text-slate-900 transition-colors group-hover:text-blue-600"
										>
											{plan.name}
										</h3>
										<span
											class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium capitalize
											{plan.status === 'complete'
												? 'bg-emerald-100 text-emerald-700'
												: plan.status === 'flagged'
													? 'bg-red-100 text-red-700'
													: 'bg-amber-100 text-amber-700'}"
										>
											{plan.status}
										</span>
									</div>
									<p class="mt-1 text-sm text-slate-500">
										{#if plan.beds || plan.baths}
											{plan.beds ?? 0} beds · {plan.baths ?? 0} baths
											{#if plan.heated_sf}· {plan.heated_sf.toLocaleString()} SF{/if}
										{:else}
											No specs added
										{/if}
									</p>
								</div>
								<ArrowRight
									class="h-4 w-4 text-slate-400 transition-colors group-hover:text-blue-600"
								/>
							</a>
						{/each}
					{/if}
				</div>
			</div>

			<!-- Recent Activity -->
			<div class="overflow-hidden rounded-xl border border-slate-200 bg-white">
				<div class="flex items-center justify-between border-b border-slate-200 p-6">
					<h2 class="text-lg font-semibold text-slate-900">Recent Activity</h2>
					<a
						href="/activity"
						class="flex items-center gap-1 text-sm font-medium text-blue-600 transition-colors hover:text-blue-700"
					>
						View All
						<ArrowRight class="h-4 w-4" />
					</a>
				</div>
				<div class="divide-y divide-slate-100">
					{#if !recentActivity?.length}
						<div class="p-8 text-center text-slate-500">No recent activity.</div>
					{:else}
						{#each recentActivity as activity}
							<div class="flex items-start gap-3 p-4">
								<div
									class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-slate-100"
								>
									<span class="text-sm font-medium text-slate-600">
										{activity.user?.name?.charAt(0).toUpperCase() ?? '?'}
									</span>
								</div>
								<div class="min-w-0 flex-1">
									<p class="text-sm text-slate-900">
										<span class="font-medium">{activity.user?.name ?? 'Unknown'}</span>
										<span class="text-slate-600">{formatAction(activity.action)}</span>
										{#if activity.plan}
											<span class="font-medium text-blue-600">{activity.plan.name}</span>
										{/if}
									</p>
									<div class="mt-1 flex items-center gap-1 text-xs text-slate-500">
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
		<div class="rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 p-6 text-white">
			<h2 class="mb-2 text-lg font-semibold">Quick Actions</h2>
			<p class="mb-4 text-blue-100">Get started with these common tasks</p>
			<div class="flex flex-wrap gap-3">
				<a
					href="/plans"
					class="inline-flex items-center justify-center rounded-lg bg-white px-4 py-2 text-sm font-medium text-blue-700 transition-colors hover:bg-blue-50"
				>
					View All Plans
				</a>
				<a
					href="/plans/new"
					class="inline-flex items-center justify-center rounded-lg border border-blue-400 bg-blue-500 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-400"
				>
					<Plus class="mr-2 h-4 w-4" />
					Add New Plan
				</a>
				<button
					onclick={() => (exportModalOpen = true)}
					class="inline-flex items-center justify-center rounded-lg border border-blue-400 bg-transparent px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-500"
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
