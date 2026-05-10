<script lang="ts">
  import type { TournamentMatchup } from "$lib/types/tournament";
  import type { Song } from "$lib/types/song";
  import Play from "lucide-svelte/icons/play";
import VideoOff from "lucide-svelte/icons/video-off";
import CheckCircle2 from "lucide-svelte/icons/check-circle-2";

  import { getSongName } from "$lib/song-utils";

  interface Props {
    matchup: TournamentMatchup;
    canVote?: boolean;
    loading?: boolean;
    onvoteRequest?: (
      matchupId: number,
      songId: number,
      song: Song | undefined,
    ) => void;
    onpreview?: (song: Song | undefined) => void;
  }

  let {
    matchup,
    canVote = false,
    loading = false,
    onvoteRequest,
    onpreview,
  }: Props = $props();

  function handleVote(songId: number | null) {
    if (!canVote || !songId || loading || matchup.user_voted_song_id) return;
    const song = songId === matchup.song1_id ? matchup.song1 : matchup.song2;
    onvoteRequest?.(matchup.id, songId, song);
  }

  let totalVotes = $derived(matchup.song1_votes + matchup.song2_votes);
  let song1Percent = $derived(
    totalVotes > 0 ? Math.round((matchup.song1_votes / totalVotes) * 100) : 0,
  );
  let song2Percent = $derived(
    totalVotes > 0 ? Math.round((matchup.song2_votes / totalVotes) * 100) : 0,
  );

  let hasPreview1 = $derived(
    matchup.song1?.song_variants?.some(
      (v: any) => !!(v.video?.local_url || v.video?.embed_url),
    ) ?? false,
  );
  let hasPreview2 = $derived(
    matchup.song2?.song_variants?.some(
      (v: any) => !!(v.video?.local_url || v.video?.embed_url),
    ) ?? false,
  );

  // Helper for song names
  function songDisplayName(song: Song | undefined) {
    if (!song) return "Waiting...";
    const name = getSongName(song);
    return name === "N/A" ? "Untitled" : name;
  }
</script>

<div
  class="group relative flex flex-col gap-1 w-[320px] rounded-2xl p-1 transition-all duration-300
  {matchup.is_active
    ? 'bg-linear-to-b from-primary/30 to-transparent shadow-[0_0_20px_rgba(255,78,80,0.1)]'
    : 'bg-white/5'}"
>
  <!-- Slot 1 -->
  <div
    class="relative overflow-hidden rounded-xl bg-surface-container border border-white/5 p-4 flex flex-col gap-3 transition-all
    {matchup.winner_song_id === matchup.song1_id
      ? 'border-yellow-500/50 bg-yellow-500/5 ring-1 ring-yellow-500/20'
      : ''}"
  >
    <div class="flex justify-between items-start gap-4">
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-1">
          <span
            class="text-[9px] font-black uppercase tracking-widest px-1.5 py-0.5 rounded bg-primary/10 text-primary border border-primary/10"
          >
            {matchup.song1?.type || "OP"}
            {matchup.song1?.theme_num || ""}
          </span>
          {#if matchup.user_voted_song_id === matchup.song1_id}
            <span
              class="flex items-center gap-1 text-[9px] font-black uppercase tracking-widest text-green-500"
            >
              <CheckCircle2 size={10} />
              Voted
            </span>
          {/if}
        </div>
        <h4
          class="text-sm font-bold text-on-surface truncate group-hover:text-primary transition-colors"
          title={songDisplayName(matchup.song1)}
        >
          {songDisplayName(matchup.song1)}
        </h4>
        <p class="text-[10px] text-on-surface-variant truncate opacity-60">
          {#if matchup.song1?.artists}
            {matchup.song1.artists.map((a) => a.name).join(", ")}
          {:else}
            Artists Info
          {/if}
        </p>
      </div>

      <div class="flex gap-2">
        {#if matchup.song1}
          <button
            class="p-2 rounded-lg bg-surface-highest text-on-surface-variant hover:bg-primary hover:text-white transition-all disabled:opacity-20"
            onclick={() => onpreview?.(matchup.song1)}
            disabled={!hasPreview1}
          >
            {#if hasPreview1}
              <Play size={14} fill="currentColor" />
            {:else}
              <VideoOff size={14} />
            {/if}
          </button>
        {/if}
      </div>
    </div>

    {#if !matchup.is_active && totalVotes > 0}
      <div class="w-full h-1.5 bg-surface-highest rounded-full overflow-hidden">
        <div class="h-full bg-yellow-500 transition-all duration-1000" style="width: {song1Percent}%"></div>
      </div>
      <div class="flex justify-between text-[10px] font-black text-on-surface-variant opacity-50 uppercase tracking-widest">
        <span>{song1Percent}%</span>
        <span>{matchup.song1_votes} Votes</span>
      </div>
    {/if}

    {#if canVote && matchup.is_active && matchup.song1_id && !matchup.user_voted_song_id}
      <button
        class="w-full py-2 bg-primary text-white text-xs font-black uppercase tracking-widest rounded-lg hover:shadow-lg hover:shadow-primary/20 transition-all mt-1"
        onclick={() => handleVote(matchup.song1_id)}
        disabled={loading}
      >
        Vote
      </button>
    {/if}
  </div>

  <!-- Divider -->
  <div class="flex items-center justify-center -my-3 relative z-10">
    <div class="px-3 py-1 rounded-full bg-surface-highest border border-white/10 text-[9px] font-black text-on-surface-variant tracking-[0.2em] uppercase shadow-lg">
      VS
    </div>
  </div>

  <!-- Slot 2 -->
  <div
    class="relative overflow-hidden rounded-xl bg-surface-container border border-white/5 p-4 flex flex-col gap-3 transition-all
    {matchup.winner_song_id === matchup.song2_id
      ? 'border-yellow-500/50 bg-yellow-500/5 ring-1 ring-yellow-500/20'
      : ''}"
  >
    <div class="flex justify-between items-start gap-4">
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-1">
          <span
            class="text-[9px] font-black uppercase tracking-widest px-1.5 py-0.5 rounded bg-primary/10 text-primary border border-primary/10"
          >
            {matchup.song2?.type || "ED"}
            {matchup.song2?.theme_num || ""}
          </span>
          {#if matchup.user_voted_song_id === matchup.song2_id}
            <span
              class="flex items-center gap-1 text-[9px] font-black uppercase tracking-widest text-green-500"
            >
              <CheckCircle2 size={10} />
              Voted
            </span>
          {/if}
        </div>
        <h4
          class="text-sm font-bold text-on-surface truncate group-hover:text-primary transition-colors"
          title={songDisplayName(matchup.song2)}
        >
          {songDisplayName(matchup.song2)}
        </h4>
        <p class="text-[10px] text-on-surface-variant truncate opacity-60">
          {#if matchup.song2?.artists}
            {matchup.song2.artists.map((a) => a.name).join(", ")}
          {:else}
            Artists Info
          {/if}
        </p>
      </div>

      <div class="flex gap-2">
        {#if matchup.song2}
          <button
            class="p-2 rounded-lg bg-surface-highest text-on-surface-variant hover:bg-primary hover:text-white transition-all disabled:opacity-20"
            onclick={() => onpreview?.(matchup.song2)}
            disabled={!hasPreview2}
          >
            {#if hasPreview2}
              <Play size={14} fill="currentColor" />
            {:else}
              <VideoOff size={14} />
            {/if}
          </button>
        {/if}
      </div>
    </div>

    {#if !matchup.is_active && totalVotes > 0}
      <div class="w-full h-1.5 bg-surface-highest rounded-full overflow-hidden">
        <div class="h-full bg-yellow-500 transition-all duration-1000" style="width: {song2Percent}%"></div>
      </div>
      <div class="flex justify-between text-[10px] font-black text-on-surface-variant opacity-50 uppercase tracking-widest">
        <span>{song2Percent}%</span>
        <span>{matchup.song2_votes} Votes</span>
      </div>
    {/if}

    {#if canVote && matchup.is_active && matchup.song2_id && !matchup.user_voted_song_id}
      <button
        class="w-full py-2 bg-primary text-white text-xs font-black uppercase tracking-widest rounded-lg hover:shadow-lg hover:shadow-primary/20 transition-all mt-1"
        onclick={() => handleVote(matchup.song2_id)}
        disabled={loading}
      >
        Vote
      </button>
    {/if}
  </div>

  {#if matchup.ends_at && matchup.is_active}
    <div class="text-[9px] font-black text-on-surface-variant/40 uppercase tracking-widest text-center mt-2 px-4 italic">
      Ends {new Date(matchup.ends_at).toLocaleDateString()}
    </div>
  {/if}
</div>
