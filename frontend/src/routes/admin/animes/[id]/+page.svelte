<script lang="ts">
  import type { PageData } from "./$types";
  import api from "$lib/api";
  import { createTrustedHTML } from "$lib/trusted";
  import { toastState } from "$lib/state/toast.svelte";
  import { invalidateAll } from "$app/navigation";
  import { getApiErrorMessage } from "$lib/api-errors";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import ExternalLink from "lucide-svelte/icons/external-link";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";

  let { data } = $props<{ data: PageData }>();
  let anime = $derived(data.anime);

  let isSyncing = $state(false);

  async function updateAnimeData() {
    if (isSyncing) return;
    isSyncing = true;
    try {
      await api.post(`/admin/animes/${anime.id}/sync`);
      toastState.addToast("Anime data synchronized with AniList", "success");
      await invalidateAll();
    } catch (err: any) {
      console.error(err);
      toastState.addToast(
        getApiErrorMessage(err, "Failed to sync anime data"),
        "error",
      );
    } finally {
      isSyncing = false;
    }
  }
</script>

<!-- Hero Banner & Cover -->
<div
  class="relative w-full h-[250px] mb-20 rounded-3xl bg-zinc-900 border border-outline-variant shadow-xl overflow-visible hidden md:block"
>
  <div class="absolute inset-0 overflow-hidden rounded-3xl bg-zinc-950">
    {#if anime.banner_url}
      <OptimizedImage
        src={anime.banner_url}
        sources={anime.banner_sources}
        alt="{anime.title} banner"
        class="w-full h-full object-cover opacity-60"
        sizes="100vw"
      />
      <div
        class="absolute inset-0 bg-linear-to-t from-zinc-950 via-zinc-900/40 to-transparent"
      ></div>
    {/if}
  </div>

  <div
    class="absolute -bottom-16 left-10 flex items-end gap-6 z-10 w-full pr-20"
  >
    <!-- Cover -->
    <div
      class="w-40 h-60 rounded-2xl border-4 border-zinc-950 shadow-2xl bg-zinc-900 overflow-hidden shrink-0 relative"
    >
      {#if anime.cover_url}
        <OptimizedImage
          src={anime.cover_url}
          sources={anime.cover_sources}
          alt="{anime.title} cover"
          class="w-full h-full object-cover"
          sizes="160px"
        />
      {:else}
        <div
          class="w-full h-full flex items-center justify-center text-zinc-700"
        >
          No Cover
        </div>
      {/if}
      <div class="absolute top-3 right-3">
        <span
          class="px-2 py-1 text-[10px] font-bold rounded-lg {anime.status
            ? 'bg-green-500 text-on-surface'
            : 'bg-zinc-800 text-zinc-400'} shadow-lg"
        >
          {anime.status ? "ACTIVE" : "INACTIVE"}
        </span>
      </div>
    </div>

    <!-- Hub Header Title (Overlay on banner space) -->
    <div class="mb-4 flex-1">
      <div class="flex items-center gap-3">
        <h2 class="text-2xl font-bold text-on-surface drop-shadow-lg">
          {anime.title}
        </h2>
        <button
          onclick={updateAnimeData}
          disabled={isSyncing}
          class="p-2 bg-surface-highest hover:bg-white/20 border border-outline-variant text-on-surface rounded-lg transition-colors disabled:opacity-50"
          title="Sync from AniList"
        >
          <RefreshCw size={14} class={isSyncing ? 'animate-spin' : ''} />

        </button>
      </div>
      <p class="text-on-surface-variant text-sm drop-shadow-md">
        {anime.format?.name || "Unknown Format"} • {anime.season?.name || ""}
        {anime.year?.name || ""}
      </p>
    </div>
  </div>
</div>

<div class="grid grid-cols-1 lg:grid-cols-4 gap-8">
  <!-- Main Column -->
  <div class="lg:col-span-3 space-y-8">
    <!-- Description -->
    <div
      class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm"
    >
      <h2 class="text-xl font-semibold text-on-surface mb-4">Synopsis</h2>
      <div class="text-on-surface-variant leading-relaxed max-w-none text-sm space-y-4">
        {#if anime.description}
          {@html createTrustedHTML(anime.description)}
        {:else}
          <p class="text-on-surface-variant/40 italic">No description provided.</p>
        {/if}
      </div>
    </div>
  </div>

  <!-- Sidebar Column -->
  <div class="space-y-6">
    <div
      class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm"
    >
      <h3
        class="text-sm font-semibold text-on-surface uppercase tracking-wider mb-4 border-b border-outline-variant pb-2"
      >
        Information
      </h3>

      <dl class="space-y-4 text-sm">
        <div>
          <dt class="text-on-surface-variant/40 mb-1">Format</dt>
          <dd class="text-gray-200 font-medium">
            {anime.format?.name || "Unknown"}
          </dd>
        </div>
        <div>
          <dt class="text-on-surface-variant/40 mb-1">Season</dt>
          <dd class="text-gray-200 font-medium">
            {anime.season?.name || "Unknown"}
            {anime.year?.name || ""}
          </dd>
        </div>
        <div>
          <dt class="text-on-surface-variant/40 mb-1">Studios</dt>
          <dd class="text-gray-200 font-medium">
            {#if anime.studios && anime.studios.length > 0}
              {anime.studios.map((s: any) => s.name).join(", ")}
            {:else}
              <span class="text-gray-600">-</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-on-surface-variant/40 mb-1">Producers</dt>
          <dd class="text-gray-200 font-medium">
            {#if anime.producers && anime.producers.length > 0}
              {anime.producers.map((s: any) => s.name).join(", ")}
            {:else}
              <span class="text-gray-600">-</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-on-surface-variant/40 mb-1">Genres</dt>
          <dd class="text-gray-200 font-medium">
            {#if anime.genres && anime.genres.length > 0}
              <div class="flex flex-wrap gap-1 mt-1">
                {#each anime.genres as genre}
                  <span
                    class="px-2 py-0.5 bg-surface-highest border border-outline-variant rounded text-xs text-on-surface-variant"
                    >{genre.name}</span
                  >
                {/each}
              </div>
            {:else}
              <span class="text-gray-600">-</span>
            {/if}
          </dd>
        </div>
      </dl>
    </div>

    <!-- External Links -->
    {#if anime.external_links && anime.external_links.length > 0}
      <div
        class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm flex flex-col gap-2"
      >
        <h3
          class="text-sm font-semibold text-on-surface uppercase tracking-wider mb-2 border-b border-outline-variant pb-2"
        >
          External Links
        </h3>
        {#each anime.external_links as link}
          <a
            href={link.url}
            target="_blank"
            rel="noopener noreferrer"
            class="flex items-center gap-2 p-2 hover:bg-surface-highest rounded-lg transition-colors border border-transparent hover:border-outline-variant group"
          >
            <span
              class="w-6 h-6 rounded bg-black/20 text-on-surface-variant/70 flex items-center justify-center shrink-0 group-hover:text-blue-400 transition-colors"
            >
              <ExternalLink size={14} />
            </span>
            <span
              class="text-sm text-on-surface-variant group-hover:text-on-surface transition-colors"
              >{link.name}</span
            >
          </a>
        {/each}
      </div>
    {/if}
  </div>
</div>
