<script lang="ts">
  import { page } from "$app/state";
  import { authState } from "$lib/state/auth.svelte";
  import {
    getSongArtistNames,
    getSongName,
    getFormattedScore,
  } from "$lib/song-utils";
  import { goto } from "$app/navigation";
  import { toastState } from "$lib/state/toast.svelte";
  import RatingModal from "$lib/components/RatingModal.svelte";
  import ReportModal from "$lib/components/ReportModal.svelte";
  import CommentReportModal from "$lib/components/CommentReportModal.svelte";
  import PlaylistModal from "$lib/components/PlaylistModal.svelte";
  import SEO from "$lib/components/SEO.svelte";
  import api from "$lib/api";
  import type { Song, Artist, SongVariant as Variant } from "$lib/types/song";
  interface User {
    badges: any;
    id: number;
    name: string;
    avatar_url: string;
    score_format: string;
    role: string;
  }

  interface Comment {
    id: number;
    content: string;
    user_id: number;
    created_at: string;
    user: User;
    replies?: Comment[];
    is_liked?: boolean;
    is_disliked?: boolean;
    likes_count?: number;
    dislikes_count?: number;
  }

  let { data } = $props<{
    data: { song: Song; comments: any[]; related: Song[] };
  }>();

  // svelte-ignore state_referenced_locally
  let currentSong: Song = $state(data.song);
  // svelte-ignore state_referenced_locally
  let relatedSongs: Song[] = $state(data.related);

  let selectedVariantIndex = $state(0);
  let selectedVariant = $derived(
    currentSong.song_variants?.[selectedVariantIndex],
  );

  // svelte-ignore state_referenced_locally
  let comments: Comment[] = $state(data.comments);

  $effect(() => {
    currentSong = data.song;
    relatedSongs = data.related;
    comments = data.comments;
    selectedVariantIndex = 0;
  });

  let newCommentText = $state("");
  let replyText = $state("");
  let replyingToId: number | null = $state(null);

  function changeVariant(index: number) {
    selectedVariantIndex = index;
  }

  let showRatingModal = $state(false);
  let showReportModal = $state(false);
  let showPlaylistModal = $state(false);
  let showCommentReportModal = $state(false);
  let reportingCommentId = $state<number | null>(null);

  function openCommentReportModal(id: number) {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    reportingCommentId = id;
    showCommentReportModal = true;
  }

  function handleRatingClick() {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    showRatingModal = true;
  }

  function handlePlaylistClick() {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    showPlaylistModal = true;
  }

  function getAutoplayUrl(url: string | undefined) {
    if (!url) return "";
    const separator = url.includes("?") ? "&" : "?";
    return `${url}${separator}autoplay=1&muted=1`;
  }
  let videoElement: HTMLVideoElement | undefined = $state();

  async function toggleLike() {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    try {
      const response = await api.post(`/interactions/reactions`, {
        song_id: currentSong.id,
        type: 1,
      });
      if (response.data.success || response.status === 200) {
        currentSong.is_liked = !currentSong.is_liked;
        if (currentSong.is_liked) {
          currentSong.is_disliked = false;
          toastState.addToast("Song liked!", "success");
        } else {
          toastState.addToast("Reaction removed", "info");
        }
        currentSong.likes_count = response.data.likesCount;
        currentSong.dislikes_count = response.data.dislikesCount;
      }
    } catch (error: any) {
      console.error("Error liking song:", error);
      toastState.addToast(
        error.response?.data?.message || "Failed to like song",
        "error",
      );
    }
  }

  async function toggleDislike() {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    try {
      const response = await api.post(`/interactions/reactions`, {
        song_id: currentSong.id,
        type: -1,
      });
      if (response.data.success || response.status === 200) {
        currentSong.is_disliked = !currentSong.is_disliked;
        if (currentSong.is_disliked) {
          currentSong.is_liked = false;
          toastState.addToast("Song disliked", "info");
        } else {
          toastState.addToast("Reaction removed", "info");
        }
        currentSong.likes_count = response.data.likesCount;
        currentSong.dislikes_count = response.data.dislikesCount;
      }
    } catch (error: any) {
      console.error("Error disliking song:", error);
      toastState.addToast(
        error.response?.data?.message || "Failed to dislike song",
        "error",
      );
    }
  }

  async function toggleFavorite() {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    try {
      const response = await api.post(`/interactions/favorites`, {
        entity_id: currentSong.id,
        entity_type: "song",
      });
      if (
        response.data.success ||
        response.status === 200 ||
        response.status === 201
      ) {
        currentSong.is_favorited =
          response.data.favorite || response.data.favorited;
        if (currentSong.is_favorited) {
          toastState.addToast("Added to favorites!", "success");
        } else {
          toastState.addToast("Removed from favorites", "info");
        }
      }
    } catch (error: any) {
      console.error("Error toggling favorite:", error);
      toastState.addToast(
        error.response?.data?.message || "Failed to update favorites",
        "error",
      );
    }
  }

  function reportSong() {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    if (currentSong.is_reported) return;
    showReportModal = true;
  }

  async function fetchComments(songId: number) {
    if (!songId) return;
    try {
      const resp = await api.get(`/songs/${songId}/comments`);
      comments = resp.data.data;
    } catch (e) {
      console.error(e);
    }
  }

  async function postComment() {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    if (!newCommentText.trim()) return;
    try {
      const resp = await api.post("/comments", {
        entity_id: currentSong.id,
        entity_type: "song",
        content: newCommentText,
      });
      const newComment = resp.data.data;
      if (!newComment.user && authState.user) {
        newComment.user = authState.user;
      }
      comments = comments ? [newComment, ...comments] : [newComment];
      newCommentText = "";
    } catch (e) {
      console.error(e);
    }
  }

  async function postReply(commentId: number) {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    if (!replyText.trim()) return;
    try {
      const resp = await api.post(`/comments`, {
        entity_id: currentSong.id,
        entity_type: "song",
        content: replyText,
        parent_id: commentId,
      });
      const newReply = resp.data.data;
      if (!newReply.user && authState.user) {
        newReply.user = authState.user;
      }

      const parentIndex = (comments || []).findIndex((c) => c.id === commentId);
      if (parentIndex !== -1) {
        const parent = comments[parentIndex];
        const updatedReplies = parent.replies
          ? [...parent.replies, newReply]
          : [newReply];

        // Use Svelte 5 reactive reassignment to trigger update
        comments[parentIndex] = { ...parent, replies: updatedReplies };
      }
      replyingToId = null;
      replyText = "";
    } catch (e) {
      console.error(e);
    }
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
    }, 100); // Sube el volumen cada 100ms
  }

  async function deleteComment(id: number, parentId: number | null = null) {
    if (!confirm("Are you sure you want to delete this comment?")) return;
    try {
      await api.delete(`/comments/${id}`);
      if (parentId) {
        const parent = comments.find((c) => c.id === parentId);
        if (parent && parent.replies) {
          parent.replies = parent.replies.filter((r) => r.id !== id);
        }
      } else {
        comments = comments.filter((c) => c.id !== id);
      }
    } catch (error) {
      console.error("Error deleting comment:", error);
    }
  }

  async function toggleCommentLike(
    commentId: number,
    parentId: number | null = null,
  ) {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    try {
      let targetComment;
      let parentIndex = -1;
      let replyIndex = -1;

      if (parentId) {
        const pIdx = comments.findIndex((c) => c.id === parentId);
        if (pIdx !== -1) {
          const parent = comments[pIdx];
          if (parent.replies) {
            const rIdx = parent.replies.findIndex(
              (r: any) => r.id === commentId,
            );
            if (rIdx !== -1) targetComment = parent.replies[rIdx];
            parentIndex = pIdx;
            replyIndex = rIdx;
          }
        }
      } else {
        parentIndex = comments.findIndex((c) => c.id === commentId);
        if (parentIndex !== -1) targetComment = comments[parentIndex];
      }

      if (!targetComment) return;

      const type = targetComment.is_liked ? 0 : 1;
      const res = await api.post("/interactions/reactions", {
        entity_id: targetComment.id,
        entity_type: "comment",
        type: type,
      });

      if (res.data.success || res.status === 200) {
        targetComment.likes_count = res.data.likesCount;
        targetComment.dislikes_count = res.data.dislikesCount;
        targetComment.is_liked = type === 1;
        targetComment.is_disliked = false;

        if (parentId && parentIndex !== -1 && comments[parentIndex]?.replies) {
          const p = comments[parentIndex];
          if (p.replies && replyIndex !== -1) {
            p.replies[replyIndex] = targetComment;
            comments[parentIndex] = { ...p };
          }
        } else if (parentIndex !== -1) {
          comments[parentIndex] = targetComment;
          comments = [...comments]; // Trigger reactivity
        }
      }
    } catch (e) {
      console.error(e);
      toastState.addToast("Failed to update reaction", "error");
    }
  }

  async function toggleCommentDislike(
    commentId: number,
    parentId: number | null = null,
  ) {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    try {
      let targetComment;
      let parentIndex = -1;
      let replyIndex = -1;

      if (parentId) {
        parentIndex = comments.findIndex((c) => c.id === parentId);
        const parent = comments[parentIndex];
        if (parentIndex !== -1 && parent?.replies) {
          replyIndex = parent.replies.findIndex((r: any) => r.id === commentId);
          if (replyIndex !== -1) targetComment = parent.replies[replyIndex];
        }
      } else {
        parentIndex = comments.findIndex((c) => c.id === commentId);
        if (parentIndex !== -1) targetComment = comments[parentIndex];
      }

      if (!targetComment) return;

      const type = targetComment.is_disliked ? 0 : -1;
      const res = await api.post("/interactions/reactions", {
        entity_id: targetComment.id,
        entity_type: "comment",
        type: type,
      });

      if (res.data.success || res.status === 200) {
        targetComment.likes_count = res.data.likesCount;
        targetComment.dislikes_count = res.data.dislikesCount;
        targetComment.is_liked = false;
        targetComment.is_disliked = type === -1;

        if (parentId && parentIndex !== -1) {
          const parent = comments[parentIndex];
          if (parent?.replies && replyIndex !== -1) {
            parent.replies[replyIndex] = targetComment;
            comments[parentIndex] = { ...parent };
          }
        } else if (parentIndex !== -1) {
          comments[parentIndex] = targetComment;
          comments = [...comments];
        }
      }
    } catch (e) {
      console.error(e);
      toastState.addToast("Failed to update reaction", "error");
    }
  }

  $effect(() => {
    currentSong = data.song;
    relatedSongs = data.related;
    selectedVariantIndex = 0;

    if (data.song?.id) {
      fetchComments(data.song.id);
    }
  });
</script>

<SEO
  title="{getSongName(currentSong)} - {currentSong.anime?.title} - AniRank"
  description="Listen to and rate '{getSongName(
    currentSong,
  )}' ({currentSong.type}{currentSong.theme_num || ''}) by {getSongArtistNames(
    currentSong.artists,
  )} from the anime {currentSong.anime?.title}."
  image={`${page.url.origin}/api/og/song/${currentSong.id}`}
  type="music.song"
/>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-8 space-y-8">
  <div class="grid grid-cols-1 lg:grid-cols-12 gap-8 pb-12">
    <!-- Main Content (Left) -->
    <div class="lg:col-span-8 space-y-2">
      <!-- Video Player -->
      <div
        class="relative w-full aspect-video rounded-3xl overflow-hidden bg-black video-shadow group border border-white/5"
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
              src={selectedVariant.video.local_url}
              class="w-full h-full"
              controls
              autoplay
              onplay={fadeInVolume}
            >
              <track kind="captions" />
            </video>
          {/if}
        {:else}
          <div
            class="absolute inset-0 bg-cover bg-center opacity-50"
            style="background-image: url('{currentSong.anime?.banner_url ||
              currentSong.anime?.cover_url ||
              'https://placehold.co/1280x720/2a2136/white?text=No+Art'}');"
          ></div>
          <div
            class="absolute inset-0 flex flex-col items-center justify-center p-6 text-center"
          >
            <span class="material-symbols-outlined text-white/20 text-6xl mb-4"
              >videocam_off</span
            >
            <span class="text-white/60 font-bold text-lg"
              >{(currentSong.song_variants?.length ?? 0) > 0
                ? "No video available for this variant"
                : "No video versions available for this theme song"}</span
            >
            {#if currentSong.song_variants?.length === 0}
              <p class="text-white/30 text-sm mt-2 max-w-md">
                We don't have a video file or embed for this theme yet. If you
                have it, you can contribute it on our community server!
              </p>
            {/if}
          </div>
        {/if}
      </div>

      <!-- Song Info Header -->
      <div class="space-y-2">
        <div class="flex items-center gap-3">
          <div class="flex bg-white/5 rounded-full p-1 border border-white/5">
            <span
              class="px-3 py-1.5 rounded-full text-[10px] font-black bg-primary text-white uppercase tracking-wider"
            >
              {currentSong.type}
              {currentSong.theme_num || ""}
            </span>
          </div>
          <!-- Version Selector -->
          {#if currentSong.song_variants && currentSong.song_variants.length > 1}
            <div class="flex bg-white/5 rounded-full p-1 border border-white/5">
              {#each currentSong.song_variants as variant, i}
                <button
                  class="px-3 py-1.5 rounded-full text-[10px] font-bold transition-all {selectedVariantIndex ===
                  i
                    ? 'bg-primary text-white'
                    : 'hover:bg-white/10 text-white/60'}"
                  onclick={() => changeVariant(i)}
                  title="Select version {variant.version_number}"
                  aria-label="Select version {variant.version_number}"
                >
                  V{variant.version_number}
                </button>
              {/each}
            </div>
          {/if}
        </div>
        <h1 class="text-2xl md:text-3xl font-black text-white tracking-tight">
          {getSongName(currentSong)}
        </h1>
      </div>
      <div class="space-x-2">
        {#if currentSong.artists?.length}
          {#each currentSong.artists as artist}
            {#if artist?.status === false}
              <span
                class="text-white/40 text-md font-bold uppercase tracking-wider"
                >N/A</span
              >
            {:else}
              <a
                href="/artists/{artist.slug}"
                class="text-white/40 text-md font-bold uppercase tracking-wider"
                title="View artist profile: {artist.name}">{artist.name}</a
              >
            {/if}
          {/each}
        {:else}
          <span class="text-white/40 text-md font-bold uppercase tracking-wider"
            >Without artists</span
          >
        {/if}
      </div>
      <div class="flex items-center gap-2">
        <a
          href="/animes/{currentSong.anime?.slug}"
          class="text-white/40 hover:text-primary text-md font-medium transition-colors"
          title="View anime: {currentSong.anime?.title}"
        >
          {currentSong.anime?.title}
        </a>
      </div>

      <!-- Meta Info Bar -->
      <div
        class="bg-surface-dark rounded-2xl p-5 border border-white/10 flex flex-wrap items-center justify-between gap-6"
      >
        <div class="flex items-center gap-8">
          <!-- Integrated Views & Interactions -->
          <div class="flex flex-col min-w-[120px] max-w-[160px]">
            <div class="text-sm text-white/90 tracking-wide mb-1">
              {currentSong.views.toLocaleString()} views
            </div>

            <div
              class="h-[6px] w-full bg-[#d0b3e6] rounded-full overflow-hidden flex mb-1.5"
            >
              <div
                class="h-full bg-[#874fb3] transition-all duration-500 ease-out"
                style="width: {(currentSong.likes_count || 0) +
                  (currentSong.dislikes_count || 0) >
                0
                  ? ((currentSong.likes_count || 0) /
                      ((currentSong.likes_count || 0) +
                        (currentSong.dislikes_count || 0))) *
                    100
                  : 50}%"
              ></div>
            </div>

            <div class="flex items-center justify-between px-0.5">
              <button
                class="flex items-center gap-1 transition-colors {currentSong.is_liked
                  ? 'text-primary'
                  : 'text-white/80 hover:text-white'}"
                onclick={toggleLike}
                title={currentSong.is_liked ? "Remove like" : "Like this theme"}
                aria-label={currentSong.is_liked
                  ? "Unlike theme"
                  : "Like theme"}
              >
                <span
                  class="material-symbols-outlined text-[14px] {currentSong.is_liked
                    ? 'filled'
                    : ''}">thumb_up</span
                >
                <span class="text-[10px] font-bold"
                  >{currentSong.likes_count || 0}</span
                >
              </button>

              <button
                class="flex items-center gap-1 transition-colors {currentSong.is_disliked
                  ? 'text-red-500'
                  : 'text-white/80 hover:text-white'}"
                onclick={toggleDislike}
                title={currentSong.is_disliked
                  ? "Remove dislike"
                  : "Dislike this theme"}
                aria-label={currentSong.is_disliked
                  ? "Undislike theme"
                  : "Dislike theme"}
              >
                <span
                  class="material-symbols-outlined text-[14px] {currentSong.is_disliked
                    ? 'filled'
                    : ''}">thumb_down</span
                >
                <span class="text-[10px] font-bold"
                  >{currentSong.dislikes_count || 0}</span
                >
              </button>
            </div>
          </div>

          <div class="h-10 w-px bg-white/10"></div>

          <button
            onclick={handleRatingClick}
            class="flex flex-col items-start hover:bg-white/5 px-3 py-1.5 rounded-xl transition-all group active:scale-95"
            title="Rate this theme"
            aria-label="Rate this theme"
          >
            <span
              class="text-white/40 text-[10px] font-bold uppercase tracking-wider group-hover:text-primary transition-colors"
              >Rating</span
            >
            <div class="flex items-center gap-2">
              <span class="text-yellow-400 font-bold text-lg"
                >{getFormattedScore(
                  currentSong.average_rating,
                  authState.user?.score_format,
                )}</span
              >
              <span
                class="material-symbols-outlined filled text-[16px] text-yellow-400 group-hover:rotate-12 transition-transform"
                >star</span
              >
            </div>
          </button>
        </div>

        <div class="flex items-center gap-3">
          <button
            class="w-10 h-10 flex items-center justify-center bg-white/5 hover:bg-white/10 border border-white/5 rounded-full transition-colors {currentSong.is_favorited
              ? 'text-pink-500'
              : 'text-white/60'}"
            onclick={toggleFavorite}
            title={currentSong.is_favorited
              ? "Remove from favorites"
              : "Add to favorites"}
            aria-label={currentSong.is_favorited
              ? "Remove from favorites"
              : "Add to favorites"}
          >
            <span
              class="material-symbols-outlined text-[20px] {currentSong.is_favorited
                ? 'filled'
                : ''}">favorite</span
            >
          </button>
          <button
            class="w-10 h-10 flex items-center justify-center bg-white/5 hover:bg-white/10 border border-white/5 rounded-full transition-colors text-white/60 hover:text-primary"
            onclick={handlePlaylistClick}
            title="Add to Playlist"
            aria-label="Add this theme to a playlist"
          >
            <span class="material-symbols-outlined text-[20px]"
              >playlist_add</span
            >
          </button>
          <button
            class="w-10 h-10 flex items-center justify-center bg-white/5 hover:bg-white/10 border border-white/5 rounded-full transition-colors {currentSong.is_reported
              ? 'text-red-500 opacity-50 cursor-not-allowed'
              : 'text-white/60 hover:text-red-400'}"
            onclick={reportSong}
            disabled={currentSong.is_reported}
            title={currentSong.is_reported ? "Already reported" : "Report Song"}
            aria-label={currentSong.is_reported
              ? "Already reported"
              : "Report this theme"}
          >
            <span
              class="material-symbols-outlined text-[20px] {currentSong.is_reported
                ? 'filled'
                : ''}">report</span
            >
          </button>
        </div>
      </div>
      <!-- Comments Section -->
      <div class="space-y-6 pt-4">
        <h2 class="text-2xl font-bold flex items-center gap-3">
          <span class="material-symbols-outlined text-primary">forum</span>
          Comments
        </h2>

        <!-- New Comment Input -->
        <div class="flex gap-4">
          <div
            class="w-10 h-10 rounded-full bg-white/10 overflow-hidden shrink-0"
          >
            {#if authState.isAuthenticated && authState.user}
              <img
                src={authState.user.avatar_url ||
                  "https://api.dicebear.com/7.x/notionists/svg?seed=" +
                    authState.user.name}
                alt="{authState.user.name}'s avatar"
                title="{authState.user.name}'s avatar"
                class="w-full h-full object-cover"
              />
            {:else}
              <span
                class="material-symbols-outlined text-white/40 w-full h-full flex items-center justify-center"
                >person</span
              >
            {/if}
          </div>
          <div class="flex-1 flex flex-col gap-2">
            <textarea
              bind:value={newCommentText}
              placeholder={authState.isAuthenticated
                ? "Add a comment..."
                : "Sign in to comment..."}
              class="w-full bg-white/5 border border-white/10 rounded-xl p-3 text-sm text-white focus:outline-none focus:border-primary transition-colors resize-none disabled:opacity-50"
              rows="2"
              disabled={!authState.isAuthenticated}
              aria-label="Add a new comment"
            ></textarea>
            <div class="flex justify-end">
              <button
                onclick={postComment}
                disabled={!newCommentText.trim() || !authState.isAuthenticated}
                class="bg-primary hover:bg-primary/80 text-white font-bold text-xs px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
                title="Post comment"
                aria-label="Post comment"
              >
                Comment
              </button>
            </div>
          </div>
        </div>

        <!-- Comments List -->
        <div class="space-y-4 pb-12">
          {#each comments as comment (comment.id)}
            <div class="flex gap-4">
              <div
                class="w-10 h-10 rounded-full bg-white/10 overflow-hidden shrink-0"
              >
                <img
                  src={comment.user?.avatar_url ||
                    "https://api.dicebear.com/7.x/notionists/svg?seed=" +
                      comment.user?.name}
                  alt={comment.user?.name}
                  title={comment.user?.name}
                  class="w-full h-full object-cover"
                />
              </div>
              <div class="flex-1 space-y-2">
                <div
                  class="bg-white/5 border border-white/10 rounded-2xl rounded-tl-sm p-4"
                >
                  <div class="flex justify-between items-start mb-1">
                    <div class="flex items-center gap-2">
                      <span class="font-bold text-sm text-white/90"
                        >{comment.user?.name || "Unknown User"}</span
                      >
                      {#if comment.user?.badges}
                        {#each comment.user.badges as badge}
                          <img
                            src={badge.image_url}
                            alt={badge.name}
                            class="w-4 h-4"
                            title={badge.name}
                          />
                        {/each}
                      {/if}
                      <span class="text-xs text-white/40"
                        >{comment.created_at
                          ? new Date(comment.created_at).toLocaleDateString()
                          : "Just now"}</span
                      >
                    </div>
                  </div>
                  <p class="text-sm text-white/80 whitespace-pre-wrap">
                    {comment.content}
                  </p>
                </div>
                <div class="flex justify-between text-md">
                  <div class="flex gap-2 items-center">
                    <button
                      onclick={() => toggleCommentLike(comment.id)}
                      class="flex items-center gap-1 transition-colors {comment.is_liked
                        ? 'text-primary'
                        : 'text-white/20 hover:text-white'}"
                      title={comment.is_liked
                        ? "Unlike comment"
                        : "Like comment"}
                      aria-label={comment.is_liked
                        ? "Unlike comment"
                        : "Like comment"}
                    >
                      <span
                        class="material-symbols-outlined {comment.is_liked
                          ? 'filled'
                          : ''}">thumb_up</span
                      >
                      {comment.likes_count || 0}
                    </button>
                    <button
                      onclick={() => toggleCommentDislike(comment.id)}
                      class="flex items-center gap-1 transition-colors {comment.is_disliked
                        ? 'text-red-500'
                        : 'text-white/20 hover:text-white'}"
                      title={comment.is_disliked
                        ? "Undislike comment"
                        : "Dislike comment"}
                      aria-label={comment.is_disliked
                        ? "Undislike comment"
                        : "Dislike comment"}
                    >
                      <span
                        class="material-symbols-outlined {comment.is_disliked
                          ? 'filled'
                          : ''}">thumb_down</span
                      >
                      {comment.dislikes_count || 0}
                    </button>
                    <button
                      onclick={() => {
                        replyingToId =
                          replyingToId === comment.id ? null : comment.id;
                        replyText = "";
                      }}
                      class="text-white/40 hover:text-white tracking-wider transition-colors"
                      title="Reply to comment"
                      aria-label="Reply to comment"
                      >Reply
                      <span class="material-symbols-outlined">reply</span>
                    </button>
                  </div>

                  <div class="flex gap-2">
                    <button
                      onclick={() => openCommentReportModal(comment.id)}
                      class="shrink-0 p-1 hover:bg-white/10 text-white/40 hover:text-white transition-colors"
                      title="Report Comment"
                      aria-label="Report comment"
                    >
                      <span class="material-symbols-outlined text-[16px]"
                        >flag</span
                      >
                    </button>

                    {#if authState.user && (authState.user.id === comment.user_id || authState.isAdmin)}
                      <button
                        onclick={() => deleteComment(comment.id)}
                        class="text-white/20 hover:text-red-400 transition-colors"
                        title="Delete Comment"
                        aria-label="Delete comment"
                      >
                        <span class="material-symbols-outlined text-[16px]"
                          >delete</span
                        >
                      </button>
                    {/if}
                  </div>
                </div>

                <!-- Inline Reply Input -->
                {#if replyingToId === comment.id}
                  <div class="flex gap-3 mt-3">
                    <div class="flex-1 flex gap-2 items-start">
                      <textarea
                        bind:value={replyText}
                        placeholder="Write a reply..."
                        class="w-full bg-white/5 border border-white/10 rounded-xl p-2 text-sm text-white focus:outline-none focus:border-primary transition-colors resize-none"
                        rows="1"
                        aria-label="Write a reply"
                      ></textarea>
                      <button
                        onclick={() => postReply(comment.id)}
                        disabled={!replyText.trim()}
                        class="bg-white/10 hover:bg-primary text-white font-bold text-xs px-3 py-2 rounded-lg transition-colors disabled:opacity-50 shrink-0"
                        title="Send reply"
                        aria-label="Send reply"
                      >
                        Send
                      </button>
                    </div>
                  </div>
                {/if}

                <!-- Replies List -->
                {#if comment.replies && comment.replies.length > 0}
                  <div class="space-y-3 mt-3 border-l-2 border-white/5 pl-4">
                    {#each comment.replies as reply (reply.id)}
                      <div class="flex gap-3">
                        <div
                          class="w-8 h-8 rounded-full bg-white/10 overflow-hidden shrink-0"
                        >
                          <img
                            src={reply.user?.avatar_url ||
                              "https://api.dicebear.com/7.x/notionists/svg?seed=" +
                                reply.user?.name}
                            alt={reply.user?.name}
                            title={reply.user?.name}
                            class="w-full h-full object-cover"
                          />
                        </div>
                        <div class="flex-1 space-y-1">
                          <div
                            class="bg-white/5 border border-white/10 rounded-2xl rounded-tl-sm p-3"
                          >
                            <div class="flex justify-between items-start mb-1">
                              <div class="flex items-center gap-2">
                                <span class="font-bold text-xs text-white/90"
                                  >{reply.user?.name || "Unknown User"}</span
                                >
                                {#if reply.user?.badges}
                                  <div class="flex gap-1 ml-1">
                                    {#each reply.user.badges as badge}
                                      <img
                                        src={badge.image_url}
                                        alt={badge.name}
                                        class="w-3.5 h-3.5"
                                        title={badge.name}
                                      />
                                    {/each}
                                  </div>
                                {/if}
                                <span class="text-[10px] text-white/40"
                                  >{reply.created_at
                                    ? new Date(
                                        reply.created_at,
                                      ).toLocaleDateString()
                                    : "Just now"}</span
                                >
                              </div>
                              <div class="flex items-center gap-2">
                                {#if authState.user && (authState.user.id === reply.user_id || authState.isAdmin)}
                                  <button
                                    onclick={() =>
                                      deleteComment(reply.id, comment.id)}
                                    class="text-white/20 hover:text-red-400 transition-colors"
                                    title="Delete Reply"
                                    aria-label="Delete reply"
                                  >
                                    <span
                                      class="material-symbols-outlined text-[14px]"
                                      >delete</span
                                    >
                                  </button>
                                {/if}
                                <button
                                  onclick={() =>
                                    openCommentReportModal(reply.id)}
                                  class="text-white/20 hover:text-primary transition-colors"
                                  title="Report Reply"
                                  aria-label="Report reply"
                                >
                                  <span
                                    class="material-symbols-outlined text-[14px]"
                                    >flag</span
                                  >
                                </button>
                              </div>
                            </div>
                            <p
                              class="text-[13px] text-white/80 whitespace-pre-wrap"
                            >
                              {reply.content}
                            </p>
                            <div class="flex gap-4 items-center mt-2">
                              <button
                                onclick={() =>
                                  toggleCommentLike(reply.id, comment.id)}
                                class="flex items-center gap-1 transition-colors {reply.is_liked
                                  ? 'text-primary'
                                  : 'text-white/20 hover:text-white'}"
                                title={reply.is_liked
                                  ? "Unlike reply"
                                  : "Like reply"}
                                aria-label={reply.is_liked
                                  ? "Unlike reply"
                                  : "Like reply"}
                              >
                                <span
                                  class="material-symbols-outlined text-[14px] {reply.is_liked
                                    ? 'filled'
                                    : ''}">thumb_up</span
                                >
                                <span class="text-xs"
                                  >{reply.likes_count || 0}</span
                                >
                              </button>
                              <button
                                onclick={() =>
                                  toggleCommentDislike(reply.id, comment.id)}
                                class="flex items-center gap-1 transition-colors {reply.is_disliked
                                  ? 'text-red-500'
                                  : 'text-white/20 hover:text-white'}"
                              >
                                <span
                                  class="material-symbols-outlined text-[14px] {reply.is_disliked
                                    ? 'filled'
                                    : ''}">thumb_down</span
                                >
                                <span class="text-xs"
                                  >{reply.dislikes_count || 0}</span
                                >
                              </button>
                            </div>
                          </div>
                        </div>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            </div>
          {:else}
            <div class="text-center py-8">
              <span
                class="material-symbols-outlined text-[48px] text-white/10 mb-2"
                >speaker_notes</span
              >
              <p class="text-white/40 font-bold text-sm">No comments yet</p>
              <p class="text-white/20 text-xs mt-1">
                Be the first to share your thoughts on this song!
              </p>
            </div>
          {/each}
        </div>
      </div>
    </div>

    <!-- Sidebar (Right - YouTube Style) -->
    <div class="lg:col-span-4 space-y-4">
      <div>
        <h2
          class="text-lg font-bold flex items-center gap-2 mb-4 text-white/90"
        >
          <span class="material-symbols-outlined text-primary text-[20px]"
            >playlist_play</span
          >
          More from this series
        </h2>
      </div>
      <div class="space-y-3">
        {#each relatedSongs as related}
          <a
            href="/songs/{currentSong.anime?.slug}/{related.slug}"
            class="flex gap-3 group bg-white/0 hover:bg-white/5 p-2 rounded-xl transition-all"
            title="View theme: {getSongName(related)}"
          >
            <div
              class="w-32 aspect-video rounded-lg overflow-hidden shrink-0 border border-white/5"
            >
              <img
                src={currentSong.anime?.cover_url ||
                  "https://placehold.co/400x225/2a2136/white?text=No+Art"}
                alt={getSongName(related)}
                title={getSongName(related)}
                class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500"
              />
            </div>
            <div class="flex-1 min-w-0 flex flex-col justify-center">
              <div class="flex items-center gap-1.5 mb-1">
                <span
                  class="bg-primary/20 text-primary text-[8px] font-black px-1.5 py-0.5 rounded uppercase"
                  >{related.type}</span
                >
                <span
                  class="text-[10px] text-yellow-400 font-bold flex items-center gap-0.5"
                >
                  <span class="material-symbols-outlined filled text-[10px]"
                    >star</span
                  >
                  {getFormattedScore(
                    related.average_rating,
                    authState.user?.score_format,
                  )}
                </span>
              </div>
              <h4
                class="text-sm font-bold text-white group-hover:text-primary transition-colors line-clamp-1"
              >
                {getSongName(related)}
              </h4>
              <p class="text-[10px] text-white/40 line-clamp-1">
                by {getSongArtistNames(related.artists)}
              </p>
            </div>
          </a>
        {:else}
          <div
            class="text-center py-8 text-white/20 text-xs italic border border-dashed border-white/5 rounded-xl"
          >
            No other themes found for this series.
          </div>
        {/each}
      </div>
    </div>
  </div>
</main>

<RatingModal
  show={showRatingModal}
  song={currentSong}
  onClose={() => (showRatingModal = false)}
  onRated={(newData) => {
    if (newData.average) {
      currentSong.average_rating = newData.average;

      const format = authState.user?.score_format || "POINT_10_DECIMAL";
      let multiplier = 10;
      if (format === "POINT_5") multiplier = 20;
      if (format === "POINT_100") multiplier = 1;

      currentSong.average_rating = newData.average;
    }
    if (newData.rating !== undefined) {
      currentSong.user_rating = newData.rating;
    }
  }}
/>

<ReportModal
  show={showReportModal}
  song={currentSong}
  onClose={() => (showReportModal = false)}
  onSuccess={() => {
    currentSong.is_reported = true;
  }}
/>

{#if showCommentReportModal && reportingCommentId}
  <CommentReportModal
    show={showCommentReportModal}
    commentId={reportingCommentId}
    onClose={() => {
      showCommentReportModal = false;
      reportingCommentId = null;
    }}
  />
{/if}

<PlaylistModal
  show={showPlaylistModal}
  song={currentSong}
  onClose={() => (showPlaylistModal = false)}
/>

<style>
  .video-shadow {
    box-shadow: 0 0 100px -20px rgba(127, 19, 236, 0.3);
  }
  .material-symbols-outlined.filled {
    font-variation-settings:
      "FILL" 1,
      "wght" 400,
      "GRAD" 0,
      "opsz" 24;
  }
</style>
