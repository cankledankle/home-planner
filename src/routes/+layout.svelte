<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { QueryClient, QueryClientProvider } from '@tanstack/svelte-query';
	import { theme } from '$lib/stores/theme';

	const queryClient = new QueryClient();
	let { children } = $props();

	// Initialize theme store (needed for all pages including login)
	theme.subscribe(() => {});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<QueryClientProvider client={queryClient}>
	{@render children()}
</QueryClientProvider>
