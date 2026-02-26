<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import {
		Dialog,
		DialogContent,
		DialogHeader,
		DialogTitle,
		DialogDescription,
		DialogFooter
	} from '$lib/components/ui/dialog';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import {
		exportData,
		type ExportOptions,
		type ExportType,
		type ExportPreset,
		type ExportScope
	} from '$lib/api';
	import { toast } from 'svelte-sonner';
	import {
		FileSpreadsheet,
		Archive,
		Download,
		Loader2,
		Check,
		Filter,
		List,
		Layers
	} from '@lucide/svelte';

	interface Props {
		open: boolean;
		selectedPlanIds?: string[];
		currentFilters?: {
			search?: string;
			status?: string;
			type?: string;
			style?: string;
		};
		onClose: () => void;
	}

	let { open = $bindable(), selectedPlanIds = [], currentFilters = {}, onClose }: Props = $props();

	// Export configuration
	let exportType: ExportType = $state('csv');
	let exportPreset: ExportPreset = $state('general');
	let exportScope: ExportScope = $state('all');
	let selectedFields: string[] = $state([]);
	let selectedCategories: string[] = $state(['website', 'reference', 'technical', '3d', 'other']);
	let exporting = $state(false);

	const exportPresets: { id: ExportPreset; name: string; description: string; fields: string[] }[] =
		[
			{
				id: 'wp-all-import',
				name: 'WP All Import',
				description: 'Optimized for WordPress All Import plugin',
				fields: [
					'id',
					'name',
					'slug',
					'status',
					'type',
					'style',
					'beds',
					'baths',
					'heated_sf',
					'poster_url'
				]
			},
			{
				id: 'general',
				name: 'General',
				description: 'Standard export with all metadata',
				fields: [
					'id',
					'name',
					'slug',
					'status',
					'type',
					'style',
					'beds',
					'baths',
					'heated_sf',
					'total_sf',
					'notes'
				]
			},
			{
				id: 'minimal',
				name: 'Minimal',
				description: 'Basic plan info only',
				fields: ['id', 'name', 'slug', 'status']
			}
		];

	const allFields = [
		{ name: 'id', label: 'ID' },
		{ name: 'name', label: 'Name' },
		{ name: 'slug', label: 'Slug' },
		{ name: 'status', label: 'Status' },
		{ name: 'type', label: 'Type' },
		{ name: 'style', label: 'Style' },
		{ name: 'beds', label: 'Beds' },
		{ name: 'baths', label: 'Baths' },
		{ name: 'half_baths', label: 'Half Baths' },
		{ name: 'main_sf', label: 'Main SF' },
		{ name: 'upper_sf', label: 'Upper SF' },
		{ name: 'lower_sf', label: 'Lower SF' },
		{ name: 'porch_deck_sf', label: 'Porch/Deck SF' },
		{ name: 'garage_sf', label: 'Garage SF' },
		{ name: 'heated_sf', label: 'Heated SF' },
		{ name: 'total_sf', label: 'Total SF' },
		{ name: 'notes', label: 'Notes' },
		{ name: 'poster_url', label: 'Poster URL' },
		{ name: 'render_front_url', label: 'Render URL' }
	];

	const categories = [
		{ id: 'website', label: 'Website Images', description: 'Renders, elevations, floor plans' },
		{ id: 'reference', label: 'Reference Files', description: 'Client materials and references' },
		{
			id: 'technical',
			label: 'Technical Drawings',
			description: 'Construction and detail drawings'
		},
		{ id: '3d', label: '3D Assets', description: 'Models and 3D files' },
		{ id: 'other', label: 'Other Files', description: 'Miscellaneous files' }
	];

	function handlePresetChange(presetId: ExportPreset) {
		exportPreset = presetId;
		if (presetId !== 'custom') {
			const preset = exportPresets.find((p) => p.id === presetId);
			if (preset) {
				selectedFields = [...preset.fields];
			}
		}
	}

	function toggleField(fieldName: string) {
		if (selectedFields.includes(fieldName)) {
			selectedFields = selectedFields.filter((f) => f !== fieldName);
		} else {
			selectedFields = [...selectedFields, fieldName];
		}
		exportPreset = 'custom';
	}

	function toggleCategory(categoryId: string) {
		if (selectedCategories.includes(categoryId)) {
			selectedCategories = selectedCategories.filter((c) => c !== categoryId);
		} else {
			selectedCategories = [...selectedCategories, categoryId];
		}
	}

	async function handleExport() {
		if (exportType === 'csv' && selectedFields.length === 0) {
			toast.error('Please select at least one field to export');
			return;
		}
		if (exportType === 'zip' && selectedCategories.length === 0) {
			toast.error('Please select at least one file category');
			return;
		}

		exporting = true;
		try {
			const options: ExportOptions = {
				type: exportType,
				scope: exportScope,
				...(exportType === 'csv' && {
					preset: exportPreset,
					fields: selectedFields
				}),
				...(exportType === 'zip' && {
					categories: selectedCategories
				}),
				...(exportScope === 'selected' && {
					plan_ids: selectedPlanIds
				}),
				...(exportScope === 'filtered' && {
					filters: currentFilters
				})
			};

			const blob = await exportData(options);

			// Create download link
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `plans-export-${new Date().toISOString().split('T')[0]}.${exportType === 'csv' ? 'csv' : 'zip'}`;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			window.URL.revokeObjectURL(url);

			toast.success(`Export downloaded successfully`);
			onClose();
		} catch (err) {
			toast.error('Export failed: ' + (err instanceof Error ? err.message : 'Unknown error'));
		} finally {
			exporting = false;
		}
	}

	function getScopeLabel() {
		switch (exportScope) {
			case 'all':
				return 'All Plans';
			case 'selected':
				return `${selectedPlanIds.length} Selected Plan${selectedPlanIds.length === 1 ? '' : 's'}`;
			case 'filtered':
				return 'Filtered Results';
		}
	}

	// Initialize selected fields when opening
	$effect(() => {
		if (open) {
			const preset = exportPresets.find((p) => p.id === exportPreset);
			if (preset && exportPreset !== 'custom') {
				selectedFields = [...preset.fields];
			}
			// Default scope based on selection
			if (selectedPlanIds.length > 0) {
				exportScope = 'selected';
			} else {
				exportScope = 'all';
			}
		}
	});
</script>

<Dialog bind:open onOpenChange={(v) => !v && onClose()}>
	<DialogContent class="max-h-[90vh] max-w-2xl overflow-y-auto">
		<DialogHeader>
			<DialogTitle class="flex items-center gap-2">
				{#if exportType === 'csv'}
					<FileSpreadsheet class="h-5 w-5 text-emerald-600" />
				{:else}
					<Archive class="h-5 w-5 text-blue-600" />
				{/if}
				Export Plans
			</DialogTitle>
			<DialogDescription>
				Export your plans as CSV or download all associated files as ZIP
			</DialogDescription>
		</DialogHeader>

		<div class="space-y-6 py-4">
			<!-- Export Type Selection -->
			<div>
				<label class="mb-3 block text-sm font-medium text-slate-700">Export Type</label>
				<div class="grid grid-cols-2 gap-3">
					<button
						class="flex items-center gap-3 rounded-lg border p-4 text-left transition-colors {exportType ===
						'csv'
							? 'border-emerald-500 bg-emerald-50'
							: 'border-slate-200 hover:border-slate-300'}"
						onclick={() => (exportType = 'csv')}
					>
						<FileSpreadsheet
							class="h-5 w-5 {exportType === 'csv' ? 'text-emerald-600' : 'text-slate-400'}"
						/>
						<div>
							<p class="font-medium {exportType === 'csv' ? 'text-emerald-900' : 'text-slate-700'}">
								CSV Export
							</p>
							<p class="text-xs {exportType === 'csv' ? 'text-emerald-700' : 'text-slate-500'}">
								Spreadsheet with plan data
							</p>
						</div>
					</button>
					<button
						class="flex items-center gap-3 rounded-lg border p-4 text-left transition-colors {exportType ===
						'zip'
							? 'border-blue-500 bg-blue-50'
							: 'border-slate-200 hover:border-slate-300'}"
						onclick={() => (exportType = 'zip')}
					>
						<Archive class="h-5 w-5 {exportType === 'zip' ? 'text-blue-600' : 'text-slate-400'}" />
						<div>
							<p class="font-medium {exportType === 'zip' ? 'text-blue-900' : 'text-slate-700'}">
								ZIP Export
							</p>
							<p class="text-xs {exportType === 'zip' ? 'text-blue-700' : 'text-slate-500'}">
								Download all files
							</p>
						</div>
					</button>
				</div>
			</div>

			<!-- Scope Selection -->
			<div>
				<label class="mb-3 block text-sm font-medium text-slate-700">Export Scope</label>
				<div class="space-y-2">
					<button
						class="flex w-full items-center gap-3 rounded-lg border p-3 text-left transition-colors {exportScope ===
						'all'
							? 'border-blue-500 bg-blue-50'
							: 'border-slate-200 hover:border-slate-300'}"
						onclick={() => (exportScope = 'all')}
					>
						<Layers class="h-4 w-4 {exportScope === 'all' ? 'text-blue-600' : 'text-slate-400'}" />
						<div class="flex-1">
							<p class="font-medium {exportScope === 'all' ? 'text-blue-900' : 'text-slate-700'}">
								All Plans
							</p>
							<p class="text-xs {exportScope === 'all' ? 'text-blue-700' : 'text-slate-500'}">
								Export your entire collection
							</p>
						</div>
						{#if exportScope === 'all'}
							<Check class="h-4 w-4 text-blue-600" />
						{/if}
					</button>

					{#if selectedPlanIds.length > 0}
						<button
							class="flex w-full items-center gap-3 rounded-lg border p-3 text-left transition-colors {exportScope ===
							'selected'
								? 'border-blue-500 bg-blue-50'
								: 'border-slate-200 hover:border-slate-300'}"
							onclick={() => (exportScope = 'selected')}
						>
							<List
								class="h-4 w-4 {exportScope === 'selected' ? 'text-blue-600' : 'text-slate-400'}"
							/>
							<div class="flex-1">
								<p
									class="font-medium {exportScope === 'selected'
										? 'text-blue-900'
										: 'text-slate-700'}"
								>
									Selected Plans ({selectedPlanIds.length})
								</p>
								<p
									class="text-xs {exportScope === 'selected' ? 'text-blue-700' : 'text-slate-500'}"
								>
									Only currently selected plans
								</p>
							</div>
							{#if exportScope === 'selected'}
								<Check class="h-4 w-4 text-blue-600" />
							{/if}
						</button>
					{/if}

					{#if currentFilters && (currentFilters.search || currentFilters.status || currentFilters.type || currentFilters.style)}
						<button
							class="flex w-full items-center gap-3 rounded-lg border p-3 text-left transition-colors {exportScope ===
							'filtered'
								? 'border-blue-500 bg-blue-50'
								: 'border-slate-200 hover:border-slate-300'}"
							onclick={() => (exportScope = 'filtered')}
						>
							<Filter
								class="h-4 w-4 {exportScope === 'filtered' ? 'text-blue-600' : 'text-slate-400'}"
							/>
							<div class="flex-1">
								<p
									class="font-medium {exportScope === 'filtered'
										? 'text-blue-900'
										: 'text-slate-700'}"
								>
									Filtered Results
								</p>
								<p
									class="text-xs {exportScope === 'filtered' ? 'text-blue-700' : 'text-slate-500'}"
								>
									Current filter and search results
								</p>
							</div>
							{#if exportScope === 'filtered'}
								<Check class="h-4 w-4 text-blue-600" />
							{/if}
						</button>
					{/if}
				</div>
			</div>

			{#if exportType === 'csv'}
				<!-- CSV Presets -->
				<div>
					<label class="mb-3 block text-sm font-medium text-slate-700">Export Preset</label>
					<div class="space-y-2">
						{#each exportPresets as preset}
							<button
								class="flex w-full items-start gap-3 rounded-lg border p-3 text-left transition-colors {exportPreset ===
								preset.id
									? 'border-emerald-500 bg-emerald-50'
									: 'border-slate-200 hover:border-slate-300'}"
								onclick={() => handlePresetChange(preset.id)}
							>
								<div class="flex-1">
									<p
										class="font-medium {exportPreset === preset.id
											? 'text-emerald-900'
											: 'text-slate-700'}"
									>
										{preset.name}
									</p>
									<p
										class="text-xs {exportPreset === preset.id
											? 'text-emerald-700'
											: 'text-slate-500'}"
									>
										{preset.description} • {preset.fields.length} fields
									</p>
								</div>
								{#if exportPreset === preset.id}
									<Check class="h-4 w-4 text-emerald-600" />
								{/if}
							</button>
						{/each}
						<button
							class="flex w-full items-start gap-3 rounded-lg border p-3 text-left transition-colors {exportPreset ===
							'custom'
								? 'border-emerald-500 bg-emerald-50'
								: 'border-slate-200 hover:border-slate-300'}"
							onclick={() => handlePresetChange('custom')}
						>
							<div class="flex-1">
								<p
									class="font-medium {exportPreset === 'custom'
										? 'text-emerald-900'
										: 'text-slate-700'}"
								>
									Custom
								</p>
								<p
									class="text-xs {exportPreset === 'custom'
										? 'text-emerald-700'
										: 'text-slate-500'}"
								>
									Select your own fields • {selectedFields.length} selected
								</p>
							</div>
							{#if exportPreset === 'custom'}
								<Check class="h-4 w-4 text-emerald-600" />
							{/if}
						</button>
					</div>
				</div>

				<!-- Custom Field Selection -->
				{#if exportPreset === 'custom'}
					<div>
						<label class="mb-3 block text-sm font-medium text-slate-700">
							Select Fields ({selectedFields.length} selected)
						</label>
						<div
							class="grid grid-cols-2 gap-2 rounded-lg border border-slate-200 p-3 sm:grid-cols-3"
						>
							{#each allFields as field}
								<label class="flex cursor-pointer items-center gap-2">
									<Checkbox
										checked={selectedFields.includes(field.name)}
										onCheckedChange={() => toggleField(field.name)}
									/>
									<span class="text-sm text-slate-700">{field.label}</span>
								</label>
							{/each}
						</div>
					</div>
				{/if}
			{/if}

			{#if exportType === 'zip'}
				<!-- ZIP Category Selection -->
				<div>
					<label class="mb-3 block text-sm font-medium text-slate-700">
						File Categories ({selectedCategories.length} selected)
					</label>
					<div class="space-y-2">
						{#each categories as category}
							<label
								class="flex cursor-pointer items-start gap-3 rounded-lg border border-slate-200 p-3 hover:bg-slate-50"
							>
								<Checkbox
									checked={selectedCategories.includes(category.id)}
									onCheckedChange={() => toggleCategory(category.id)}
									class="mt-0.5"
								/>
								<div class="flex-1">
									<p class="font-medium text-slate-700">{category.label}</p>
									<p class="text-xs text-slate-500">{category.description}</p>
								</div>
							</label>
						{/each}
					</div>
				</div>
			{/if}
		</div>

		<DialogFooter>
			<Button variant="outline" onclick={onClose} disabled={exporting}>Cancel</Button>
			<Button onclick={handleExport} disabled={exporting}>
				{#if exporting}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
					Exporting...
				{:else}
					<Download class="mr-2 h-4 w-4" />
					Download {exportType.toUpperCase()}
				{/if}
			</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>
