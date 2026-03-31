<script lang="ts">
  import type { PageData } from "./$types";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { invalidateAll } from "$app/navigation";
  import { getApiErrorMessage } from "$lib/api-errors";

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
      toastState.addToast(getApiErrorMessage(err, "Failed to sync anime data"), "error");
    } finally {
      isSyncing = false;
    }
  }
</script>

<!-- Hero Banner & Cover -->
<div
  class="relative w-full h-[250px] mb-20 rounded-3xl bg-zinc-900 border border-white/10 shadow-xl overflow-visible hidden md:block"
>
  <div class="absolute inset-0 overflow-hidden rounded-3xl bg-zinc-950">
    {#if anime.banner_url}
      <img
        src={anime.banner_url}
        alt="{anime.title} banner"
        class="w-full h-full object-cover opacity-60"
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
        <img
          src={anime.cover_url}
          alt="{anime.title} cover"
          class="w-full h-full object-cover"
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
            ? 'bg-green-500 text-white'
            : 'bg-zinc-800 text-zinc-400'} shadow-lg backdrop-blur-md"
        >
          {anime.status ? "ACTIVE" : "INACTIVE"}
        </span>
      </div>
    </div>

    <!-- Hub Header Title (Overlay on banner space) -->
    <div class="mb-4 flex-1">
      <div class="flex items-center gap-3">
        <h2 class="text-2xl font-bold text-white drop-shadow-lg">
          {anime.title}
        </h2>
        <button
          onclick={updateAnimeData}
          disabled={isSyncing}
          class="p-2 bg-white/10 hover:bg-white/20 border border-white/10 text-white rounded-lg transition-colors disabled:opacity-50"
          title="Sync from AniList"
        >
          <span
            class="material-symbols-outlined text-sm {isSyncing
              ? 'animate-spin'
              : ''}">refresh</span
          >
        </button>
      </div>
      <p class="text-gray-300 text-sm drop-shadow-md">
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
      class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm"
    >
      <h2 class="text-xl font-semibold text-white mb-4">Synopsis</h2>
      <div class="text-gray-300 leading-relaxed max-w-none text-sm space-y-4">
        {#if anime.description}
          {@html anime.description}
        {:else}
          <p class="text-gray-500 italic">No description provided.</p>
        {/if}
      </div>
    </div>
  </div>

  <!-- Sidebar Column -->
  <div class="space-y-6">
    <div
      class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm"
    >
      <h3
        class="text-sm font-semibold text-white uppercase tracking-wider mb-4 border-b border-white/10 pb-2"
      >
        Information
      </h3>

      <dl class="space-y-4 text-sm">
        <div>
          <dt class="text-gray-500 mb-1">Format</dt>
          <dd class="text-gray-200 font-medium">
            {anime.format?.name || "Unknown"}
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 mb-1">Season</dt>
          <dd class="text-gray-200 font-medium">
            {anime.season?.name || "Unknown"}
            {anime.year?.name || ""}
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 mb-1">Studios</dt>
          <dd class="text-gray-200 font-medium">
            {#if anime.studios && anime.studios.length > 0}
              {anime.studios.map((s: any) => s.name).join(", ")}
            {:else}
              <span class="text-gray-600">-</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 mb-1">Producers</dt>
          <dd class="text-gray-200 font-medium">
            {#if anime.producers && anime.producers.length > 0}
              {anime.producers.map((s: any) => s.name).join(", ")}
            {:else}
              <span class="text-gray-600">-</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 mb-1">Genres</dt>
          <dd class="text-gray-200 font-medium">
            {#if anime.genres && anime.genres.length > 0}
              <div class="flex flex-wrap gap-1 mt-1">
                {#each anime.genres as genre}
                  <span
                    class="px-2 py-0.5 bg-white/5 border border-white/10 rounded text-xs text-gray-300"
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
        class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm flex flex-col gap-2"
      >
        <h3
          class="text-sm font-semibold text-white uppercase tracking-wider mb-2 border-b border-white/10 pb-2"
        >
          External Links
        </h3>
        {#each anime.external_links as link}
          <a
            href={link.url}
            target="_blank"
            rel="noopener noreferrer"
            class="flex items-center gap-2 p-2 hover:bg-white/5 rounded-lg transition-colors border border-transparent hover:border-white/10 group"
          >
            <span
              class="w-6 h-6 rounded bg-black/20 text-gray-400 flex items-center justify-center shrink-0 group-hover:text-blue-400 transition-colors"
            >
              <svg
                class="w-3.5 h-3.5"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                ><path
                  d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"
                ></path><polyline points="15 3 21 3 21 9"></polyline><line
                  x1="10"
                  y1="14"
                  x2="21"
                  y2="3"
                ></line></svg
              >
            </span>
            <span
              class="text-sm text-gray-300 group-hover:text-white transition-colors"
              >{link.name}</span
            >
          </a>
        {/each}
      </div>
    {/if}
  </div>
</div>
