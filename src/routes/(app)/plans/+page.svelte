<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import {
		Dialog,
		DialogContent,
		DialogHeader,
		DialogTitle,
		DialogTrigger,
		DialogFooter,
		DialogDescription
	} from '$lib/components/ui/dialog';
	import {
		Plus,
		Search,
		LayoutGrid,
		List,
		ChevronLeft,
		ChevronRight,
		Filter,
		X,
		ChevronDown,
		Download,
		Trash2,
		Flag,
		FlagOff,
		CheckSquare,
		Square,
		FileDown,
		Archive
	} from '@lucide/svelte';
	import { getPlans, createPlan, deletePlan, flagPlan, unflagPlan } from '$lib/api';
	import { toast } from 'svelte-sonner';
	import type { Plan, PaginatedResponse } from '$lib/types';
	import ExportModal from '$lib/components/plan/ExportModal.svelte';
	import EmptyState from '$lib/components/ui/empty-state/EmptyState.svelte';
	import TableSkeleton from '$lib/components/ui/skeleton/TableSkeleton.svelte';
	import ConfirmationDialog from '$lib/components/ui/dialog/ConfirmationDialog.svelte';

	// View state
	let viewMode: 'table' | 'grid' = $state('table');

	// Search and filters
	let searchQuery = $state('');
	let debouncedSearch = $state('');
	let statusFilter = $state('');
	let typeFilter = $state('');
	let styleFilter = $state('');
	let showFilters = $state(false);

	// Sorting
	let sortField = $state('name');
	let sortOrder = $state<'asc' | 'desc'>('asc');

	// Pagination
	let currentPage = $state(1);
	let itemsPerPage = $state(20);

	// Data
	let plans = $state<PaginatedResponse<Plan> | null>(null);
	let loading = $state(true);

	// New plan modal
	let newPlanOpen = $state(false);
	let newPlanName = $state('');
	let creating = $state(false);

	// Bulk selection
	let selectedIds = $state<Set<string>>(new Set());
	let bulkActionLoading = $state(false);
	let confirmDeleteOpen = $state(false);
	let confirmDeleteCount = $state(0);

	// Export modal
	let exportModalOpen = $state(false);

	// Debounce search
	let searchTimeout: ReturnType<typeof setTimeout>;
	$effect(() => {
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			debouncedSearch = searchQuery;
			currentPage = 1;
		}, 300);
	});

	// Fetch plans when filters change
	$effect(() => {
		fetchPlans();
	});

	async function fetchPlans() {
		loading = true;
		try {
			plans = await getPlans({
				search: debouncedSearch,
				status: statusFilter,
				type: typeFilter,
				style: styleFilter,
				sort: sortField,
				order: sortOrder,
				page: currentPage,
				limit: itemsPerPage
			});
		} catch (err) {
			toast.error('Failed to load plans');
		} finally {
			loading = false;
		}
	}

	async function handleCreatePlan() {
		if (!newPlanName.trim()) {
			toast.error('Plan name is required');
			return;
		}

		creating = true;
		try {
			await createPlan({ name: newPlanName.trim() });
			toast.success('Plan created successfully');
			newPlanOpen = false;
			newPlanName = '';
			fetchPlans();
		} catch (err) {
			toast.error('Failed to create plan');
		} finally {
			creating = false;
		}
	}

	function clearFilters() {
		searchQuery = '';
		debouncedSearch = '';
		statusFilter = '';
		typeFilter = '';
		styleFilter = '';
		currentPage = 1;
	}

	function handleSort(field: string) {
		if (sortField === field) {
			sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
		} else {
			sortField = field;
			sortOrder = 'asc';
		}
	}

	const hasActiveFilters = $derived(debouncedSearch || statusFilter || typeFilter || styleFilter);
	const selectedCount = $derived(selectedIds.size);
	const allSelected = $derived(
		!!plans?.data?.length && plans.data.every((p) => selectedIds.has(p.id))
	);
	const someSelected = $derived(selectedIds.size > 0 && !allSelected);

	function toggleSelection(id: string) {
		const newSet = new Set(selectedIds);
		if (newSet.has(id)) {
			newSet.delete(id);
		} else {
			newSet.add(id);
		}
		selectedIds = newSet;
	}

	function toggleSelectAll() {
		if (allSelected) {
			selectedIds = new Set();
		} else {
			selectedIds = new Set(plans?.data?.map((p) => p.id) ?? []);
		}
	}

	function clearSelection() {
		selectedIds = new Set();
	}

	async function handleBulkDelete() {
		confirmDeleteCount = selectedCount;
		confirmDeleteOpen = true;
	}

	async function confirmBulkDelete() {
		bulkActionLoading = true;
		confirmDeleteOpen = false;

		const ids = Array.from(selectedIds);
		let successCount = 0;
		let errorCount = 0;

		for (const id of ids) {
			try {
				await deletePlan(id);
				successCount++;
			} catch (err) {
				errorCount++;
			}
		}

		if (successCount > 0) {
			toast.success(`${successCount} plan${successCount === 1 ? '' : 's'} deleted`);
		}
		if (errorCount > 0) {
			toast.error(`Failed to delete ${errorCount} plan${errorCount === 1 ? '' : 's'}`);
		}

		selectedIds = new Set();
		await fetchPlans();
		bulkActionLoading = false;
	}

	async function handleBulkFlag() {
		bulkActionLoading = true;
		const ids = Array.from(selectedIds);
		let successCount = 0;
		let errorCount = 0;

		for (const id of ids) {
			try {
				await flagPlan(id);
				successCount++;
			} catch (err) {
				errorCount++;
			}
		}

		if (successCount > 0) {
			toast.success(`${successCount} plan${successCount === 1 ? '' : 's'} flagged`);
		}
		if (errorCount > 0) {
			toast.error(`Failed to flag ${errorCount} plan${errorCount === 1 ? '' : 's'}`);
		}

		selectedIds = new Set();
		await fetchPlans();
		bulkActionLoading = false;
	}

	async function handleBulkUnflag() {
		bulkActionLoading = true;
		const ids = Array.from(selectedIds);
		let successCount = 0;
		let errorCount = 0;

		for (const id of ids) {
			try {
				await unflagPlan(id);
				successCount++;
			} catch (err) {
				errorCount++;
			}
		}

		if (successCount > 0) {
			toast.success(`${successCount} plan${successCount === 1 ? '' : 's'} unflagged`);
		}
		if (errorCount > 0) {
			toast.error(`Failed to unflag ${errorCount} plan${errorCount === 1 ? '' : 's'}`);
		}

		selectedIds = new Set();
		await fetchPlans();
		bulkActionLoading = false;
	}
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-3xl font-bold text-slate-900">Plans</h1>
			<p class="mt-1 text-slate-600">
				{plans?.meta.total ?? 0} plans in your collection
			</p>
		</div>
		<Dialog bind:open={newPlanOpen}>
			<DialogTrigger class="w-full sm:w-auto">
				<Button class="w-full sm:w-auto">
					<Plus class="mr-2 h-4 w-4" />
					New Plan
				</Button>
			</DialogTrigger>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Create New Plan</DialogTitle>
				</DialogHeader>
				<div class="space-y-4 pt-4">
					<div>
						<label for="plan-name" class="mb-2 block text-sm font-medium text-slate-700">
							Plan Name *
						</label>
						<Input
							id="plan-name"
							bind:value={newPlanName}
							placeholder="e.g., Abilene"
							onkeydown={(e) => e.key === 'Enter' && handleCreatePlan()}
						/>
					</div>
					<div class="flex justify-end gap-3">
						<Button variant="outline" onclick={() => (newPlanOpen = false)}>Cancel</Button>
						<Button onclick={handleCreatePlan} disabled={creating}>
							{#if creating}
								Creating...
							{:else}
								Create Plan
							{/if}
						</Button>
					</div>
				</div>
			</DialogContent>
		</Dialog>
	</div>

	<!-- Search and Controls -->
	<div class="flex flex-col gap-4 lg:flex-row">
		<div class="relative max-w-md flex-1">
			<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-slate-400" />
			<Input placeholder="Search plans..." class="pl-10" bind:value={searchQuery} />
		</div>

		<div class="flex items-center gap-2">
			<Button
				variant="outline"
				size="sm"
				class={showFilters ? 'bg-slate-100' : ''}
				onclick={() => (showFilters = !showFilters)}
			>
				<Filter class="mr-2 h-4 w-4" />
				Filters
				{#if hasActiveFilters}
					<span class="ml-2 h-2 w-2 rounded-full bg-blue-600"></span>
				{/if}
			</Button>

			<div class="flex items-center overflow-hidden rounded-lg border">
				<Button
					variant="ghost"
					size="sm"
					class={viewMode === 'table' ? 'bg-slate-100' : ''}
					onclick={() => (viewMode = 'table')}
				>
					<List class="h-4 w-4" />
				</Button>
				<Button
					variant="ghost"
					size="sm"
					class={viewMode === 'grid' ? 'bg-slate-100' : ''}
					onclick={() => (viewMode = 'grid')}
				>
					<LayoutGrid class="h-4 w-4" />
				</Button>
			</div>
		</div>
	</div>

	<!-- Filters Panel -->
	{#if showFilters}
		<div class="space-y-4 rounded-lg border border-slate-200 bg-slate-50 p-4">
			<div class="flex items-center justify-between">
				<h3 class="font-medium text-slate-900">Filters</h3>
				{#if hasActiveFilters}
					<Button variant="ghost" size="sm" onclick={clearFilters}>
						<X class="mr-1 h-4 w-4" />
						Clear all
					</Button>
				{/if}
			</div>
			<div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
				<div>
					<label for="status-filter" class="mb-2 block text-sm font-medium text-slate-700"
						>Status</label
					>
					<select
						id="status-filter"
						class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
						bind:value={statusFilter}
					>
						<option value="">All Statuses</option>
						<option value="complete">Complete</option>
						<option value="incomplete">Incomplete</option>
						<option value="flagged">Flagged</option>
					</select>
				</div>
				<div>
					<label for="type-filter" class="mb-2 block text-sm font-medium text-slate-700">Type</label
					>
					<select
						id="type-filter"
						class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
						bind:value={typeFilter}
					>
						<option value="">All Types</option>
						<option value="single_level">Single Level</option>
						<option value="multi_level">Multi Level</option>
					</select>
				</div>
				<div>
					<label for="style-filter" class="mb-2 block text-sm font-medium text-slate-700"
						>Style</label
					>
					<select
						id="style-filter"
						class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
						bind:value={styleFilter}
					>
						<option value="">All Styles</option>
						<option value="cabin">Cabin</option>
						<option value="lodge">Lodge</option>
						<option value="modern">Modern</option>
						<option value="ranch">Ranch</option>
						<option value="farmhouse">Farmhouse</option>
					</select>
				</div>
			</div>
		</div>
	{/if}

	<!-- Bulk Action Bar -->
	{#if selectedCount > 0}
		<div
			class="sticky top-0 z-10 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 rounded-lg border border-blue-200 bg-blue-50 p-3 shadow-sm"
		>
			<div class="flex items-center gap-3">
				<Checkbox
					checked={allSelected}
					indeterminate={someSelected}
					onCheckedChange={toggleSelectAll}
				/>
				<span class="text-sm font-medium text-blue-900">
					{selectedCount} selected
				</span>
			</div>
			<div class="flex flex-wrap items-center gap-2">
				<Button
					variant="outline"
					size="sm"
					class="border-blue-200 bg-white hover:bg-blue-100"
					onclick={handleBulkFlag}
					disabled={bulkActionLoading}
				>
					<Flag class="mr-1.5 h-3.5 w-3.5" />
					<span class="hidden sm:inline">Flag</span>
				</Button>
				<Button
					variant="outline"
					size="sm"
					class="border-blue-200 bg-white hover:bg-blue-100"
					onclick={handleBulkUnflag}
					disabled={bulkActionLoading}
				>
					<FlagOff class="mr-1.5 h-3.5 w-3.5" />
					<span class="hidden sm:inline">Unflag</span>
				</Button>
				<div class="mx-1 h-6 w-px bg-blue-200 hidden sm:block"></div>
				<Button
					variant="outline"
					size="sm"
					class="border-blue-200 bg-white hover:bg-blue-100"
					onclick={() => (exportModalOpen = true)}
					disabled={bulkActionLoading}
				>
					<Download class="mr-1.5 h-3.5 w-3.5" />
					<span class="hidden sm:inline">Export</span>
				</Button>
				<Button
					variant="outline"
					size="sm"
					class="border-blue-200 bg-white hover:bg-blue-100"
					disabled={bulkActionLoading}
				>
					<Archive class="mr-1.5 h-3.5 w-3.5" />
					<span class="hidden sm:inline">ZIP</span>
				</Button>
				<div class="mx-1 h-6 w-px bg-blue-200 hidden sm:block"></div>
				<Button
					variant="outline"
					size="sm"
					class="border-red-200 bg-white text-red-600 hover:bg-red-50 hover:text-red-700"
					onclick={handleBulkDelete}
					disabled={bulkActionLoading}
				>
					<Trash2 class="mr-1.5 h-3.5 w-3.5" />
					<span class="hidden sm:inline">Delete</span>
				</Button>
				<Button variant="ghost" size="sm" onclick={clearSelection} disabled={bulkActionLoading}>
					<X class="h-4 w-4" />
				</Button>
			</div>
		</div>
	{/if}

	<!-- Content -->
	{#if loading}
		<div class="rounded-lg border border-slate-200 bg-white">
			<TableSkeleton count={5} />
		</div>
	{:else if plans?.data?.length === 0}
		<EmptyState
			icon={Search}
			title={hasActiveFilters ? 'No plans match your filters' : 'No plans yet'}
			description={hasActiveFilters
				? "Try adjusting your search or filters to find what you're looking for."
				: 'Get started by creating your first home plan.'}
			actionLabel={hasActiveFilters ? 'Clear Filters' : 'Create First Plan'}
			onAction={hasActiveFilters ? clearFilters : () => newPlanOpen = true}
		/>
	{:else}
		{#if viewMode === 'table'}
			<!-- Table View - isolated overflow container -->
			<div class="relative rounded-lg border border-slate-200 bg-white">
				<div class="overflow-x-auto" style="max-width: 100vw;">
					<table class="min-w-[900px] text-left text-sm">
						<thead class="border-b border-slate-200 bg-slate-50 font-medium text-slate-600">
							<tr>
								<th class="w-10 px-4 py-3">
									<Checkbox
										checked={allSelected}
										indeterminate={someSelected}
										onCheckedChange={toggleSelectAll}
									/>
								</th>
								<th
									class="cursor-pointer px-4 py-3 hover:bg-slate-100"
									onclick={() => handleSort('name')}
								>
									<div class="flex items-center gap-1">
										Name
										{#if sortField === 'name'}
											<ChevronDown class="h-4 w-4 {sortOrder === 'desc' ? 'rotate-180' : ''}" />
										{/if}
									</div>
								</th>
								<th class="px-4 py-3">Status</th>
								<th class="px-4 py-3">Type</th>
								<th class="px-4 py-3">Style</th>
								<th
									class="cursor-pointer px-4 py-3 hover:bg-slate-100"
									onclick={() => handleSort('beds')}
								>
									<div class="flex items-center gap-1">
										Beds
										{#if sortField === 'beds'}
											<ChevronDown class="h-4 w-4 {sortOrder === 'desc' ? 'rotate-180' : ''}" />
										{/if}
									</div>
								</th>
								<th class="px-4 py-3">Baths</th>
								<th
									class="cursor-pointer px-4 py-3 hover:bg-slate-100"
									onclick={() => handleSort('heated_sf')}
								>
									<div class="flex items-center gap-1">
										Heated SF
										{#if sortField === 'heated_sf'}
											<ChevronDown class="h-4 w-4 {sortOrder === 'desc' ? 'rotate-180' : ''}" />
										{/if}
									</div>
								</th>
								<th
									class="cursor-pointer px-4 py-3 hover:bg-slate-100"
									onclick={() => handleSort('updated_at')}
								>
									<div class="flex items-center gap-1">
										Updated
										{#if sortField === 'updated_at'}
											<ChevronDown class="h-4 w-4 {sortOrder === 'desc' ? 'rotate-180' : ''}" />
										{/if}
									</div>
								</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-100">
							{#each plans?.data ?? [] as plan}
								<tr
									class="transition-colors hover:bg-slate-50 {selectedIds.has(plan.id)
										? 'bg-blue-50/50'
										: ''}"
								>
									<td class="px-4 py-3">
										<Checkbox
											checked={selectedIds.has(plan.id)}
											onCheckedChange={() => toggleSelection(plan.id)}
										/>
									</td>
									<td class="px-4 py-3">
										<a
											href="/plans/{plan.id}"
											class="font-medium text-blue-600 hover:text-blue-700"
										>
											{plan.name}
										</a>
									</td>
									<td class="px-4 py-3">
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
									</td>
									<td class="px-4 py-3 text-slate-600 capitalize"
										>{plan.type?.replace('_', ' ') ?? '-'}</td
									>
									<td class="px-4 py-3 text-slate-600 capitalize">{plan.style ?? '-'}</td>
									<td class="px-4 py-3 text-slate-600">{plan.beds ?? '-'}</td>
									<td class="px-4 py-3 text-slate-600">
										{plan.baths ?? '-'}{plan.half_baths ? `.${plan.half_baths}` : ''}
									</td>
									<td class="px-4 py-3 text-slate-600">
										{plan.heated_sf?.toLocaleString() ?? '-'}
									</td>
									<td class="px-4 py-3 text-xs text-slate-500">
										{new Date(plan.updated_at ?? '').toLocaleDateString()}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{:else}
			<!-- Grid View -->
			<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
				{#each plans?.data ?? [] as plan}
					<div
						class="group relative overflow-hidden rounded-lg border border-slate-200 bg-white transition-all hover:shadow-md {selectedIds.has(
							plan.id
						)
							? 'border-blue-500 ring-2 ring-blue-500'
							: ''}"
					>
						<!-- Selection Checkbox -->
						<div class="absolute top-3 left-3 z-10">
							<Checkbox
								checked={selectedIds.has(plan.id)}
								onCheckedChange={() => toggleSelection(plan.id)}
								class="bg-white shadow-sm"
							/>
						</div>
						<a href="/plans/{plan.id}" class="block">
							<div class="flex aspect-[4/3] items-center justify-center bg-slate-100">
								<span class="text-sm text-slate-400">No image</span>
							</div>
							<div class="p-4">
								<div class="flex items-start justify-between gap-2">
									<h3
										class="truncate font-medium text-slate-900 transition-colors group-hover:text-blue-600"
									>
										{plan.name}
									</h3>
									<span
										class="inline-flex flex-shrink-0 items-center rounded-full px-2 py-0.5 text-xs font-medium capitalize
									{plan.status === 'complete'
											? 'bg-emerald-100 text-emerald-700'
											: plan.status === 'flagged'
												? 'bg-red-100 text-red-700'
												: 'bg-amber-100 text-amber-700'}"
									>
										{plan.status}
									</span>
								</div>
								<p class="mt-2 text-sm text-slate-600">
									{#if plan.beds || plan.baths}
										{plan.beds ?? 0} beds · {plan.baths ?? 0} baths
									{:else}
										No specs
									{/if}
								</p>
								{#if plan.heated_sf}
									<p class="mt-1 text-sm text-slate-500">{plan.heated_sf.toLocaleString()} SF</p>
								{/if}
							</div>
						</a>
					</div>
				{/each}
			</div>
		{/if}

		<!-- Pagination -->
		{#if plans?.meta && plans.meta.total_pages > 1}
			<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between pt-4">
				<p class="text-sm text-slate-600 text-center sm:text-left">
					Showing {(plans.meta.page - 1) * plans.meta.limit + 1} - {Math.min(
						plans.meta.page * plans.meta.limit,
						plans.meta.total
					)} of {plans.meta.total} plans
				</p>
				<div class="flex items-center justify-center gap-2">
					<Button
						variant="outline"
						size="sm"
						disabled={plans.meta.page === 1}
						onclick={() => currentPage--}
					>
						<ChevronLeft class="mr-1 h-4 w-4" />
						<span class="hidden sm:inline">Previous</span>
					</Button>
					<span class="px-2 text-sm text-slate-600">
						Page {plans.meta.page} of {plans.meta.total_pages}
					</span>
					<Button
						variant="outline"
						size="sm"
						disabled={plans.meta.page === plans.meta.total_pages}
						onclick={() => currentPage++}
					>
						<span class="hidden sm:inline">Next</span>
						<ChevronRight class="ml-1 h-4 w-4" />
					</Button>
				</div>
			</div>
		{/if}
	{/if}
</div>

<!-- Bulk Delete Confirmation Dialog -->
<ConfirmationDialog
	bind:open={confirmDeleteOpen}
	title="Delete {confirmDeleteCount} Plan{confirmDeleteCount === 1 ? '' : 's'}?"
	description="This action cannot be undone. These plans will be permanently removed from the system."
	confirmLabel="Delete"
	cancelLabel="Cancel"
	confirmVariant="destructive"
	onConfirm={confirmBulkDelete}
	onCancel={() => confirmDeleteOpen = false}
/>

<!-- Export Modal -->
<ExportModal
	bind:open={exportModalOpen}
	selectedPlanIds={Array.from(selectedIds)}
	currentFilters={{
		search: debouncedSearch,
		status: statusFilter,
		type: typeFilter,
		style: styleFilter
	}}
	onClose={() => (exportModalOpen = false)}
/>
