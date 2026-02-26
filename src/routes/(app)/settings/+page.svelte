<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		Dialog,
		DialogContent,
		DialogHeader,
		DialogTitle,
		DialogTrigger,
		DialogFooter
	} from '$lib/components/ui/dialog';

	import { auth } from '$lib/stores';
	import { getUsers, createUser, updateUser, deleteUser, getSFTPStatus, getUserSFTP, generateSFTP, rotateSFTP, revokeSFTP, updateSFTPPermission, deleteSFTP } from '$lib/api';
	import { toast } from 'svelte-sonner';
	import {
		Users,
		Plus,
		Edit2,
		Trash2,
		Shield,
		User as UserIcon,
		Loader2,
		X,
		Check,
		FileSpreadsheet,
		Download,
		ChevronDown,
		ChevronUp,
		Server,
		Key,
		RefreshCw,
		Lock,
		Unlock,
		Eye,
		EyeOff,
		Copy
	} from '@lucide/svelte';
	import type { User, SFTPCredentials, SFTPStatus } from '$lib/types';
	import EmptyState from '$lib/components/ui/empty-state/EmptyState.svelte';
	import ConfirmationDialog from '$lib/components/ui/dialog/ConfirmationDialog.svelte';

	// Data
	let users = $state<User[]>([]);
	let loading = $state(true);

	// Dialog states
	let createDialogOpen = $state(false);
	let editDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);
	let selectedUser = $state<User | null>(null);

	// Form states
	let formData = $state({
		name: '',
		email: '',
		role: 'editor' as 'admin' | 'editor',
		password: ''
	});
	let formLoading = $state(false);

	// SFTP states
	let sftpStatus = $state<SFTPStatus | null>(null);
	let sftpCredentials = $state<Record<string, SFTPCredentials>>({});
	let sftpLoading = $state(false);
	let sftpSelectedUser = $state<User | null>(null);
	let sftpDialogOpen = $state(false);
	let showPassword = $state(false);
	let newCredentials = $state<SFTPCredentials | null>(null);

	onMount(() => {
		loadUsers();
		loadSFTPStatus();
	});

	async function loadSFTPStatus() {
		try {
			sftpStatus = await getSFTPStatus();
		} catch (err) {
			sftpStatus = { configured: false, message: 'SFTP service unavailable' };
		}
	}

	async function loadUserSFTP(userId: string) {
		if (!sftpStatus?.configured) return;
		try {
			const creds = await getUserSFTP(userId);
			if (creds) {
				sftpCredentials[userId] = creds;
			}
		} catch (err) {
			// User doesn't have SFTP credentials yet
		}
	}

	function openSFTPDialog(user: User) {
		sftpSelectedUser = user;
		newCredentials = null;
		showPassword = false;
		sftpDialogOpen = true;
		loadUserSFTP(user.id);
	}

	async function handleGenerateSFTP(permission: 'read' | 'readwrite') {
		if (!sftpSelectedUser) return;
		sftpLoading = true;
		try {
			const creds = await generateSFTP(sftpSelectedUser.id, permission);
			sftpCredentials[sftpSelectedUser.id] = creds;
			newCredentials = creds;
			toast.success('SFTP credentials generated');
		} catch (err) {
			toast.error('Failed to generate SFTP credentials');
		} finally {
			sftpLoading = false;
		}
	}

	async function handleRotateSFTP() {
		if (!sftpSelectedUser) return;
		sftpLoading = true;
		try {
			const creds = await rotateSFTP(sftpSelectedUser.id);
			sftpCredentials[sftpSelectedUser.id] = creds;
			newCredentials = creds;
			showPassword = true;
			toast.success('SFTP password rotated');
		} catch (err) {
			toast.error('Failed to rotate SFTP credentials');
		} finally {
			sftpLoading = false;
		}
	}

	async function handleRevokeSFTP() {
		if (!sftpSelectedUser) return;
		sftpLoading = true;
		try {
			await revokeSFTP(sftpSelectedUser.id);
			sftpCredentials[sftpSelectedUser.id] = {
				...sftpCredentials[sftpSelectedUser.id],
				permission: 'read'
			};
			toast.success('SFTP access revoked');
		} catch (err) {
			toast.error('Failed to revoke SFTP access');
		} finally {
			sftpLoading = false;
		}
	}

	async function handleUpdateSFTPPermission(permission: 'read' | 'readwrite') {
		if (!sftpSelectedUser) return;
		sftpLoading = true;
		try {
			await updateSFTPPermission(sftpSelectedUser.id, permission);
			sftpCredentials[sftpSelectedUser.id] = {
				...sftpCredentials[sftpSelectedUser.id],
				permission
			};
			toast.success(`SFTP permission updated to ${permission}`);
		} catch (err) {
			toast.error('Failed to update SFTP permission');
		} finally {
			sftpLoading = false;
		}
	}

	async function handleDeleteSFTP() {
		if (!sftpSelectedUser) return;
		sftpLoading = true;
		try {
			await deleteSFTP(sftpSelectedUser.id);
			delete sftpCredentials[sftpSelectedUser.id];
			toast.success('SFTP credentials deleted');
			sftpDialogOpen = false;
		} catch (err) {
			toast.error('Failed to delete SFTP credentials');
		} finally {
			sftpLoading = false;
		}
	}

	function copyToClipboard(text: string) {
		navigator.clipboard.writeText(text);
		toast.success('Copied to clipboard');
	}

	async function loadUsers() {
		loading = true;
		try {
			users = await getUsers();
		} catch (err) {
			toast.error('Failed to load users');
		} finally {
			loading = false;
		}
	}

	function resetForm() {
		formData = {
			name: '',
			email: '',
			role: 'editor',
			password: ''
		};
	}

	function openCreateDialog() {
		resetForm();
		createDialogOpen = true;
	}

	function openEditDialog(user: User) {
		selectedUser = user;
		formData = {
			name: user.name,
			email: user.email,
			role: user.role,
			password: ''
		};
		editDialogOpen = true;
	}

	function openDeleteDialog(user: User) {
		selectedUser = user;
		deleteDialogOpen = true;
	}

	async function handleCreate() {
		if (!formData.name || !formData.email || !formData.password) {
			toast.error('Please fill in all required fields');
			return;
		}

		formLoading = true;
		try {
			const result = await createUser({
				name: formData.name,
				email: formData.email,
				role: formData.role,
				password: formData.password
			});
			toast.success('User created successfully');
			if (result && 'sftp_credentials' in result && result.sftp_credentials) {
				toast.success('SFTP credentials generated automatically');
			}
			createDialogOpen = false;
			resetForm();
			await loadUsers();
		} catch (err) {
			toast.error('Failed to create user');
		} finally {
			formLoading = false;
		}
	}

	async function handleUpdate() {
		if (!selectedUser) return;

		formLoading = true;
		try {
			const updateData: Partial<User> & { password?: string } = {
				name: formData.name,
				email: formData.email,
				role: formData.role
			};

			// Only include password if it was provided
			if (formData.password) {
				updateData.password = formData.password;
			}

			await updateUser(selectedUser.id, updateData);
			toast.success('User updated successfully');
			editDialogOpen = false;
			selectedUser = null;
			resetForm();
			await loadUsers();
		} catch (err) {
			toast.error('Failed to update user');
		} finally {
			formLoading = false;
		}
	}

	async function handleDelete() {
		if (!selectedUser) return;

		formLoading = true;
		try {
			await deleteUser(selectedUser.id);
			toast.success('User deleted successfully');
			deleteDialogOpen = false;
			selectedUser = null;
			await loadUsers();
		} catch (err) {
			toast.error('Failed to delete user');
		} finally {
			formLoading = false;
		}
	}

	function getRoleIcon(role: string) {
		return role === 'admin' ? Shield : UserIcon;
	}

	function getRoleColor(role: string) {
		return role === 'admin' ? 'bg-purple-100 text-purple-700' : 'bg-slate-100 text-slate-700';
	}

	// Export preset reference
	let expandedPreset = $state<string | null>(null);

	import {
		EXPORT_PRESETS,
		EXPORT_PRESET_FIELDS,
		EXPORT_FIELDS
	} from '$lib/api';

	const exportPresets = [
		{
			id: EXPORT_PRESETS.wpAllImport,
			name: 'WP All Import',
			description: 'Optimized for WordPress All Import plugin with all required fields and image slots',
			fields: EXPORT_PRESET_FIELDS[EXPORT_PRESETS.wpAllImport]
		},
		{
			id: 'general',
			name: 'General',
			description: 'Standard export with all plan metadata and file URLs',
			fields: [
				'id',
				'name',
				'slug',
				'status',
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
				'heated_sf',
				'total_sf',
				'notes'
			]
		},
		{
			id: 'minimal',
			name: 'Minimal',
			description: 'Basic plan info only - name, slug, and status',
			fields: ['id', 'name', 'slug', 'status']
		}
	];

	const allExportFields = [
		{ name: 'id', label: 'ID', description: 'Unique plan identifier' },
		{ name: 'name', label: 'Name', description: 'Plan display name' },
		{ name: 'slug', label: 'Slug', description: 'URL-friendly identifier' },
		{ name: 'status', label: 'Status', description: 'complete, incomplete, or flagged' },
		{ name: 'type', label: 'Type', description: 'single_level or multi_level' },
		{ name: 'style', label: 'Style', description: 'cabin, lodge, modern, ranch, farmhouse' },
		{ name: 'beds', label: 'Beds', description: 'Number of bedrooms' },
		{ name: 'baths', label: 'Baths', description: 'Number of full bathrooms' },
		{ name: 'half_baths', label: 'Half Baths', description: 'Number of half bathrooms' },
		{ name: 'main_sf', label: 'Main SF', description: 'Main floor square footage' },
		{ name: 'upper_sf', label: 'Upper SF', description: 'Upper floor square footage' },
		{ name: 'lower_sf', label: 'Lower SF', description: 'Lower floor square footage' },
		{ name: 'porch_deck_sf', label: 'Porch/Deck SF', description: 'Porch and deck square footage' },
		{ name: 'garage_sf', label: 'Garage SF', description: 'Garage square footage' },
		{
			name: 'garage_apartment_sf',
			label: 'Garage Apt SF',
			description: 'Garage apartment square footage'
		},
		{
			name: 'unfinished_sf',
			label: 'Unfinished SF',
			description: 'Unfinished space square footage'
		},
		{ name: 'heated_sf', label: 'Heated SF', description: 'Total heated square footage' },
		{ name: 'total_sf', label: 'Total SF', description: 'Total square footage' },
		{ name: 'notes', label: 'Notes', description: 'Plan notes and comments' },
		{ name: 'poster_url', label: 'Poster URL', description: 'URL to poster image' },
		{ name: 'render_front_url', label: 'Render Front URL', description: 'URL to front render' }
	];

	function togglePreset(presetId: string) {
		expandedPreset = expandedPreset === presetId ? null : presetId;
	}
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-3xl font-bold text-slate-900">Settings</h1>
			<p class="mt-1 text-slate-600">Manage users and system settings</p>
		</div>
	</div>

	<!-- Admin-only: User Management -->
	{#if $auth?.role === 'admin'}
		<div class="rounded-lg border border-slate-200 bg-white">
			<!-- Section Header -->
			<div class="flex items-center justify-between border-b border-slate-200 p-6">
				<div class="flex items-center gap-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100">
						<Users class="h-5 w-5 text-blue-600" />
					</div>
					<div>
						<h2 class="text-lg font-semibold text-slate-900">User Management</h2>
						<p class="text-sm text-slate-600">Manage system users and permissions</p>
					</div>
				</div>
				<Button onclick={openCreateDialog}>
					<Plus class="mr-2 h-4 w-4" />
					Add User
				</Button>
			</div>

			<!-- Users List -->
			{#if loading}
				<div class="p-8 text-center">
					<Loader2 class="mx-auto h-8 w-8 animate-spin text-slate-400" />
					<p class="mt-2 text-sm text-slate-500">Loading users...</p>
				</div>
			{:else if users.length === 0}
				<div class="p-8 text-center">
					<Users class="mx-auto mb-4 h-12 w-12 text-slate-300" />
					<h3 class="mb-2 text-lg font-medium text-slate-900">No users found</h3>
					<p class="text-slate-600">Add your first user to get started</p>
				</div>
			{:else}
				<div class="divide-y divide-slate-100">
				{#each users as user}
					{@const RoleIcon = getRoleIcon(user.role)}
					<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 p-4 hover:bg-slate-50">
						<div class="flex items-center gap-3 min-w-0">
							<div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-slate-100">
								<span class="text-sm font-medium text-slate-600">
									{user.name.charAt(0).toUpperCase()}
								</span>
							</div>
							<div class="min-w-0 flex-1">
								<p class="font-medium text-slate-900 truncate">{user.name}</p>
								<p class="text-sm text-slate-500 truncate">{user.email}</p>
							</div>
						</div>
						<div class="flex items-center justify-between sm:justify-end gap-2 sm:gap-3">
							<div class="flex items-center gap-2">
								<span
									class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium {getRoleColor(
										user.role
									)}"
								>
									<RoleIcon class="h-3 w-3" />
									{user.role}
								</span>
								{#if user.id === $auth?.id}
									<span class="text-xs text-slate-400">(You)</span>
								{/if}
							</div>
							<div class="flex items-center gap-1">
								<Button
									variant="ghost"
									size="sm"
									onclick={() => openEditDialog(user)}
									disabled={user.id === $auth?.id}
									title={user.id === $auth?.id ? 'Cannot edit yourself' : 'Edit user'}
								>
									<Edit2 class="h-4 w-4" />
								</Button>
								<Button
									variant="ghost"
									size="sm"
									class="text-red-600 hover:text-red-700"
									onclick={() => openDeleteDialog(user)}
									disabled={user.id === $auth?.id}
									title={user.id === $auth?.id ? 'Cannot delete yourself' : 'Delete user'}
								>
									<Trash2 class="h-4 w-4" />
								</Button>
							</div>
						</div>
					</div>
				{/each}
				</div>
			{/if}
		</div>

		<!-- SFTP Credentials Management -->
		{#if sftpStatus?.configured}
			<div class="rounded-lg border border-slate-200 bg-white">
				<div class="border-b border-slate-200 p-6">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-3">
							<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-orange-100">
								<Server class="h-5 w-5 text-orange-600" />
							</div>
							<div>
								<h2 class="text-lg font-semibold text-slate-900">SFTP Access</h2>
								<p class="text-sm text-slate-600">Manage SFTP credentials for user file access</p>
							</div>
						</div>
						{#if sftpStatus.healthy}
							<span class="inline-flex items-center gap-1 rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-700">
								<Check class="h-3 w-3" />
								Connected
							</span>
						{:else}
							<span class="inline-flex items-center gap-1 rounded-full bg-red-100 px-2.5 py-0.5 text-xs font-medium text-red-700">
								<X class="h-3 w-3" />
								Disconnected
							</span>
						{/if}
					</div>
				</div>
				{#if loading}
					<div class="p-8 text-center">
						<Loader2 class="mx-auto h-8 w-8 animate-spin text-slate-400" />
						<p class="mt-2 text-sm text-slate-500">Loading users...</p>
					</div>
				{:else}
				<div class="divide-y divide-slate-100">
					{#each users as user}
						<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 p-4 hover:bg-slate-50">
							<div class="flex items-center gap-3 min-w-0">
								<div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-slate-100">
									<span class="text-sm font-medium text-slate-600">
										{user.name.charAt(0).toUpperCase()}
									</span>
								</div>
								<div class="min-w-0 flex-1">
									<p class="font-medium text-slate-900 truncate">{user.name}</p>
									<p class="text-sm text-slate-500 truncate">{user.email}</p>
								</div>
								{#if sftpCredentials[user.id]}
									<span class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium {sftpCredentials[user.id].permission === 'readwrite' ? 'bg-blue-100 text-blue-700' : 'bg-slate-100 text-slate-700'}">
										<Key class="h-3 w-3" />
										{sftpCredentials[user.id].permission === 'readwrite' ? 'Read/Write' : 'Read Only'}
									</span>
								{:else}
									<span class="inline-flex items-center gap-1 rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-medium text-slate-500">
										No SFTP
									</span>
								{/if}
							</div>
							<Button
								variant="outline"
								size="sm"
								onclick={() => openSFTPDialog(user)}
								class="w-full sm:w-auto"
							>
								<Key class="mr-2 h-4 w-4" />
								{sftpCredentials[user.id] ? 'Manage' : 'Create'}
							</Button>
						</div>
					{/each}
				</div>
				{/if}
			</div>
		{/if}

		<!-- Export Preset Reference -->
		<div class="rounded-lg border border-slate-200 bg-white">
			<div class="border-b border-slate-200 p-6">
				<div class="flex items-center gap-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-100">
						<FileSpreadsheet class="h-5 w-5 text-emerald-600" />
					</div>
					<div>
						<h2 class="text-lg font-semibold text-slate-900">Export Preset Reference</h2>
						<p class="text-sm text-slate-600">Available CSV export presets and field definitions</p>
					</div>
				</div>
			</div>
			<div class="divide-y divide-slate-100">
				{#each exportPresets as preset}
					<div class="p-4">
						<button
							class="flex w-full items-center justify-between text-left"
							onclick={() => togglePreset(preset.id)}
						>
							<div>
								<h3 class="font-medium text-slate-900">{preset.name}</h3>
								<p class="text-sm text-slate-500">{preset.description}</p>
							</div>
							{#if expandedPreset === preset.id}
								<ChevronUp class="h-5 w-5 text-slate-400" />
							{:else}
								<ChevronDown class="h-5 w-5 text-slate-400" />
							{/if}
						</button>
						{#if expandedPreset === preset.id}
							<div class="mt-3 rounded-lg bg-slate-50 p-3">
								<p class="mb-2 text-xs font-medium tracking-wider text-slate-500 uppercase">
									Included Fields ({preset.fields.length})
								</p>
								<div class="flex flex-wrap gap-2">
									{#each preset.fields as field}
										<span
											class="rounded-full border border-slate-200 bg-white px-2.5 py-1 text-xs font-medium text-slate-600 shadow-sm"
										>
											{field}
										</span>
									{/each}
								</div>
							</div>
						{/if}
					</div>
				{/each}
			</div>
			<div class="border-t border-slate-200 p-4">
				<p class="mb-3 text-sm font-medium text-slate-700">All Available Fields</p>
				<div class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
					{#each allExportFields as field}
						<div class="rounded-md bg-slate-50 p-2">
							<p class="text-sm font-medium text-slate-700">{field.label}</p>
							<p class="text-xs text-slate-500">{field.description}</p>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{:else}
		<div class="rounded-lg border border-slate-200 bg-white p-8 text-center">
			<Shield class="mx-auto mb-4 h-12 w-12 text-slate-300" />
			<h3 class="mb-2 text-lg font-medium text-slate-900">Admin Access Required</h3>
			<p class="text-slate-600">Only administrators can access system settings</p>
		</div>
	{/if}
</div>

<!-- Create User Dialog -->
<Dialog bind:open={createDialogOpen}>
	<DialogContent class="sm:max-w-md">
		<DialogHeader>
			<DialogTitle>Create New User</DialogTitle>
		</DialogHeader>
		<div class="space-y-4 py-4">
			<div class="space-y-2">
				<Label for="name">Name</Label>
				<Input
					id="name"
					bind:value={formData.name}
					placeholder="Enter user's name"
					disabled={formLoading}
				/>
			</div>
			<div class="space-y-2">
				<Label for="email">Email</Label>
				<Input
					id="email"
					type="email"
					bind:value={formData.email}
					placeholder="user@example.com"
					disabled={formLoading}
				/>
			</div>
			<div class="space-y-2">
				<Label for="role">Role</Label>
				<select
					id="role"
					bind:value={formData.role}
					class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
					disabled={formLoading}
				>
					<option value="editor">Editor</option>
					<option value="admin">Administrator</option>
				</select>
			</div>
			<div class="space-y-2">
				<Label for="password">Password</Label>
				<Input
					id="password"
					type="password"
					bind:value={formData.password}
					placeholder="Set a secure password"
					disabled={formLoading}
				/>
			</div>
		</div>
		<DialogFooter>
			<Button variant="outline" onclick={() => (createDialogOpen = false)} disabled={formLoading}>
				Cancel
			</Button>
			<Button onclick={handleCreate} disabled={formLoading}>
				{#if formLoading}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
				{/if}
				Create User
			</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>

<!-- Edit User Dialog -->
<Dialog bind:open={editDialogOpen}>
	<DialogContent class="sm:max-w-md">
		<DialogHeader>
			<DialogTitle>Edit User</DialogTitle>
		</DialogHeader>
		<div class="space-y-4 py-4">
			<div class="space-y-2">
				<Label for="edit-name">Name</Label>
				<Input
					id="edit-name"
					bind:value={formData.name}
					placeholder="Enter user's name"
					disabled={formLoading}
				/>
			</div>
			<div class="space-y-2">
				<Label for="edit-email">Email</Label>
				<Input
					id="edit-email"
					type="email"
					bind:value={formData.email}
					placeholder="user@example.com"
					disabled={formLoading}
				/>
			</div>
			<div class="space-y-2">
				<Label for="edit-role">Role</Label>
				<select
					id="edit-role"
					bind:value={formData.role}
					class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
					disabled={formLoading}
				>
					<option value="editor">Editor</option>
					<option value="admin">Administrator</option>
				</select>
			</div>
			<div class="space-y-2">
				<Label for="edit-password">
					New Password
					<span class="text-xs text-slate-500">(leave blank to keep current)</span>
				</Label>
				<Input
					id="edit-password"
					type="password"
					bind:value={formData.password}
					placeholder="Enter new password"
					disabled={formLoading}
				/>
			</div>
		</div>
		<DialogFooter>
			<Button variant="outline" onclick={() => (editDialogOpen = false)} disabled={formLoading}>
				Cancel
			</Button>
			<Button onclick={handleUpdate} disabled={formLoading}>
				{#if formLoading}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
				{/if}
				Save Changes
			</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>

<!-- Delete User Dialog -->
<ConfirmationDialog
	bind:open={deleteDialogOpen}
	title="Delete User"
	description="Are you sure you want to delete {selectedUser?.name}? This action cannot be undone."
	confirmLabel="Delete User"
	cancelLabel="Cancel"
	confirmVariant="destructive"
	onConfirm={handleDelete}
	onCancel={() => deleteDialogOpen = false}
/>

<!-- SFTP Credentials Dialog -->
<Dialog bind:open={sftpDialogOpen}>
	<DialogContent class="sm:max-w-lg">
		<DialogHeader>
			<DialogTitle>SFTP Credentials</DialogTitle>
		</DialogHeader>
		<div class="space-y-4 py-4">
			{#if sftpSelectedUser}
				<div class="flex items-center gap-3 rounded-lg bg-slate-50 p-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-full bg-slate-200">
						<span class="text-sm font-medium text-slate-600">
							{sftpSelectedUser.name.charAt(0).toUpperCase()}
						</span>
					</div>
					<div>
						<p class="font-medium text-slate-900">{sftpSelectedUser.name}</p>
						<p class="text-sm text-slate-500">{sftpSelectedUser.email}</p>
					</div>
				</div>

				{#if newCredentials}
					<!-- New/Rotated Credentials Display -->
					<div class="rounded-lg border border-green-200 bg-green-50 p-4">
						<p class="mb-3 text-sm font-medium text-green-800">New SFTP Credentials Generated</p>
						<div class="space-y-2">
							<div class="flex items-center justify-between">
								<span class="text-sm text-slate-600">Host:</span>
								<div class="flex items-center gap-2">
									<code class="rounded bg-white px-2 py-1 text-sm">{newCredentials.host}</code>
									<Button variant="ghost" size="sm" class="h-6 w-6 p-0" onclick={() => copyToClipboard(newCredentials!.host)}>
										<Copy class="h-3 w-3" />
									</Button>
								</div>
							</div>
							<div class="flex items-center justify-between">
								<span class="text-sm text-slate-600">Port:</span>
								<code class="rounded bg-white px-2 py-1 text-sm">{newCredentials.port}</code>
							</div>
							<div class="flex items-center justify-between">
								<span class="text-sm text-slate-600">Username:</span>
								<div class="flex items-center gap-2">
									<code class="rounded bg-white px-2 py-1 text-sm">{newCredentials.username}</code>
									<Button variant="ghost" size="sm" class="h-6 w-6 p-0" onclick={() => copyToClipboard(newCredentials!.username)}>
										<Copy class="h-3 w-3" />
									</Button>
								</div>
							</div>
							{#if newCredentials.password}
								<div class="flex items-center justify-between">
									<span class="text-sm text-slate-600">Password:</span>
									<div class="flex items-center gap-2">
										<code class="rounded bg-white px-2 py-1 text-sm">
											{showPassword ? newCredentials.password : '••••••••••••••••'}
										</code>
										<Button variant="ghost" size="sm" class="h-6 w-6 p-0" onclick={() => showPassword = !showPassword}>
											{#if showPassword}
												<EyeOff class="h-3 w-3" />
											{:else}
												<Eye class="h-3 w-3" />
											{/if}
										</Button>
										<Button variant="ghost" size="sm" class="h-6 w-6 p-0" onclick={() => copyToClipboard(newCredentials!.password || '')}>
											<Copy class="h-3 w-3" />
										</Button>
									</div>
								</div>
							{/if}
							<div class="flex items-center justify-between">
								<span class="text-sm text-slate-600">Permission:</span>
								<span class="rounded-full px-2 py-0.5 text-xs font-medium {newCredentials.permission === 'readwrite' ? 'bg-blue-100 text-blue-700' : 'bg-slate-100 text-slate-700'}">
									{newCredentials.permission === 'readwrite' ? 'Read/Write' : 'Read Only'}
								</span>
							</div>
						</div>
						<p class="mt-3 text-xs text-green-700">
							Please copy these credentials now. The password will not be shown again.
						</p>
					</div>
				{:else if sftpCredentials[sftpSelectedUser.id]}
					<!-- Existing Credentials Management -->
					<div class="rounded-lg border border-slate-200 bg-slate-50 p-4">
						<p class="mb-3 text-sm font-medium text-slate-900">SFTP Connection Details</p>
						<div class="space-y-2">
							<div class="flex items-center justify-between">
								<span class="text-sm text-slate-600">Host:</span>
								<div class="flex items-center gap-2">
									<code class="rounded bg-white px-2 py-1 text-sm">{sftpCredentials[sftpSelectedUser.id].host}</code>
									<Button variant="ghost" size="sm" class="h-6 w-6 p-0" onclick={() => copyToClipboard(sftpCredentials[sftpSelectedUser!.id].host)}>
										<Copy class="h-3 w-3" />
									</Button>
								</div>
							</div>
							<div class="flex items-center justify-between">
								<span class="text-sm text-slate-600">Port:</span>
								<code class="rounded bg-white px-2 py-1 text-sm">{sftpCredentials[sftpSelectedUser.id].port}</code>
							</div>
							<div class="flex items-center justify-between">
								<span class="text-sm text-slate-600">Username:</span>
								<div class="flex items-center gap-2">
									<code class="rounded bg-white px-2 py-1 text-sm">{sftpCredentials[sftpSelectedUser.id].username}</code>
									<Button variant="ghost" size="sm" class="h-6 w-6 p-0" onclick={() => copyToClipboard(sftpCredentials[sftpSelectedUser!.id].username)}>
										<Copy class="h-3 w-3" />
									</Button>
								</div>
							</div>
							<div class="flex items-center justify-between">
								<span class="text-sm text-slate-600">Permission:</span>
								<span class="rounded-full px-2 py-0.5 text-xs font-medium {sftpCredentials[sftpSelectedUser.id].permission === 'readwrite' ? 'bg-blue-100 text-blue-700' : 'bg-slate-100 text-slate-700'}">
									{sftpCredentials[sftpSelectedUser.id].permission === 'readwrite' ? 'Read/Write' : 'Read Only'}
								</span>
							</div>
						</div>
					</div>

					<div class="grid grid-cols-2 gap-2">
						<Button variant="outline" onclick={() => handleRotateSFTP()} disabled={sftpLoading}>
							<RefreshCw class="mr-2 h-4 w-4" />
							Rotate Password
						</Button>
						{#if sftpCredentials[sftpSelectedUser.id].permission === 'readwrite'}
							<Button variant="outline" onclick={() => handleUpdateSFTPPermission('read')} disabled={sftpLoading}>
								<Lock class="mr-2 h-4 w-4" />
								Set Read-Only
							</Button>
						{:else}
							<Button variant="outline" onclick={() => handleUpdateSFTPPermission('readwrite')} disabled={sftpLoading}>
								<Unlock class="mr-2 h-4 w-4" />
								Allow Write
							</Button>
						{/if}
					</div>

					<div class="grid grid-cols-2 gap-2">
						<Button variant="outline" onclick={() => handleRevokeSFTP()} disabled={sftpLoading}>
							<Lock class="mr-2 h-4 w-4" />
							Revoke Access
						</Button>
						<Button variant="destructive" onclick={() => handleDeleteSFTP()} disabled={sftpLoading}>
							<Trash2 class="mr-2 h-4 w-4" />
							Delete Credentials
						</Button>
					</div>
				{:else}
					<!-- Create New Credentials -->
					<p class="text-sm text-slate-600">This user doesn't have SFTP credentials yet.</p>
					<div class="grid grid-cols-2 gap-2">
						<Button variant="outline" onclick={() => handleGenerateSFTP('read')} disabled={sftpLoading}>
							<Lock class="mr-2 h-4 w-4" />
							Create Read-Only
						</Button>
						<Button onclick={() => handleGenerateSFTP('readwrite')} disabled={sftpLoading}>
							<Unlock class="mr-2 h-4 w-4" />
							Create Read/Write
						</Button>
					</div>
				{/if}
			{/if}
		</div>
		<DialogFooter>
			<Button variant="outline" onclick={() => (sftpDialogOpen = false)} disabled={sftpLoading}>
				Close
			</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>
