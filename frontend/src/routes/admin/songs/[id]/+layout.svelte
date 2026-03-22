<script lang="ts">
  import type { LayoutData } from "./$types";
  import { getSongName } from "$lib/song-utils";
  import { page } from "$app/stores";

  let { data, children } = $props<{ data: LayoutData; children: any }>();
  let song = $derived(data.song);

  const tabs = $derived([
    { label: "Overview", href: `/admin/songs/${song.id}`, match: (path: string) => path === `/admin/songs/${song.id}` },
    { label: "Edit Info", href: `/admin/songs/${song.id}/edit`, match: (path: string) => path === `/admin/songs/${song.id}/edit` },
    { label: "Variants", href: `/admin/songs/${song.id}/variants`, match: (path: string) => path.startsWith(`/admin/songs/${song.id}/variants`) },
  ]);
</script>

<svelte:head>
  <title>{getSongName(song)} | Song Hub</title>
</svelte:head>

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <a
      href="/admin/songs"
      aria-label="Back to Songs"
      class="text-gray-400 hover:text-white transition-colors p-2 -ml-2 rounded-lg hover:bg-white/5"
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
    <h1 class="text-3xl font-bold tracking-tight text-white line-clamp-1">
      {getSongName(song)}
    </h1>
    <span
      class="inline-flex items-center px-2.5 py-1 rounded-md text-sm font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20 ml-2"
    >
      {song.type} {song.theme_num}
    </span>
  </div>
  
  {#if song.anime}
    <div class="flex items-center gap-2 ml-10 text-sm text-gray-400">
        <span>From:</span>
        <a href="/admin/animes/{song.anime.id}" class="text-anirank-primary hover:underline font-medium">
            {song.anime.title}
        </a>
    </div>
  {/if}
</div>

<!-- Tabs Navigation -->
<div class="flex items-center gap-2 border-b border-white/5 mb-8 overflow-x-auto pb-px">
  {#each tabs as tab}
    {@const active = tab.match($page.url.pathname)}
    <a
      href={tab.href}
      class="px-6 py-3 text-sm font-medium transition-all relative whitespace-nowrap {active ? 'text-white' : 'text-gray-400 hover:text-white'}"
    >
      {tab.label}
      {#if active}
        <div class="absolute bottom-0 left-0 right-0 h-0.5 bg-anirank-primary rounded-full"></div>
      {/if}
    </a>
  {/each}
</div>

<main class="min-h-[400px]">
  {@render children()}
</main>
