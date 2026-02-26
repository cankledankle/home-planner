<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from 'svelte-sonner';
	import { createPlan } from '$lib/api';
	import { ArrowLeft, Save, Loader2, Home } from '@lucide/svelte';

	// Form state
	let name = $state('');
	let type = $state<'single_level' | 'multi_level'>('single_level');
	let style = $state<'cabin' | 'lodge' | 'modern' | 'ranch' | 'farmhouse'>('cabin');
	let beds = $state<number | undefined>(undefined);
	let baths = $state<number | undefined>(undefined);
	let halfBaths = $state<number | undefined>(undefined);
	let mainSF = $state<number | undefined>(undefined);
	let upperSF = $state<number | undefined>(undefined);
	let lowerSF = $state<number | undefined>(undefined);
	let totalSF = $state<number | undefined>(undefined);
	let notes = $state('');

	let saving = $state(false);
	let errors = $state<Record<string, string>>({});

	const planTypes = [
		{ value: 'single_level', label: 'Single Level' },
		{ value: 'multi_level', label: 'Multi Level' }
	] as const;

	const planStyles = [
		{ value: 'cabin', label: 'Cabin' },
		{ value: 'lodge', label: 'Lodge' },
		{ value: 'modern', label: 'Modern' },
		{ value: 'ranch', label: 'Ranch' },
		{ value: 'farmhouse', label: 'Farmhouse' }
	] as const;

	function validate(): boolean {
		errors = {};
		if (!name.trim()) {
			errors.name = 'Plan name is required';
		}
		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validate()) return;

		saving = true;
		try {
			const plan = await createPlan({
				name: name.trim(),
				type,
				style,
				beds: beds || undefined,
				baths: baths || undefined,
				half_baths: halfBaths || undefined,
				main_sf: mainSF || undefined,
				upper_sf: upperSF || undefined,
				lower_sf: lowerSF || undefined,
				total_sf: totalSF || undefined,
				notes: notes.trim() || undefined
			});
			toast.success('Plan created successfully');
			goto(`/plans/${plan.id}`);
		} catch (err) {
			toast.error('Failed to create plan');
			console.error(err);
		} finally {
			saving = false;
		}
	}

	function handleCancel() {
		goto('/plans');
	}
</script>

<svelte:head>
	<title>New Home Plan - Natural Element Homes</title>
</svelte:head>

<div class="container mx-auto max-w-4xl p-6">
	<!-- Header -->
	<div class="mb-6 flex items-center justify-between">
		<div class="flex items-center gap-4">
			<Button variant="outline" size="icon" onclick={handleCancel}>
				<ArrowLeft class="h-4 w-4" />
			</Button>
			<div>
				<h1 class="text-2xl font-bold text-slate-900">New Home Plan</h1>
				<p class="text-sm text-slate-600">Create a new home plan</p>
			</div>
		</div>
		<Button onclick={handleSubmit} disabled={saving}>
			{#if saving}
				<Loader2 class="mr-2 h-4 w-4 animate-spin" />
				Saving...
			{:else}
				<Save class="mr-2 h-4 w-4" />
				Create Plan
			{/if}
		</Button>
	</div>

	<!-- Form -->
	<div class="space-y-6">
		<!-- Basic Information -->
		<div class="rounded-lg border border-slate-200 bg-white p-6">
			<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold text-slate-900">
				<Home class="h-5 w-5 text-blue-500" />
				Basic Information
			</h2>
			<div class="grid gap-6 md:grid-cols-2">
				<div class="md:col-span-2">
					<span class="mb-2 block text-sm font-medium text-slate-700">
						Plan Name <span class="text-red-500">*</span>
					</span>
					<Input
						type="text"
						bind:value={name}
						placeholder="Enter plan name (e.g., Mountain Retreat)"
						class={errors.name ? 'border-red-500' : ''}
					/>
					{#if errors.name}
						<p class="mt-1 text-sm text-red-500">{errors.name}</p>
					{/if}
				</div>

				<div>
					<span class="mb-2 block text-sm font-medium text-slate-700">Plan Type</span>
					<div class="flex flex-wrap gap-2">
						{#each planTypes as pt}
							<button
								type="button"
								class="rounded-full px-4 py-2 text-sm font-medium transition-colors {type ===
								pt.value
									? 'bg-blue-500 text-white'
									: 'border border-slate-200 bg-white text-slate-700 hover:bg-slate-50'}"
								onclick={() => (type = pt.value)}
							>
								{pt.label}
							</button>
						{/each}
					</div>
				</div>

				<div>
					<span class="mb-2 block text-sm font-medium text-slate-700">Style</span>
					<div class="flex flex-wrap gap-2">
						{#each planStyles as ps}
							<button
								type="button"
								class="rounded-full px-4 py-2 text-sm font-medium transition-colors {style ===
								ps.value
									? 'bg-emerald-500 text-white'
									: 'border border-slate-200 bg-white text-slate-700 hover:bg-slate-50'}"
								onclick={() => (style = ps.value)}
							>
								{ps.label}
							</button>
						{/each}
					</div>
				</div>
			</div>
		</div>

		<!-- Room Counts -->
		<div class="rounded-lg border border-slate-200 bg-white p-6">
			<h2 class="mb-4 text-lg font-semibold text-slate-900">Room Counts</h2>
			<div class="grid gap-6 sm:grid-cols-3">
				<div>
					<span class="mb-2 block text-sm font-medium text-slate-700">Bedrooms</span>
					<Input
						type="number"
						bind:value={beds}
						placeholder="e.g., 3"
						min="0"
					/>
				</div>
				<div>
					<span class="mb-2 block text-sm font-medium text-slate-700">Full Baths</span>
					<Input
						type="number"
						bind:value={baths}
						placeholder="e.g., 2"
						min="0"
					/>
				</div>
				<div>
					<span class="mb-2 block text-sm font-medium text-slate-700">Half Baths</span>
					<Input
						type="number"
						bind:value={halfBaths}
						placeholder="e.g., 1"
						min="0"
					/>
				</div>
			</div>
		</div>

		<!-- Square Footage -->
		<div class="rounded-lg border border-slate-200 bg-white p-6">
			<h2 class="mb-4 text-lg font-semibold text-slate-900">Square Footage</h2>
			<div class="grid gap-6 sm:grid-cols-2 md:grid-cols-4">
				<div>
					<span class="mb-2 block text-sm font-medium text-slate-700">Main Floor</span>
					<Input
						type="number"
						bind:value={mainSF}
						placeholder="e.g., 1500"
						min="0"
					/>
				</div>
				<div>
					<span class="mb-2 block text-sm font-medium text-slate-700">Upper Floor</span>
					<Input
						type="number"
						bind:value={upperSF}
						placeholder="e.g., 800"
						min="0"
					/>
				</div>
				<div>
					<span class="mb-2 block text-sm font-medium text-slate-700">Lower Floor</span>
					<Input
						type="number"
						bind:value={lowerSF}
						placeholder="e.g., 600"
						min="0"
					/>
				</div>
				<div>
					<span class="mb-2 block text-sm font-medium text-slate-700">Total</span>
					<Input
						type="number"
						bind:value={totalSF}
						placeholder="e.g., 2900"
						min="0"
					/>
				</div>
			</div>
		</div>

		<!-- Notes -->
		<div class="rounded-lg border border-slate-200 bg-white p-6">
			<h2 class="mb-4 text-lg font-semibold text-slate-900">Notes</h2>
			<textarea
				bind:value={notes}
				placeholder="Add any notes about this plan..."
				class="min-h-[120px] w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
			></textarea>
		</div>

		<!-- Actions -->
		<div class="flex justify-end gap-3">
			<Button variant="outline" onclick={handleCancel}>Cancel</Button>
			<Button onclick={handleSubmit} disabled={saving}>
				{#if saving}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
					Creating...
				{:else}
					<Save class="mr-2 h-4 w-4" />
					Create Plan
				{/if}
			</Button>
		</div>
	</div>
</div>
