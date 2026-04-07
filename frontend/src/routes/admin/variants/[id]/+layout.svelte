<script lang="ts">
  import type { LayoutData } from "./$types";
  import { getSongName } from "$lib/song-utils";
  import { page } from "$app/stores";

  let { data, children } = $props<{ data: LayoutData; children: any }>();
  let variant = $derived(data.variant);
  let song = $derived(variant.song);
  let anime = $derived(song?.anime);

  const tabs = $derived([
    { label: "Overview", href: `/admin/variants/${variant.id}`, match: (path: string) => path === `/admin/variants/${variant.id}` },
    { label: "Edit Info", href: `/admin/variants/${variant.id}/edit`, match: (path: string) => path === `/admin/variants/${variant.id}/edit` },
    { label: "Video Source", href: `/admin/variants/${variant.id}/video`, match: (path: string) => path.startsWith(`/admin/variants/${variant.id}/video`) },
  ]);
</script>

<svelte:head>
  <title>{variant.slug} | Variant Hub</title>
</svelte:head>

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <a
      href={song ? `/admin/songs/${song.id}/variants` : "/admin/variants"}
      aria-label="Back to Variants"
      class="text-on-surface-variant/70 hover:text-on-surface transition-colors p-2 -ml-2 rounded-lg hover:bg-surface-highest"
    >
      <svg
        class="w-5 h-5"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M10 19l-7-7m0 0l7-7m-7 7h18"
        />
      </svg>
    </a>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface line-clamp-1 uppercase">
      {variant.slug}
    </h1>
    <span
      class="inline-flex items-center px-2.5 py-1 rounded-md text-sm font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 ml-2"
    >
      v{variant.version_number}
    </span>
  </div>
  
  <div class="flex flex-col gap-1 ml-10 text-xs text-on-surface-variant/40">
    {#if song}
      <div class="flex items-center gap-2">
          <span class="uppercase font-bold tracking-tighter opacity-50">Song:</span>
          <a href="/admin/songs/{song.id}" class="text-primary hover:underline font-medium">
              {getSongName(song)}
          </a>
      </div>
    {/if}
    {#if anime}
      <div class="flex items-center gap-2">
          <span class="uppercase font-bold tracking-tighter opacity-50">Anime:</span>
          <a href="/admin/animes/{anime.id}" class="text-on-surface-variant/70 hover:underline">
              {anime.title}
          </a>
      </div>
    {/if}
  </div>
</div>

<!-- Tabs Navigation -->
<div class="flex items-center gap-2 border-b border-outline-variant mb-8 overflow-x-auto pb-px">
  {#each tabs as tab}
    {@const active = tab.match($page.url.pathname)}
    <a
      href={tab.href}
      class="px-6 py-3 text-sm font-medium transition-all relative whitespace-nowrap {active ? 'text-on-surface' : 'text-on-surface-variant/70 hover:text-on-surface'}"
    >
      {tab.label}
      {#if active}
        <div class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full"></div>
      {/if}
    </a>
  {/each}
</div>

<main class="min-h-[400px]">
  {@render children()}
</main>
