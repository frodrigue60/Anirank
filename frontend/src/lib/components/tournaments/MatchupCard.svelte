<script lang="ts">
  import type { TournamentMatchup } from "$lib/types/tournament";
  import type { Song } from "$lib/types/song";

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
</script>

<div class="matchup-card {matchup.is_active ? 'active' : 'inactive'}">
  <div
    class="song-slot {matchup.winner_song_id &&
    matchup.winner_song_id === matchup.song1_id
      ? 'winner'
      : ''}"
  >
    <div class="song-info">
      <div class="song-header">
        <span class="type-tag"
          >{matchup.song1?.type || ""} {matchup.song1?.theme_num || ""}</span
        >
        {#if matchup.user_voted_song_id === matchup.song1_id}
          <span class="voted-tag">Voted</span>
        {/if}
        <span class="name truncate" title={matchup.song1?.song_romaji}
          >{matchup.song1?.song_romaji ||
            matchup.song1?.song_en ||
            matchup.song1?.song_jp ||
            "Waiting..."}</span
        >
      </div>
      {#if matchup.song1?.artists}
        <div
          class="artists truncate"
          title={matchup.song1.artists.map((a) => a.name).join(", ")}
        >
          {#each matchup.song1.artists as artist, i}
            <span
              >{artist.name}{i < matchup.song1.artists.length - 1
                ? ", "
                : ""}</span
            >
          {/each}
        </div>
      {/if}
      {#if matchup.song1?.anime}
        <div class="anime-title truncate" title={matchup.song1.anime.title}>
          <span class="text-xs opacity-50 italic"
            >{matchup.song1.anime.title}</span
          >
        </div>
      {/if}
      {#if !matchup.is_active && totalVotes > 0}
        <span class="votes">{matchup.song1_votes} ({song1Percent}%)</span>
      {/if}
    </div>
    <div class="actions">
      {#if matchup.song1}
        <button
          class="preview-btn"
          onclick={() => onpreview?.(matchup.song1)}
          disabled={!hasPreview1}
          title={hasPreview1 ? "Preview" : "No preview available"}
        >
          <span class="material-symbols-outlined">
            {hasPreview1 ? "play_arrow" : "videocam_off"}
          </span>
        </button>
      {/if}
      {#if canVote && matchup.is_active && matchup.song1_id}
        <button
          class="vote-btn"
          onclick={() => handleVote(matchup.song1_id)}
          disabled={loading || !!matchup.user_voted_song_id}
        >
          {matchup.user_voted_song_id ? "Voted" : "Vote"}
        </button>
      {/if}
    </div>
  </div>

  <div class="vs-divider">VS</div>

  <div
    class="song-slot {matchup.winner_song_id &&
    matchup.winner_song_id === matchup.song2_id
      ? 'winner'
      : ''}"
  >
    <div class="song-info">
      <div class="song-header">
        <span class="type-tag"
          >{matchup.song2?.type || ""} {matchup.song2?.theme_num || ""}</span
        >
        {#if matchup.user_voted_song_id === matchup.song2_id}
          <span class="voted-tag">Voted</span>
        {/if}
        <span class="name truncate" title={matchup.song2?.song_romaji}
          >{matchup.song2?.song_romaji ||
            matchup.song2?.song_en ||
            matchup.song2?.song_jp ||
            "Waiting..."}</span
        >
      </div>
      {#if matchup.song2?.artists}
        <div
          class="artists truncate"
          title={matchup.song2.artists.map((a) => a.name).join(", ")}
        >
          {#each matchup.song2.artists as artist, i}
            <span
              >{artist.name}{i < matchup.song2.artists.length - 1
                ? ", "
                : ""}</span
            >
          {/each}
        </div>
      {/if}
      {#if matchup.song2?.anime}
        <div class="anime-title truncate" title={matchup.song2.anime.title}>
          <span class="text-xs opacity-50 italic"
            >{matchup.song2.anime.title}</span
          >
        </div>
      {/if}
      {#if !matchup.is_active && totalVotes > 0}
        <span class="votes">{matchup.song2_votes} ({song2Percent}%)</span>
      {/if}
    </div>
    <div class="actions">
      {#if matchup.song2}
        <button
          class="preview-btn"
          onclick={() => onpreview?.(matchup.song2)}
          disabled={!hasPreview2}
          title={hasPreview2 ? "Preview" : "No preview available"}
        >
          <span class="material-symbols-outlined">
            {hasPreview2 ? "play_arrow" : "videocam_off"}
          </span>
        </button>
      {/if}
      {#if canVote && matchup.is_active && matchup.song2_id}
        <button
          class="vote-btn"
          onclick={() => handleVote(matchup.song2_id)}
          disabled={loading || !!matchup.user_voted_song_id}
        >
          {matchup.user_voted_song_id ? "Voted" : "Vote"}
        </button>
      {/if}
    </div>
  </div>

  {#if matchup.ends_at && matchup.is_active}
    <div class="timer">
      Ends at: {new Date(matchup.ends_at).toLocaleString()}
    </div>
  {/if}
</div>

<style>
  .matchup-card {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-width: 320px;
    transition: all 0.2s ease;
  }

  .matchup-card.active {
    border-color: var(--primary-color, #ff4e50);
    box-shadow: 0 0 15px rgba(255, 78, 80, 0.1);
  }

  .song-slot {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.2);
    gap: 12px;
  }

  .song-slot.winner {
    background: rgba(255, 215, 0, 0.1);
    border: 1px solid rgba(255, 215, 0, 0.3);
  }

  .song-info {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
  }

  .song-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .type-tag {
    font-size: 0.65rem;
    font-weight: 900;
    text-transform: uppercase;
    color: var(--primary-color, #ff4e50);
    background: rgba(255, 78, 80, 0.1);
    padding: 2px 6px;
    border-radius: 4px;
    white-space: nowrap;
  }

  .voted-tag {
    font-size: 0.65rem;
    font-weight: 900;
    text-transform: uppercase;
    color: #4caf50;
    background: rgba(76, 175, 80, 0.1);
    padding: 2px 6px;
    border-radius: 4px;
    white-space: nowrap;
  }

  .name {
    font-weight: 700;
    font-size: 0.95rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .artists {
    font-size: 0.75rem;
    opacity: 0.5;
    margin-top: 2px;
  }

  .anime-title {
    margin-top: 1px;
    line-height: 1.2;
  }

  .truncate {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: block;
  }

  .votes {
    font-size: 0.8rem;
    opacity: 0.7;
    margin-top: 4px;
  }

  .actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .vs-divider {
    text-align: center;
    font-weight: 800;
    font-size: 0.8rem;
    opacity: 0.5;
    letter-spacing: 2px;
  }

  .vote-btn {
    background: var(--primary-color, #ff4e50);
    color: white;
    border: none;
    padding: 6px 14px;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: bold;
    cursor: pointer;
    transition: transform 0.1s;
  }

  .vote-btn:hover:not(:disabled) {
    transform: scale(1.05);
  }

  .preview-btn {
    background: rgba(255, 255, 255, 0.05);
    color: white;
    border: 1px solid rgba(255, 255, 255, 0.1);
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s;
  }

  .preview-btn:hover {
    background: var(--primary-color, #ff4e50);
    border-color: var(--primary-color, #ff4e50);
  }

  .preview-btn .material-symbols-outlined {
    font-size: 18px;
  }

  .preview-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
    background: rgba(255, 255, 255, 0.02);
  }

  .preview-btn:disabled:hover {
    background: rgba(255, 255, 255, 0.02);
    border-color: rgba(255, 255, 255, 0.1);
  }

  .vote-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .timer {
    font-size: 0.75rem;
    opacity: 0.6;
    text-align: center;
    margin-top: 4px;
  }
</style>
