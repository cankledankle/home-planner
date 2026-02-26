<script lang="ts">
	interface Props {
		count?: number;
	}

	let { count = 5 }: Props = $props();
</script>

<div class="space-y-2">
	{#each Array(count) as _, i}
		<div class="flex items-center gap-4 p-4">
			<div class="h-4 w-4 animate-pulse rounded bg-slate-200"></div>
			<div class="flex-1">
				<div class="h-4 w-3/4 animate-pulse rounded bg-slate-200"></div>
			</div>
			<div class="h-6 w-20 animate-pulse rounded bg-slate-200"></div>
			<div class="h-4 w-16 animate-pulse rounded bg-slate-200"></div>
			<div class="h-4 w-16 animate-pulse rounded bg-slate-200"></div>
			<div class="h-4 w-12 animate-pulse rounded bg-slate-200"></div>
		</div>
	{/each}
</div>