<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import BulkImageUpload from '$lib/components/import/BulkImageUpload.svelte';

	import { toast } from 'svelte-sonner';
	import {
		Upload,
		FileSpreadsheet,
		CheckCircle2,
		AlertCircle,
		Loader2,
		ArrowRight,
		ArrowLeft,
		RefreshCw,
		X,
		Check,
		Download
	} from '@lucide/svelte';
	import {
		previewCsvImport,
		importCsv,
		getRecentImports,
		type ImportPreview,
		type ImportResult
	} from '$lib/api';
	import type { Plan } from '$lib/types';

	// Step management
	let currentStep = $state(1);
	const totalSteps = 5;

	// Recently imported plans for bulk image upload
	let recentPlans = $state<Plan[]>([]);

	// File upload
	let selectedFile = $state<File | null>(null);
	let isDragging = $state(false);

	// Import preview data
	let preview = $state<ImportPreview | null>(null);
	let loading = $state(false);

	// Column mapping
	let columnMapping = $state<Record<string, string>>({});
	let importMode = $state<'create' | 'update' | 'upsert'>('create');

	// Import result
	let importResult = $state<ImportResult | null>(null);

	// Available plan fields for mapping
	const planFields = [
		{ value: 'name', label: 'Plan Name', required: true },
		{ value: 'type', label: 'Type' },
		{ value: 'style', label: 'Style' },
		{ value: 'beds', label: 'Bedrooms' },
		{ value: 'baths', label: 'Bathrooms' },
		{ value: 'half_baths', label: 'Half Baths' },
		{ value: 'main_sf', label: 'Main Floor SF' },
		{ value: 'upper_sf', label: 'Upper Floor SF' },
		{ value: 'lower_sf', label: 'Lower Floor SF' },
		{ value: 'porch_deck_sf', label: 'Porch/Deck SF' },
		{ value: 'garage_sf', label: 'Garage SF' },
		{ value: 'garage_apartment_sf', label: 'Garage Apartment SF' },
		{ value: 'unfinished_sf', label: 'Unfinished SF' },
		{ value: 'heated_sf', label: 'Heated SF' },
		{ value: 'total_sf', label: 'Total SF' },
		{ value: 'notes', label: 'Notes' }
	];

	function handleFileSelect(event: Event) {
		const input = event.target as HTMLInputElement;
		if (input.files && input.files[0]) {
			selectedFile = input.files[0];
		}
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		isDragging = false;
		if (event.dataTransfer?.files && event.dataTransfer.files[0]) {
			const file = event.dataTransfer.files[0];
			if (file.name.endsWith('.csv')) {
				selectedFile = file;
			} else {
				toast.error('Please upload a CSV file');
			}
		}
	}

	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		isDragging = true;
	}

	function handleDragLeave(event: DragEvent) {
		event.preventDefault();
		isDragging = false;
	}

	async function handleUpload() {
		if (!selectedFile) {
			toast.error('Please select a file');
			return;
		}

		loading = true;
		try {
			preview = await previewCsvImport(selectedFile);
			// Initialize mapping with suggestions
			columnMapping = { ...preview.suggested_mapping };
			currentStep = 2;
		} catch (err) {
			toast.error('Failed to preview CSV file');
		} finally {
			loading = false;
		}
	}

	function handleMapping() {
		// Validate that name is mapped
		const hasNameMapping = Object.values(columnMapping).includes('name');
		if (!hasNameMapping) {
			toast.error('Please map at least the Plan Name column');
			return;
		}
		currentStep = 3;
	}

	async function handleImport() {
		if (!selectedFile) return;

		loading = true;
		try {
			// Convert mapping to API format
			const mapping: Record<string, string | null> = {};
			preview?.columns.forEach((col) => {
				mapping[col] = columnMapping[col] || null;
			});

			importResult = await importCsv(selectedFile, {
				mode: importMode,
				mapping
			});

			// Fetch recently imported plans for bulk image upload
			try {
				recentPlans = await getRecentImports();
			} catch (err) {
				console.error('Failed to fetch recent imports:', err);
				recentPlans = [];
			}

			currentStep = 4;
		} catch (err) {
			toast.error('Import failed');
		} finally {
			loading = false;
		}
	}

	function resetImport() {
		currentStep = 1;
		selectedFile = null;
		preview = null;
		columnMapping = {};
		importResult = null;
	}

	function goToStep(step: number) {
		if (step < currentStep) {
			currentStep = step;
		}
	}

	function downloadExampleCsv() {
		// Example plan data
		const exampleData = [
			{
				name: 'Abilene',
				type: 'single_level',
				style: 'cabin',
				beds: 3,
				baths: 2,
				half_baths: 1,
				main_sf: 1200,
				upper_sf: 0,
				lower_sf: 0,
				porch_deck_sf: 400,
				garage_sf: 480,
				garage_apartment_sf: 0,
				unfinished_sf: 0,
				heated_sf: 1200,
				total_sf: 1680,
				notes: 'Example cabin plan with front porch'
			},
			{
				name: 'Angler',
				type: 'single_level',
				style: 'cabin',
				beds: 2,
				baths: 1,
				half_baths: 0,
				main_sf: 750,
				upper_sf: 0,
				lower_sf: 0,
				porch_deck_sf: 300,
				garage_sf: 0,
				garage_apartment_sf: 0,
				unfinished_sf: 0,
				heated_sf: 750,
				total_sf: 1050,
				notes: 'Compact cabin with screened porch'
			},
			{
				name: 'Arrowhead Lodge',
				type: 'multi_level',
				style: 'lodge',
				beds: 4,
				baths: 3,
				half_baths: 1,
				main_sf: 1800,
				upper_sf: 800,
				lower_sf: 600,
				porch_deck_sf: 600,
				garage_sf: 600,
				garage_apartment_sf: 450,
				unfinished_sf: 200,
				heated_sf: 2600,
				total_sf: 3400,
				notes: 'Spacious lodge with finished basement and garage apartment'
			}
		];

		// Create CSV content
		const headers = [
			'name',
			'type',
			'style',
			'beds',
			'baths',
			'half_baths',
			'main_sf',
			'upper_sf',
			'lower_sf',
			'porch_deck_sf',
			'garage_sf',
			'garage_apartment_sf',
			'unfinished_sf',
			'heated_sf',
			'total_sf',
			'notes'
		];

		const csvRows = [
			headers.join(','),
			...exampleData.map((row) =>
				headers
					.map((header) => {
						const value = row[header as keyof typeof row];
						if (typeof value === 'string' && value.includes(',')) {
							return `"${value.replace(/"/g, '""')}"`;
						}
						return value ?? '';
					})
					.join(',')
			)
		];

		const csvContent = '\uFEFF' + csvRows.join('\n');
		const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
		const link = document.createElement('a');
		link.href = URL.createObjectURL(blob);
		link.download = 'example-plans-import.csv';
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
		URL.revokeObjectURL(link.href);

		toast.success('Example CSV downloaded');
	}
</script>

<div class="mx-auto max-w-4xl space-y-6">
	<!-- Header -->
	<div class="text-center">
		<h1 class="text-3xl font-bold text-card-foreground">Import Plans</h1>
		<p class="mt-2 text-muted-foreground">Import home plans from a CSV file</p>
	</div>

	<!-- Step Indicator -->
	<div class="relative">
		<div class="absolute top-5 left-0 h-0.5 w-full -translate-y-1/2 bg-border"></div>
		<div class="relative flex justify-between">
			{#each [1, 2, 3, 4, 5] as step}
				<button
					class="flex h-10 w-10 items-center justify-center rounded-full border-2 text-sm font-medium transition-colors {currentStep >
					step
						? 'border-green-500 bg-green-500 text-white'
						: currentStep === step
							? 'border-blue-600 bg-blue-600 text-white'
							: 'border-input bg-card text-muted-foreground'}"
					onclick={() => goToStep(step)}
					disabled={step > currentStep}
				>
					{#if currentStep > step}
						<Check class="h-5 w-5" />
					{:else}
						{step}
					{/if}
				</button>
			{/each}
		</div>
		<div class="mt-2 flex justify-between text-xs text-muted-foreground">
			<span class="w-10 text-center sm:w-auto">Upload</span>
			<span class="w-10 text-center sm:w-auto">Mapping</span>
			<span class="w-10 text-center sm:w-auto">Review</span>
			<span class="w-10 text-center sm:w-auto">Results</span>
			<span class="w-10 text-center sm:w-auto">Images</span>
		</div>
	</div>

	<!-- Step Content -->
	<div class="rounded-lg border border-border bg-card p-6">
		{#if currentStep === 1}
			<!-- Step 1: Upload -->
			<div class="space-y-6">
				<h2 class="text-xl font-semibold text-card-foreground">Upload CSV File</h2>

				<div
					class="rounded-lg border-2 border-dashed p-8 text-center transition-colors {isDragging
						? 'border-blue-500 bg-blue-500/10'
						: selectedFile
							? 'border-green-500 bg-green-500/10'
							: 'border-input hover:border-muted-foreground'}"
					ondrop={handleDrop}
					ondragover={handleDragOver}
					ondragleave={handleDragLeave}
					role="button"
					tabindex="0"
				>
					{#if selectedFile}
						<div class="flex flex-col items-center gap-3">
							<div class="rounded-full bg-green-500/20 p-3">
								<FileSpreadsheet class="h-6 w-6 text-green-600" />
							</div>
							<div>
								<p class="font-medium text-card-foreground">{selectedFile.name}</p>
								<p class="text-sm text-muted-foreground">
									{(selectedFile.size / 1024).toFixed(1)} KB
								</p>
							</div>
							<Button variant="outline" size="sm" onclick={() => (selectedFile = null)}>
								<X class="mr-2 h-4 w-4" />
								Remove
							</Button>
						</div>
					{:else}
						<div class="flex flex-col items-center gap-3">
							<div class="rounded-full bg-blue-500/20 p-3">
								<Upload class="h-6 w-6 text-blue-600" />
							</div>
							<div>
								<p class="font-medium text-card-foreground">Drop your CSV file here</p>
								<p class="text-sm text-muted-foreground">or click to browse</p>
							</div>
							<Button
								variant="outline"
								onclick={() => document.getElementById('csv-upload')?.click()}
							>
								Select File
							</Button>
							<input
								id="csv-upload"
								type="file"
								accept=".csv"
								class="hidden"
								onchange={handleFileSelect}
							/>
						</div>
					{/if}
				</div>

				<div class="rounded-lg bg-muted p-4">
					<div class="flex flex-col gap-4">
						<div>
							<h3 class="mb-2 font-medium text-card-foreground">CSV Format Requirements</h3>
							<ul class="space-y-1 text-sm text-muted-foreground">
								<li>• First row must contain column headers</li>
								<li>• Required column: Plan Name</li>
								<li>• Optional columns: Type, Style, Beds, Baths, Heated SF, etc.</li>
								<li>• Maximum file size: 10MB</li>
							</ul>
						</div>
						<div class="flex-shrink-0">
							<Button
								variant="outline"
								size="sm"
								onclick={downloadExampleCsv}
								class="w-full sm:w-auto"
							>
								<Download class="mr-2 h-4 w-4" />
								Download Example CSV
							</Button>
						</div>
					</div>
				</div>

				<div class="flex justify-end">
					<Button onclick={handleUpload} disabled={!selectedFile || loading}>
						{#if loading}
							<Loader2 class="mr-2 h-4 w-4 animate-spin" />
							Analyzing...
						{:else}
							Continue
							<ArrowRight class="ml-2 h-4 w-4" />
						{/if}
					</Button>
				</div>
			</div>
		{:else if currentStep === 2}
			<!-- Step 2: Column Mapping -->
			<div class="space-y-6">
				<div class="flex items-center justify-between">
					<h2 class="text-xl font-semibold text-card-foreground">Map Columns</h2>
					<div class="text-sm text-muted-foreground">
						{preview?.total_rows ?? 0} rows found
					</div>
				</div>

				<!-- Import Mode -->
				<div class="space-y-2">
					<span class="text-sm font-medium">Import Mode</span>
					<div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
						<label
							class="flex cursor-pointer items-center gap-2 rounded-lg border border-border p-3 transition-colors hover:bg-muted {importMode ===
							'create'
								? 'border-blue-500 bg-blue-50'
								: ''}"
						>
							<input type="radio" bind:group={importMode} value="create" class="h-4 w-4" />
							<div>
								<p class="font-medium text-card-foreground">Create New</p>
								<p class="text-xs text-muted-foreground">Only create new plans</p>
							</div>
						</label>
						<label
							class="flex cursor-pointer items-center gap-2 rounded-lg border border-border p-3 transition-colors hover:bg-muted {importMode ===
							'update'
								? 'border-blue-500 bg-blue-50'
								: ''}"
						>
							<input type="radio" bind:group={importMode} value="update" class="h-4 w-4" />
							<div>
								<p class="font-medium text-card-foreground">Update Existing</p>
								<p class="text-xs text-muted-foreground">Only update existing plans</p>
							</div>
						</label>
						<label
							class="flex cursor-pointer items-center gap-2 rounded-lg border border-border p-3 transition-colors hover:bg-muted {importMode ===
							'upsert'
								? 'border-blue-500 bg-blue-50'
								: ''}"
						>
							<input type="radio" bind:group={importMode} value="upsert" class="h-4 w-4" />
							<div>
								<p class="font-medium text-card-foreground">Create or Update</p>
								<p class="text-xs text-muted-foreground">Create new or update existing</p>
							</div>
						</label>
					</div>
				</div>

				<!-- Column Mapping -->
				<div class="space-y-4">
					<h3 class="font-medium text-card-foreground">Column Mapping</h3>
					<p class="text-sm text-muted-foreground">
						Match your CSV columns to plan fields. At minimum, Plan Name is required.
					</p>

					<div class="space-y-3">
						{#each preview?.columns ?? [] as column}
							<div
								class="flex flex-col gap-3 rounded-lg border border-border p-4 sm:flex-row sm:items-center sm:gap-4"
							>
								<div class="min-w-0 flex-1">
									<p class="truncate font-medium text-card-foreground">{column}</p>
									<p class="truncate text-sm text-muted-foreground">
										Example: {preview?.preview[0]?.[column] ?? 'N/A'}
									</p>
								</div>
								<div class="flex flex-shrink-0 items-center gap-2">
									<span class="hidden text-muted-foreground sm:inline">→</span>
									<select
										bind:value={columnMapping[column]}
										class="w-full rounded-md border border-input px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none sm:w-auto sm:min-w-[200px]"
									>
										<option value="">— Skip this column —</option>
										{#each planFields as field}
											<option value={field.value}>
												{field.label}
												{#if field.required}(required){/if}
											</option>
										{/each}
									</select>
									{#if columnMapping[column] === 'name'}
										<span class="flex-shrink-0 rounded-full bg-green-100 p-1">
											<Check class="h-4 w-4 text-green-600" />
										</span>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				</div>

				<div class="flex justify-between">
					<Button variant="outline" onclick={() => (currentStep = 1)}>
						<ArrowLeft class="mr-2 h-4 w-4" />
						Back
					</Button>
					<Button onclick={handleMapping}>
						Continue
						<ArrowRight class="ml-2 h-4 w-4" />
					</Button>
				</div>
			</div>
		{:else if currentStep === 3}
			<!-- Step 3: Review -->
			<div class="space-y-6">
				<h2 class="text-xl font-semibold text-card-foreground">Review Import</h2>

				<div class="space-y-4">
					<div class="rounded-lg bg-muted p-4">
						<h3 class="mb-2 font-medium text-card-foreground">Import Summary</h3>
						<div class="grid grid-cols-2 gap-4 text-sm">
							<div>
								<span class="text-muted-foreground">Mode:</span>
								<span class="ml-2 font-medium capitalize">{importMode}</span>
							</div>
							<div>
								<span class="text-muted-foreground">Total Rows:</span>
								<span class="ml-2 font-medium">{preview?.total_rows}</span>
							</div>
						</div>
					</div>

					<div class="rounded-lg bg-muted p-4">
						<h3 class="mb-2 font-medium text-card-foreground">Column Mapping</h3>
						<div class="space-y-1 text-sm">
							{#each Object.entries(columnMapping) as [csvCol, planField]}
								{#if planField}
									<div class="flex items-center gap-2">
										<span class="text-muted-foreground">{csvCol}</span>
										<span class="text-muted-foreground">→</span>
										<span class="font-medium text-card-foreground">
											{planFields.find((f) => f.value === planField)?.label ?? planField}
										</span>
									</div>
								{/if}
							{/each}
						</div>
					</div>

					<div class="rounded-lg bg-blue-50 p-4">
						<h3 class="mb-2 font-medium text-blue-900">Preview (First 3 Rows)</h3>
						<div class="overflow-x-auto">
							<table class="min-w-full text-sm">
								<thead>
									<tr class="border-b border-blue-200">
										{#each Object.values(columnMapping).filter(Boolean) as field}
											<th class="px-2 py-1 text-left font-medium text-blue-700">
												{planFields.find((f) => f.value === field)?.label ?? field}
											</th>
										{/each}
									</tr>
								</thead>
								<tbody>
									{#each preview?.preview.slice(0, 3) ?? [] as row}
										<tr class="border-b border-blue-100 last:border-0">
											{#each Object.entries(columnMapping) as [csvCol, planField]}
												{#if planField}
													<td class="px-2 py-1 text-blue-900">{row[csvCol] ?? ''}</td>
												{/if}
											{/each}
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>
				</div>

				<div class="flex justify-between">
					<Button variant="outline" onclick={() => (currentStep = 2)}>
						<ArrowLeft class="mr-2 h-4 w-4" />
						Back
					</Button>
					<Button onclick={handleImport} disabled={loading}>
						{#if loading}
							<Loader2 class="mr-2 h-4 w-4 animate-spin" />
							Importing...
						{:else}
							<CheckCircle2 class="mr-2 h-4 w-4" />
							Confirm Import
						{/if}
					</Button>
				</div>
			</div>
		{:else if currentStep === 4}
			<!-- Step 4: Results -->
			<div class="space-y-6 text-center">
				{#if importResult}
					{#if importResult.errors.length === 0}
						<div
							class="mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-green-100"
						>
							<CheckCircle2 class="h-10 w-10 text-green-600" />
						</div>
					{:else}
						<div
							class="mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-amber-100"
						>
							<AlertCircle class="h-10 w-10 text-amber-600" />
						</div>
					{/if}

					<h2 class="text-xl font-semibold text-card-foreground">Import Complete</h2>

					<div class="grid grid-cols-2 gap-4">
						<div class="rounded-lg bg-green-50 p-4">
							<p class="text-sm text-green-600">Created</p>
							<p class="text-2xl font-bold text-green-700">{importResult.imported}</p>
						</div>
						<div class="rounded-lg bg-blue-50 p-4">
							<p class="text-sm text-blue-600">Updated</p>
							<p class="text-2xl font-bold text-blue-700">{importResult.updated}</p>
						</div>
					</div>

					{#if importResult.errors.length > 0}
						<div class="rounded-lg border border-red-200 bg-red-50 p-4 text-left">
							<h3 class="mb-2 font-medium text-red-900">Errors ({importResult.errors.length})</h3>
							<div class="max-h-48 space-y-1 overflow-y-auto text-sm">
								{#each importResult.errors as error}
									<div class="flex items-center gap-2 text-red-700">
										<AlertCircle class="h-4 w-4" />
										<span>Row {error.row}: {error.message}</span>
									</div>
								{/each}
							</div>
						</div>
					{/if}

					<div class="flex justify-center gap-3">
						<Button variant="outline" onclick={() => (window.location.href = '/plans')}>
							View Plans
						</Button>
						{#if recentPlans.length > 0}
							<Button onclick={() => (currentStep = 5)}>
								Upload Images
								<ArrowRight class="ml-2 h-4 w-4" />
							</Button>
						{:else}
							<Button onclick={resetImport}>
								<RefreshCw class="mr-2 h-4 w-4" />
								Import Another File
							</Button>
						{/if}
					</div>
				{/if}
			</div>
		{:else if currentStep === 5}
			<!-- Step 5: Bulk Image Upload -->
			{#if recentPlans.length > 0}
				<BulkImageUpload
					plans={recentPlans}
					onComplete={() => {
						toast.success('Import process complete!');
						resetImport();
					}}
					onBack={() => (currentStep = 4)}
					onSkip={() => {
						toast.success('Skipped image upload');
						resetImport();
					}}
				/>
			{:else}
				<div class="space-y-6 text-center">
					<AlertCircle class="mx-auto h-12 w-12 text-amber-600" />
					<h2 class="text-xl font-semibold text-card-foreground">No Recent Plans</h2>
					<p class="text-muted-foreground">No plans were found from the recent import.</p>
					<Button onclick={resetImport}>
						<RefreshCw class="mr-2 h-4 w-4" />
						Start Over
					</Button>
				</div>
			{/if}
		{/if}
	</div>
</div>
