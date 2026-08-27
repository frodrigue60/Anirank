<script lang="ts">
  import { page } from "$app/state";
  import { onMount } from "svelte";
  import api from "$lib/api";
  import type { Tournament } from "$lib/types/tournament";
  import TournamentBracket from "$lib/components/tournaments/TournamentBracket.svelte";
  import { authState } from "$lib/state/auth.svelte";
  import { toastState } from "$lib/state/toast.svelte";
  import type { Song } from "$lib/types/song";
  import { getSongName } from "$lib/song-utils";
  import SEO from "$lib/components/SEO.svelte";
  import Trophy from "lucide-svelte/icons/trophy";
import X from "lucide-svelte/icons/x";
import Play from "lucide-svelte/icons/play";
import Music from "lucide-svelte/icons/music";
import Info from "lucide-svelte/icons/info";
import Users from "lucide-svelte/icons/users";
import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
import VideoOff from "lucide-svelte/icons/video-off";;

  let tournament = $state<Tournament | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  // Preview logic
  let previewSong = $state<Song | null>(null);
  let selectedVariantIndex = $state(0);
  let videoElement = $state<HTMLVideoElement | undefined>();

  // Vote confirmation logic
  let confirmingVote = $state<{
    matchupId: number | string;
    songId: number | string;
    songName: string;
  } | null>(null);

  onMount(async () => {
    try {
      const slug = page.params.slug;
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

  function handleVoteConfirm(data: { matchupId: number | string, songId: number | string, song: Song | undefined }) {
    const { matchupId, songId, song } = data;
    confirmingVote = {
      matchupId,
      songId,
      songName: getSongName(song),
    };
  }

  async function executeVote() {
    if (!confirmingVote) return;
    try {
      loading = true;
      console.log("EXECUTING VOTE FOR MATCHUP:", confirmingVote.matchupId);
      await api.post(`/tournaments/matchups/${confirmingVote.matchupId}/vote`, {
        song_id: confirmingVote.songId,
      });

      // Refresh tournament data
      const slug = page.params.slug;
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
  description={tournament?.description ||
    "Vote in this anime theme song tournament on AniRank."}
/>

<main class="w-full min-h-screen pb-24">
  {#if loading && !tournament}
    <div
      class="flex flex-col items-center justify-center min-h-[400px] gap-4 opacity-50"
    >
      <div
        class="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin"
      ></div>
      <p class="text-sm font-black uppercase tracking-widest">
        Loading bracket...
      </p>
    </div>
  {:else if error}
    <div
      class="flex flex-col items-center justify-center min-h-[400px] gap-4 text-red-500"
    >
      <Info size={48} />
      <p class="font-bold">{error}</p>
    </div>
  {:else if tournament}
    <!-- Header Section -->
    <header
      class="relative pt-20 pb-16 px-6 overflow-hidden border-b border-white/5"
    >
      <div
        class="absolute inset-0 bg-linear-to-b from-primary/10 via-transparent to-transparent pointer-events-none"
      ></div>

      <div
        class="max-w-[1440px] mx-auto relative z-10 flex flex-col items-center text-center"
      >
        <span
          class="px-4 py-1.5 rounded-full text-[10px] font-black uppercase tracking-[0.2em] mb-6 shadow-lg
          {tournament.status === 'active' ? 'bg-primary text-white' : ''}
          {tournament.status === 'completed' ? 'bg-green-500 text-white' : ''}
          {tournament.status === 'draft'
            ? 'bg-surface-highest text-on-surface-variant'
            : ''}"
        >
          {tournament.status}
        </span>

        <h1
          class="text-5xl lg:text-7xl font-black mb-6 tracking-tighter bg-linear-to-b from-white to-white/60 bg-clip-text text-transparent"
        >
          {tournament.name}
        </h1>

        <p
          class="text-on-surface-variant text-lg max-w-2xl leading-relaxed opacity-70"
        >
          {tournament.description || ""}
        </p>

        {#if tournament.status === "completed" && tournament.winner_song_id}
          <!-- Champion Section -->
          <div class="mt-16 flex flex-col items-center">
            <div class="flex items-center gap-3 mb-6">
              <Trophy class="text-yellow-500 w-8 h-8" />
              <span
                class="text-xl font-black uppercase tracking-[0.2em] text-yellow-500"
                >Tournament Champion</span
              >
            </div>

            <a
              href={`/animes/${tournament.winner?.anime?.slug}/${tournament.winner?.slug}`}
              class="group relative w-full max-w-[700px] aspect-21/9 rounded-md overflow-hidden bg-surface-container border border-yellow-500/20 shadow-[0_0_50px_rgba(234,179,8,0.1)] hover:scale-[1.02] transition-all duration-500"
            >
              <div
                class="absolute inset-0 bg-cover bg-center group-hover:scale-110 transition-transform duration-700"
                style="background-image: url({tournament.winner?.anime
                  ?.banner_url})"
              ></div>
              <div
                class="absolute inset-0 bg-linear-to-t from-black via-black/40 to-transparent"
              ></div>

              <div
                class="absolute inset-0 p-8 flex flex-col justify-end text-left"
              >
                <div
                  class="px-3 py-1 bg-yellow-500 text-black text-[10px] font-black uppercase tracking-widest w-fit rounded-lg mb-3"
                >
                  {tournament.winner?.type}
                  {tournament.winner?.theme_num || ""}
                </div>
                <h2
                  class="text-3xl lg:text-4xl font-black text-white mb-2 group-hover:text-yellow-400 transition-colors"
                >
                  {getSongName(tournament.winner)}
                </h2>
                <div class="flex flex-col gap-1">
                  <p class="text-white/60 font-medium">
                    {#if tournament.winner?.artists && tournament.winner.artists.length > 0}
                      {tournament.winner.artists
                        .map((artist) => artist.name)
                        .join(", ")}
                    {:else}
                      Artists Info
                    {/if}
                  </p>
                  <p
                    class="text-yellow-500/80 font-black text-xs uppercase tracking-widest"
                  >
                    {tournament.winner?.anime?.title || "Anime Title"}
                  </p>
                </div>
              </div>

              <!-- Decoration -->
              <div
                class="absolute -inset-px border-2 border-yellow-500/10 rounded-md pointer-events-none"
              ></div>
            </a>
          </div>
        {/if}
      </div>
    </header>

    <!-- Bracket Area -->
    <section class="mt-12">
      <div class="max-w-[1440px] mx-auto">
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
            handleVoteConfirm({ matchupId, songId, song });
          }}
        />
      </div>
    </section>
  {/if}
</main>

<!-- Modern Modals -->
{#if previewSong}
  <div
    class="fixed inset-0 z-100 flex items-center justify-center p-6 bg-black/80 backdrop-blur-xl animate-in fade-in duration-300"
    role="dialog"
    aria-modal="true"
  >
    <div
      class="fixed inset-0"
      onclick={() => (previewSong = null)}
      onkeydown={(e) => e.key === "Escape" && (previewSong = null)}
      role="button"
      tabindex="0"
      aria-label="Close"
    ></div>

    <div
      class="relative w-full max-w-[1100px] bg-surface-container rounded-md border border-white/5 overflow-hidden shadow-2xl scale-in animate-in zoom-in-95 duration-300"
    >
      <button
        class="absolute top-6 right-6 z-20 w-12 h-12 flex items-center justify-center rounded-full bg-black/40 text-white hover:bg-primary transition-all"
        onclick={() => (previewSong = null)}
      >
        <X size={24} />
      </button>

      <div class="flex flex-col lg:flex-row">
        <!-- Video Part -->
        <div class="w-full lg:w-[65%] aspect-video bg-black relative">
          {#if selectedVariant?.video}
            {#if selectedVariant.video.local_url}
              <video
                bind:this={videoElement}
                src={selectedVariant.video.local_url}
                class="w-full h-full object-contain"
                controls
                autoplay
                onplay={fadeInVolume}
              >
                <track kind="captions" />
              </video>
            {:else}
              <div
                class="flex flex-col items-center justify-center h-full gap-4 text-white/20"
              >
                <VideoOff size={48} />
                <p>No video source available</p>
              </div>
            {/if}
          {:else}
            <div
              class="flex flex-col items-center justify-center h-full gap-4 text-white/20"
            >
              <VideoOff size={48} />
              <p>No video available for this variant</p>
            </div>
          {/if}
        </div>

        <!-- Info Part -->
        <div class="flex-1 p-10 flex flex-col gap-8 bg-surface-container/50">
          <div>
            <div class="flex flex-wrap items-center gap-3 mb-4">
              <span
                class="px-3 py-1 bg-primary/10 text-primary border border-primary/20 rounded-lg text-xs font-black uppercase tracking-widest"
              >
                {previewSong.type}
                {previewSong.theme_num || ""}
              </span>

              {#if (previewSong.song_variants?.length ?? 0) > 1}
                <div
                  class="flex gap-1 bg-surface-highest p-1 rounded-md border border-white/5"
                >
                  {#each previewSong.song_variants || [] as v, i}
                    <button
                      class="px-3 py-1.5 rounded-sm text-[10px] font-black tracking-widest transition-all
                      {selectedVariantIndex === i
                        ? 'bg-primary text-white shadow-lg shadow-primary/20'
                        : 'text-on-surface-variant hover:text-on-surface'}"
                      onclick={() => (selectedVariantIndex = i)}
                    >
                      V{v.version_number}
                    </button>
                  {/each}
                </div>
              {/if}
            </div>

            <h2 class="text-3xl font-black text-on-surface mb-2 leading-tight">
              {getSongName(previewSong)}
            </h2>
            <p class="text-on-surface-variant flex items-center gap-2">
              <Users size={14} class="text-primary" />
              {previewSong.artists?.map((a) => a.name).join(", ")}
            </p>
          </div>

          <div class="mt-auto pt-8 border-t border-on-surface-variant/5">
            <p
              class="text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant opacity-40 mb-1"
            >
              Featured In
            </p>
            <p class="text-on-surface font-black text-lg">
              {previewSong.anime?.title}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Vote Confirmation -->
{#if confirmingVote}
  <div
    class="fixed inset-0 z-100 flex items-center justify-center p-6 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200"
    role="dialog"
  >
    <div
      class="w-full max-w-[440px] bg-surface-container rounded-md border border-white/5 p-10 shadow-2xl scale-in animate-in zoom-in-95 duration-200"
    >
      <div class="flex flex-col items-center text-center gap-6">
        <div
          class="w-20 h-20 bg-primary/10 rounded-full flex items-center justify-center text-primary border border-primary/20"
        >
          <CheckCircle2 size={40} />
        </div>

        <div>
          <h3 class="text-2xl font-black text-on-surface mb-3">Confirm Vote</h3>
          <p class="text-on-surface-variant leading-relaxed">
            Are you sure you want to cast your vote for <br />
            <span class="text-on-surface font-black"
              >"{confirmingVote.songName}"</span
            >?
          </p>
        </div>

        <div class="flex w-full gap-3 mt-4">
          <button
            class="flex-1 py-4 bg-surface-highest text-on-surface-variant font-black uppercase tracking-widest text-xs rounded-sm border border-white/5 hover:bg-surface-highest/80 transition-all"
            onclick={() => (confirmingVote = null)}
          >
            Cancel
          </button>
          <button
            class="flex-1 py-4 bg-primary text-white font-black uppercase tracking-widest text-xs rounded-sm shadow-lg shadow-primary/20 hover:shadow-xl hover:shadow-primary/30 transition-all disabled:opacity-50"
            onclick={executeVote}
            disabled={loading}
          >
            {loading ? "Voting..." : "Confirm"}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  /* Removed large legacy style block in favor of Tailwind CSS */
</style>
