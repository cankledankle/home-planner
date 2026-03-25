<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { toast } from 'svelte-sonner';
	import {
		Upload,
		Image,
		CheckCircle2,
		AlertCircle,
		Loader2,
		X,
		ArrowRight,
		ArrowLeft,
		Wand2,
		RefreshCw,
		FileImage,
		HelpCircle
	} from '@lucide/svelte';
	import type { Plan } from '$lib/types';
	import { WEBSITE_SLOTS } from '$lib/api/contracts';
	import { bulkUploadFiles } from '$lib/api';

	interface Props {
		plans: Plan[];
		onComplete: () => void;
		onBack: () => void;
		onSkip: () => void;
	}

	let { plans, onComplete, onBack, onSkip }: Props = $props();

	// File upload state
	let uploadedFiles = $state<
		Array<{
			file: File;
			id: string;
			matchedPlan?: Plan;
			suggestedSlot?: string;
			assignedSlot?: string;
			confidence: number;
			status: 'pending' | 'uploading' | 'success' | 'error';
			message?: string;
		}>
	>([]);

	let isDragging = $state(false);
	let isUploading = $state(false);
	let uploadProgress = $state(0);
	let showHelp = $state(false);

	// Available slots
	const slots = [
		{ value: WEBSITE_SLOTS.renderFront, label: 'Render Front', icon: '🏠' },
		{ value: WEBSITE_SLOTS.elevationFront, label: 'Elevation Front', icon: '⬆️' },
		{ value: WEBSITE_SLOTS.elevationLeft, label: 'Elevation Left', icon: '⬅️' },
		{ value: WEBSITE_SLOTS.elevationRear, label: 'Elevation Rear', icon: '⬇️' },
		{ value: WEBSITE_SLOTS.elevationRight, label: 'Elevation Right', icon: '➡️' },
		{ value: WEBSITE_SLOTS.floorPlanMain, label: 'Floor Plan Main', icon: '📋' },
		{ value: WEBSITE_SLOTS.floorPlanUpper, label: 'Floor Plan Upper', icon: '📋' },
		{ value: WEBSITE_SLOTS.floorPlanLower, label: 'Floor Plan Lower', icon: '📋' },
		{ value: WEBSITE_SLOTS.poster, label: 'Poster', icon: '🖼️' }
	];

	// Parse filename for smart matching
	function parseFilename(filename: string): {
		planSlug?: string;
		slot?: string;
		confidence: number;
	} {
		// Remove extension
		const name = filename.replace(/\.[^/.]+$/, '').toLowerCase();

		// Pattern: {plan-slug}--{slot-type}--{view} or {plan-slug}--{slot}
		const parts = name.split('--');

		if (parts.length >= 2) {
			const planSlug = parts[0];
			const slotPart = parts[1];

			// Try to match slot
			let matchedSlot: string | undefined;
			let confidence = 0;

			// Direct slot name match
			for (const slot of slots) {
				const slotName = slot.value.toLowerCase();
				if (slotPart === slotName || slotPart.includes(slotName.replace(/-/g, ''))) {
					matchedSlot = slot.value;
					confidence = 100;
					break;
				}
			}

			// Partial match with keywords
			if (!matchedSlot) {
				const keywords: Record<string, string[]> = {
					[WEBSITE_SLOTS.renderFront]: ['render', 'front', 'main', 'hero'],
					[WEBSITE_SLOTS.elevationFront]: ['elevation', 'front', 'front-elevation'],
					[WEBSITE_SLOTS.elevationLeft]: ['elevation', 'left', 'left-elevation'],
					[WEBSITE_SLOTS.elevationRear]: [
						'elevation',
						'rear',
						'back',
						'rear-elevation',
						'back-elevation'
					],
					[WEBSITE_SLOTS.elevationRight]: ['elevation', 'right', 'right-elevation'],
					[WEBSITE_SLOTS.floorPlanMain]: ['floor', 'plan', 'main', 'main-floor'],
					[WEBSITE_SLOTS.floorPlanUpper]: [
						'floor',
						'plan',
						'upper',
						'second',
						'upper-floor',
						'second-floor'
					],
					[WEBSITE_SLOTS.floorPlanLower]: ['floor', 'plan', 'lower', 'basement', 'lower-floor'],
					[WEBSITE_SLOTS.poster]: ['poster', 'print', 'marketing']
				};

				for (const [slotValue, words] of Object.entries(keywords)) {
					const matches = words.filter((w) => slotPart.includes(w)).length;
					if (matches > 0) {
						const matchConfidence = Math.min(matches * 25, 75);
						if (matchConfidence > confidence) {
							matchedSlot = slotValue;
							confidence = matchConfidence;
						}
					}
				}
			}

			return { planSlug, slot: matchedSlot, confidence };
		}

		return { confidence: 0 };
	}

	// Find matching plan from slug
	function findPlanBySlug(slug: string): Plan | undefined {
		return plans.find(
			(p) =>
				p.slug.toLowerCase() === slug.toLowerCase() ||
				p.name.toLowerCase().replace(/\s+/g, '-') === slug.toLowerCase()
		);
	}

	// Auto-match all files
	function autoMatchFiles() {
		let matchCount = 0;
		uploadedFiles = uploadedFiles.map((fileData) => {
			if (fileData.assignedSlot) return fileData; // Skip already assigned

			const parsed = parseFilename(fileData.file.name);
			const matchedPlan = parsed.planSlug ? findPlanBySlug(parsed.planSlug) : undefined;

			if (matchedPlan && parsed.slot && parsed.confidence >= 50) {
				matchCount++;
			}

			return {
				...fileData,
				matchedPlan,
				suggestedSlot: parsed.slot,
				confidence: parsed.confidence,
				assignedSlot: parsed.confidence >= 75 ? parsed.slot : fileData.assignedSlot
			};
		});

		toast.success(`Auto-matched ${matchCount} files`);
	}

	// Handle file drop
	function handleDrop(event: DragEvent) {
		event.preventDefault();
		isDragging = false;

		if (event.dataTransfer?.files) {
			const files = Array.from(event.dataTransfer.files).filter(
				(f) => f.type.startsWith('image/') || f.name.match(/\.(jpg|jpeg|png)$/i)
			);
			addFiles(files);
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

	function handleFileSelect(event: Event) {
		const input = event.target as HTMLInputElement;
		if (input.files) {
			const files = Array.from(input.files).filter(
				(f) => f.type.startsWith('image/') || f.name.match(/\.(jpg|jpeg|png)$/i)
			);
			addFiles(files);
		}
	}

	function addFiles(files: File[]) {
		const newFiles = files.map((file) => ({
			file,
			id: Math.random().toString(36).substring(7),
			confidence: 0,
			status: 'pending' as const
		}));

		uploadedFiles = [...uploadedFiles, ...newFiles];

		// Auto-try to match new files
		setTimeout(() => {
			uploadedFiles = uploadedFiles.map((fileData) => {
				if (fileData.matchedPlan) return fileData; // Already matched

				const parsed = parseFilename(fileData.file.name);
				const matchedPlan = parsed.planSlug ? findPlanBySlug(parsed.planSlug) : undefined;

				return {
					...fileData,
					matchedPlan,
					suggestedSlot: parsed.slot,
					confidence: parsed.confidence,
					assignedSlot: parsed.confidence >= 75 ? parsed.slot : fileData.assignedSlot
				};
			});
		}, 0);
	}

	function removeFile(id: string) {
		uploadedFiles = uploadedFiles.filter((f) => f.id !== id);
	}

	function assignSlot(fileId: string, slot: string) {
		uploadedFiles = uploadedFiles.map((f) => (f.id === fileId ? { ...f, assignedSlot: slot } : f));
	}

	function assignPlan(fileId: string, planId: string) {
		const plan = plans.find((p) => p.id === planId);
		uploadedFiles = uploadedFiles.map((f) => (f.id === fileId ? { ...f, matchedPlan: plan } : f));
	}

	// Batch assign selected files to a slot
	let selectedFileIds = $state<string[]>([]);
	let batchSlot = $state('');

	function toggleFileSelection(fileId: string) {
		if (selectedFileIds.includes(fileId)) {
			selectedFileIds = selectedFileIds.filter((id) => id !== fileId);
		} else {
			selectedFileIds = [...selectedFileIds, fileId];
		}
	}

	function batchAssign() {
		if (!batchSlot || selectedFileIds.length === 0) return;

		uploadedFiles = uploadedFiles.map((f) =>
			selectedFileIds.includes(f.id) ? { ...f, assignedSlot: batchSlot } : f
		);

		selectedFileIds = [];
		batchSlot = '';
		toast.success(`Assigned ${selectedFileIds.length} files to slot`);
	}

	async function handleUpload() {
		const readyFiles = uploadedFiles.filter((f) => f.matchedPlan && f.assignedSlot);

		if (readyFiles.length === 0) {
			toast.error('Please assign plans and slots to files first');
			return;
		}

		isUploading = true;
		uploadProgress = 0;

		try {
			const files = readyFiles.map((f) => f.file);
			const metadata = readyFiles.map((f) => ({
				plan_id: f.matchedPlan!.id,
				slot: f.assignedSlot!
			}));

			// Mark files as uploading
			uploadedFiles = uploadedFiles.map((f) =>
				readyFiles.find((rf) => rf.id === f.id) ? { ...f, status: 'uploading' } : f
			);

			const result = await bulkUploadFiles(files, metadata);

			// Update file statuses based on results
			uploadedFiles = uploadedFiles.map((f) => {
				const resultItem = result.results.find(
					(r) => r.plan_id === f.matchedPlan?.id && r.slot === f.assignedSlot
				);
				if (resultItem) {
					return {
						...f,
						status: resultItem.success ? 'success' : 'error',
						message: resultItem.message
					};
				}
				return f;
			});

			const successCount = result.summary.success;
			const failedCount = result.summary.failed;

			if (failedCount === 0) {
				toast.success(`Successfully uploaded ${successCount} images`);
			} else {
				toast.error(`${successCount} uploaded, ${failedCount} failed`);
			}

			uploadProgress = 100;
		} catch (err) {
			toast.error('Bulk upload failed: ' + (err as Error).message);
			uploadedFiles = uploadedFiles.map((f) =>
				f.status === 'uploading' ? { ...f, status: 'error' } : f
			);
		} finally {
			isUploading = false;
		}
	}

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
		return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
	}

	const readyCount = $derived(uploadedFiles.filter((f) => f.matchedPlan && f.assignedSlot).length);
	const totalCount = $derived(uploadedFiles.length);
	const successCount = $derived(uploadedFiles.filter((f) => f.status === 'success').length);
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h2 class="text-xl font-semibold text-card-foreground">Upload Images</h2>
			<p class="text-sm text-muted-foreground">
				Assign images to the {plans.length} imported plans
			</p>
		</div>
		<div class="flex items-center gap-2">
			<Button variant="ghost" size="sm" onclick={() => (showHelp = !showHelp)}>
				<HelpCircle class="mr-1 h-4 w-4" />
				Help
			</Button>
		</div>
	</div>

	<!-- Help Panel -->
	{#if showHelp}
		<div class="rounded-lg bg-muted p-4 text-sm">
			<h4 class="mb-2 font-medium text-card-foreground">Smart Filename Matching</h4>
			<p class="mb-2 text-muted-foreground">
				Files are automatically matched to plans and slots based on filename patterns:
			</p>
			<ul class="list-inside list-disc space-y-1 text-muted-foreground">
				<li><code>plan-slug--render-front.jpg</code> → Render Front slot</li>
				<li><code>plan-slug--elevation-front.jpg</code> → Elevation Front slot</li>
				<li><code>plan-slug--floor-plan-main.jpg</code> → Floor Plan Main slot</li>
				<li><code>plan-slug--poster.jpg</code> → Poster slot</li>
			</ul>
			<p class="mt-2 text-muted-foreground">
				Use <strong>--</strong> (double dash) to separate plan name from slot type.
			</p>
		</div>
	{/if}

	<!-- Drop Zone -->
	{#if !isUploading}
		<div
			class="rounded-lg border-2 border-dashed p-8 text-center transition-colors {isDragging
				? 'border-blue-500 bg-blue-500/10'
				: 'border-input hover:border-muted-foreground'}"
			ondrop={handleDrop}
			ondragover={handleDragOver}
			ondragleave={handleDragLeave}
			role="button"
			tabindex="0"
		>
			<div class="flex flex-col items-center gap-3">
				<div class="rounded-full bg-blue-500/20 p-3">
					<Upload class="h-6 w-6 text-blue-600" />
				</div>
				<div>
					<p class="font-medium text-card-foreground">Drop image files here</p>
					<p class="text-sm text-muted-foreground">or click to browse</p>
				</div>
				<Button
					variant="outline"
					onclick={() => document.getElementById('bulk-image-upload')?.click()}
				>
					Select Files
				</Button>
				<input
					id="bulk-image-upload"
					type="file"
					accept="image/*"
					multiple
					class="hidden"
					onchange={handleFileSelect}
				/>
			</div>
		</div>
	{/if}

	<!-- Batch Assignment Bar -->
	{#if selectedFileIds.length > 0 && !isUploading}
		<div class="flex items-center gap-4 rounded-lg bg-muted p-4">
			<span class="text-sm font-medium text-foreground">
				{selectedFileIds.length} file{selectedFileIds.length > 1 ? 's' : ''} selected
			</span>
			<select
				bind:value={batchSlot}
				class="rounded-md border border-input px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
			>
				<option value="">Select slot...</option>
				{#each slots as slot}
					<option value={slot.value}>{slot.icon} {slot.label}</option>
				{/each}
			</select>
			<Button size="sm" onclick={batchAssign} disabled={!batchSlot}>Assign to Slot</Button>
			<Button variant="ghost" size="sm" onclick={() => (selectedFileIds = [])}>Clear</Button>
		</div>
	{/if}

	<!-- File List -->
	{#if uploadedFiles.length > 0}
		<div class="space-y-3">
			<div class="flex items-center justify-between">
				<h3 class="font-medium text-card-foreground">
					Files ({readyCount}/{totalCount} ready)
				</h3>
				{#if !isUploading}
					<Button variant="outline" size="sm" onclick={autoMatchFiles}>
						<Wand2 class="mr-1 h-4 w-4" />
						Auto-Match All
					</Button>
				{/if}
			</div>

			<div class="max-h-96 space-y-2 overflow-y-auto">
				{#each uploadedFiles as fileData (fileData.id)}
					<div
						class="flex items-center gap-3 rounded-lg border border-border p-3 {fileData.status ===
						'success'
							? 'border-green-500/30 bg-green-500/10'
							: fileData.status === 'error'
								? 'border-red-500/30 bg-red-500/10'
								: ''}"
					>
						<!-- Selection checkbox -->
						{#if !isUploading && fileData.status !== 'success'}
							<input
								type="checkbox"
								checked={selectedFileIds.includes(fileData.id)}
								onchange={() => toggleFileSelection(fileData.id)}
								class="h-4 w-4 rounded border-input"
							/>
						{/if}

						<!-- File icon -->
						<div class="flex-shrink-0">
							{#if fileData.status === 'uploading'}
								<Loader2 class="h-5 w-5 animate-spin text-blue-600" />
							{:else if fileData.status === 'success'}
								<CheckCircle2 class="h-5 w-5 text-green-600" />
							{:else if fileData.status === 'error'}
								<AlertCircle class="h-5 w-5 text-red-600" />
							{:else}
								<FileImage class="h-5 w-5 text-muted-foreground" />
							{/if}
						</div>

						<!-- File info -->
						<div class="min-w-0 flex-1">
							<p class="truncate font-medium text-card-foreground">{fileData.file.name}</p>
							<p class="text-xs text-muted-foreground">
								{formatFileSize(fileData.file.size)}
								{#if fileData.confidence > 0}
									<span
										class="ml-2 {fileData.confidence >= 75
											? 'text-green-600'
											: fileData.confidence >= 50
												? 'text-amber-600'
												: 'text-muted-foreground'}"
									>
										{fileData.confidence}% match
									</span>
								{/if}
							</p>
							{#if fileData.message}
								<p class="mt-1 text-xs text-muted-foreground">{fileData.message}</p>
							{/if}
						</div>

						<!-- Plan assignment -->
						{#if fileData.status !== 'success' && fileData.status !== 'uploading'}
							<select
								value={fileData.matchedPlan?.id ?? ''}
								onchange={(e) => assignPlan(fileData.id, (e.target as HTMLSelectElement).value)}
								class="w-40 rounded-md border border-input px-2 py-1 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
							>
								<option value="">Select plan...</option>
								{#each plans as plan}
									<option value={plan.id}>{plan.name}</option>
								{/each}
							</select>
						{:else if fileData.matchedPlan}
							<span class="text-sm text-muted-foreground">{fileData.matchedPlan.name}</span>
						{/if}

						<!-- Slot assignment -->
						{#if fileData.status !== 'success' && fileData.status !== 'uploading'}
							<select
								value={fileData.assignedSlot ?? ''}
								onchange={(e) => assignSlot(fileData.id, (e.target as HTMLSelectElement).value)}
								class="w-44 rounded-md border border-input px-2 py-1 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 {fileData.suggestedSlot &&
								!fileData.assignedSlot
									? 'border-amber-500/50 bg-amber-500/10'
									: ''}"
							>
								<option value="">Select slot...</option>
								{#each slots as slot}
									<option value={slot.value}>{slot.icon} {slot.label}</option>
								{/each}
							</select>
						{:else if fileData.assignedSlot}
							<span class="text-sm text-muted-foreground">
								{slots.find((s) => s.value === fileData.assignedSlot)?.icon}
								{slots.find((s) => s.value === fileData.assignedSlot)?.label}
							</span>
						{/if}

						<!-- Remove button -->
						{#if !isUploading && fileData.status !== 'uploading'}
							<button
								onclick={() => removeFile(fileData.id)}
								class="text-muted-foreground hover:text-red-600"
							>
								<X class="h-4 w-4" />
							</button>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Progress -->
	{#if isUploading}
		<div class="space-y-2">
			<div class="flex justify-between text-sm">
				<span class="text-muted-foreground">Uploading...</span>
				<span class="font-medium">{uploadProgress}%</span>
			</div>
			<div class="h-2 rounded-full bg-muted">
				<div
					class="h-2 rounded-full bg-blue-600 transition-all"
					style="width: {uploadProgress}%"
				></div>
			</div>
		</div>
	{/if}

	<!-- Summary -->
	{#if successCount > 0 && !isUploading}
		<div class="rounded-lg bg-green-500/10 p-4 text-center">
			<CheckCircle2 class="mx-auto mb-2 h-8 w-8 text-green-600" />
			<p class="font-medium text-green-600">
				{successCount} image{successCount > 1 ? 's' : ''} uploaded successfully
			</p>
		</div>
	{/if}

	<!-- Actions -->
	<div class="flex justify-between">
		<Button variant="outline" onclick={onBack} disabled={isUploading}>
			<ArrowLeft class="mr-2 h-4 w-4" />
			Back
		</Button>
		<div class="flex gap-3">
			{#if !isUploading && successCount === 0}
				<Button variant="ghost" onclick={onSkip}>Skip</Button>
			{/if}
			{#if isUploading}
				<Button disabled>
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
					Uploading...
				</Button>
			{:else if successCount === totalCount && totalCount > 0}
				<Button onclick={onComplete}>
					Complete
					<ArrowRight class="ml-2 h-4 w-4" />
				</Button>
			{:else}
				<Button onclick={handleUpload} disabled={readyCount === 0 || isUploading}>
					<Upload class="mr-2 h-4 w-4" />
					Upload {readyCount} File{readyCount !== 1 ? 's' : ''}
				</Button>
			{/if}
		</div>
	</div>
</div>
