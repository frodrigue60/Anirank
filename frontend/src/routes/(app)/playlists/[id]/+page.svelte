<script lang="ts">
  import api from "$lib/api";
  import { page } from "$app/state";
  import { authState } from "$lib/state/auth.svelte";
  import { getSongName, getFormattedScore } from "$lib/song-utils";
  import { goto } from "$app/navigation";
  import ReportModal from "$lib/components/ReportModal.svelte";
  import type { Song } from "$lib/types/song";
  import { toastState } from "$lib/state/toast.svelte";
  import Play from "lucide-svelte/icons/play";
  import Star from "lucide-svelte/icons/star";
  import Heart from "lucide-svelte/icons/heart";
  import ThumbsUp from "lucide-svelte/icons/thumbs-up";
  import ThumbsDown from "lucide-svelte/icons/thumbs-down";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import AlertTriangle from "lucide-svelte/icons/alert-triangle";
  import Share2 from "lucide-svelte/icons/share-2";
  import Music2 from "lucide-svelte/icons/music-2";
  import AudioLines from "lucide-svelte/icons/audio-lines";
  import Shuffle from "lucide-svelte/icons/shuffle";
  import SkipBack from "lucide-svelte/icons/skip-back";
  import SkipForward from "lucide-svelte/icons/skip-forward";
  import PlayCircle from "lucide-svelte/icons/play-circle";
  import PauseCircle from "lucide-svelte/icons/pause-circle";
  import Pause from "lucide-svelte/icons/pause";
  import ListMinus from "lucide-svelte/icons/list-minus";
  import Share from "lucide-svelte/icons/share";

  import SEO from "$lib/components/SEO.svelte";
  import { PUBLIC_API_URL } from "$lib/api";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let playlist = $state(data.playlist);
  // svelte-ignore state_referenced_locally
  let songs = $state((playlist.songs || []) as Song[]);
  // svelte-ignore state_referenced_locally
  let currentSong = $state(songs[0] || null);
  let openMenuId = $state<number | null>(null);

  let selectedVariantIndex = $state(0);
  let selectedVariant = $derived(
    currentSong?.song_variants?.[selectedVariantIndex],
  );

  let showReportModal = $state(false);

  function reportSong() {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    if (!currentSong) return;
    showReportModal = true;
  }

  function shareSong() {
    if (!currentSong?.anime) return;
    const url = `${window.location.origin}/animes/${currentSong.anime.slug}/${currentSong.slug}`;
    navigator.clipboard
      .writeText(url)
      .then(() => {
        toastState.addToast("Link copied to clipboard!", "success");
      })
      .catch((err) => {
        console.error("Failed to copy text: ", err);
        toastState.addToast("Failed to copy link", "error");
      });
  }

  let videoElement: HTMLVideoElement | undefined = $state();
  let maxVolume = $state(1); // Default max volume

  import { browser } from "$app/environment";

  let isDesktop = $state(true);

  $effect(() => {
    if (!browser) return;
    const mql = window.matchMedia("(min-width: 1024px)");
    isDesktop = mql.matches;
    const handler = (e: MediaQueryListEvent) => (isDesktop = e.matches);
    mql.addEventListener("change", handler);
    return () => mql.removeEventListener("change", handler);
  });

  $effect(() => {
    if (browser) {
      const storedVolume = localStorage.getItem("anirank_volume");
      if (storedVolume !== null) {
        maxVolume = parseFloat(storedVolume);
      }
    }
  });

  function updateMaxVolume(event: Event) {
    const target = event.target as HTMLInputElement;
    maxVolume = parseFloat(target.value);
    if (browser) {
      localStorage.setItem("anirank_volume", maxVolume.toString());
    }
    if (videoElement) {
      videoElement.volume = maxVolume; // Update immediately if playing
    }
  }

  let isPaused = $state(false);
  let isShuffle = $state(false);
  let shuffleSequence: number[] = $state([]);
  let currentShuffleIndex = $state(0);

  function generateShuffleSequence() {
    let indices = songs.map((_: any, i: number) => i);
    for (let i = indices.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [indices[i], indices[j]] = [indices[j], indices[i]];
    }
    return indices;
  }

  function toggleShuffle() {
    isShuffle = !isShuffle;
    if (isShuffle && songs.length > 0) {
      shuffleSequence = generateShuffleSequence();

      if (currentSong) {
        const currentRealIndex = songs.findIndex(
          (s: any) => s.id === currentSong?.id,
        );
        if (currentRealIndex !== -1) {
          shuffleSequence = shuffleSequence.filter(
            (idx) => idx !== currentRealIndex,
          );
          shuffleSequence.unshift(currentRealIndex);
        }
      }
      currentShuffleIndex = 0;
    }
  }

  function togglePlay() {
    if (!videoElement) return;
    if (isPaused) {
      videoElement.play();
    } else {
      videoElement.pause();
    }
  }

  function playNext() {
    if (songs.length === 0) return;

    if (isShuffle) {
      currentShuffleIndex++;
      if (currentShuffleIndex >= shuffleSequence.length) {
        // We reached the end of the random sequence
        shuffleSequence = generateShuffleSequence();
        currentShuffleIndex = 0;
        playSong(songs[shuffleSequence[0]]);

        // Pause to indicate the playlist ended, waiting for user Play
        setTimeout(() => {
          if (videoElement) {
            videoElement.pause();
            isPaused = true;
          }
        }, 100);
        return;
      }
      playSong(songs[shuffleSequence[currentShuffleIndex]]);
    } else {
      const currentIndex = songs.findIndex(
        (s: any) => s.id === currentSong?.id,
      );
      if (currentIndex === -1) return;
      const nextIndex = (currentIndex + 1) % songs.length;
      playSong(songs[nextIndex]);
    }
  }

  function playPrev() {
    if (songs.length === 0) return;

    if (isShuffle) {
      if (currentShuffleIndex > 0) {
        currentShuffleIndex--;
        playSong(songs[shuffleSequence[currentShuffleIndex]]);
      } else {
        if (videoElement) {
          videoElement.currentTime = 0;
        }
      }
    } else {
      const currentIndex = songs.findIndex(
        (s: any) => s.id === currentSong?.id,
      );
      if (currentIndex === -1) return;
      const prevIndex = (currentIndex - 1 + songs.length) % songs.length;
      playSong(songs[prevIndex]);
    }
  }

  $effect(() => {
    // Reset variant index when song changes to ensure we play the first variant
    if (currentSong) {
      selectedVariantIndex = 0;
    }
  });

  function toggleMenu(id: number, e: MouseEvent) {
    e.stopPropagation();
    openMenuId = openMenuId === id ? null : id;
  }

  function closeMenu() {
    openMenuId = null;
  }

  async function removeFromPlaylist(songId: number) {
    try {
      await api.delete(`/playlists/${playlist.id}/songs/${songId}`);
      songs = songs.filter((s: any) => s.id !== songId);
      openMenuId = null;
      toastState.addToast("Song removed from playlist", "info");
    } catch (e: any) {
      console.error("Failed to remove song", e);
      toastState.addToast(
        e.response?.data?.message || "Failed to remove song",
        "error",
      );
    }
  }

  function playSong(song: any) {
    currentSong = song;
    selectedVariantIndex = 0;

    if (isShuffle) {
      const idx = songs.findIndex((s: any) => s.id === song.id);
      const sequenceIdx = shuffleSequence.indexOf(idx);
      if (sequenceIdx !== -1) {
        currentShuffleIndex = sequenceIdx;
      }
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
      if (volume >= maxVolume) {
        videoElement.volume = maxVolume;
        clearInterval(interval);
      } else {
        videoElement.volume = volume;
      }
    }, 100);
  }

  async function toggleLike() {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    if (!currentSong) return;

    try {
      const resp = await api.post(`/interactions/reactions`, {
        entity_id: currentSong.id,
        entity_type: "song",
        type: "like",
      });
      if (resp.data.success) {
        currentSong.is_liked = !currentSong.is_liked;
        if (currentSong.is_liked) currentSong.is_disliked = false;
        currentSong.likes_count = resp.data.likesCount;
        currentSong.dislikes_count = resp.data.dislikesCount;
        toastState.addToast(
          currentSong.is_liked ? "Song liked!" : "Like removed!",
          "success",
        );
      }
    } catch (e: any) {
      console.error(e);
      toastState.addToast(
        e.response?.data?.message || "Failed to update like status",
        "error",
      );
    }
  }

  async function toggleDislike() {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    if (!currentSong) return;

    try {
      const resp = await api.post(`/interactions/reactions`, {
        entity_id: currentSong.id,
        entity_type: "song",
        type: "dislike",
      });
      if (resp.data.success) {
        currentSong.is_disliked = !currentSong.is_disliked;
        if (currentSong.is_disliked) currentSong.is_liked = false;
        currentSong.likes_count = resp.data.likesCount;
        currentSong.dislikes_count = resp.data.dislikesCount;
        toastState.addToast(
          currentSong.is_disliked ? "Song disliked!" : "Dislike removed!",
          "success",
        );
      }
    } catch (e: any) {
      console.error(e);
      toastState.addToast(
        e.response?.data?.message || "Failed to update dislike status",
        "error",
      );
    }
  }

  async function toggleFavorite() {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    if (!currentSong) return;
    try {
      const resp = await api.post(`/interactions/favorites`, {
        entity_id: currentSong.id,
        entity_type: "song",
      });
      if (resp.data.success || resp.status === 200 || resp.status === 201) {
        currentSong.is_favorited = resp.data.favorited || resp.data.favorite;
        toastState.addToast(
          currentSong.is_favorited
            ? "Added to favorites!"
            : "Removed from favorites",
          "success",
        );
      }
    } catch (e: any) {
      console.error(e);
      toastState.addToast(
        e.response?.data?.message || "Failed to update favorites",
        "error",
      );
    }
  }
</script>

<SEO
  title={`${playlist.name} - Playlist by ${playlist.user?.name || "User"} - AniRank`}
  description={`Listen to the "${playlist.name}" playlist curated by ${playlist.user?.name || "User"} on AniRank. Featuring ${songs.length} anime theme songs.`}
  image={`${PUBLIC_API_URL}/og/playlist/${playlist.id}`}
/>

<svelte:window onclick={closeMenu} />

<!-- ==================== DESKTOP LAYOUT (lg+) ==================== -->
{#if isDesktop}
  <div
    class="flex h-[calc(100vh-64px)] overflow-hidden bg-background-dark antialiased"
  >
    <!-- Left Sidebar: Now Playing -->
    <aside
      class="w-[60%] flex flex-col p-6 overflow-y-auto border-r border-white/5 bg-linear-to-b from-background-dark to-surface-darker"
    >
      <div
        class="relative aspect-video rounded-md overflow-hidden bg-black shadow-2xl mb-6 group neon-border"
      >
        {#if selectedVariant?.video}
          {#if selectedVariant.video.type === "embed"}
            <iframe
              src={getAutoplayUrl(selectedVariant.video.embed_url)}
              class="w-full h-full"
              frameborder="0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowfullscreen
              title="Song Video"
            ></iframe>
          {:else}
            <video
              bind:this={videoElement}
              bind:paused={isPaused}
              src={selectedVariant.video.local_url}
              class="w-full h-full"
              controls
              autoplay
              muted
              onplay={fadeInVolume}
              onended={playNext}
            >
              <track kind="captions" />
            </video>
          {/if}
        {:else if currentSong}
          <OptimizedImage
            src={currentSong.anime?.banner_url}
            sources={currentSong.anime?.banner_sources}
            alt={currentSong.title}
            class="w-full h-full object-cover"
            sizes="(max-width: 1024px) 100vw, 60vw"
          />
          <div
            class="absolute inset-0 bg-black/40 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
          >
            <Play size={64} fill="currentColor" />
          </div>
        {:else}
          <div
            class="w-full h-full flex items-center justify-center text-white/5 bg-surface-dark"
          >
            <Play size={64} fill="currentColor" />
          </div>
        {/if}
      </div>

      <div class="flex flex-col gap-6">
        <div class="flex justify-between items-start">
          <div class="flex flex-col">
            <h1
              class="text-3xl font-black text-white tracking-tight leading-tight mb-1 neon-text"
            >
              {getSongName(currentSong)}
            </h1>
            <span class="text-lg font-bold text-primary">
              {currentSong?.anime?.title || "No Anime info"}
            </span>
            {#if currentSong?.artists}
              {#each currentSong.artists as artist}
                <span class="text-white/60 font-medium">
                  {artist.name}
                </span>
              {/each}
            {/if}
          </div>
          {#if currentSong}
            <div class="flex flex-col items-end gap-1">
              <div
                class="flex items-center gap-1.5 text-yellow-400 font-black text-xl"
              >
                <Star size={20} class="fill-yellow-400" />
                <span
                  >{getFormattedScore(
                    currentSong.average_rating,
                    authState.user?.score_format,
                  )}</span
                >
              </div>
            </div>
          {/if}
        </div>

        <div
          class="flex items-center justify-between p-4 rounded-md bg-white/5 border border-white/10"
        >
          <div class="flex items-center gap-6">
            <button
              class="flex flex-col items-center gap-1 group"
              onclick={toggleFavorite}
            >
                <Heart
                  size={20}
                  class={currentSong?.is_favorited ? "fill-primary text-primary" : "text-white/40 group-hover:text-primary"}
                />
              <span
                class="text-[10px] font-bold uppercase tracking-tighter {currentSong?.is_favorited
                  ? 'text-primary'
                  : 'text-white/40'}"
              >
                Favorite
              </span>
            </button>
            <button
              class="flex flex-col items-center gap-1 group"
              onclick={toggleLike}
            >
                <ThumbsUp
                  size={20}
                  class={currentSong?.is_liked ? "fill-primary text-primary" : "text-white/40 group-hover:text-primary"}
                />
              <span
                class="text-[10px] font-bold uppercase tracking-tighter {currentSong?.is_liked
                  ? 'text-primary'
                  : 'text-white/40'}"
              >
                Like
              </span>
            </button>
            <button
              class="flex flex-col items-center gap-1 group"
              onclick={toggleDislike}
            >
                <ThumbsDown
                  size={20}
                  class={currentSong?.is_disliked ? "fill-primary text-primary" : "text-white/40 group-hover:text-primary"}
                />
              <span
                class="text-[10px] font-bold uppercase tracking-tighter {currentSong?.is_disliked
                  ? 'text-primary'
                  : 'text-white/40'}"
              >
                Dislike
              </span>
            </button>
          </div>
          <div class="flex items-center gap-4">
            <button
              onclick={() => currentSong && removeFromPlaylist(currentSong.id)}
              class="w-10 h-10 rounded-full border border-white/10 flex items-center justify-center text-red-400 hover:bg-red-500/10 transition-colors"
              title="Remove from Playlist"
            >
              <ListMinus size={18} />
            </button>
            <button
              onclick={reportSong}
              class="w-10 h-10 rounded-full border border-white/10 flex items-center justify-center text-white hover:bg-white/5 transition-colors"
              title="Report Song"
            >
              <AlertTriangle size={18} />
            </button>
            <button
              onclick={shareSong}
              class="w-10 h-10 rounded-full border border-white/10 flex items-center justify-center text-white hover:bg-white/5 transition-colors"
              title="Share Song"
            >
              <Share2 size={18} />
            </button>
          </div>
        </div>
      </div>
    </aside>

    <!-- Right Section: Song List -->
    <section class="w-[40%] flex flex-col bg-surface-darker overflow-hidden">
      <div
        class="p-6 flex items-center justify-between border-b border-white/5"
      >
        <h2 class="text-xl font-bold flex items-center gap-3">
          <span class="w-1 h-6 bg-primary rounded-full"></span>
          {playlist.name}
        </h2>
        <div class="flex items-center gap-4">
          <div
            class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-surface-darker/50 border border-white/5"
          >
            <Music2 size={14} class="text-primary" />
            <span class="text-xs font-bold text-white/60"
              >{songs.length} Songs</span
            >
          </div>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto p-6 space-y-3 custom-scrollbar">
        {#each songs as song, i (song.id)}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            class="group flex items-center gap-4 p-3 rounded-md transition-all cursor-pointer justify-between {currentSong?.id ===
            song.id
              ? 'bg-primary/10 border border-primary/30'
              : 'hover:bg-white/5 border border-transparent hover:border-white/5'}"
            onclick={() => playSong(song)}
          >
            <div class="flex items-center gap-4 flex-1 min-w-0">
              {#if currentSong?.id === song.id}
                <div
                  class="absolute -left-1 w-1 h-8 bg-primary rounded-full"
                ></div>
                <div class="w-6 text-center text-primary font-black">
                  <AudioLines size={18} />
                </div>
              {:else}
                <div
                  class="w-6 text-center text-white/20 font-bold text-sm group-hover:text-primary"
                >
                  {i + 1}
                </div>
              {/if}
              <div
                class="w-20 h-14 rounded-lg overflow-hidden shrink-0 border border-white/5"
              >
                <OptimizedImage
                  src={song.anime?.cover_url}
                  sources={song.anime?.cover_sources}
                  alt={getSongName(song)}
                  class="w-full h-full object-cover {currentSong?.id === song.id
                    ? ''
                    : 'opacity-80 group-hover:opacity-100'}"
                  sizes="80px"
                />
              </div>

              <div class="flex flex-col min-w-0">
                <h4
                  class="font-bold truncate text-base {currentSong?.id ===
                  song.id
                    ? 'text-white neon-text'
                    : 'text-white/90'}"
                >
                  {getSongName(song)}
                </h4>
                <p class="text-[10px] text-white/40 line-clamp-1">
                  by {song.artists?.map((a) => a.name).join(", ") ||
                    "Unknown Artist"}
                </p>
                <span class="text-xs text-white/50 truncate">
                  {song.anime?.title || "Anime"}
                </span>
              </div>
            </div>

            <div class="flex items-center shrink-0">
              <div
                class="flex items-center gap-1 text-yellow-400 font-bold text-sm"
              >
                <Star size={14} class="fill-yellow-400" />
                {getFormattedScore(
                  song.average_rating,
                  authState.user?.score_format,
                )}
              </div>
            </div>
          </div>
        {/each}

        {#if songs.length === 0}
          <div
            class="py-20 flex flex-col items-center justify-center text-center opacity-40"
          >
            <Music2 size={64} class="font-thin" />
            <p class="mt-4 font-bold">This playlist is empty</p>
          </div>
        {/if}
      </div>

      <!-- Player Controls (Footer of Section) -->
      <div
        class="p-4 border-t border-white/5 flex items-center justify-between gap-4"
      >
        <!-- Left: Column (Hidden volume spacer) -->
        <div class="flex-1 min-w-0"></div>

        <!-- Center: Playback Controls -->
        <div class="flex items-center gap-4 shrink-0">
          <button
            onclick={toggleShuffle}
            class="w-12 h-12 rounded-sm bg-surface-dark border border-white/10 flex items-center justify-center hover:bg-white/5 transition-colors {isShuffle
              ? 'text-primary'
              : 'text-white/60 hover:text-white'}"
            title="Shuffle"
          >
            <Shuffle size={24} />
          </button>
          <button
            onclick={playPrev}
            class="w-12 h-12 rounded-sm bg-surface-dark border border-white/10 flex items-center justify-center hover:bg-white/5 text-white/60 hover:text-white transition-colors"
          >
            <SkipBack size={24} fill="currentColor" />
          </button>
          <button
            onclick={togglePlay}
            class="w-12 h-12 rounded-sm bg-primary flex items-center justify-center hover:bg-primary/80 shadow-lg shadow-primary/20 text-white transition-all transform hover:scale-105 active:scale-95"
          >
            {#if isPaused}
              <Play size={24} fill="currentColor" />
            {:else}
              <Pause size={24} fill="currentColor" />
            {/if}
          </button>
          <button
            onclick={playNext}
            class="w-12 h-12 rounded-sm bg-surface-dark border border-white/10 flex items-center justify-center hover:bg-white/5 text-white/60 hover:text-white transition-colors"
          >
            <SkipForward size={24} fill="currentColor" />
          </button>
        </div>

        <!-- Right: Desktop Spacer/Actions -->
        <div class="flex-1 min-w-0 flex justify-end">
          <!-- Puedes añadir botones extras aquí si es necesario -->
        </div>
      </div>
    </section>
  </div>
{:else}
  <!-- ==================== MOBILE LAYOUT (<lg) ==================== -->
  <div
    class="flex flex-col h-[calc(100vh-64px)] bg-background-dark antialiased overflow-hidden"
  >
    <!-- STICKY TOP: Video + Song Details -->
    <div
      class="shrink-0 sticky top-0 z-30 bg-background-dark border-b border-white/5"
    >
      <!-- Video -->
      <div class="relative aspect-video w-full overflow-hidden bg-black">
        {#if selectedVariant?.video}
          {#if selectedVariant.video.type === "embed"}
            <iframe
              src={getAutoplayUrl(selectedVariant.video.embed_url)}
              class="w-full h-full"
              frameborder="0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowfullscreen
              title="Song Video"
            ></iframe>
          {:else}
            <video
              bind:this={videoElement}
              bind:paused={isPaused}
              src={selectedVariant.video.local_url}
              class="w-full h-full"
              controls
              autoplay
              muted
              onplay={fadeInVolume}
              onended={playNext}
            >
              <track kind="captions" />
            </video>
          {/if}
        {:else if currentSong}
          <OptimizedImage
            src={currentSong.anime?.banner_url}
            sources={currentSong.anime?.banner_sources}
            alt={currentSong.title}
            class="w-full h-full object-cover"
            sizes="(max-width: 1024px) 100vw, 60vw"
          />
        {:else}
          <div
            class="w-full h-full flex items-center justify-center text-white/5 bg-surface-dark"
          >
            <Play size={48} fill="currentColor" />
          </div>
        {/if}
      </div>

      <!-- Song Details (compact for mobile) -->
      <div class="px-4 py-3 flex flex-col gap-3">
        <div class="flex flex-col min-w-0 flex-1">
          <h1 class="text-base font-black text-white truncate neon-text">
            {getSongName(currentSong)}
          </h1>
          <span class="text-sm font-bold text-primary truncate">
            {currentSong?.anime?.title || "No Anime info"}
          </span>
          {#if currentSong?.artists}
            <span class="text-xs text-white/60 truncate">
              {currentSong.artists.map((a: any) => a.name).join(", ")}
            </span>
          {/if}
        </div>
        <div class="flex items-center gap-3 shrink-0">
          {#if currentSong}
            <div
              class="flex items-center gap-1 text-yellow-400 font-black text-base"
            >
              <Star size={16} class="fill-yellow-400" />
              <span
                >{getFormattedScore(
                  currentSong.average_rating,
                  authState.user?.score_format,
                )}</span
              >
            </div>
          {/if}
          <div class="flex items-center gap-1">
            <button
              class="w-8 h-8 rounded-full flex items-center justify-center transition-colors {currentSong?.is_favorited
                ? 'text-primary'
                : 'text-white/40'}"
              onclick={toggleFavorite}
            >
              <Heart
                size={18}
                class={currentSong?.is_favorited ? "fill-primary text-primary" : ""}
              />
            </button>
            <button
              class="w-8 h-8 rounded-full flex items-center justify-center transition-colors {currentSong?.is_liked
                ? 'text-primary'
                : 'text-white/40'}"
              onclick={toggleLike}
            >
              <ThumbsUp
                size={18}
                class={currentSong?.is_liked ? "fill-primary text-primary" : ""}
              />
            </button>
            <button
              class="w-8 h-8 rounded-full flex items-center justify-center transition-colors {currentSong?.is_disliked
                ? 'text-primary'
                : 'text-white/40'}"
              onclick={toggleDislike}
            >
              <ThumbsDown
                size={18}
                class={currentSong?.is_disliked ? "fill-primary text-primary" : ""}
              />
            </button>

            <button
              onclick={shareSong}
              class="w-8 h-8 rounded-full flex items-center justify-center text-white/40 transition-colors"
            >
              <Share2 size={18} />
            </button>
            <button
              onclick={() => currentSong && removeFromPlaylist(currentSong.id)}
              class="w-8 h-8 rounded-full flex items-center justify-center text-red-400 transition-colors"
              title="Remove from Playlist"
            >
              <ListMinus size={18} />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- SCROLLABLE MIDDLE: Playlist -->
    <div class="flex-1 overflow-y-auto custom-scrollbar">
      <div
        class="px-4 py-3 flex items-center justify-between border-b border-white/5 bg-surface-darker/50"
      >
        <h2 class="text-sm font-bold flex items-center gap-2 text-white/80">
          <span class="w-1 h-5 bg-primary rounded-full"></span>
          {playlist.name}
        </h2>
        <span class="text-xs font-bold text-white/40">{songs.length} songs</span
        >
      </div>

      <div class="px-3 py-2 space-y-1">
        {#each songs as song, i (song.id)}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            class="group flex items-center gap-3 p-2.5 rounded-md transition-all cursor-pointer {currentSong?.id ===
            song.id
              ? 'bg-primary/10 border border-primary/30'
              : 'border border-transparent active:bg-white/5'}"
            onclick={() => playSong(song)}
          >
            <div class="flex items-center gap-3 flex-1 min-w-0">
              {#if currentSong?.id === song.id}
                <div class="w-5 text-center text-primary font-black shrink-0">
                  <AudioLines size={16} />
                </div>
              {:else}
                <div
                  class="w-5 text-center text-white/20 font-bold text-xs shrink-0"
                >
                  {i + 1}
                </div>
              {/if}
              <div
                class="w-12 h-12 rounded-lg overflow-hidden shrink-0 border border-white/5"
              >
                <OptimizedImage
                  src={song.anime?.cover_url}
                  sources={song.anime?.cover_sources}
                  alt={song.title}
                  class="w-full h-full object-cover"
                  sizes="48px"
                />
              </div>

              <div class="flex flex-col min-w-0">
                <h4
                  class="font-bold truncate text-sm {currentSong?.id === song.id
                    ? 'text-white neon-text'
                    : 'text-white/90'}"
                >
                  {getSongName(song)}
                </h4>
                <span class="text-[11px] text-white/50 truncate">
                  {song.artists?.map((a: any) => a.name).join(", ") ||
                    "Unknown"} ·
                  {song.anime?.title || "Anime"}
                </span>
              </div>
            </div>

            <div class="flex items-center shrink-0">
              <div
                class="flex items-center gap-0.5 text-yellow-400 font-bold text-xs"
              >
                <Star size={12} class="fill-yellow-400" />
                {getFormattedScore(
                  song.average_rating,
                  authState.user?.score_format,
                )}
              </div>
            </div>
          </div>
        {/each}

        {#if songs.length === 0}
          <div
            class="py-16 flex flex-col items-center justify-center text-center opacity-40"
          >
            <Music2 size={48} class="font-thin" />
            <p class="mt-3 font-bold text-sm">This playlist is empty</p>
          </div>
        {/if}
      </div>
    </div>

    <!-- STICKY BOTTOM: Player Controls -->
    <div
      class="shrink-0 sticky bottom-0 z-30 border-t border-white/5 px-4 py-3 safe-area-bottom"
    >
      <div class="flex items-center justify-between gap-2">
        <button
          onclick={toggleShuffle}
          class="w-10 h-10 rounded-full flex items-center justify-center transition-colors border border-white/5 {isShuffle
            ? 'text-primary'
            : 'text-white/40'}"
          title="Shuffle"
        >
          <Shuffle size={22} />
        </button>

        <button
          onclick={playPrev}
          class="w-10 h-10 rounded-full flex items-center justify-center text-white/60 transition-colors border border-white/5"
        >
          <SkipBack size={24} fill="currentColor" />
        </button>

        <button
          onclick={togglePlay}
          class="w-14 h-14 rounded-full bg-primary flex items-center justify-center shadow-lg shadow-primary/30 text-white transition-all active:scale-95"
        >
          {#if isPaused}
            <Play size={28} fill="currentColor" />
          {:else}
            <Pause size={28} fill="currentColor" />
          {/if}
        </button>

        <button
          onclick={playNext}
          class="w-10 h-10 rounded-full flex items-center justify-center text-white/60 transition-colors border border-white/5"
        >
          <SkipForward size={24} fill="currentColor" />
        </button>

        <button
          onclick={reportSong}
          class="w-10 h-10 rounded-full flex items-center justify-center text-white/40 border border-white/5 transition-colors"
          title="Report"
        >
          <AlertTriangle size={20} />
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showReportModal}
  <ReportModal
    show={showReportModal}
    song={currentSong}
    onClose={() => (showReportModal = false)}
    onSuccess={() =>
      toastState.addToast("Report submitted successfully!", "success")}
  />
{/if}

<style>
  .neon-border {
    box-shadow:
      0 0 15px rgba(127, 19, 236, 0.3),
      inset 0 0 5px rgba(127, 19, 236, 0.2);
  }
  .neon-text {
    text-shadow: 0 0 8px rgba(178, 77, 255, 0.6);
  }
  .custom-scrollbar::-webkit-scrollbar {
    width: 6px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: rgba(21, 13, 29, 0.5);
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: #4a3b59;
    border-radius: 10px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: #7f13ec;
  }
  .safe-area-bottom {
    padding-bottom: max(0.75rem, env(safe-area-inset-bottom));
  }
</style>
