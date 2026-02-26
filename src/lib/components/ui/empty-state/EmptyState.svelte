<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import type { Component } from 'svelte';

	interface Props {
		title: string;
		description?: string;
		icon?: Component;
		actionLabel?: string;
		secondaryActionLabel?: string;
		onAction?: () => void;
		onSecondaryAction?: () => void;
	}

	let { 
		title, 
		description, 
		icon: Icon, 
		actionLabel,
		secondaryActionLabel,
		onAction,
		onSecondaryAction
	}: Props = $props();
</script>

<div class="flex flex-col items-center justify-center rounded-lg border border-slate-200 bg-white p-12 text-center">
	{#if Icon}
		<div class="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-slate-100">
			<Icon class="h-8 w-8 text-slate-400" />
		</div>
	{/if}
	
	<h3 class="mb-2 text-lg font-medium text-slate-900">
		{title}
	</h3>
	
	{#if description}
		<p class="mb-6 max-w-md text-slate-600">
			{description}
		</p>
	{/if}
	
	<div class="flex gap-3">
		{#if actionLabel && onAction}
			<Button onclick={onAction}>{actionLabel}</Button>
		{/if}
		{#if secondaryActionLabel && onSecondaryAction}
			<Button variant="outline" onclick={onSecondaryAction}>{secondaryActionLabel}</Button>
		{/if}
	</div>
</div>