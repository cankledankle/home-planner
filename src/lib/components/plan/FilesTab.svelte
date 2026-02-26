<script lang="ts">
	import { Button } from '$lib/components/ui/button';
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
		Loader2
	} from '@lucide/svelte';
	import { getPlanFiles, uploadFiles, deleteFile, getFileUrl } from '$lib/api';
	import type { File as FileType } from '$lib/types';

	interface Props {
		planId: string;
	}

	let { planId }: Props = $props();

	// Category tabs
	type Category = 'reference' | 'technical' | '3d' | 'other';
	let activeCategory = $state<Category>('reference');

	// Files data
	let files = $state<{
		reference: FileType[];
		technical: FileType[];
		'3d': FileType[];
		other: FileType[];
	}>({
		reference: [],
		technical: [],
		'3d': [],
		other: []
	});

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
				reference: data.reference || [],
				technical: data.technical || [],
				'3d': data['3d'] || [],
				other: data.other || []
			};
		} catch (err) {
			toast.error('Failed to load files');
		} finally {
			loading = false;
		}
	}

	function getCategoryIcon(category: Category) {
		switch (category) {
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

	async function uploadFilesList(fileList: File[]) {
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
		reference: 'Reference Images',
		technical: 'Technical Drawings',
		'3d': '3D Models',
		other: 'Other Files'
	};

	const currentFiles = $derived(files[activeCategory] || []);
</script>

<div class="space-y-6">
	<!-- Category Tabs -->
	<div class="border-b border-slate-200">
		<nav class="flex gap-1">
			{#each Object.entries(categoryLabels) as [category, label]}
				<button
					class="border-b-2 px-4 py-3 text-sm font-medium transition-colors {activeCategory ===
					category
						? 'border-blue-600 text-blue-600'
						: 'border-transparent text-slate-600 hover:text-slate-900'}"
					onclick={() => (activeCategory = category as Category)}
				>
					<svelte:component
						this={getCategoryIcon(category as Category)}
						class="mr-2 inline h-4 w-4"
					/>
					{label}
					<span class="ml-2 rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-600">
						{files[category as Category]?.length || 0}
					</span>
				</button>
			{/each}
		</nav>
	</div>

	<!-- Upload Zone -->
	<div
		class="rounded-lg border-2 border-dashed p-8 text-center transition-colors {isDragging
			? 'border-blue-500 bg-blue-50'
			: 'border-slate-300 hover:border-slate-400'}"
		ondrop={handleDrop}
		ondragover={handleDragOver}
		ondragleave={handleDragLeave}
	>
		{#if uploading}
			<div class="flex flex-col items-center gap-3">
				<Loader2 class="h-8 w-8 animate-spin text-blue-600" />
				<p class="text-sm font-medium text-slate-700">Uploading files...</p>
			</div>
		{:else}
			<div class="flex flex-col items-center gap-3">
				<div class="rounded-full bg-blue-100 p-3">
					<Upload class="h-6 w-6 text-blue-600" />
				</div>
				<div>
					<p class="text-sm font-medium text-slate-700">Drop files here or click to upload</p>
					<p class="mt-1 text-xs text-slate-500">Maximum file size: 500MB</p>
				</div>
				<Button variant="outline" onclick={() => document.getElementById('file-upload')?.click()}>
					Select Files
				</Button>
				<input id="file-upload" type="file" multiple class="hidden" onchange={handleFileSelect} />
			</div>
		{/if}
	</div>

	<!-- File List -->
	<div class="rounded-lg border border-slate-200 bg-white">
		{#if loading}
			<div class="p-8 text-center">
				<Loader2 class="mx-auto h-8 w-8 animate-spin text-slate-400" />
				<p class="mt-2 text-sm text-slate-500">Loading files...</p>
			</div>
		{:else if currentFiles.length === 0}
			<div class="p-8 text-center">
				<svelte:component
					this={getCategoryIcon(activeCategory)}
					class="mx-auto mb-4 h-12 w-12 text-slate-300"
				/>
				<h3 class="mb-2 text-lg font-medium text-slate-900">No files yet</h3>
				<p class="text-slate-600">
					Upload {categoryLabels[activeCategory].toLowerCase()} to get started
				</p>
			</div>
		{:else}
			<div class="divide-y divide-slate-100">
				{#each currentFiles as file}
					<div class="flex items-center gap-4 p-4 hover:bg-slate-50">
						<div
							class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-slate-100"
						>
							<svelte:component
								this={getCategoryIcon(activeCategory)}
								class="h-5 w-5 text-slate-600"
							/>
						</div>
						<div class="min-w-0 flex-1">
							<p class="truncate font-medium text-slate-900">
								{file.filename}
							</p>
							<p class="text-xs text-slate-500">
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
