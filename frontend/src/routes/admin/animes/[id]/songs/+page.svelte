<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { getApiErrorMessage } from "$lib/api-errors";
  import { getSongName } from "$lib/song-utils";
  import type { PageData } from "./$types";
  import AlertTriangle from "lucide-svelte/icons/alert-triangle";
  import Plus from "lucide-svelte/icons/plus";
  import Video from "lucide-svelte/icons/video";
  import Edit2 from "lucide-svelte/icons/edit-2";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import Music from "lucide-svelte/icons/music";

  let { data } = $props<{ data: PageData }>();
  let anime = $derived(data.anime);
  // svelte-ignore state_referenced_locally
  let songs = $state(data.anime.songs || []);

  $effect(() => {
    songs = data.anime.songs || [];
  });

  async function deleteSong(id: number) {
    if (
      !confirm(
        "Are you sure you want to delete this song? This will also remove its variants/videos.",
      )
    )
      return;

    try {
      await api.delete(`/admin/songs/${id}`);
      toastState.addToast("Song deleted successfully", "success");
      await invalidateAll();
    } catch (err: any) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to delete song"),
        "error",
      );
    }
  }

  async function handleStatusChange(id: number, currentStatus: boolean) {
    const index = songs.findIndex((s: any) => s.id === id);
    if (index === -1) return;

    // Optimistic UI
    const prevStatus = songs[index].status;
    songs[index].status = !currentStatus;

    try {
      await api.patch(`/admin/songs/${id}/status`);
      toastState.addToast("Status updated successfully", "success");
    } catch (err: any) {
      // Rollback
      songs[index].status = prevStatus;
      console.error(err);
      toastState.addToast(
        getApiErrorMessage(err, "Failed to update status"),
        "error",
      );
    }
  }
</script>

<div class="space-y-4">
  {#if !anime.season_id || !anime.year_id}
    <div class="bg-amber-500/10 border border-amber-500/20 rounded-2xl p-4 flex items-start gap-3">
      <AlertTriangle size={20} class="text-amber-500 mt-0.5" />
      <div>
        <h3 class="text-amber-500 font-medium text-sm">Missing Information</h3>
        <p class="text-amber-400/80 text-sm mt-1">This anime must have a Season and Year assigned before you can create songs. Please edit the anime info first.</p>
      </div>
    </div>
  {/if}

  <div
    class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden shadow-sm shadow-black/20"
  >
    <div class="p-6 border-b border-outline-variant flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
      <div>
        <h2 class="text-xl font-semibold text-on-surface">Songs Management</h2>
        <p class="text-sm text-on-surface-variant/70 mt-1">
          Manage themes, openings and endings for this anime.
        </p>
      </div>
      {#if !anime.season_id || !anime.year_id}
        <button
          disabled
          class="text-sm px-4 py-2 bg-surface-highest border border-outline-variant text-on-surface-variant/40 rounded-lg cursor-not-allowed flex items-center gap-2"
          title="Season and Year required"
        >
          <Plus size={16} />
          Add Song
        </button>
      {:else}
        <a
          href="/admin/songs/create?anime_id={anime.id}"
          class="text-sm px-4 py-2 bg-primary hover:bg-primary-container text-on-surface rounded-lg transition-colors flex items-center gap-2 shadow-lg shadow-anirank-primary/20"
        >
          <Plus size={16} />
          Add Song
        </a>
      {/if}
    </div>

    <div class="overflow-x-auto">
      {#if songs.length > 0}
        <table class="w-full text-left text-sm text-on-surface-variant">
          <thead
            class="text-xs text-on-surface-variant/70 uppercase bg-black/20 border-b border-outline-variant"
          >
            <tr>
              <th class="px-6 py-4 font-semibold">Type</th>
              <th class="px-6 py-4 font-semibold">Title</th>
              <th class="px-6 py-4 font-semibold hidden md:table-cell">Artists</th
              >
              <th class="px-6 py-4 font-semibold">Status</th>
              <th class="px-6 py-4 font-semibold text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            {#each songs as song}
              <tr class="hover:bg-white/2 transition-colors">
                <td class="px-6 py-4 whitespace-nowrap">
                  <span
                    class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20"
                  >
                    {song.type}
                    {song.theme_num}
                  </span>
                </td>
                <td
                  class="px-6 py-4 font-medium text-on-surface max-w-[250px] truncate"
                  title={getSongName(song)}
                >
                  <div class="flex flex-col">
                    <span class="text-on-surface"
                      >{getSongName(song)}</span
                    >
                    <span class="text-[10px] text-on-surface-variant/40 font-mono"
                      >ID: #{song.id}</span
                    >
                  </div>
                </td>
                <td
                  class="px-6 py-4 text-xs text-on-surface-variant/40 max-w-[200px] truncate hidden md:table-cell"
                  title={song.artists
                    ? song.artists.map((a: any) => a.name).join(", ")
                    : ""}
                >
                  {song.artists
                    ? song.artists.map((a: any) => a.name).join(", ")
                    : "-"}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <button
                    onclick={() => handleStatusChange(song.id, song.status)}
                    class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border transition-all duration-200 hover:scale-105 active:scale-95 {song.status ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20 hover:bg-emerald-500/20' : 'bg-orange-500/10 text-orange-400 border-orange-500/20 hover:bg-orange-500/20'}"
                    title="Toggle status"
                  >
                    <div class="w-1.5 h-1.5 rounded-full mr-2 {song.status ? 'bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.5)]' : 'bg-orange-400 shadow-[0_0_8px_rgba(251,146,60,0.5)]'}"></div>
                    {song.status ? "Active" : "Inactive"}
                  </button>
                </td>
                <td class="px-6 py-4 text-right">
                  <div class="flex items-center justify-end gap-2">
                    <a
                      href="/admin/songs/{song.id}/variants"
                      class="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500 hover:text-on-surface transition-all border border-emerald-500/20 hover:border-emerald-500"
                      title="Variants"
                      >
                       <Video size={16} />
                      </a
                    >
                    <a
                      href="/admin/songs/{song.id}/edit"
                      class="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-surface-highest text-on-surface-variant/70 hover:bg-surface-highest hover:text-on-surface transition-all border border-outline-variant hover:border-outline-variant"
                      title="Edit Song"
                      >
                       <Edit2 size={16} />
                      </a
                    >
                    <button
                      onclick={() => deleteSong(song.id)}
                      class="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-surface-highest text-on-surface-variant/70 hover:bg-rose-500/10 hover:text-rose-400 transition-all border border-outline-variant hover:border-rose-500/20"
                      title="Delete Song"
                      >
                       <Trash2 size={16} />
                      </button
                    >
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {:else}
        <div class="px-6 py-16 flex flex-col items-center justify-center text-center">
          <div class="w-16 h-16 bg-surface-highest rounded-2xl flex items-center justify-center mb-4 border border-outline-variant">
              <Music size={32} class="text-on-surface-variant/40" />
          </div>
          <h3 class="text-on-surface font-medium mb-1">No songs registered</h3>
          <p class="text-sm text-on-surface-variant/40 max-w-sm mb-6">Create the first song for this anime, like an Opening or Ending theme.</p>
          {#if !anime.season_id || !anime.year_id}
            <button
              disabled
              class="px-5 py-2.5 bg-surface-highest border border-outline-variant text-on-surface-variant/40 rounded-xl text-sm font-medium cursor-not-allowed flex items-center gap-2"
            >
              <Plus size={16} />
              Create First Song
            </button>
          {:else}
            <a
              href="/admin/songs/create?anime_id={anime.id}"
              class="px-5 py-2.5 bg-primary hover:bg-primary-container text-on-surface rounded-xl text-sm font-medium transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
            >
              <Plus size={16} />
              Create First Song
            </a>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>
