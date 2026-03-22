<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import api from "$lib/api";
  import type { Tournament } from "$lib/types/tournament";
  import TournamentBracket from "$lib/components/tournaments/TournamentBracket.svelte";
  import { authState } from "$lib/state/auth.svelte";
  import { toastState } from "$lib/state/toast.svelte";
  import type { Song } from "$lib/types/song";
  import { getSongName } from "$lib/song-utils";
  import SEO from "$lib/components/SEO.svelte";

  let tournament = $state<Tournament | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  // Preview logic
  let previewSong = $state<Song | null>(null);
  let selectedVariantIndex = $state(0);
  let videoElement = $state<HTMLVideoElement | undefined>();

  // Vote confirmation logic
  let confirmingVote = $state<{
    matchupId: number;
    songId: number;
    songName: string;
  } | null>(null);

  onMount(async () => {
    try {
      const slug = $page.params.slug;
      const response = await api.get(`/tournaments/${slug}`);
      tournament = response.data.data;
    } catch (err) {
      console.error("Error fetching tournament:", err);
      error = "Could not load tournament details.";
    } finally {
      loading = false;
    }
  });

  function handlePreview(event: any) {
    previewSong = event.detail.song;
    selectedVariantIndex = 0;
  }

  function handleVoteConfirm(event: any) {
    const { matchupId, songId, song } = event.detail;
    confirmingVote = {
      matchupId,
      songId,
      songName: getSongName(song),
    };
  }

  async function executeVote() {
    if (!confirmingVote) return;
    try {
      loading = true; // Use a local loading state or disable button
      await api.post(`/tournaments/matchups/${confirmingVote.matchupId}/vote`, {
        song_id: confirmingVote.songId,
      });

      // Refresh tournament data or update local state
      const slug = $page.params.slug;
      const response = await api.get(`/tournaments/${slug}`);
      tournament = response.data.data;

      toastState.addToast(`Voted for ${confirmingVote.songName}`, "success");
    } catch (err: any) {
      console.error("Error voting:", err);
      toastState.addToast(
        err.response?.data?.message || "Failed to submit vote",
        "error",
      );
    } finally {
      loading = false;
      confirmingVote = null;
    }
  }

  function getAutoplayUrl(url: string | undefined) {
    if (!url) return "";
    const separator = url.includes("?") ? "&" : "?";
    return `${url}${separator}autoplay=1&muted=1`;
  }

  function fadeInVolume() {
    if (!videoElement) return;
    videoElement.volume = 0;
    videoElement.muted = false;
    let volume = 0;
    const interval = setInterval(() => {
      if (!videoElement) {
        clearInterval(interval);
        return;
      }
      volume += 0.05;
      if (volume >= 1) {
        videoElement.volume = 1;
        clearInterval(interval);
      } else {
        videoElement.volume = volume;
      }
    }, 100);
  }

  let canVote = $derived(!!authState.user && tournament?.status === "active");
  let selectedVariant = $derived(
    previewSong?.song_variants?.[selectedVariantIndex],
  );
</script>

<SEO 
  title={tournament ? `${tournament.name} Tournament` : "Tournament"} 
  description={tournament?.description || "Vote in this anime theme song tournament on AniRank."} 
/>

<div class="tournament-show">
  {#if loading && !tournament}
    <div class="loading">Loading tournament tree...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if tournament}
    <div class="header-section">
      <div class="container">
        <div class="status-tag {tournament.status}">{tournament.status}</div>
        <h1>{tournament.name}</h1>
        <p class="desc">{tournament.description || ""}</p>

        {#if tournament.status === "completed" && tournament.winner_song_id}
          <div class="champion-section flex flex-col items-center">
            <div class="champion-label flex items-center gap-2 mb-4">
              <span class="material-symbols-outlined text-yellow-500 text-3xl">trophy</span>
              <span class="text-xl font-black uppercase tracking-[0.2em] text-yellow-500/80">Tournament Champion</span>
            </div>
            
            <a 
              href={`/songs/${tournament.winner?.anime?.slug}/${tournament.winner?.slug}`}
              class="champion-card-wrapper group"
            >
              <div class="champion-card" style="--banner-url: url({tournament.winner?.anime?.banner_url})">
                <div class="champion-banner"></div>
                <div class="champion-details">
                  <div class="song-type-badge">
                    {tournament.winner?.type} {tournament.winner?.theme_num || ""}
                  </div>
                  <h2 class="song-title">
                    {getSongName(tournament.winner)}
                  </h2>
                  <div class="artist-list">
                    {#if tournament.winner?.artists && tournament.winner.artists.length > 0}
                      {tournament.winner.artists
                        .map((artist) => artist.name)
                        .join(", ")}
                    {:else}
                      Artists Info
                    {/if}
                  </div>
                  <div class="anime-title">
                    {tournament.winner?.anime?.title || "Anime Title"}
                  </div>
                </div>
              </div>
              <div class="card-glow"></div>
            </a>
          </div>
        {/if}
      </div>
    </div>

    <div class="bracket-section">
      <div class="container">
        <TournamentBracket
          {tournament}
          {canVote}
          onpreview={(song) => {
            if (song) {
              previewSong = song;
              selectedVariantIndex = 0;
            }
          }}
          onvoteRequest={(matchupId, songId, song) => {
            handleVoteConfirm({ detail: { matchupId, songId, song } });
          }}
        />
      </div>
    </div>
  {/if}
</div>

<!-- Preview Modal -->
{#if previewSong}
  <div
    class="modal-overlay"
    onclick={() => (previewSong = null)}
    onkeydown={(e) => e.key === "Escape" && (previewSong = null)}
    role="button"
    tabindex="0"
    aria-label="Close preview"
  >
    <div class="preview-modal" onclick={(e) => e.stopPropagation()} role="none">
      <button class="close-btn" onclick={() => (previewSong = null)}>
        <span class="material-symbols-outlined">close</span>
      </button>

      <div class="video-container">
        {#if selectedVariant?.video}
          {#if selectedVariant.video.local_url}
            <video
              bind:this={videoElement}
              src={selectedVariant.video.local_url}
              class="w-full h-full"
              controls
              autoplay
              onplay={fadeInVolume}
            >
              <track kind="captions" />
            </video>
          {:else if selectedVariant.video.embed_url}
            <iframe
              src={getAutoplayUrl(selectedVariant.video.embed_url)}
              class="w-full h-full"
              frameborder="0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowfullscreen
              title="Song Video"
            ></iframe>
          {:else}
            <div class="no-video">
              <span class="material-symbols-outlined">videocam_off</span>
              <p>No video source available (S3 or Embed)</p>
            </div>
          {/if}
        {:else}
          <div class="no-video">
            <span class="material-symbols-outlined">videocam_off</span>
            <p>No video available for this variant</p>
          </div>
        {/if}
      </div>

      <div class="preview-info">
        <div class="preview-header">
          <div class="tags">
            <span class="type-tag"
              >{previewSong.type} {previewSong.theme_num || ""}</span
            >
            {#if (previewSong.song_variants?.length ?? 0) > 1}
              <div class="versions">
                {#each previewSong.song_variants || [] as v, i}
                  <button
                    class="v-btn {selectedVariantIndex === i ? 'active' : ''}"
                    onclick={() => (selectedVariantIndex = i)}
                  >
                    V{v.version_number}
                  </button>
                {/each}
              </div>
            {/if}
          </div>
          <h2>{getSongName(previewSong)}</h2>
          {#if previewSong.artists}
            <p class="artists">
              By: {previewSong.artists.map((a) => a.name).join(", ")}
            </p>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Vote Confirmation Modal -->
{#if confirmingVote}
  <div
    class="modal-overlay"
    onclick={() => (confirmingVote = null)}
    onkeydown={(e) => e.key === "Escape" && (confirmingVote = null)}
    role="button"
    tabindex="0"
    aria-label="Cancel vote"
  >
    <div class="confirm-modal" onclick={(e) => e.stopPropagation()} role="none">
      <h3>Confirm your Vote</h3>
      <p>
        Are you sure you want to vote for <strong
          >{confirmingVote.songName}</strong
        >?
      </p>
      <div class="confirm-actions">
        <button class="cancel-btn" onclick={() => (confirmingVote = null)}
          >Cancel</button
        >
        <button class="confirm-btn" onclick={executeVote} disabled={loading}>
          {loading ? "Submitting..." : "Confirm Vote"}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .tournament-show {
    min-height: 100vh;
    padding-bottom: 80px;
  }

  .header-section {
    background: linear-gradient(to bottom, rgba(255, 78, 80, 0.1), transparent);
    padding: 60px 0;
    text-align: center;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 0 20px;
  }

  .status-tag {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 20px;
    font-size: 0.75rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 15px;
  }

  .status-tag.active {
    background: #4caf50;
    color: white;
  }
  .status-tag.draft {
    background: #ff9800;
    color: white;
  }
  .status-tag.completed {
    background: #2196f3;
    color: white;
  }

  h1 {
    font-size: 3rem;
    font-weight: 900;
    margin-bottom: 10px;
    background: linear-gradient(45deg, #fff, #ccc);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .desc {
    font-size: 1.1rem;
    color: rgba(255, 255, 255, 0.6);
    max-width: 700px;
    margin: 0 auto;
    line-height: 1.6;
  }

  .champion-section {
    margin-top: 40px;
  }

  .champion-card-wrapper {
    position: relative;
    display: block;
    text-decoration: none;
    transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  .champion-card-wrapper:hover {
    transform: scale(1.02);
  }

  .champion-card {
    position: relative;
    z-index: 2;
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    background: rgba(26, 26, 26, 0.8);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(255, 215, 0, 0.3);
    border-radius: 24px;
    overflow: hidden;
    width: 600px;
    height: 250px;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
  }

  .champion-banner {
    position: absolute;
    inset: 0;
    background-image: var(--banner-url);
    background-size: cover;
    background-position: center 20%;
    z-index: -1;
    transition: transform 0.5s ease;
  }

  .champion-card-wrapper:hover .champion-banner {
    transform: scale(1.1);
  }

  .champion-card::after {
    content: '';
    position: absolute;
    inset: 0;
    background: linear-gradient(to top, rgba(26, 26, 26, 1) 10%, rgba(26, 26, 26, 0.4) 50%, rgba(26, 26, 26, 0.1));
    z-index: 0;
  }

  .champion-details {
    position: relative;
    z-index: 1;
    padding: 30px;
    text-align: left;
  }

  .song-type-badge {
    display: inline-block;
    color: #ffd700;
    font-weight: 900;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 2px;
    margin-bottom: 8px;
  }

  .champion-details h2.song-title {
    font-size: 1.8rem;
    font-weight: 900;
    line-height: 1.1;
    margin-bottom: 6px;
    background: linear-gradient(to right, #fff, #ffd700);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .artist-list {
    font-size: 1rem;
    color: rgba(255, 255, 255, 0.6);
    margin-bottom: 10px;
    font-weight: 500;
  }

  .anime-title {
    font-size: 0.9rem;
    color: #ffd700;
    font-weight: 700;
    opacity: 0.8;
  }

  .card-glow {
    position: absolute;
    inset: -20px;
    background: radial-gradient(circle at center, rgba(255, 215, 0, 0.15) 0%, transparent 70%);
    z-index: 1;
    opacity: 0;
    transition: opacity 0.3s;
  }

  .champion-card-wrapper:hover .card-glow {
    opacity: 1;
  }

  .bracket-section {
    padding: 60px 0;
  }

  /* Modal Styles */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(10px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 20px;
  }

  .preview-modal {
    background: #1a1a1a;
    width: 100%;
    max-width: 1000px;
    border-radius: 20px;
    overflow: hidden;
    position: relative;
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  }

  .video-container {
    width: 100%;
    aspect-ratio: 16/9;
    background: #000;
    position: relative;
  }

  .no-video {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: rgba(255, 255, 255, 0.3);
  }

  .close-btn {
    position: absolute;
    top: 20px;
    right: 20px;
    background: rgba(0, 0, 0, 0.5);
    border: none;
    color: white;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    z-index: 10;
    transition: background 0.2s;
  }

  .close-btn:hover {
    background: #ff4e50;
  }

  .preview-info {
    padding: 30px;
  }

  .preview-header {
    margin-bottom: 20px;
  }

  .tags {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .type-tag {
    color: var(--primary-color, #ff4e50);
    font-weight: 900;
    text-transform: uppercase;
    font-size: 0.8rem;
    letter-spacing: 1px;
    background: rgba(255, 78, 80, 0.1);
    padding: 4px 10px;
    border-radius: 6px;
  }

  .versions {
    display: flex;
    gap: 8px;
  }

  .v-btn {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: white;
    padding: 4px 10px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.8rem;
    transition: all 0.2s;
  }

  .v-btn.active {
    background: var(--primary-color, #ff4e50);
    border-color: var(--primary-color, #ff4e50);
  }

  h2 {
    font-size: 1.8rem;
    font-weight: 800;
    margin-bottom: 5px;
  }
  .artists {
    color: rgba(255, 255, 255, 0.6);
    font-size: 1rem;
  }

  /* Confirm Modal */
  .confirm-modal {
    background: #1a1a1a;
    padding: 40px;
    border-radius: 20px;
    text-align: center;
    max-width: 450px;
    width: 100%;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .confirm-modal h3 {
    font-size: 1.5rem;
    font-weight: 800;
    margin-bottom: 15px;
  }
  .confirm-modal p {
    color: rgba(255, 255, 255, 0.7);
    margin-bottom: 30px;
    line-height: 1.6;
  }

  .confirm-actions {
    display: flex;
    gap: 15px;
  }

  .confirm-actions button {
    flex: 1;
    padding: 14px;
    border-radius: 12px;
    font-weight: 800;
    cursor: pointer;
    transition: all 0.2s;
  }

  .cancel-btn {
    background: rgba(255, 255, 255, 0.05);
    color: white;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }
  .cancel-btn:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .confirm-btn {
    background: var(--primary-color, #ff4e50);
    color: white;
    border: none;
  }
  .confirm-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 10px 20px rgba(255, 78, 80, 0.3);
  }
  .confirm-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
  }

  .loading,
  .error {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 400px;
    font-size: 1.2rem;
    opacity: 0.6;
  }

  .error {
    color: #f44336;
  }
</style>
