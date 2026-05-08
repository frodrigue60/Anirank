<script lang="ts">
  import type { LayoutData } from "./$types";
  import { page } from "$app/stores";
  import { adminNav } from "$lib/state/admin-nav.svelte";
  import AdminBreadcrumb from "$lib/components/admin/AdminBreadcrumb.svelte";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";

  let { data, children } = $props<{ data: LayoutData; children: any }>();
  let anime = $derived(data.anime);

  let tabs = $derived([
    {
      label: "Overview",
      href: `/admin/animes/${anime.id}`,
      match: (path: string) => path === `/admin/animes/${anime.id}`,
    },
    {
      label: "Edit Info",
      href: `/admin/animes/${anime.id}/edit`,
      match: (path: string) => path === `/admin/animes/${anime.id}/edit`,
    },
    {
      label: "Songs",
      href: `/admin/animes/${anime.id}/songs`,
      match: (path: string) =>
        path.startsWith(`/admin/animes/${anime.id}/songs`),
    },
  ]);

  // Build breadcrumb context for this anime
  let crumbs = $derived([
    { label: "Animes", href: "/admin/animes", type: "list" as const },
    { label: anime.title, href: `/admin/animes/${anime.id}`, type: "anime" as const },
  ]);

  // Set navigation context when this layout mounts
  $effect(() => {
    adminNav.setContext(crumbs);
  });

  let backUrl = $derived(adminNav.getBackUrl("anime", "/admin/animes"));
</script>

<svelte:head>
  <title>{anime.title} | Admin Hub</title>
</svelte:head>

<AdminBreadcrumb {crumbs} />

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <a
      href={backUrl}
      aria-label="Back to Animes"
      class="text-on-surface-variant/70 hover:text-on-surface transition-colors p-2 -ml-2 rounded-lg hover:bg-surface-highest"
    >
      <ArrowLeft size={20} />
    </a>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface line-clamp-1">
      {anime.title}
    </h1>
    {#if anime.anilist_id}
      <a
        href="https://anilist.co/anime/{anime.anilist_id}"
        target="_blank"
        rel="noopener noreferrer"
        class="p-2 ml-2 bg-blue-500/10 text-blue-400 hover:bg-blue-500/20 rounded-lg transition-colors flex shrink-0"
        title="View on AniList"
      >
        <svg
          class="w-5 h-5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
          <polyline points="15 3 21 3 21 9" />
          <line x1="10" y1="14" x2="21" y2="3" />
        </svg>
      </a>
    {/if}
  </div>
  <p class="text-on-surface-variant/70 ml-10">Anime Management Hub</p>
</div>

<!-- Tabs Navigation -->
<div
  class="flex items-center gap-2 border-b border-outline-variant mb-8 overflow-x-auto pb-px"
>
  {#each tabs as tab}
    {@const active = tab.match($page.url.pathname)}
    <a
      href={tab.href}
      class="px-6 py-3 text-sm font-medium transition-all relative whitespace-nowrap {active
        ? 'text-on-surface'
        : 'text-on-surface-variant/70 hover:text-on-surface'}"
    >
      {tab.label}
      {#if active}
        <div
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full"
        ></div>
      {/if}
    </a>
  {/each}
</div>

<main class="min-h-[400px]">
  {@render children()}
</main>
