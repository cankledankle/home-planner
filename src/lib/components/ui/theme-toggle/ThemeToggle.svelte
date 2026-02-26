<script lang="ts">
	import { theme, type Theme } from '$lib/stores/theme';
	import { Sun, Moon, Monitor, ChevronDown } from '@lucide/svelte';

	let isOpen = $state(false);

	const themes: { value: Theme; label: string; icon: typeof Sun }[] = [
		{ value: 'system', label: 'System', icon: Monitor },
		{ value: 'light', label: 'Light', icon: Sun },
		{ value: 'dark', label: 'Dark', icon: Moon }
	];

	function handleSelect(selectedTheme: Theme) {
		theme.setTheme(selectedTheme);
		isOpen = false;
	}

	function handleKeydown(event: KeyboardEvent, selectedTheme: Theme) {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			handleSelect(selectedTheme);
		}
	}

	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (!target.closest('.theme-toggle-container')) {
			isOpen = false;
		}
	}
</script>

<svelte:window onclick={handleClickOutside} />

<div class="theme-toggle-container relative">
	<button
		type="button"
		onclick={() => (isOpen = !isOpen)}
		class="flex w-full items-center justify-between gap-2 rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
		aria-expanded={isOpen}
		aria-haspopup="listbox"
	>
		<span class="flex items-center gap-2">
			{#if $theme.theme === 'system'}
				<Monitor class="h-4 w-4" />
			{:else if $theme.theme === 'light'}
				<Sun class="h-4 w-4" />
			{:else}
				<Moon class="h-4 w-4" />
			{/if}
			Theme
		</span>
		<ChevronDown class="h-4 w-4 transition-transform {isOpen ? 'rotate-180' : ''}" />
	</button>

	{#if isOpen}
		<ul
			class="absolute right-0 bottom-full left-0 mb-1 overflow-hidden rounded-lg border border-border bg-popover shadow-lg"
			role="listbox"
			aria-label="Select theme"
		>
			{#each themes as themeOption}
				{@const Icon = themeOption.icon}
				<li>
					<button
						type="button"
						role="option"
						aria-selected={$theme.theme === themeOption.value}
						class="flex w-full items-center gap-2 px-3 py-2 text-sm transition-colors hover:bg-accent {$theme.theme ===
						themeOption.value
							? 'bg-accent text-accent-foreground'
							: 'text-popover-foreground'}"
						onclick={() => handleSelect(themeOption.value)}
						onkeydown={(e) => handleKeydown(e, themeOption.value)}
					>
						<Icon class="h-4 w-4" />
						{themeOption.label}
						{#if $theme.theme === themeOption.value}
							<span class="ml-auto text-xs text-muted-foreground">Active</span>
						{/if}
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
