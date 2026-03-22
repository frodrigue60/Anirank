<script lang="ts">
  import { fade } from "svelte/transition";
  import { authState } from "$lib/state/auth.svelte";
  import api from "$lib/api";
  import { invalidateAll } from "$app/navigation";

  let { playlist, profile, openEditModal, handleDelete } = $props();

  async function togglePrivacy() {
    try {
      await api.put(`/playlists/${playlist.id}`, {
        is_public: !playlist.is_public,
        name: playlist.name, // name is required by validation
      });
      await invalidateAll();
    } catch (e) {
      console.error("Failed to toggle privacy", e);
    }
  }
</script>

<a href={`/playlists/${playlist.id}`} class="group cursor-pointer">
  <div
    class="aspect-square rounded-2xl overflow-hidden glass-effect border-none mb-3 relative"
  >
    <!-- svelte-ignore a11y_missing_attribute -->
    <img
      class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
      alt={playlist.name}
      title={playlist.name}
      src={playlist.banner_url ||
        "https://placehold.co/400x400/1e1e28/a855f7?text=Playlist"}
    />
    <div
      class="absolute inset-0 bg-primary/20 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center"
    >
      <span class="material-symbols-outlined text-white text-5xl"
        >play_circle</span
      >
    </div>
  </div>
  <h4
    class="text-slate-100 font-bold group-hover:text-primary transition-colors"
  >
    {playlist.name}
  </h4>
  <p class="text-slate-400 text-sm">{playlist.song_count} Tracks</p>
</a>
