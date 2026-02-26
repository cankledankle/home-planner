// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

declare module 'svelte-sonner' {
	import { SvelteComponent } from 'svelte';

	export class Toaster extends SvelteComponent<{
		position?:
			| 'top-left'
			| 'top-right'
			| 'bottom-left'
			| 'bottom-right'
			| 'top-center'
			| 'bottom-center';
		richColors?: boolean;
	}> {}

	export function toast(message: string, options?: { description?: string }): void;
	export namespace toast {
		function success(message: string, options?: { description?: string }): void;
		function error(message: string, options?: { description?: string }): void;
	}
}

export {};
