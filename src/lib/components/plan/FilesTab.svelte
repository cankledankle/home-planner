<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import {
		Dialog,
		DialogContent,
		DialogHeader,
		DialogTitle,
		DialogDescription
	} from '$lib/components/ui/dialog';
	import { toast } from 'svelte-sonner';
	import {
		FileText,
		Upload,
		X,
		Download,
		Trash2,
		Image,
		File,
		Box,
		MoreHorizontal,
		Loader2,
		Globe,
		Check,
		Trash,
		AlertTriangle
	} from '@lucide/svelte';
	import { getPlanFiles, uploadFiles, deleteFile, getFileUrl, uploadWebsiteFile } from '$lib/api';
	import type { File as FileType } from '$lib/types';

	interface Props {
		planId: string;
		onUploadComplete?: () => void;
	}

	let { planId, onUploadComplete }: Props = $props();

	// Category tabs
	type Category = 'website' | 'reference' | 'technical' | '3d' | 'other';
	let activeCategory = $state<Category>('reference');

	// Website slots
	const websiteSlots = [
		{ key: 'render-front', label: 'Render Front' },
		{ key: 'elevation-front', label: 'Elevation Front' },
		{ key: 'elevation-left', label: 'Elevation Left' },
		{ key: 'elevation-rear', label: 'Elevation Rear' },
		{ key: 'elevation-right', label: 'Elevation Right' },
		{ key: 'floor-plan-main', label: 'Floor Plan Main' },
		{ key: 'floor-plan-upper', label: 'Floor Plan Upper' },
		{ key: 'floor-plan-lower', label: 'Floor Plan Lower' },
		{ key: 'poster', label: 'Poster' }
	];

	// Files data
	let files = $state<{
		website: FileType[];
		reference: FileType[];
		technical: FileType[];
		'3d': FileType[];
		other: FileType[];
	}>({
		website: [],
		reference: [],
		technical: [],
		'3d': [],
		other: []
	});

	// Website slot files
	let websiteSlotFiles = $state<Record<string, FileType | null>>({});

	// Slot image URLs
	let slotImageUrls = $state<Record<string, string | null>>({});

	// Bulk upload for website slots
	type UploadStatus = 'pending' | 'uploading' | 'complete' | 'error';
	interface SelectedFile {
		file: globalThis.File;
		slot: string | null;
		id: string;
		status: UploadStatus;
		errorMessage?: string;
	}
	let bulkUploadOpen = $state(false);
	let selectedFiles = $state<SelectedFile[]>([]);
	let bulkUploading = $state(false);

	// Overwrite confirmation
	let overwriteConfirmOpen = $state(false);
	let slotsToOverwrite = $state<{ slot: string; label: string; existingFile: FileType }[]>([]);

	let loading = $state(true);
	let uploadProgress = $state<Record<string, number>>({});
	let isDragging = $state(false);
	let uploading = $state(false);

	// Load files on mount and when category changes
	$effect(() => {
		loadFiles();
	});

	async function loadFiles() {
		loading = true;
		try {
			const data = await getPlanFiles(planId);
			files = {
				website: data.website
					? Object.values(data.website).filter((f): f is FileType => Boolean(f))
					: [],
				reference: data.reference || [],
				technical: data.technical || [],
				'3d': data['3d'] || [],
				other: data.other || []
			};

			// Store website slot files for grid display
			websiteSlotFiles = data.website || {};

			// Load signed URLs for website slot images
			if (data.website) {
				await loadSlotImageUrls(data.website);
			}
		} catch (err) {
			toast.error('Failed to load files');
		} finally {
			loading = false;
		}
	}

	async function loadSlotImageUrls(websiteFiles: Record<string, FileType | null>) {
		const urls: Record<string, string | null> = {};

		for (const [slotKey, file] of Object.entries(websiteFiles)) {
			if (file) {
				try {
					const result = await getFileUrl(file.id);
					urls[slotKey] = result.url;
				} catch (err) {
					urls[slotKey] = null;
				}
			}
		}

		slotImageUrls = urls;
	}

	function getCategoryIcon(category: Category) {
		switch (category) {
			case 'website':
				return Globe;
			case 'reference':
				return Image;
			case 'technical':
				return FileText;
			case '3d':
				return Box;
			default:
				return File;
		}
	}

	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 Bytes';
		const k = 1024;
		const sizes = ['Bytes', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	async function handleFileSelect(event: Event) {
		const input = event.target as HTMLInputElement;
		if (input.files && input.files.length > 0) {
			await uploadFilesList(Array.from(input.files));
		}
	}

	async function handleDrop(event: DragEvent) {
		event.preventDefault();
		isDragging = false;

		if (event.dataTransfer?.files) {
			await uploadFilesList(Array.from(event.dataTransfer.files));
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

	async function uploadFilesList(fileList: globalThis.File[]) {
		if (fileList.length === 0) return;

		uploading = true;
		const uploadId = Math.random().toString(36).substring(7);
		uploadProgress[uploadId] = 0;

		try {
			// Simulate progress (actual progress tracking would require XMLHttpRequest)
			const progressInterval = setInterval(() => {
				if (uploadProgress[uploadId] < 90) {
					uploadProgress[uploadId] += 10;
				}
			}, 200);

			await uploadFiles(planId, activeCategory, fileList);

			clearInterval(progressInterval);
			uploadProgress[uploadId] = 100;

			toast.success(`${fileList.length} file(s) uploaded successfully`);
			await loadFiles();
		} catch (err) {
			toast.error('Failed to upload files');
		} finally {
			delete uploadProgress[uploadId];
			uploading = false;
		}
	}

	async function handleDownload(file: FileType) {
		try {
			const { url } = await getFileUrl(file.id);
			window.open(url, '_blank');
		} catch (err) {
			toast.error('Failed to get download URL');
		}
	}

	async function handleDelete(file: FileType) {
		if (!confirm(`Are you sure you want to delete "${file.filename}"?`)) return;

		try {
			await deleteFile(file.id);
			toast.success('File deleted successfully');
			await loadFiles();
		} catch (err) {
			toast.error('Failed to delete file');
		}
	}

	const categoryLabels: Record<Category, string> = {
		website: 'Website Images',
		reference: 'Reference Images',
		technical: 'Technical Drawings',
		'3d': '3D Models',
		other: 'Other Files'
	};

	const currentFiles = $derived(files[activeCategory] || []);

	// Bulk upload functions for website slots
	function openBulkUpload() {
		selectedFiles = [];
		bulkUploadOpen = true;
	}

	function handleBulkFileSelect(event: Event) {
		const input = event.target as HTMLInputElement;
		if (input.files && input.files.length > 0) {
			const newFiles: SelectedFile[] = Array.from(input.files).map((file) => ({
				file,
				slot: null,
				id: Math.random().toString(36).substring(7),
				status: 'pending'
			}));
			selectedFiles = [...selectedFiles, ...newFiles];
		}
		input.value = '';
	}

	function handleBulkDrop(event: DragEvent) {
		event.preventDefault();
		isDragging = false;

		if (event.dataTransfer?.files) {
			const newFiles: SelectedFile[] = Array.from(event.dataTransfer.files).map((file) => ({
				file,
				slot: null,
				id: Math.random().toString(36).substring(7),
				status: 'pending'
			}));
			selectedFiles = [...selectedFiles, ...newFiles];
		}
	}

	function removeSelectedFile(id: string) {
		selectedFiles = selectedFiles.filter((f) => f.id !== id);
	}

	function assignSlot(fileId: string, slotKey: string) {
		selectedFiles = selectedFiles.map((f) => (f.id === fileId ? { ...f, slot: slotKey } : f));
	}

	function checkForExistingFiles() {
		const filesWithSlots = selectedFiles.filter((f) => f.slot);
		const existing: { slot: string; label: string; existingFile: FileType }[] = [];

		for (const f of filesWithSlots) {
			if (f.slot && websiteSlotFiles[f.slot]) {
				const slotLabel = websiteSlots.find((s) => s.key === f.slot)?.label || f.slot;
				existing.push({
					slot: f.slot,
					label: slotLabel,
					existingFile: websiteSlotFiles[f.slot]!
				});
			}
		}

		return existing;
	}

	function initiateBulkUpload() {
		const filesWithSlots = selectedFiles.filter((f) => f.slot);
		if (filesWithSlots.length === 0) {
			toast.error('Please assign at least one file to a slot');
			return;
		}

		// Check for duplicate slot assignments
		const slotCounts: Record<string, number> = {};
		for (const f of filesWithSlots) {
			if (f.slot) {
				slotCounts[f.slot] = (slotCounts[f.slot] || 0) + 1;
			}
		}
		const duplicates = Object.entries(slotCounts).filter(([_, count]) => count > 1);
		if (duplicates.length > 0) {
			const slots = duplicates
				.map(([slot]) => websiteSlots.find((s) => s.key === slot)?.label || slot)
				.join(', ');
			toast.error(`Multiple files assigned to: ${slots}`);
			return;
		}

		// Check for existing files
		const existing = checkForExistingFiles();
		if (existing.length > 0) {
			slotsToOverwrite = existing;
			overwriteConfirmOpen = true;
		} else {
			executeBulkUpload();
		}
	}

	async function executeBulkUpload() {
		overwriteConfirmOpen = false;
		bulkUploading = true;

		const filesWithSlots = selectedFiles.filter((f) => f.slot);
		const total = filesWithSlots.length;
		let completed = 0;
		let failed = 0;

		for (const selectedFile of filesWithSlots) {
			if (!selectedFile.slot) continue;

			// Update status to uploading
			selectedFiles = selectedFiles.map((f) =>
				f.id === selectedFile.id ? { ...f, status: 'uploading' } : f
			);

			try {
				await uploadWebsiteFile(planId, selectedFile.slot, selectedFile.file);
				completed++;

				// Update status to complete
				selectedFiles = selectedFiles.map((f) =>
					f.id === selectedFile.id ? { ...f, status: 'complete' } : f
				);
			} catch (err) {
				failed++;
				const errorMsg = err instanceof Error ? err.message : 'Upload failed';

				// Update status to error
				selectedFiles = selectedFiles.map((f) =>
					f.id === selectedFile.id ? { ...f, status: 'error', errorMessage: errorMsg } : f
				);
			}
		}

		bulkUploading = false;

		if (failed === 0) {
			toast.success(`${completed} file(s) uploaded successfully`);
			// Call parent callback to refresh plan data
			onUploadComplete?.();
			// Keep modal open briefly so user can see completion status
			setTimeout(() => {
				bulkUploadOpen = false;
				selectedFiles = [];
				loadFiles();
			}, 1500);
		} else if (completed === 0) {
			toast.error('All uploads failed');
		} else {
			toast.error(`${failed} file(s) failed, ${completed} succeeded`);
		}
	}

	function getStatusIcon(status: UploadStatus) {
		switch (status) {
			case 'uploading':
				return Loader2;
			case 'complete':
				return Check;
			case 'error':
				return AlertTriangle;
			default:
				return null;
		}
	}

	function getStatusColor(status: UploadStatus): string {
		switch (status) {
			case 'uploading':
				return 'text-blue-500';
			case 'complete':
				return 'text-green-500';
			case 'error':
				return 'text-red-500';
			default:
				return '';
		}
	}

	function handleBulkUpload() {
		initiateBulkUpload();
	}

	function getSlotLabel(slotKey: string | null): string {
		if (!slotKey) return 'Select slot...';
		const slot = websiteSlots.find((s) => s.key === slotKey);
		return slot?.label || slotKey;
	}

	function getAvailableSlots(
		currentFileId: string
	): { key: string; label: string; occupied: boolean }[] {
		const occupiedSlots = new Set(
			selectedFiles.filter((f) => f.id !== currentFileId && f.slot).map((f) => f.slot)
		);
		return websiteSlots.map((slot) => ({
			...slot,
			occupied: occupiedSlots.has(slot.key)
		}));
	}
</script>

<div class="space-y-6">
	<!-- Category Tabs -->
	<div class="border-b border-border">
		<nav class="flex gap-1">
			{#each Object.entries(categoryLabels) as [category, label]}
				{@const IconComponent = getCategoryIcon(category as Category)}
				<button
					class="border-b-2 px-4 py-3 text-sm font-medium transition-colors {activeCategory ===
					category
						? 'border-blue-600 text-blue-600'
						: 'border-transparent text-muted-foreground hover:text-card-foreground'}"
					onclick={() => (activeCategory = category as Category)}
				>
					<IconComponent class="mr-2 inline h-4 w-4" />
					{label}
					<span class="ml-2 rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
						{files[category as Category]?.length || 0}
					</span>
				</button>
			{/each}
		</nav>
	</div>

	<!-- Upload Zone -->
	{#if activeCategory === 'website'}
		<!-- Website Slots Grid -->
		<div class="rounded-lg border border-border bg-card p-4">
			<div class="mb-4 flex items-center justify-between">
				<h3 class="text-sm font-medium text-foreground">Website Image Slots</h3>
				<Button size="sm" onclick={openBulkUpload}>
					<Upload class="mr-2 h-4 w-4" />
					Bulk Upload to Slots
				</Button>
			</div>
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
				{#each websiteSlots as slot}
					{@const file = websiteSlotFiles[slot.key]}
					{@const imageUrl = slotImageUrls[slot.key]}
					<div
						class="relative aspect-square overflow-hidden rounded-lg border border-border bg-muted"
					>
						{#if file && imageUrl}
							<img src={imageUrl} alt={slot.label} class="h-full w-full object-cover" />
							<div
								class="absolute inset-0 flex items-end justify-between bg-gradient-to-t from-black/60 to-transparent p-2"
							>
								<span class="text-xs font-medium text-white">{slot.label}</span>
							</div>
						{:else}
							<div class="flex h-full flex-col items-center justify-center p-2 text-center">
								<Image class="h-8 w-8 text-muted-foreground" />
								<span class="mt-1 text-[10px] text-muted-foreground">{slot.label}</span>
							</div>
						{/if}
					</div>
				{/each}
			</div>
		</div>

		<!-- Upload Area -->
		<div
			class="rounded-lg border-2 border-dashed p-8 text-center transition-colors {isDragging
				? 'border-blue-500 bg-blue-500/10'
				: 'border-input hover:border-slate-400'}"
			role="button"
			tabindex="0"
			aria-label="Upload files drop zone"
			ondrop={handleDrop}
			ondragover={handleDragOver}
			ondragleave={handleDragLeave}
		>
			{#if uploading}
				<div class="flex flex-col items-center gap-3">
					<Loader2 class="h-8 w-8 animate-spin text-blue-600" />
					<p class="text-sm font-medium text-foreground">Uploading files...</p>
				</div>
			{:else}
				<div class="flex flex-col items-center gap-3">
					<div class="rounded-full bg-blue-500/20 p-3">
						<Upload class="h-6 w-6 text-blue-600" />
					</div>
					<div>
						<p class="text-sm font-medium text-foreground">Drop files here or click to upload</p>
						<div class="mt-1 space-y-0.5 text-xs text-muted-foreground">
							<p>Max size: 50MB • Images will be optimized (max 4000px, JPEG 90%)</p>
							<p class="text-amber-600">PNG files with transparency will be stored as "Other"</p>
						</div>
					</div>
					<Button variant="outline" onclick={() => document.getElementById('file-upload')?.click()}>
						Select Files
					</Button>
					<input id="file-upload" type="file" multiple class="hidden" onchange={handleFileSelect} />
				</div>
			{/if}
		</div>
	{:else}
		<div
			class="rounded-lg border-2 border-dashed p-8 text-center transition-colors {isDragging
				? 'border-blue-500 bg-blue-500/10'
				: 'border-input hover:border-slate-400'}"
			role="button"
			tabindex="0"
			aria-label="Upload files drop zone"
			ondrop={handleDrop}
			ondragover={handleDragOver}
			ondragleave={handleDragLeave}
		>
			{#if uploading}
				<div class="flex flex-col items-center gap-3">
					<Loader2 class="h-8 w-8 animate-spin text-blue-600" />
					<p class="text-sm font-medium text-foreground">Uploading files...</p>
				</div>
			{:else}
				<div class="flex flex-col items-center gap-3">
					<div class="rounded-full bg-blue-500/20 p-3">
						<Upload class="h-6 w-6 text-blue-600" />
					</div>
					<div>
						<p class="text-sm font-medium text-foreground">Drop files here or click to upload</p>
						<p class="mt-1 text-xs text-muted-foreground">Maximum file size: 50MB</p>
					</div>
					<Button variant="outline" onclick={() => document.getElementById('file-upload')?.click()}>
						Select Files
					</Button>
					<input id="file-upload" type="file" multiple class="hidden" onchange={handleFileSelect} />
				</div>
			{/if}
		</div>
	{/if}

	<!-- File List -->
	<div class="rounded-lg border border-border bg-card">
		{#if loading}
			<div class="p-8 text-center">
				<Loader2 class="mx-auto h-8 w-8 animate-spin text-muted-foreground" />
				<p class="mt-2 text-sm text-muted-foreground">Loading files...</p>
			</div>
		{:else if currentFiles.length === 0}
			{@const EmptyIcon = getCategoryIcon(activeCategory)}
			<div class="p-8 text-center">
				<EmptyIcon class="mx-auto mb-4 h-12 w-12 text-muted-foreground" />
				<h3 class="mb-2 text-lg font-medium text-card-foreground">No files yet</h3>
				<p class="text-muted-foreground">
					Upload {categoryLabels[activeCategory].toLowerCase()} to get started
				</p>
			</div>
		{:else}
			<div class="divide-y divide-slate-100">
				{#each currentFiles as file}
					{@const FileIcon = getCategoryIcon(activeCategory)}
					<div class="flex items-center gap-4 p-4 hover:bg-muted/50">
						<div
							class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-muted"
						>
							<FileIcon class="h-5 w-5 text-muted-foreground" />
						</div>
						<div class="min-w-0 flex-1">
							<p class="truncate font-medium text-card-foreground">
								{file.filename}
							</p>
							<p class="text-xs text-muted-foreground">
								{formatFileSize(file.size_bytes)} • Uploaded {formatDate(file.uploaded_at)}
							</p>
						</div>
						<div class="flex items-center gap-2">
							<Button
								variant="ghost"
								size="sm"
								onclick={() => handleDownload(file)}
								title="Download"
							>
								<Download class="h-4 w-4" />
							</Button>
							<Button
								variant="ghost"
								size="sm"
								class="text-red-600 hover:text-red-700"
								onclick={() => handleDelete(file)}
								title="Delete"
							>
								<Trash2 class="h-4 w-4" />
							</Button>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- Bulk Upload Dialog for Website Slots -->
<Dialog bind:open={bulkUploadOpen}>
	<DialogContent class="max-h-[90vh] overflow-hidden sm:max-w-4xl">
		<DialogHeader>
			<DialogTitle>Bulk Upload to Website Slots</DialogTitle>
			<DialogDescription>
				Upload multiple images and assign them to specific website slots
			</DialogDescription>
		</DialogHeader>

		<!-- File Drop Zone for Bulk Upload -->
		<div
			class="rounded-lg border-2 border-dashed p-6 text-center transition-colors {isDragging
				? 'border-blue-500 bg-blue-500/10'
				: 'border-input hover:border-slate-400'}"
			role="button"
			tabindex="0"
			aria-label="Bulk upload drop zone"
			ondrop={handleBulkDrop}
			ondragover={handleDragOver}
			ondragleave={handleDragLeave}
		>
			<div class="flex flex-col items-center gap-2">
				<div class="rounded-full bg-blue-500/20 p-2">
					<Upload class="h-5 w-5 text-blue-600" />
				</div>
				<p class="text-sm font-medium text-foreground">Drop files here or click to select</p>
				<div class="space-y-0.5 text-center text-xs text-muted-foreground">
					<p>Max 5MB per image • Will be resized to max 4000px • JPEG output</p>
					<p class="text-amber-600">PNG with transparency → stored as "Other"</p>
				</div>
				<Button
					size="sm"
					variant="outline"
					onclick={() => document.getElementById('bulk-file-upload')?.click()}
				>
					Select Files
				</Button>
				<input
					id="bulk-file-upload"
					type="file"
					multiple
					accept="image/jpeg,image/png,.jpg,.jpeg,.png"
					class="hidden"
					onchange={handleBulkFileSelect}
				/>
			</div>
		</div>

		<!-- Selected Files with Slot Assignment -->
		{#if selectedFiles.length > 0}
			<div class="mt-6 space-y-3">
				<h4 class="text-sm font-medium text-foreground">
					Selected Files ({selectedFiles.length})
				</h4>
				<div class="max-h-[50vh] space-y-2 overflow-y-auto pr-2">
					{#each selectedFiles as selectedFile (selectedFile.id)}
						{@const StatusIcon = getStatusIcon(selectedFile.status)}
						<div
							class="flex items-center gap-3 overflow-hidden rounded-lg border border-border p-3 {selectedFile.status ===
							'error'
								? 'border-red-500/30 bg-red-500/10'
								: ''} {selectedFile.status === 'complete'
								? 'border-green-500/30 bg-green-500/10'
								: ''}"
						>
							<div
								class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg bg-muted"
							>
								{#if selectedFile.status === 'uploading'}
									<Loader2 class="h-6 w-6 animate-spin text-blue-500" />
								{:else}
									<Image class="h-6 w-6 text-muted-foreground" />
								{/if}
							</div>
							<div class="min-w-0 flex-1 overflow-hidden">
								<p class="truncate text-sm font-medium text-card-foreground">
									{selectedFile.file.name}
								</p>
								<p class="text-xs text-muted-foreground">
									{formatFileSize(selectedFile.file.size)}
								</p>
								{#if selectedFile.status === 'error' && selectedFile.errorMessage}
									<p class="mt-1 text-xs text-red-600">{selectedFile.errorMessage}</p>
								{/if}
							</div>
							<div class="flex flex-shrink-0 items-center gap-2">
								<select
									value={selectedFile.slot || ''}
									onchange={(e) =>
										assignSlot(selectedFile.id, (e.target as HTMLSelectElement).value)}
									disabled={selectedFile.status === 'uploading' ||
										selectedFile.status === 'complete'}
									class="w-[140px] rounded-md border border-input bg-card px-3 py-1.5 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none disabled:bg-muted disabled:text-muted-foreground"
								>
									<option value="">Select slot...</option>
									{#each getAvailableSlots(selectedFile.id) as slotOption}
										<option value={slotOption.key} disabled={slotOption.occupied}>
											{slotOption.label}{slotOption.occupied ? ' (assigned)' : ''}
										</option>
									{/each}
								</select>
								{#if StatusIcon}
									<StatusIcon
										class="h-4 w-4 flex-shrink-0 {getStatusColor(
											selectedFile.status
										)} {selectedFile.status === 'uploading' ? 'animate-spin' : ''}"
									/>
								{/if}
								<Button
									size="sm"
									variant="ghost"
									class="flex-shrink-0 text-red-600 hover:text-red-700"
									onclick={() => removeSelectedFile(selectedFile.id)}
									disabled={selectedFile.status === 'uploading'}
								>
									<Trash class="h-4 w-4" />
								</Button>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Upload Progress -->
		{#if bulkUploading}
			<div class="mt-6">
				<div class="flex items-center gap-2">
					<Loader2 class="h-4 w-4 animate-spin text-blue-600" />
					<p class="text-sm text-muted-foreground">Uploading files...</p>
				</div>
			</div>
		{/if}

		<!-- Dialog Footer -->
		<div class="mt-6 flex justify-end gap-2">
			<Button
				variant="outline"
				onclick={() => {
					bulkUploadOpen = false;
					selectedFiles = [];
				}}
				disabled={bulkUploading}
			>
				Cancel
			</Button>
			<Button
				onclick={handleBulkUpload}
				disabled={selectedFiles.length === 0 ||
					bulkUploading ||
					selectedFiles.every((f) => f.status === 'complete')}
			>
				{#if bulkUploading}
					Uploading...
				{:else}
					Upload {selectedFiles.filter((f) => f.slot && f.status !== 'complete').length > 0
						? `${selectedFiles.filter((f) => f.slot && f.status !== 'complete').length} Files`
						: 'Files'}
				{/if}
			</Button>
		</div>
	</DialogContent>
</Dialog>

<!-- Overwrite Confirmation Dialog -->
<Dialog bind:open={overwriteConfirmOpen}>
	<DialogContent class="sm:max-w-md">
		<DialogHeader>
			<DialogTitle class="flex items-center gap-2">
				<AlertTriangle class="h-5 w-5 text-amber-500" />
				Confirm Overwrite
			</DialogTitle>
			<DialogDescription>
				The following slots already have files. Uploading will replace the existing files.
			</DialogDescription>
		</DialogHeader>

		<div class="mt-4 space-y-2">
			{#each slotsToOverwrite as { slot, label, existingFile }}
				<div class="flex items-center gap-3 rounded-md bg-amber-500/10 p-3 text-sm">
					<div
						class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded bg-amber-500/20"
					>
						<Image class="h-4 w-4 text-amber-600" />
					</div>
					<div class="min-w-0 flex-1">
						<p class="truncate font-medium text-amber-600">{label}</p>
						<p class="truncate text-xs text-amber-600">Current: {existingFile.filename}</p>
					</div>
				</div>
			{/each}
		</div>

		<div class="mt-6 flex justify-end gap-2">
			<Button variant="outline" onclick={() => (overwriteConfirmOpen = false)}>Cancel</Button>
			<Button variant="destructive" onclick={executeBulkUpload}>Overwrite Files</Button>
		</div>
	</DialogContent>
</Dialog>
