<script lang="ts">
  import { fade } from "svelte/transition";
  import { authState } from "$lib/state/auth.svelte";
  import api from "$lib/api";
  import { invalidateAll } from "$app/navigation";

  let { playlist, profile, openEditModal, handleDelete, onTogglePrivacy } =
    $props();

  async function togglePrivacy() {
    try {
      const nextPublic = !playlist.is_public;
      await api.put(`/playlists/${playlist.id}`, {
        is_public: nextPublic,
        name: playlist.name,
      });

      if (onTogglePrivacy) {
        onTogglePrivacy(playlist.id, nextPublic);
      }
    } catch (e) {
      console.error("Failed to toggle privacy", e);
    }
  }
</script>

<div
  in:fade={{ duration: 300 }}
  class="group relative bg-surface-dark/30 rounded-md border border-white/5 p-5 hover:bg-surface-dark/50 transition-all flex flex-col gap-4 overflow-hidden aspect-video"
>
  <!-- Background Image -->
  <div
    class="absolute inset-0 bg-cover bg-center transition-transform duration-500 group-hover:scale-105"
    style="background-image: url('{playlist.banner_url ||
      'https://placehold.co/1280x720/2a2136/white?text=No+Songs'}'); filter:brightness(0.5)"
  ></div>

  <div
    class="absolute inset-0 bg-linear-to-t from-black via-black/40 to-transparent"
  ></div>

  {#if authState.user?.id === profile?.id}
    <div
      class="absolute top-0 left-0 right-0 p-2 z-10 flex items-center justify-between pointer-events-none"
    >
      <button
        onclick={togglePrivacy}
        class="text-white/70 hover:text-primary bg-white/5 hover:bg-white/10 rounded-sm px-3 py-2 transition-all border border-white/5 pointer-events-auto flex items-center gap-1.5"
        title="Toggle Privacy"
      >
        <span class="material-symbols-outlined text-sm">
          {playlist.is_public ? "public" : "lock"}
        </span>
        <span class="text-[10px] font-bold uppercase tracking-wider">
          {playlist.is_public ? "public" : "private"}
        </span>
      </button>
      <div class="flex items-center gap-2 pointer-events-auto">
        <button
          onclick={() => openEditModal(playlist)}
          class="text-white/70 hover:text-primary bg-white/5 hover:bg-white/10 rounded-sm p-2 transition-all border border-white/5"
          title="Edit Playlist"
        >
          <span class="material-symbols-outlined text-[20px]">edit</span>
        </button>
        <button
          onclick={() => handleDelete(playlist.id)}
          class="text-white/70 hover:text-red-500 bg-white/5 hover:bg-white/10 rounded-sm p-2 transition-all border border-white/5"
          title="Delete Playlist"
        >
          <span class="material-symbols-outlined text-[20px]">delete</span>
        </button>
      </div>
    </div>
  {/if}

  <!-- Bottom Info -->
  <div class="absolute bottom-0 left-0 right-0 p-5 flex flex-col gap-2">
    <div class="flex items-center gap-2 mb-2">
      <span class="text-slate-400 text-[10px] flex items-center gap-1">
        <span class="material-symbols-outlined text-xs">music_note</span>
        {playlist.song_count || 0} Songs
      </span>
    </div>

    <a
      href="/playlists/{playlist.id}"
      class="text-xl font-bold text-white mb-1 group-hover:text-primary transition-colors uppercase truncate"
    >
      {playlist.name}
    </a>

    <p class="text-slate-300 text-xs font-medium opacity-80">
      Created by <span class="text-slate-400 font-semibold"
        >{profile?.name}</span
      >
    </p>
  </div>
</div>
