<script lang="ts">
  import type { LayoutData } from "./$types";
  import { getSongName } from "$lib/song-utils";
  import { page } from "$app/stores";
  import { adminNav } from "$lib/state/admin-nav.svelte";
  import AdminBreadcrumb from "$lib/components/admin/AdminBreadcrumb.svelte";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";

  let { data, children } = $props<{ data: LayoutData; children: any }>();
  let variant = $derived(data.variant);
  let song = $derived(variant.song);
  let anime = $derived(song?.anime);

  const tabs = $derived([
    { label: "Overview", href: `/admin/variants/${variant.id}`, match: (path: string) => path === `/admin/variants/${variant.id}` },
    { label: "Edit Info", href: `/admin/variants/${variant.id}/edit`, match: (path: string) => path === `/admin/variants/${variant.id}/edit` },
    { label: "Video Source", href: `/admin/variants/${variant.id}/video`, match: (path: string) => path.startsWith(`/admin/variants/${variant.id}/video`) },
  ]);

  // Build full breadcrumb chain: Anime → Song → Variant
  let crumbs = $derived.by(() => {
    const trail: any[] = [];
    if (anime) {
      trail.push({ label: "Animes", href: "/admin/animes", type: "list" as const });
      trail.push({
        label: anime.title,
        href: `/admin/animes/${anime.id}`,
        type: "anime" as const,
      });
    }
    if (song) {
      trail.push({
        label: getSongName(song),
        href: `/admin/songs/${song.id}`,
        type: "song" as const,
      });
      trail.push({
        label: "Variants",
        href: `/admin/songs/${song.id}/variants`,
        type: "list" as const,
      });
    } else {
      trail.push({ label: "Variants", href: "/admin/variants", type: "list" as const });
    }
    trail.push({
      label: variant.slug,
      href: `/admin/variants/${variant.id}`,
      type: "variant" as const,
    });
    return trail;
  });

  $effect(() => {
    adminNav.setContext(crumbs);
  });

  // Smart back: go to the parent song's variants tab if available
  let backUrl = $derived(
    song
      ? `/admin/songs/${song.id}/variants`
      : "/admin/variants"
  );
</script>

<svelte:head>
  <title>{variant.slug} | Variant Hub</title>
</svelte:head>

<AdminBreadcrumb {crumbs} />

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <a
      href={backUrl}
      aria-label="Back to Variants"
      class="text-on-surface-variant/70 hover:text-on-surface transition-colors p-2 -ml-2 rounded-lg hover:bg-surface-highest"
    >
      <ArrowLeft size={20} />
    </a>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface line-clamp-1 uppercase">
      {variant.slug}
    </h1>
    <span
      class="inline-flex items-center px-2.5 py-1 rounded-md text-sm font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 ml-2"
    >
      v{variant.version_number}
    </span>
    {#if !variant.status}
      <span
        class="inline-flex items-center px-2.5 py-1 rounded-md text-[10px] font-bold bg-white/5 text-on-surface-variant/40 border border-outline-variant/30 ml-2 uppercase tracking-widest"
      >
        Draft
      </span>
    {/if}

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
              {anime.title} ({anime.year?.name || ""} {anime.season?.name || ""})
          </a>

      </div>
    {/if}
    
    <div class="flex items-center gap-4 mt-1">
      {#if variant.episodes}
        <div class="flex items-center gap-1.5">
          <span class="uppercase font-bold tracking-tighter opacity-50">Episodes:</span>
          <span class="text-on-surface-variant/80">{variant.episodes}</span>
        </div>
      {/if}

      <div class="flex items-center gap-2">
        {#if variant.spoiler}
          <span class="bg-amber-500/10 text-amber-400 text-[10px] px-1.5 py-0.5 rounded border border-amber-500/20 font-bold uppercase tracking-wider">Spoiler</span>
        {/if}
        {#if variant.nsfw}
          <span class="bg-rose-500/10 text-rose-400 text-[10px] px-1.5 py-0.5 rounded border border-rose-500/20 font-bold uppercase tracking-wider">NSFW</span>
        {/if}
      </div>
    </div>
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
