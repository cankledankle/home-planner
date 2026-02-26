<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import {
		Dialog,
		DialogContent,
		DialogHeader,
		DialogTitle,
		DialogDescription,
		DialogFooter
	} from '$lib/components/ui/dialog';
	import { toast } from 'svelte-sonner';
	import {
		getPlan,
		updatePlan,
		deletePlan,
		flagPlan,
		unflagPlan,
		getPlanActivities,
		duplicatePlan,
		getFileUrl,
		deleteFile,
		uploadWebsiteFile
	} from '$lib/api';
	import { auth } from '$lib/stores';
	import type { PlanWithFiles, Activity, PaginatedResponse, File as FileType } from '$lib/types';
	import {
		ArrowLeft,
		Edit2,
		Flag,
		Trash2,
		Copy,
		Save,
		X,
		CheckCircle2,
		AlertCircle,
		Image,
		FileText,
		History,
		Upload,
		Replace,
		Eye,
		Loader2
	} from '@lucide/svelte';
	import FilesTab from '$lib/components/plan/FilesTab.svelte';
	import ConfirmationDialog from '$lib/components/ui/dialog/ConfirmationDialog.svelte';

	let plan = $state<PlanWithFiles | null>(null);
	let loading = $state(true);
	let editing = $state(false);

	// Get active tab from URL query param, default to 'overview'
	let activeTab = $derived<'overview' | 'files' | 'activity'>(
		($page.url.searchParams.get('tab') as 'overview' | 'files' | 'activity') || 'overview'
	);
	let saving = $state(false);
	let activities = $state<PaginatedResponse<Activity> | null>(null);

	// Edit form state
	let editForm = $state<Partial<PlanWithFiles>>({});

	// Image preview modal
	let previewOpen = $state(false);
	let previewFile = $state<FileType | null>(null);
	let previewUrl = $state<string | null>(null);
	let previewLoading = $state(false);

	// Slot upload modal
	let uploadSlotOpen = $state(false);
	let uploadSlotKey = $state<string>('');
	let uploadSlotLabel = $state<string>('');
	let uploadFile = $state<File | null>(null);
	let uploadLoading = $state(false);

	// Slot image URLs cache
	let slotImageUrls = $state<Record<string, string | null>>({});

	// Slot delete confirmation
	let deleteSlotOpen = $state(false);
	let deleteSlotFile = $state<FileType | null>(null);
	let deleteSlotLoading = $state(false);

	// Plan delete confirmation
	let deletePlanOpen = $state(false);

	const planId = $derived($page.params.id!);

	// Warn about unsaved changes when leaving page while editing
	function handleBeforeUnload(e: BeforeUnloadEvent) {
		if (editing) {
			e.preventDefault();
			e.returnValue = '';
		}
	}

	onMount(() => {
		window.addEventListener('beforeunload', handleBeforeUnload);
	});

	onDestroy(() => {
		window.removeEventListener('beforeunload', handleBeforeUnload);
	});

	onMount(async () => {
		if (!planId) {
			goto('/plans');
			return;
		}
		await loadPlan();
	});

	async function loadPlan() {
		if (!planId) return;
		loading = true;
		try {
			plan = await getPlan(planId);
			editForm = { ...plan };
			// Load signed URLs for all website slot images
			await loadSlotImageUrls();
		} catch (err) {
			toast.error('Failed to load plan');
			goto('/plans');
		} finally {
			loading = false;
		}
	}

	async function loadSlotImageUrls() {
		if (!plan) return;

		const urls: Record<string, string | null> = {};
		const slotFiles = plan.files.website;

		// Load URLs for all filled slots
		for (const [slotKey, file] of Object.entries(slotFiles)) {
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

	async function loadActivities() {
		if (activities || !planId) return;
		try {
			activities = await getPlanActivities(planId, 1, 20);
		} catch (err) {
			toast.error('Failed to load activity');
		}
	}

	$effect(() => {
		if (activeTab === 'activity') {
			loadActivities();
		}
	});

	async function handleSave() {
		if (!plan || !planId) return;
		saving = true;
		try {
			await updatePlan(planId, {
				name: editForm.name,
				type: editForm.type,
				style: editForm.style,
				beds: editForm.beds,
				baths: editForm.baths,
				half_baths: editForm.half_baths,
				main_sf: editForm.main_sf,
				upper_sf: editForm.upper_sf,
				lower_sf: editForm.lower_sf,
				porch_deck_sf: editForm.porch_deck_sf,
				garage_sf: editForm.garage_sf,
				garage_apartment_sf: editForm.garage_apartment_sf,
				unfinished_sf: editForm.unfinished_sf,
				heated_sf: editForm.heated_sf,
				total_sf: editForm.total_sf,
				notes: editForm.notes
			});
			toast.success('Plan updated successfully');
			await loadPlan();
			editing = false;
		} catch (err) {
			toast.error('Failed to update plan');
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		if (!planId) return;
		deletePlanOpen = true;
	}

	async function confirmDelete() {
		if (!planId) return;
		try {
			await deletePlan(planId);
			toast.success('Plan deleted successfully');
			goto('/plans');
		} catch (err) {
			toast.error('Failed to delete plan');
		}
	}

	async function handleFlag() {
		if (!planId) return;
		try {
			if (plan?.status === 'flagged') {
				await unflagPlan(planId);
				toast.success('Plan unflagged');
			} else {
				await flagPlan(planId);
				toast.success('Plan flagged for review');
			}
			await loadPlan();
		} catch (err) {
			toast.error('Failed to update flag status');
		}
	}

	async function handleDuplicate() {
		if (!planId) return;
		const newName = prompt('Enter name for the duplicated plan:', `${plan?.name} Copy`);
		if (!newName) return;
		try {
			const newPlan = await duplicatePlan(planId, newName);
			toast.success('Plan duplicated successfully');
			goto(`/plans/${newPlan.id}`);
		} catch (err) {
			toast.error('Failed to duplicate plan');
		}
	}

	function cancelEdit() {
		editing = false;
		editForm = { ...plan };
	}

	function getStatusIcon(status: string) {
		return status === 'complete' ? CheckCircle2 : status === 'flagged' ? Flag : AlertCircle;
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'complete':
				return 'bg-emerald-100 text-emerald-700 border-emerald-200';
			case 'flagged':
				return 'bg-red-100 text-red-700 border-red-200';
			default:
				return 'bg-amber-100 text-amber-700 border-amber-200';
		}
	}

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

	// Slot management functions
	function openSlotUpload(slotKey: string, slotLabel: string, existingFile?: FileType) {
		uploadSlotKey = slotKey;
		uploadSlotLabel = slotLabel;
		uploadFile = null;
		uploadSlotOpen = true;
	}

	function handleSlotFileSelect(event: Event) {
		const input = event.target as HTMLInputElement;
		if (input.files && input.files[0]) {
			uploadFile = input.files[0];
		}
	}

	async function handleSlotUpload() {
		if (!uploadFile || !planId) return;

		uploadLoading = true;
		try {
			await uploadWebsiteFile(planId, uploadSlotKey, uploadFile);
			toast.success(`${uploadSlotLabel} uploaded successfully`);
			uploadSlotOpen = false;
			uploadFile = null;
			await loadPlan();
		} catch (err) {
			toast.error(`Failed to upload ${uploadSlotLabel}`);
		} finally {
			uploadLoading = false;
		}
	}

	async function openImagePreview(file: FileType) {
		previewFile = file;
		previewOpen = true;
		previewLoading = true;
		previewUrl = null;

		try {
			const result = await getFileUrl(file.id);
			previewUrl = result.url;
		} catch (err) {
			toast.error('Failed to load image preview');
			previewOpen = false;
		} finally {
			previewLoading = false;
		}
	}

	function openSlotDelete(file: FileType) {
		deleteSlotFile = file;
		deleteSlotOpen = true;
	}

	async function confirmSlotDelete() {
		if (!deleteSlotFile) return;

		deleteSlotLoading = true;
		try {
			await deleteFile(deleteSlotFile.id);
			toast.success('Image deleted successfully');
			deleteSlotOpen = false;
			deleteSlotFile = null;
			await loadPlan();
		} catch (err) {
			toast.error('Failed to delete image');
		} finally {
			deleteSlotLoading = false;
		}
	}
</script>

{#if loading}
	<div class="flex h-64 items-center justify-center">
		<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-slate-900"></div>
	</div>
{:else if plan}
	<div class="space-y-6">
		<!-- Header -->
		<div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
			<div class="flex items-center gap-4">
				<Button variant="ghost" size="icon" onclick={() => goto('/plans')}>
					<ArrowLeft class="h-5 w-5" />
				</Button>
				<div>
					{#if editing}
						<Input bind:value={editForm.name} class="h-auto py-1 text-2xl font-bold" />
					{:else}
						<h1 class="text-3xl font-bold text-card-foreground">{plan.name}</h1>
					{/if}
					<div class="mt-1 flex items-center gap-2">
						{#if plan.status}
							{@const StatusIcon = getStatusIcon(plan.status)}
							<Badge variant="secondary" class={getStatusColor(plan.status)}>
								<StatusIcon class="mr-1 h-3 w-3" />
								{plan.status}
							</Badge>
						{/if}
						<span class="text-sm text-muted-foreground">Slug: {plan.slug}</span>
					</div>
				</div>
			</div>

			<div class="flex items-center gap-2">
				{#if editing}
					<Button variant="outline" onclick={cancelEdit}>
						<X class="mr-2 h-4 w-4" />
						Cancel
					</Button>
					<Button onclick={handleSave} disabled={saving}>
						<Save class="mr-2 h-4 w-4" />
						{saving ? 'Saving...' : 'Save Changes'}
					</Button>
				{:else}
					<Button variant="outline" onclick={() => (editing = true)}>
						<Edit2 class="mr-2 h-4 w-4" />
						Edit
					</Button>
					<Button variant="outline" onclick={handleDuplicate}>
						<Copy class="mr-2 h-4 w-4" />
						Duplicate
					</Button>
					<Button variant="outline" onclick={handleFlag}>
						<Flag class="mr-2 h-4 w-4" />
						{plan.status === 'flagged' ? 'Unflag' : 'Flag'}
					</Button>
					{#if $auth?.role === 'admin'}
						<Button variant="destructive" onclick={handleDelete}>
							<Trash2 class="mr-2 h-4 w-4" />
							Delete
						</Button>
					{/if}
				{/if}
			</div>
		</div>

		<!-- Tabs -->
		<div class="border-b border-border">
			<nav class="flex gap-1">
				<button
					class="border-b-2 px-4 py-3 text-sm font-medium transition-colors {activeTab ===
					'overview'
						? 'border-blue-600 text-blue-600'
						: 'border-transparent text-muted-foreground hover:text-card-foreground'}"
					onclick={() => goto(`?tab=overview`, { replaceState: true })}
				>
					<Image class="mr-2 inline h-4 w-4" />
					Overview
				</button>
				<button
					class="border-b-2 px-4 py-3 text-sm font-medium transition-colors {activeTab === 'files'
						? 'border-blue-600 text-blue-600'
						: 'border-transparent text-muted-foreground hover:text-card-foreground'}"
					onclick={() => goto(`?tab=files`, { replaceState: true })}
				>
					<FileText class="mr-2 inline h-4 w-4" />
					Files
				</button>
				<button
					class="border-b-2 px-4 py-3 text-sm font-medium transition-colors {activeTab ===
					'activity'
						? 'border-blue-600 text-blue-600'
						: 'border-transparent text-muted-foreground hover:text-card-foreground'}"
					onclick={() => goto(`?tab=activity`, { replaceState: true })}
				>
					<History class="mr-2 inline h-4 w-4" />
					Activity
				</button>
			</nav>
		</div>

		<!-- Tab Content -->
		{#if activeTab === 'overview'}
			<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
				<!-- Specs Section -->
				<div class="rounded-lg border border-border bg-card p-6">
					<h2 class="mb-4 text-lg font-semibold text-card-foreground">Specifications</h2>
					<div class="grid grid-cols-2 gap-4">
						{#if editing}
							<div>
								<label for="type-input" class="mb-1 block text-sm font-medium text-foreground"
									>Type</label
								>
								<select
									id="type-input"
									bind:value={editForm.type}
									class="w-full rounded-md border px-3 py-2"
								>
									<option value={undefined}>Select type...</option>
									<option value="single_level">Single Level</option>
									<option value="multi_level">Multi Level</option>
								</select>
							</div>
							<div>
								<label for="style-input" class="mb-1 block text-sm font-medium text-foreground"
									>Style</label
								>
								<select
									id="style-input"
									bind:value={editForm.style}
									class="w-full rounded-md border px-3 py-2"
								>
									<option value={undefined}>Select style...</option>
									<option value="cabin">Cabin</option>
									<option value="lodge">Lodge</option>
									<option value="modern">Modern</option>
									<option value="ranch">Ranch</option>
									<option value="farmhouse">Farmhouse</option>
								</select>
							</div>
							<div>
								<label for="beds-input" class="mb-1 block text-sm font-medium text-foreground"
									>Beds</label
								>
								<Input id="beds-input" type="number" bind:value={editForm.beds} />
							</div>
							<div>
								<label for="baths-input" class="mb-1 block text-sm font-medium text-foreground"
									>Baths</label
								>
								<Input id="baths-input" type="number" bind:value={editForm.baths} />
							</div>
							<div>
								<label for="half-baths-input" class="mb-1 block text-sm font-medium text-foreground"
									>Half Baths</label
								>
								<Input id="half-baths-input" type="number" bind:value={editForm.half_baths} />
							</div>
							<div>
								<label for="heated-sf-input" class="mb-1 block text-sm font-medium text-foreground"
									>Heated SF</label
								>
								<Input id="heated-sf-input" type="number" bind:value={editForm.heated_sf} />
							</div>
							<div class="col-span-2">
								<label for="notes-input" class="mb-1 block text-sm font-medium text-foreground"
									>Notes</label
								>
								<textarea
									id="notes-input"
									bind:value={editForm.notes}
									class="w-full rounded-md border px-3 py-2"
									rows="3"
								></textarea>
							</div>
						{:else}
							<div>
								<span class="text-sm text-muted-foreground">Type</span>
								<p class="font-medium">{plan.type?.replace('_', ' ') ?? 'Not specified'}</p>
							</div>
							<div>
								<span class="text-sm text-muted-foreground">Style</span>
								<p class="font-medium capitalize">{plan.style ?? 'Not specified'}</p>
							</div>
							<div>
								<span class="text-sm text-muted-foreground">Beds</span>
								<p class="font-medium">{plan.beds ?? '-'}</p>
							</div>
							<div>
								<span class="text-sm text-muted-foreground">Baths</span>
								<p class="font-medium">
									{plan.baths ?? '-'}{plan.half_baths ? `.${plan.half_baths}` : ''}
								</p>
							</div>
							<div>
								<span class="text-sm text-muted-foreground">Heated SF</span>
								<p class="font-medium">{plan.heated_sf?.toLocaleString() ?? '-'}</p>
							</div>
							<div>
								<span class="text-sm text-muted-foreground">Total SF</span>
								<p class="font-medium">{plan.total_sf?.toLocaleString() ?? '-'}</p>
							</div>
							{#if plan.notes}
								<div class="col-span-2">
									<span class="text-sm text-muted-foreground">Notes</span>
									<p class="mt-1 text-foreground">{plan.notes}</p>
								</div>
							{/if}
						{/if}
					</div>
				</div>

				<!-- Website Images Section -->
				<div class="rounded-lg border border-border bg-card p-6">
					<h2 class="mb-4 text-lg font-semibold text-card-foreground">Website Images</h2>
					<div class="grid grid-cols-2 gap-4 sm:grid-cols-3">
						{#each websiteSlots as slot}
							{@const file = plan.files.website[slot.key as keyof typeof plan.files.website]}
							{#if file}
								<!-- Filled slot with image preview and actions -->
								{@const imageUrl = slotImageUrls[slot.key]}
								<div
									class="group relative aspect-square overflow-hidden rounded-lg border border-border bg-muted"
								>
									<button class="h-full w-full" onclick={() => openImagePreview(file)}>
										{#if imageUrl}
											<img src={imageUrl} alt={slot.label} class="h-full w-full object-cover" />
										{:else}
											<div class="flex h-full w-full flex-col items-center justify-center">
												<Image class="mx-auto h-12 w-12 text-muted-foreground" />
												<p class="mt-2 text-xs text-muted-foreground">Loading...</p>
											</div>
										{/if}
									</button>
									<!-- Overlay with actions -->
									<div
										class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-black/60 opacity-0 transition-opacity group-hover:opacity-100"
									>
										<p class="px-2 text-center text-xs font-medium text-white">{slot.label}</p>
										<div class="flex gap-2">
											<Button
												size="sm"
												variant="secondary"
												class="h-8 w-8 p-0"
												onclick={() => openImagePreview(file)}
											>
												<Eye class="h-4 w-4" />
											</Button>
											<Button
												size="sm"
												variant="secondary"
												class="h-8 w-8 p-0"
												onclick={() => openSlotUpload(slot.key, slot.label, file)}
											>
												<Upload class="h-4 w-4" />
											</Button>
											<Button
												size="sm"
												variant="destructive"
												class="h-8 w-8 p-0"
												onclick={() => openSlotDelete(file)}
											>
												<Trash2 class="h-4 w-4" />
											</Button>
										</div>
									</div>
								</div>
							{:else}
								<!-- Empty slot -->
								<button
									class="flex aspect-square cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-input bg-muted p-4 transition-colors hover:border-blue-400 hover:bg-blue-50"
									onclick={() => openSlotUpload(slot.key, slot.label)}
								>
									<div class="text-center">
										<div
											class="mx-auto mb-2 flex h-8 w-8 items-center justify-center rounded-full bg-slate-200 transition-colors group-hover:bg-blue-200"
										>
											<Upload class="h-4 w-4 text-muted-foreground" />
										</div>
										<p class="text-xs text-muted-foreground">{slot.label}</p>
										<p class="mt-1 text-[10px] text-muted-foreground">Click to upload</p>
									</div>
								</button>
							{/if}
						{/each}
					</div>
				</div>
			</div>
		{:else if activeTab === 'files'}
			<FilesTab {planId} onUploadComplete={loadPlan} />
		{:else if activeTab === 'activity'}
			<div class="rounded-lg border border-border bg-card">
				<div class="border-b border-border p-6">
					<h2 class="text-lg font-semibold text-card-foreground">Activity History</h2>
				</div>
				{#if activities}
					<div class="divide-y divide-slate-100">
						{#each activities.data as activity}
							<div class="flex items-start gap-3 p-4">
								<div
									class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-muted"
								>
									<span class="text-sm font-medium text-muted-foreground">
										{activity.user?.name?.charAt(0).toUpperCase() ?? '?'}
									</span>
								</div>
								<div class="flex-1">
									<p class="text-sm text-card-foreground">
										<span class="font-medium">{activity.user?.name ?? 'Unknown'}</span>
										<span class="text-muted-foreground">{activity.action.replace(/\./g, ' ')}</span>
									</p>
									<p class="mt-1 text-xs text-muted-foreground">
										{new Date(activity.created_at).toLocaleString()}
									</p>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<div class="p-8 text-center text-muted-foreground">Loading activity...</div>
				{/if}
			</div>
		{/if}
	</div>
{/if}

<!-- Image Preview Modal -->
<Dialog bind:open={previewOpen}>
	<DialogContent class="max-w-4xl">
		<DialogHeader>
			<DialogTitle>{previewFile?.filename ?? 'Image Preview'}</DialogTitle>
		</DialogHeader>
		<div class="flex aspect-video items-center justify-center overflow-hidden rounded-lg bg-muted">
			{#if previewLoading}
				<div class="flex flex-col items-center gap-2">
					<Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
					<p class="text-sm text-muted-foreground">Loading image...</p>
				</div>
			{:else if previewUrl}
				<img
					src={previewUrl}
					alt={previewFile?.filename ?? 'Preview'}
					class="max-h-full max-w-full object-contain"
				/>
			{:else}
				<div class="text-center">
					<Image class="mx-auto h-12 w-12 text-muted-foreground" />
					<p class="mt-2 text-sm text-muted-foreground">Failed to load image</p>
				</div>
			{/if}
		</div>
		<DialogFooter>
			<Button variant="outline" onclick={() => (previewOpen = false)}>Close</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>

<!-- Slot Upload Modal -->
<Dialog bind:open={uploadSlotOpen}>
	<DialogContent>
		<DialogHeader>
			<DialogTitle>
				{#if plan?.files.website[uploadSlotKey as keyof typeof plan.files.website]}
					Replace {uploadSlotLabel}
				{:else}
					Upload {uploadSlotLabel}
				{/if}
			</DialogTitle>
			<DialogDescription>
				<div class="space-y-1">
					<p>Choose an image file to upload</p>
					<p class="text-xs text-muted-foreground">
						Max 5MB • Resized to max 4000px • Output: JPEG 90% quality
					</p>
					<p class="text-xs text-amber-600">PNG with transparency → saved to "Other Files"</p>
				</div>
			</DialogDescription>
		</DialogHeader>
		<div class="space-y-4 py-4">
			<input
				type="file"
				accept=".jpg,.jpeg,.png,image/jpeg,image/png"
				onchange={handleSlotFileSelect}
				class="block w-full text-sm text-muted-foreground file:mr-4 file:rounded-full file:border-0 file:bg-blue-50 file:px-4 file:py-2 file:text-sm file:font-semibold file:text-blue-600 hover:file:bg-blue-100"
			/>
			{#if uploadFile}
				<div class="space-y-1 text-sm">
					<p class="text-muted-foreground">
						Selected: <span class="font-medium">{uploadFile.name}</span>
					</p>
					<p
						class={uploadFile.size > 5 * 1024 * 1024
							? 'font-medium text-red-600'
							: 'text-muted-foreground'}
					>
						Size: {(uploadFile.size / 1024 / 1024).toFixed(2)} MB
						{#if uploadFile.size > 5 * 1024 * 1024}
							(⚠️ Exceeds 5MB limit - will be compressed)
						{/if}
					</p>
					{#if uploadFile.name.match(/\.png$/i)}
						<p class="text-xs text-amber-600">
							⚠️ PNG file will be converted to JPEG (transparency → white background)
						</p>
					{/if}
				</div>
			{/if}
		</div>
		<DialogFooter>
			<Button variant="outline" onclick={() => (uploadSlotOpen = false)} disabled={uploadLoading}>
				Cancel
			</Button>
			<Button onclick={handleSlotUpload} disabled={!uploadFile || uploadLoading}>
				{#if uploadLoading}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
					Uploading...
				{:else}
					<Upload class="mr-2 h-4 w-4" />
					Upload
				{/if}
			</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>

<!-- Slot Delete Confirmation -->
<Dialog bind:open={deleteSlotOpen}>
	<DialogContent>
		<DialogHeader>
			<DialogTitle class="flex items-center gap-2 text-red-600">
				<Trash2 class="h-5 w-5" />
				Delete Image?
			</DialogTitle>
			<DialogDescription>
				Are you sure you want to delete "{deleteSlotFile?.filename}"? This action cannot be undone.
			</DialogDescription>
		</DialogHeader>
		<DialogFooter>
			<Button
				variant="outline"
				onclick={() => (deleteSlotOpen = false)}
				disabled={deleteSlotLoading}
			>
				Cancel
			</Button>
			<Button variant="destructive" onclick={confirmSlotDelete} disabled={deleteSlotLoading}>
				{#if deleteSlotLoading}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
					Deleting...
				{:else}
					<Trash2 class="mr-2 h-4 w-4" />
					Delete
				{/if}
			</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>

<!-- Delete Plan Confirmation -->
<ConfirmationDialog
	bind:open={deletePlanOpen}
	title="Delete Plan"
	description={`Are you sure you want to delete "${plan?.name}"? This action cannot be undone.`}
	confirmLabel="Delete Plan"
	cancelLabel="Cancel"
	confirmVariant="destructive"
	onConfirm={confirmDelete}
	onCancel={() => (deletePlanOpen = false)}
/>
