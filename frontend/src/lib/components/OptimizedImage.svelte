<script lang="ts">
	import type { ImageSource } from '$lib/types/media';

	interface Props {
		src: string | undefined | null;
		sources?: ImageSource[];
		alt?: string;
		class?: string;
		sizes?: string;
		loading?: 'lazy' | 'eager';
		fetchpriority?: 'high' | 'low' | 'auto';
		width?: number | string;
		height?: number | string;
		draggable?: boolean;
	}

	let {
		src,
		sources = [],
		alt = '',
		class: className = '',
		sizes = '100vw',
		loading = 'lazy',
		fetchpriority = 'auto',
		width,
		height,
		draggable = false
	}: Props = $props();

	// Computed srcset string
	const srcset = $derived.by(() => {
		if (!sources || sources.length === 0) return '';
		return sources.map((s) => `${s.url} ${s.width}w`).join(', ');
	});

	// Fallback to src if provided, or empty string
	const finalSrc = $derived(src || '');

	// If we have sources, we use them for srcset. The browser will pick the best one.
	// If not, we just use the original src.
</script>

{#if finalSrc}
	<img
		src={finalSrc}
		{srcset}
		{sizes}
		{alt}
		class={className}
		{loading}
		{fetchpriority}
		{width}
		{height}
		{draggable}
		decoding="async"
	/>
{:else}
	<div class="flex items-center justify-center bg-surface-lowest {className}" style:width style:height>
		<span class="material-symbols-outlined text-outline text-4xl opacity-20">image</span>
	</div>
{/if}
