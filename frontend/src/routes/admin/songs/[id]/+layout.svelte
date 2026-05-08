<script lang="ts">
  import type { LayoutData } from "./$types";
  import { getSongName } from "$lib/song-utils";
  import { page } from "$app/stores";
  import { adminNav } from "$lib/state/admin-nav.svelte";
  import AdminBreadcrumb from "$lib/components/admin/AdminBreadcrumb.svelte";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";

  let { data, children } = $props<{ data: LayoutData; children: any }>();
  let song = $derived(data.song);

  const tabs = $derived([
    { label: "Overview", href: `/admin/songs/${song.id}`, match: (path: string) => path === `/admin/songs/${song.id}` },
    { label: "Edit Info", href: `/admin/songs/${song.id}/edit`, match: (path: string) => path === `/admin/songs/${song.id}/edit` },
    { label: "Variants", href: `/admin/songs/${song.id}/variants`, match: (path: string) => path.startsWith(`/admin/songs/${song.id}/variants`) },
  ]);

  // Build breadcrumb context from the song's parent relationships
  let crumbs = $derived.by(() => {
    const trail: any[] = [
      { label: "Animes", href: "/admin/animes", type: "list" as const },
    ];
    if (song.anime) {
      trail.push({
        label: song.anime.title,
        href: `/admin/animes/${song.anime.id}`,
        type: "anime" as const,
      });
      trail.push({
        label: "Songs",
        href: `/admin/animes/${song.anime.id}/songs`,
        type: "list" as const,
      });
    } else {
      trail.push({ label: "Songs", href: "/admin/songs", type: "list" as const });
    }
    trail.push({
      label: getSongName(song),
      href: `/admin/songs/${song.id}`,
      type: "song" as const,
    });
    return trail;
  });

  // Set navigation context
  $effect(() => {
    adminNav.setContext(crumbs);
  });

  // Smart back: go to the parent anime's songs tab if available, else flat list
  let backUrl = $derived(
    song.anime
      ? `/admin/animes/${song.anime.id}/songs`
      : "/admin/songs"
  );
</script>

<svelte:head>
  <title>{getSongName(song)} | Song Hub</title>
</svelte:head>

<AdminBreadcrumb {crumbs} />

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <a
      href={backUrl}
      aria-label="Back"
      class="text-on-surface-variant/70 hover:text-on-surface transition-colors p-2 -ml-2 rounded-lg hover:bg-surface-highest"
    >
      <ArrowLeft size={20} />
    </a>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface line-clamp-1">
      {getSongName(song)}
    </h1>
    <span
      class="inline-flex items-center px-2.5 py-1 rounded-md text-sm font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20 ml-2"
    >
      {song.type} {song.theme_num}
    </span>
  </div>
  
  {#if song.anime}
    <div class="flex items-center gap-2 ml-10 text-sm text-on-surface-variant/70">
        <span>From:</span>
        <a href="/admin/animes/{song.anime.id}" class="text-primary hover:underline font-medium">
            {song.anime.title}
        </a>
    </div>
  {/if}
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
