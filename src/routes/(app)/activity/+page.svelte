<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';

	import { getActivities, getUsers, getPlans } from '$lib/api';
	import { toast } from 'svelte-sonner';
	import { Clock, Filter, ChevronLeft, ChevronRight, X, History } from '@lucide/svelte';
	import type { Activity, PaginatedResponse, User, Plan } from '$lib/types';
	import EmptyState from '$lib/components/ui/empty-state/EmptyState.svelte';
	import TableSkeleton from '$lib/components/ui/skeleton/TableSkeleton.svelte';

	// Filters
	let userFilter = $state('');
	let actionFilter = $state('');
	let planFilter = $state('');

	// Pagination
	let currentPage = $state(1);
	let itemsPerPage = $state(20);

	// Data
	let activities = $state<PaginatedResponse<Activity> | null>(null);
	let users = $state<User[]>([]);
	let plans = $state<Plan[]>([]);
	let loading = $state(true);
	let usersLoading = $state(true);
	let plansLoading = $state(true);

	// Available actions for filter
	const actionTypes = [
		{ value: 'plan.created', label: 'Plan Created' },
		{ value: 'plan.updated', label: 'Plan Updated' },
		{ value: 'plan.deleted', label: 'Plan Deleted' },
		{ value: 'plan.restored', label: 'Plan Restored' },
		{ value: 'plan.flagged', label: 'Plan Flagged' },
		{ value: 'plan.unflagged', label: 'Plan Unflagged' },
		{ value: 'plan.duplicated', label: 'Plan Duplicated' },
		{ value: 'file.uploaded', label: 'File Uploaded' },
		{ value: 'file.deleted', label: 'File Deleted' },
		{ value: 'user.login', label: 'User Login' },
		{ value: 'user.logout', label: 'User Logout' }
	];

	onMount(() => {
		loadActivities();
		loadUsers();
		loadPlans();
	});

	$effect(() => {
		loadActivities();
	});

	async function loadActivities() {
		loading = true;
		try {
			activities = await getActivities({
				user_id: userFilter,
				action: actionFilter,
				plan_id: planFilter,
				page: currentPage,
				limit: itemsPerPage
			});
		} catch (err) {
			toast.error('Failed to load activity log');
		} finally {
			loading = false;
		}
	}

	async function loadUsers() {
		try {
			users = await getUsers();
		} catch (err) {
			// Silent fail - users filter is optional
		} finally {
			usersLoading = false;
		}
	}

	async function loadPlans() {
		try {
			const response = await getPlans({ limit: 100 });
			plans = response.data;
		} catch (err) {
			// Silent fail - plans filter is optional
		} finally {
			plansLoading = false;
		}
	}

	function clearFilters() {
		userFilter = '';
		actionFilter = '';
		planFilter = '';
		currentPage = 1;
	}

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

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	const hasActiveFilters = $derived(userFilter || actionFilter || planFilter);
	const totalPages = $derived(activities?.meta?.total_pages ?? 1);
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-3xl font-bold text-card-foreground">Activity Log</h1>
			<p class="mt-1 text-muted-foreground">Track changes and actions across the system</p>
		</div>
	</div>

	<!-- Filters -->
	<div class="rounded-lg border border-border bg-card p-4">
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			<div class="space-y-2">
				<span class="text-sm font-medium text-foreground">User</span>
				<select
					bind:value={userFilter}
					class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
					disabled={usersLoading}
				>
					<option value="">All users</option>
					{#each users as user}
						<option value={user.id}>{user.name}</option>
					{/each}
				</select>
			</div>

			<div class="space-y-2">
				<span class="text-sm font-medium text-foreground">Action</span>
				<select
					bind:value={actionFilter}
					class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
				>
					<option value="">All actions</option>
					{#each actionTypes as action}
						<option value={action.value}>{action.label}</option>
					{/each}
				</select>
			</div>

			<div class="space-y-2 sm:col-span-2 lg:col-span-1">
				<span class="text-sm font-medium text-foreground">Plan</span>
				<select
					bind:value={planFilter}
					class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
					disabled={plansLoading}
				>
					<option value="">All plans</option>
					{#each plans as plan}
						<option value={plan.id}>{plan.name}</option>
					{/each}
				</select>
			</div>
		</div>
		{#if hasActiveFilters}
			<div class="mt-4 flex justify-end">
				<Button variant="ghost" size="sm" onclick={clearFilters}>
					<X class="mr-2 h-4 w-4" />
					Clear Filters
				</Button>
			</div>
		{/if}
	</div>

	<!-- Activity List -->
	<div class="rounded-lg border border-border bg-card">
		{#if loading}
			<TableSkeleton count={5} />
		{:else if !activities || activities.data.length === 0}
			<div class="p-8">
				<EmptyState
					icon={History}
					title="No activity found"
					description={hasActiveFilters
						? 'Try adjusting your filters to see more results'
						: 'Activity will appear here when users perform actions'}
					actionLabel={hasActiveFilters ? 'Clear Filters' : undefined}
					onAction={hasActiveFilters ? clearFilters : undefined}
				/>
			</div>
		{:else}
			<!-- Desktop Table View (hidden on mobile) -->
			<div class="hidden md:block">
				<!-- Table Header -->
				<div class="border-b border-border bg-muted px-6 py-3">
					<div
						class="grid grid-cols-12 gap-4 text-xs font-medium tracking-wider text-muted-foreground uppercase"
					>
						<div class="col-span-3">User</div>
						<div class="col-span-3">Action</div>
						<div class="col-span-4">Details</div>
						<div class="col-span-2">Time</div>
					</div>
				</div>

				<!-- Activity Items -->
				<div class="divide-y divide-slate-100">
					{#each activities.data as activity}
						<div class="grid grid-cols-12 items-center gap-4 px-6 py-4 hover:bg-muted">
							<!-- User -->
							<div class="col-span-3 flex items-center gap-3">
								<div
									class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-muted"
								>
									<span class="text-sm font-medium text-muted-foreground">
										{activity.user?.name?.charAt(0).toUpperCase() ?? '?'}
									</span>
								</div>
								<span class="truncate font-medium text-card-foreground">
									{activity.user?.name ?? 'Unknown'}
								</span>
							</div>

							<!-- Action -->
							<div class="col-span-3">
								<span
									class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium
									{activity.action.includes('created')
										? 'bg-green-500/15 text-green-600'
										: activity.action.includes('deleted')
											? 'bg-red-500/15 text-red-600'
											: activity.action.includes('updated')
												? 'bg-blue-500/15 text-blue-600'
												: 'bg-muted text-foreground'}"
								>
									{formatAction(activity.action)}
								</span>
							</div>

							<!-- Details -->
							<div class="col-span-4 text-sm text-muted-foreground">
								{#if activity.plan}
									<a
										href="/plans/{activity.plan.id}"
										class="font-medium text-blue-600 hover:text-blue-700"
									>
										{activity.plan.name}
									</a>
								{:else if activity.detail}
									<span class="text-muted-foreground">{JSON.stringify(activity.detail)}</span>
								{:else}
									<span class="text-muted-foreground">—</span>
								{/if}
							</div>

							<!-- Time -->
							<div
								class="col-span-2 text-sm text-muted-foreground"
								title={formatDate(activity.created_at)}
							>
								{formatTime(activity.created_at)}
							</div>
						</div>
					{/each}
				</div>
			</div>

			<!-- Mobile Card View (shown only on mobile) -->
			<div class="divide-y divide-slate-100 md:hidden">
				{#each activities.data as activity}
					<div class="p-4 hover:bg-muted">
						<div class="flex items-start justify-between gap-3">
							<div class="flex min-w-0 flex-1 items-center gap-3">
								<div
									class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-muted"
								>
									<span class="text-sm font-medium text-muted-foreground">
										{activity.user?.name?.charAt(0).toUpperCase() ?? '?'}
									</span>
								</div>
								<div class="min-w-0 flex-1">
									<p class="truncate font-medium text-card-foreground">
										{activity.user?.name ?? 'Unknown'}
									</p>
									<p class="text-xs text-muted-foreground">
										{formatTime(activity.created_at)}
									</p>
								</div>
							</div>
							<span
								class="inline-flex flex-shrink-0 items-center rounded-full px-2.5 py-0.5 text-xs font-medium
								{activity.action.includes('created')
									? 'bg-green-500/15 text-green-600'
									: activity.action.includes('deleted')
										? 'bg-red-500/15 text-red-600'
										: activity.action.includes('updated')
											? 'bg-blue-500/15 text-blue-600'
											: 'bg-muted text-foreground'}"
							>
								{formatAction(activity.action)}
							</span>
						</div>
						<div class="mt-3 pl-13">
							{#if activity.plan}
								<a
									href="/plans/{activity.plan.id}"
									class="text-sm font-medium text-blue-600 hover:text-blue-700"
								>
									{activity.plan.name}
								</a>
							{:else if activity.detail}
								<span class="text-sm text-muted-foreground">{JSON.stringify(activity.detail)}</span>
							{:else}
								<span class="text-sm text-muted-foreground">—</span>
							{/if}
						</div>
					</div>
				{/each}
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div
					class="flex flex-col gap-3 border-t border-border px-6 py-4 sm:flex-row sm:items-center sm:justify-between"
				>
					<div class="text-center text-sm text-muted-foreground sm:text-left">
						Showing {(currentPage - 1) * itemsPerPage + 1} - {Math.min(
							currentPage * itemsPerPage,
							activities.meta.total
						)} of {activities.meta.total} activities
					</div>
					<div class="flex items-center justify-center gap-2">
						<Button
							variant="outline"
							size="sm"
							disabled={currentPage === 1}
							onclick={() => (currentPage = Math.max(1, currentPage - 1))}
						>
							<ChevronLeft class="mr-1 h-4 w-4" />
							<span class="hidden sm:inline">Previous</span>
						</Button>
						<span class="px-2 text-sm text-muted-foreground">
							Page {currentPage} of {totalPages}
						</span>
						<Button
							variant="outline"
							size="sm"
							disabled={currentPage >= totalPages}
							onclick={() => (currentPage = Math.min(totalPages, currentPage + 1))}
						>
							<span class="hidden sm:inline">Next</span>
							<ChevronRight class="ml-1 h-4 w-4" />
						</Button>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div>
