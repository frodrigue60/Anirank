<script lang="ts">
  import type { PageData } from "./$types";
  import { getSongName } from "$lib/song-utils";
  import { notifySong } from "$lib/api";
  import Megaphone from "lucide-svelte/icons/megaphone";
  import { toastState } from "$lib/state/toast.svelte";
  import { getApiErrorMessage } from "$lib/api-errors";

  let { data } = $props<{ data: PageData }>();
  let song = $derived(data.song);

  const videoUrl = $derived(
    song.song_variants && song.song_variants.length > 0
      ? song.song_variants[0].video_url
      : null
  );
  let isNotifying = $state(false);

  async function handleNotify() {
    if (isNotifying) return;
    if (!confirm("¿Deseas enviar este tema a los canales de Discord?")) return;
    
    isNotifying = true;
    try {
      await notifySong(song.id);
      toastState.addToast("Anuncio enviado con éxito", "success");
    } catch (err: any) {
      toastState.addToast(getApiErrorMessage(err, "Error al enviar anuncio"), "error");
    } finally {
      isNotifying = false;
    }
  }
</script>

<div class="flex items-center justify-between mb-6">
  <div>
    <h1 class="text-2xl font-bold text-on-surface">
      {getSongName(song)}
    </h1>
    <p class="text-on-surface-variant text-sm">
      ID: {song.uuid} • {song.theme_num}
    </p>
  </div>
  <button
    onclick={handleNotify}
    disabled={isNotifying}
    class="bg-primary/10 hover:bg-primary/20 border border-primary/30 text-primary px-4 py-2 rounded-xl font-bold uppercase tracking-wider text-xs flex items-center gap-2 transition-all disabled:opacity-50"
  >
    <Megaphone size={16} class={isNotifying ? 'animate-pulse' : ''} />
    {isNotifying ? 'Sending...' : 'Announce on Discord'}
  </button>
</div>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
  <!-- Main Content -->
  <div class="lg:col-span-2 space-y-6">
    <!-- Media Player -->
    {#if videoUrl}
      <div
        class="bg-black border border-outline-variant rounded-2xl overflow-hidden shadow-xl aspect-video w-full flex items-center justify-center relative group"
      >
        <!-- svelte-ignore a11y_media_has_caption -->
        <video controls class="w-full h-full object-contain" src={videoUrl}>
          Your browser does not support the video tag.
        </video>
      </div>
    {/if}

    <!-- Alternative Titles -->
    <div
      class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm"
    >
      <h2
        class="text-lg font-semibold text-on-surface mb-4 border-b border-outline-variant pb-2"
      >
        Titles
      </h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
        <div>
          <span class="text-on-surface-variant/40 block mb-1">Romaji</span>
          <span class="text-gray-200">{song.song_romaji || "-"}</span>
        </div>
        <div>
          <span class="text-on-surface-variant/40 block mb-1">Japanese</span>
          <span class="text-gray-200">{song.song_jp || "-"}</span>
        </div>
        <div>
          <span class="text-on-surface-variant/40 block mb-1">English</span>
          <span class="text-gray-200">{song.song_en || "-"}</span>
        </div>
      </div>
    </div>

    <!-- Artists -->
    <div
      class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm"
    >
      <div
        class="flex items-center justify-between mb-4 border-b border-outline-variant pb-2"
      >
        <h2 class="text-lg font-semibold text-on-surface">Artists</h2>
      </div>
      {#if song.artists && song.artists.length > 0}
        <div class="flex flex-wrap gap-2">
          {#each song.artists as artist}
            <a
              href="/admin/artists/{artist.id}/edit"
              class="px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 border border-zinc-700 rounded-lg text-sm text-on-surface-variant hover:text-on-surface transition-colors flex items-center gap-2"
            >
              {#if artist.avatar_url}
                <img
                  src={artist.avatar_url}
                  alt={artist.name}
                  class="w-5 h-5 rounded-full object-cover"
                />
              {/if}
              {artist.name}
            </a>
          {/each}
        </div>
      {:else}
        <p class="text-on-surface-variant/40 text-sm italic">
          No artists associated with this song.
        </p>
      {/if}
    </div>
  </div>

  <!-- Sidebar Column -->
  <div class="space-y-6">
    <!-- Metadata & Taxonomy -->
    <div
      class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm"
    >
      <h3
        class="text-sm font-semibold text-on-surface uppercase tracking-wider mb-4 border-b border-outline-variant pb-2"
      >
        Details
      </h3>
      <dl class="space-y-4 text-sm">
        <div>
          <dt class="text-on-surface-variant/40 mb-1">Season</dt>
          <dd class="text-gray-200 font-medium">
            {#if song.season || song.year}
              {song.season?.name || ""} {song.year?.name || ""}
            {:else}
              <span class="text-gray-600">-</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-on-surface-variant/40 mb-1">Added On</dt>
          <dd class="text-gray-200 font-medium">
            {new Date(song.created_at).toLocaleDateString()}
          </dd>
        </div>
        <div>
          <dt class="text-on-surface-variant/40 mb-1">Last Updated</dt>
          <dd class="text-gray-200 font-medium">
            {new Date(song.updated_at).toLocaleDateString()}
          </dd>
        </div>
      </dl>
    </div>

    <!-- Metrics Stats -->
    <div
      class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm"
    >
      <h3
        class="text-sm font-semibold text-on-surface uppercase tracking-wider mb-4 border-b border-outline-variant pb-2"
      >
        Metrics
      </h3>
      <div class="grid grid-cols-2 gap-4">
        <div class="bg-surface-highest p-3 rounded-xl border border-outline-variant">
          <span class="block text-xs text-on-surface-variant/40 mb-1">Total Views</span>
          <span class="block text-xl font-semibold text-on-surface"
            >{song.views.toLocaleString()}</span
          >
        </div>
        <div class="bg-surface-highest p-3 rounded-xl border border-outline-variant">
          <span class="block text-xs text-on-surface-variant/40 mb-1">Avg Rating</span>
          <span class="block text-xl font-semibold text-on-surface"
            >{song.average_rating > 0
              ? song.average_rating.toFixed(2)
              : "-"}</span
          >
        </div>
        <div class="bg-surface-highest p-3 rounded-xl border border-outline-variant">
          <span class="block text-xs text-on-surface-variant/40 mb-1">Likes</span>
          <span class="block text-xl font-semibold text-on-surface"
            >{song.likes_count.toLocaleString()}</span
          >
        </div>
        <div class="bg-surface-highest p-3 rounded-xl border border-outline-variant">
          <span class="block text-xs text-on-surface-variant/40 mb-1">Dislikes</span>
          <span class="block text-xl font-semibold text-on-surface"
            >{song.dislikes_count.toLocaleString()}</span
          >
        </div>
      </div>
    </div>
  </div>
</div>
