<script lang="ts">
  import { fade } from "svelte/transition";
  import { authState } from "$lib/state/auth.svelte";
  import api from "$lib/api";
  import { invalidateAll } from "$app/navigation";
  import Globe from "lucide-svelte/icons/globe";
  import Lock from "lucide-svelte/icons/lock";
  import Edit2 from "lucide-svelte/icons/edit-2";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import Music from "lucide-svelte/icons/music";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";

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
  <OptimizedImage
    src={playlist.banner_url}
    sources={playlist.banner_sources}
    alt=""
    class="absolute inset-0 w-full h-full object-cover object-center transition-transform duration-500 group-hover:scale-105 brightness-50"
    sizes="(max-width: 640px) 100vw, 400px"
  />

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
        {#if playlist.is_public}
          <Globe size={14} />
        {:else}
          <Lock size={14} />
        {/if}

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
          <Edit2 size={20} />
        </button>
        <button
          onclick={() => handleDelete(playlist.id)}
          class="text-white/70 hover:text-red-500 bg-white/5 hover:bg-white/10 rounded-sm p-2 transition-all border border-white/5"
          title="Delete Playlist"
        >
          <Trash2 size={20} />
        </button>
      </div>
    </div>
  {/if}

  <!-- Bottom Info -->
  <div class="absolute bottom-0 left-0 right-0 p-5 flex flex-col gap-2">
    <div class="flex items-center gap-2 mb-2">
      <span class="text-slate-400 text-[10px] flex items-center gap-1">
        <Music size={12} />
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
