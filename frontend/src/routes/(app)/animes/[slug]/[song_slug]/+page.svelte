<script lang="ts">
  import { page } from "$app/state";
  import { authState } from "$lib/state/auth.svelte";
  import { PUBLIC_API_URL } from "$lib/api";
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
  import UserReportModal from "$lib/components/UserReportModal.svelte";
  import SEO from "$lib/components/SEO.svelte";
  import ThumbsUp from "lucide-svelte/icons/thumbs-up";
  import ThumbsDown from "lucide-svelte/icons/thumbs-down";
  import Star from "lucide-svelte/icons/star";
  import Heart from "lucide-svelte/icons/heart";
  import ListPlus from "lucide-svelte/icons/list-plus";
  import AlertTriangle from "lucide-svelte/icons/alert-triangle";
  import MessageSquare from "lucide-svelte/icons/message-square";
  import VideoOff from "lucide-svelte/icons/video-off";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import MoreVertical from "lucide-svelte/icons/more-vertical";
  import Edit2 from "lucide-svelte/icons/edit-2";
  import CornerDownRight from "lucide-svelte/icons/corner-down-right";
  import MessageCircle from "lucide-svelte/icons/message-circle";
  import UserIcon from "lucide-svelte/icons/user";
  import Flag from "lucide-svelte/icons/flag";
  import Library from "lucide-svelte/icons/library";
  import Sparkles from "lucide-svelte/icons/sparkles";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";
  import type { ImageSource } from "$lib/types/media";
  import api from "$lib/api";
  import type { Song, Artist, SongVariant as Variant } from "$lib/types/song";
  interface User {
    badges?: any;
    uuid: string;
    name: string;
    avatar_url?: string;
    avatar_sources?: ImageSource[];
    role?: string;
    truth_score?: number;
    is_shadowbanned?: boolean;
  }

  interface Comment {
    uuid: string;
    content: string;
    user_uuid: string;
    created_at: string;
    user: User;
    replies?: Comment[];
    is_liked?: boolean;
    is_disliked?: boolean;
    likes_count?: number;
    dislikes_count?: number;
    is_shadowbanned?: boolean;
  }

  function mapComment(c: any): Comment {
    const mappedUser = c.user ? { ...c.user, uuid: c.user.id } : undefined;
    return {
      ...c,
      uuid: c.id,
      user: mappedUser,
      user_uuid: mappedUser?.uuid || c.user_uuid,
      replies: c.replies ? c.replies.map(mapComment) : [],
    };
  }

  let { data } = $props<{
    data: { song: Song; comments: any[]; related: Song[] };
  }>();

  // svelte-ignore state_referenced_locally
  let currentSong: Song = $state(data.song);
  // svelte-ignore state_referenced_locally
  let relatedSongs: Song[] = $state(data.related);

  let selectedVariantIndex = $state(0);
  let selectedVariant = $derived(currentSong.variants?.[selectedVariantIndex]);
  let selectedVideoIndex = $state(0);
  let selectedVideo = $derived(
    selectedVariant?.videos?.[selectedVideoIndex] || {
      video_url: selectedVariant?.video_url,
      local_url: selectedVariant?.local_url,
      video_src: selectedVariant?.video_src,
      is_nc: selectedVariant?.is_nc,
      is_bd: selectedVariant?.is_bd,
      resolution: selectedVariant?.resolution,
      is_uncensored: selectedVariant?.is_uncensored,
      is_subbed: selectedVariant?.is_subbed,
      is_lyrics: selectedVariant?.is_lyrics,
      source: selectedVariant?.source,
      overlap: selectedVariant?.overlap,
    },
  );
  let videoError = $state(false);

  // svelte-ignore state_referenced_locally
  let comments: Comment[] = $state(data.comments?.map(mapComment) || []);

  $effect(() => {
    currentSong = data.song;
    relatedSongs = data.related;
    comments = data.comments?.map(mapComment) || [];
    selectedVariantIndex = 0;
    selectedVideoIndex = 0;
    videoError = false;
  });

  let activeElement = $state<HTMLElement | null>(null);

  function bindActive(node: HTMLElement, isSelected: boolean) {
    if (isSelected) {
      activeElement = node;
    }
    return {
      update(newIsSelected: boolean) {
        if (newIsSelected) {
          activeElement = node;
        } else if (activeElement === node) {
          activeElement = null;
        }
      },
      destroy() {
        if (activeElement === node) {
          activeElement = null;
        }
      },
    };
  }

  $effect(() => {
    if (activeElement) {
      activeElement.scrollIntoView({
        behavior: "smooth",
        block: "nearest",
      });
    }
  });

  let newCommentText = $state("");
  let replyText = $state("");
  let replyingToUuid: string | null = $state(null);

  function changeVariant(index: number) {
    selectedVariantIndex = index;
    selectedVideoIndex = 0;
    videoError = false;
  }

  let showRatingModal = $state(false);
  let showReportModal = $state(false);
  let showPlaylistModal = $state(false);
  let showCommentReportModal = $state(false);
  let showUserReportModal = $state(false);
  let reportingCommentUuid = $state<string | null>(null);
  let reportingUser = $state<any>(null);

  let editingCommentUuid = $state<string | null>(null);
  let editText = $state("");
  let openDropdownUuid = $state<string | null>(null);

  function openCommentReportModal(uuid: string) {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    reportingCommentUuid = uuid;
    showCommentReportModal = true;
    openDropdownUuid = null;
  }

  function openUserReportModal(user: any) {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    reportingUser = user;
    showUserReportModal = true;
    openDropdownUuid = null;
  }

  function startEditing(comment: Comment) {
    editingCommentUuid = comment.uuid;
    editText = comment.content;
    openDropdownUuid = null;
  }

  function cancelEditing() {
    editingCommentUuid = null;
    editText = "";
  }

  async function saveEdit() {
    if (!editingCommentUuid || !editText.trim()) return;
    try {
      await api.put(`/comments/${editingCommentUuid}`, {
        content: editText,
      });

      // Update local state
      const updateLocal = (list: Comment[] | undefined) => {
        if (!list) return false;
        for (let i = 0; i < list.length; i++) {
          if (list[i].uuid === editingCommentUuid) {
            list[i].content = editText;
            return true;
          }
          if (updateLocal(list[i].replies)) return true;
        }
        return false;
      };

      updateLocal(comments);
      comments = [...comments]; // Trigger reactivity
      editingCommentUuid = null;
      editText = "";
      toastState.addToast("Comment updated", "success");
    } catch (error) {
      console.error("Error updating comment:", error);
      toastState.addToast("Failed to update comment", "error");
    }
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
        currentSong.likes_count = response.data.data.likesCount;
        currentSong.dislikes_count = response.data.data.dislikesCount;
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
        currentSong.likes_count = response.data.data.likesCount;
        currentSong.dislikes_count = response.data.data.dislikesCount;
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

    // Optimistic Update
    const previousState = currentSong.is_favorited;
    currentSong.is_favorited = !previousState;

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
        // Sync with server response
        currentSong.is_favorited =
          response.data.data?.favorited ?? response.data.data?.favorite;

        if (currentSong.is_favorited) {
          toastState.addToast("Added to favorites!", "success");
        } else {
          toastState.addToast("Removed from favorites", "info");
        }
      } else {
        // Rollback if server doesn't report success
        currentSong.is_favorited = previousState;
      }
    } catch (error: any) {
      // Rollback on error
      currentSong.is_favorited = previousState;
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

  async function fetchComments(songId: string | number) {
    if (!songId) return;
    try {
      const resp = await api.get(`/songs/${songId}/comments`);
      comments = resp.data.data?.map(mapComment) || [];

      // Auto-scroll to comment if hash exists
      const hash = window.location.hash;
      if (hash && hash.startsWith("#comment-")) {
        const commentId = hash.slice(1);
        setTimeout(() => {
          const el = document.getElementById(commentId);
          if (el) {
            el.scrollIntoView({ behavior: "smooth", block: "center" });
            el.classList.add(
              "ring-2",
              "ring-primary",
              "rounded-md",
              "bg-primary/5",
            );
            setTimeout(() => {
              el.classList.remove("ring-2", "ring-primary", "bg-primary/5");
            }, 3000);
          }
        }, 500);
      }
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
      const newComment = mapComment(resp.data.data);
      if (!newComment.user && authState.user) {
        newComment.user = authState.user;
      }
      comments = comments ? [newComment, ...comments] : [newComment];
      newCommentText = "";
    } catch (e: any) {
      console.error(e);
      toastState.addToast(
        e.response?.data?.message || "Failed to post comment",
        "error",
      );
    }
  }

  async function postReply(commentUuid: string) {
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
        parent_id: commentUuid,
      });
      const newReply = mapComment(resp.data.data);
      if (!newReply.user && authState.user) {
        newReply.user = authState.user;
      }

      const parentIndex = (comments || []).findIndex(
        (c) => c.uuid === commentUuid,
      );
      if (parentIndex !== -1) {
        const parent = comments[parentIndex];
        const updatedReplies = parent.replies
          ? [...parent.replies, newReply]
          : [newReply];

        // Use Svelte 5 reactive reassignment to trigger update
        comments[parentIndex] = { ...parent, replies: updatedReplies };
      }
      replyingToUuid = null;
      replyText = "";
    } catch (e: any) {
      console.error(e);
      toastState.addToast(
        e.response?.data?.message || "Failed to post reply",
        "error",
      );
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

  async function deleteComment(uuid: string, parentUuid: string | null = null) {
    if (!confirm("Are you sure you want to delete this comment?")) return;
    try {
      await api.delete(`/comments/${uuid}`);
      if (parentUuid) {
        const parent = comments.find((c) => c.uuid === parentUuid);
        if (parent && parent.replies) {
          parent.replies = parent.replies.filter((r) => r.uuid !== uuid);
        }
      } else {
        comments = comments.filter((c) => c.uuid !== uuid);
      }
    } catch (error) {
      console.error("Error deleting comment:", error);
    }
  }

  async function toggleCommentLike(
    commentUuid: string,
    parentUuid: string | null = null,
  ) {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    try {
      let targetComment;
      let parentIndex = -1;
      let replyIndex = -1;

      if (parentUuid) {
        const pIdx = comments.findIndex((c) => c.uuid === parentUuid);
        if (pIdx !== -1) {
          const parent = comments[pIdx];
          if (parent.replies) {
            const rIdx = parent.replies.findIndex(
              (r: any) => r.uuid === commentUuid,
            );
            if (rIdx !== -1) targetComment = parent.replies[rIdx];
            parentIndex = pIdx;
            replyIndex = rIdx;
          }
        }
      } else {
        parentIndex = comments.findIndex((c) => c.uuid === commentUuid);
        if (parentIndex !== -1) targetComment = comments[parentIndex];
      }

      if (!targetComment) return;

      const type = targetComment.is_liked ? 0 : 1;
      const res = await api.post("/interactions/reactions", {
        entity_id: targetComment.uuid,
        entity_type: "comment",
        type: type,
      });

      if (res.data.success || res.status === 200) {
        targetComment.likes_count = res.data.data.likesCount;
        targetComment.dislikes_count = res.data.data.dislikesCount;
        targetComment.is_liked = type === 1;
        targetComment.is_disliked = false;

        if (
          parentUuid &&
          parentIndex !== -1 &&
          comments[parentIndex]?.replies
        ) {
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
    commentUuid: string,
    parentUuid: string | null = null,
  ) {
    if (!authState.isAuthenticated) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }
    try {
      let targetComment;
      let parentIndex = -1;
      let replyIndex = -1;

      if (parentUuid) {
        parentIndex = comments.findIndex((c) => c.uuid === parentUuid);
        const parent = comments[parentIndex];
        if (parentIndex !== -1 && parent?.replies) {
          replyIndex = parent.replies.findIndex(
            (r: any) => r.uuid === commentUuid,
          );
          if (replyIndex !== -1) targetComment = parent.replies[replyIndex];
        }
      } else {
        parentIndex = comments.findIndex((c) => c.uuid === commentUuid);
        if (parentIndex !== -1) targetComment = comments[parentIndex];
      }

      if (!targetComment) return;

      const type = targetComment.is_disliked ? 0 : -1;
      const res = await api.post("/interactions/reactions", {
        entity_id: targetComment.uuid,
        entity_type: "comment",
        type: type,
      });

      if (res.data.success || res.status === 200) {
        targetComment.likes_count = res.data.data.likesCount;
        targetComment.dislikes_count = res.data.data.dislikesCount;
        targetComment.is_liked = false;
        targetComment.is_disliked = type === -1;

        if (parentUuid && parentIndex !== -1) {
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
    selectedVideoIndex = 0;

    if (data.song?.id) {
      fetchComments(data.song.id);
    }
  });
</script>

<SEO
  title="{getSongName(currentSong)} - {currentSong.anime?.title} - AniRank"
  description="Listen to and rate '{getSongName(currentSong)}' ({currentSong
    .song_type?.name || currentSong.type}{currentSong.theme_num ||
    ''}) by {getSongArtistNames(
    currentSong.artists,
  )} from the anime {currentSong.anime?.title}."
  image={`${PUBLIC_API_URL}/og/song/${currentSong.anime?.slug}/${currentSong.slug}`}
  type="music.song"
/>

<main
  class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-8 space-y-8 animate-in fade-in duration-700"
>
  <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
    <!-- Main Content (Left) -->
    <div class="lg:col-span-8 space-y-4">
      <!-- Video Player -->
      <div
        class="relative w-full aspect-video rounded-md overflow-hidden bg-black group border border-outline-variant/10 shadow-2xl"
      >
        {#if selectedVideo?.video_url && !videoError}
          {#if selectedVideo.video_url.includes("youtube") || selectedVideo.video_url.includes("youtu.be")}
            <iframe
              src={getAutoplayUrl(selectedVideo.video_url)}
              class="w-full h-full"
              frameborder="0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowfullscreen
              title="Song Video"
            ></iframe>
          {:else}
            <video
              bind:this={videoElement}
              src={selectedVideo.video_url}
              class="w-full h-full"
              controls
              autoplay
              onplay={fadeInVolume}
              onerror={() => (videoError = true)}
            >
              <track kind="captions" />
            </video>
          {/if}
        {:else if selectedVariant?.embed_url}
          <iframe
            src={getAutoplayUrl(selectedVariant.embed_url)}
            class="w-full h-full"
            frameborder="0"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowfullscreen
            title="Song Video Fallback"
          ></iframe>
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
            <VideoOff size={64} class="text-on-surface-variant/20 mb-4" />
            <span class="text-on-surface-variant/60 font-bold text-lg"
              >{(currentSong.variants?.length ?? 0) > 0
                ? videoError
                  ? "Video file unreachable or not found"
                  : "No video available for this variant"
                : "No video versions available for this theme song"}</span
            >
            {#if !currentSong.variants || currentSong.variants.length === 0}
              <p class="text-on-surface-variant/30 text-sm mt-2 max-w-md">
                We don't have a video file or embed for this theme yet. If you
                have it, you can contribute it on our community server!
              </p>
            {/if}
          </div>
        {/if}
      </div>

      <!-- Details -->
      <div class="space-y-4">
        <div class="space-y-2">
          <div class="flex items-center gap-3">
            <div
              class="flex bg-surface-highest rounded-full p-1 border border-on-surface-variant/10 shadow-sm"
            >
              <span
                class="px-3 py-1.5 rounded-full text-[10px] font-black bg-primary text-white uppercase tracking-widest"
                title={currentSong.song_type?.description}
              >
                {currentSong.song_type?.name || currentSong.type}
                {currentSong.theme_num || ""}
              </span>
            </div>
            <!-- Version Selector -->
            {#if currentSong.variants && currentSong.variants.length > 1}
              <div
                class="flex bg-surface-highest rounded-full p-1 border border-on-surface-variant/10 shadow-sm"
              >
                {#each currentSong.variants as variant, i}
                  <button
                    class="px-3 py-1.5 rounded-full text-[10px] font-bold transition-all {selectedVariantIndex ===
                    i
                      ? 'bg-primary text-white shadow-lg shadow-primary/20'
                      : 'hover:bg-on-surface-variant/10 text-on-surface-variant/60'}"
                    onclick={() => changeVariant(i)}
                    title="Select version {variant.version_number}"
                    aria-label="Select version {variant.version_number}"
                  >
                    V{variant.version_number}
                  </button>
                {/each}
              </div>
            {/if}

            <!-- Video Quality/Tags Selector -->
            {#if selectedVariant?.videos && selectedVariant.videos.length > 1}
              <div
                class="flex bg-surface-highest rounded-sm p-1 border border-on-surface-variant/10 shadow-sm gap-1"
              >
                {#each selectedVariant.videos as video, i}
                  {@const tagsList = []}
                  {#if video.resolution}
                    {tagsList.push(`${video.resolution}p`)}
                  {/if}
                  {#if video.is_nc}
                    {tagsList.push("NC")}
                  {/if}
                  {#if video.is_bd}
                    {tagsList.push("BD")}
                  {/if}
                  {#if video.is_uncensored}
                    {tagsList.push("UNCEN")}
                  {/if}
                  {#if video.is_subbed}
                    {tagsList.push("SUB")}
                  {/if}
                  {#if video.is_lyrics}
                    {tagsList.push("LYRICS")}
                  {/if}
                  {@const tagText =
                    tagsList.length > 0 ? tagsList.join(" ") : `Video ${i + 1}`}
                  <button
                    class="px-3 py-1.5 rounded-sm text-[10px] font-bold transition-all {selectedVideoIndex ===
                    i
                      ? 'bg-primary text-white shadow-lg shadow-primary/20'
                      : 'hover:bg-on-surface-variant/10 text-on-surface-variant/60'}"
                    onclick={() => {
                      selectedVideoIndex = i;
                      videoError = false;
                    }}
                    title="Select video {tagText}"
                    aria-label="Select video {tagText}"
                  >
                    {tagText}
                  </button>
                {/each}
              </div>
            {/if}
          </div>
          <h1
            class="text-2xl md:text-4xl font-black text-on-surface tracking-tight"
          >
            {getSongName(currentSong)}
          </h1>
          <!-- Artists -->
          <div class="flex flex-wrap items-center gap-y-1">
            {#if currentSong.artists?.length}
              {#each currentSong.artists as artist, i}
                {#if i > 0}
                  <span class="text-on-surface-variant/40 text-sm font-bold mr-2">,</span>
                {/if}
                {#if artist?.status === false}
                  <span
                    class="text-on-surface-variant/40 text-sm font-bold uppercase tracking-widest"
                    >N/A</span
                  > 
                {:else}
                  <a
                    href="/artists/{artist.slug}"
                    class="text-on-surface-variant/60 text-sm font-bold uppercase tracking-widest hover:text-primary transition-colors"
                    title="View artist profile: {artist.name}">{artist.name}</a
                  >
                {/if}
              {/each}
            {:else}
              <span
                class="text-on-surface-variant/40 text-sm font-bold uppercase tracking-widest"
                >Without artists</span
              >
            {/if}
          </div>
          <div class="flex items-center gap-2">
            <a
              href="/animes/{currentSong.anime?.slug}"
              class="text-on-surface-variant/60 hover:text-primary text-sm font-black uppercase tracking-widest italic transition-colors"
              title="View anime: {currentSong.anime?.title}"
            >
              {currentSong.anime?.title}
            </a>
            {#if (selectedVariant?.season && selectedVariant?.year) || (currentSong.season && currentSong.year)}
              {@const displaySeason =
                selectedVariant?.season || currentSong.season}
              {@const displayYear = selectedVariant?.year || currentSong.year}
              <span
                class="px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-widest bg-surface-highest border border-on-surface-variant/10 text-on-surface-variant/60"
              >
                {displayYear?.name}
                {displaySeason?.name}
              </span>
            {/if}
          </div>
        </div>
        <!-- Meta Info Bar -->
        <div
          class="bg-surface-container rounded-md p-4 border border-outline-variant/10 flex flex-wrap items-center justify-between gap-6"
        >
          <div class="flex items-center gap-8">
            <!-- Integrated Views & Interactions -->
            <div class="flex flex-col min-w-[120px] max-w-[160px]">
              <div
                class="text-sm text-on-surface font-medium tracking-wide mb-1"
              >
                {currentSong.views.toLocaleString()} views
              </div>

              <div
                class="h-[6px] w-full bg-primary/20 rounded-full overflow-hidden flex mb-1.5"
              >
                <div
                  class="h-full bg-primary transition-all duration-500 ease-out"
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
                    : 'text-on-surface-variant hover:text-on-surface'}"
                  onclick={toggleLike}
                  title={currentSong.is_liked
                    ? "Remove like"
                    : "Like this theme"}
                  aria-label={currentSong.is_liked
                    ? "Unlike theme"
                    : "Like theme"}
                >
                  <ThumbsUp
                    size={14}
                    class={currentSong.is_liked ? "fill-primary" : ""}
                  />
                  <span class="text-[10px] font-bold"
                    >{currentSong.likes_count || 0}</span
                  >
                </button>

                <button
                  class="flex items-center gap-1 transition-colors {currentSong.is_disliked
                    ? 'text-red-500'
                    : 'text-on-surface-variant hover:text-on-surface'}"
                  onclick={toggleDislike}
                  title={currentSong.is_disliked
                    ? "Remove dislike"
                    : "Dislike this theme"}
                  aria-label={currentSong.is_disliked
                    ? "Undislike theme"
                    : "Dislike theme"}
                >
                  <ThumbsDown
                    size={14}
                    class={currentSong.is_disliked ? "fill-red-500" : ""}
                  />
                  <span class="text-[10px] font-bold"
                    >{currentSong.dislikes_count || 0}</span
                  >
                </button>
              </div>
            </div>

            <div class="h-10 w-px bg-outline-variant/20"></div>

            <button
              onclick={handleRatingClick}
              class="flex flex-col items-start hover:bg-surface-highest px-3 py-1.5 rounded-sm transition-all group active:scale-95"
              title="Rate this theme"
              aria-label="Rate this theme"
            >
              <span
                class="text-on-surface-variant/60 text-[10px] font-bold uppercase tracking-wider group-hover:text-primary transition-colors"
                >Rating</span
              >
              <div class="flex items-center gap-2">
                <span class="text-yellow-400 font-bold text-lg"
                  >{getFormattedScore(
                    currentSong.average_rating,
                    authState.user?.score_format,
                  )}</span
                >
                <Star
                  size={16}
                  class="fill-yellow-400 text-yellow-400 group-hover:rotate-12 transition-transform"
                />
              </div>
            </button>
          </div>

          <div class="flex items-center gap-3">
            <button
              class="w-10 h-10 flex items-center justify-center bg-surface-highest hover:bg-primary/10 border border-outline-variant/10 rounded-full transition-colors {currentSong.is_favorited
                ? 'text-pink-500'
                : 'text-on-surface-variant'}"
              onclick={toggleFavorite}
              title={currentSong.is_favorited
                ? "Remove from favorites"
                : "Add to favorites"}
              aria-label={currentSong.is_favorited
                ? "Remove from favorites"
                : "Add to favorites"}
            >
              <Heart
                size={20}
                class={currentSong.is_favorited ? "fill-pink-500" : ""}
              />
            </button>
            <button
              class="w-10 h-10 flex items-center justify-center bg-surface-highest hover:bg-surface-lowest border border-outline-variant/10 rounded-full transition-colors text-on-surface-variant hover:text-primary"
              onclick={handlePlaylistClick}
              title="Add to Playlist"
              aria-label="Add this theme to a playlist"
            >
              <ListPlus size={20} />
            </button>
            <button
              class="w-10 h-10 flex items-center justify-center bg-surface-highest hover:bg-red-500/10 border border-outline-variant/10 rounded-full transition-colors {currentSong.is_reported
                ? 'text-red-500 opacity-50 cursor-not-allowed'
                : 'text-on-surface-variant hover:text-red-400'}"
              onclick={reportSong}
              disabled={currentSong.is_reported}
              title={currentSong.is_reported
                ? "Already reported"
                : "Report Song"}
              aria-label={currentSong.is_reported
                ? "Already reported"
                : "Report this theme"}
            >
              <AlertTriangle
                size={20}
                class={currentSong.is_reported ? "fill-red-500" : ""}
              />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Sidebar (Right - YouTube Style) -->
    <div class="lg:col-span-4 space-y-4">
      <div>
        <h2
          class="text-lg font-bold flex items-center gap-2 mb-4 text-on-surface"
        >
          <Library size={20} class="text-primary" />
          More from this series
        </h2>
      </div>
      <div
        class="space-y-3 max-h-[480px] overflow-y-auto pr-1 scrollbar-thin scrollbar-thumb-outline-variant/30 scrollbar-track-transparent scroll-smooth"
      >
        {#each relatedSongs as related}
          {@const isSelected = related.id === currentSong.id}
          <a
            use:bindActive={isSelected}
            href="/animes/{currentSong.anime?.slug}/{related.slug}"
            class="flex gap-3 group p-2 pl-1 rounded-r-md transition-all border-l-4 {isSelected
              ? 'bg-surface-low border-primary shadow-sm'
              : 'bg-transparent hover:bg-surface-highest border-transparent'}"
            title="View theme: {getSongName(related)}"
          >
            <div
              class="w-32 aspect-video rounded-md overflow-hidden shrink-0 border border-outline-variant/10"
            >
              <OptimizedImage
                src={currentSong.anime?.cover_url}
                sources={currentSong.anime?.cover_sources}
                alt={getSongName(related)}
                class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500"
                sizes="(max-width: 1024px) 128px, 160px"
              />
            </div>
            <div class="flex-1 min-w-0 flex flex-col justify-center">
              <div class="flex items-center gap-1.5 mb-1">
                <span
                  class="bg-primary/20 text-primary text-[8px] font-black px-1.5 py-0.5 rounded uppercase"
                  >{related.slug}</span
                >
                <span
                  class="text-[10px] text-yellow-400 font-bold flex items-center gap-0.5"
                >
                  <Star size={10} class="fill-yellow-400" />
                  {getFormattedScore(
                    related.average_rating,
                    authState.user?.score_format,
                  )}
                </span>
              </div>
              <h4
                class="text-sm font-bold transition-colors line-clamp-1 {isSelected
                  ? 'text-primary'
                  : 'text-on-surface group-hover:text-primary'}"
              >
                {getSongName(related)}
              </h4>
              <p class="text-[10px] text-on-surface-variant/40 line-clamp-1">
                by {getSongArtistNames(related.artists)}
              </p>
            </div>
          </a>
        {:else}
          <div
            class="text-center py-8 text-on-surface-variant/20 text-xs italic border border-dashed border-outline-variant/10 rounded-md"
          >
            No other themes found for this series.
          </div>
        {/each}
      </div>
      <!-- Recommendation Section -->
      <div class="mt-10 space-y-4">
        <h2
          class="text-lg font-bold flex items-center gap-2 mb-4 text-on-surface"
        >
          <Star size={20} class="text-primary" />
          Recommendations
        </h2>
        <div
          class="bg-surface-low border border-outline-variant/10 rounded-md p-6 flex flex-col items-center justify-center text-center space-y-3 transition-all hover:border-primary/20 group/wip"
        >
          <div
            class="p-3 bg-primary/10 rounded-sm text-primary animate-pulse group-hover/wip:scale-110 transition-transform duration-300"
          >
            <Sparkles size={20} />
          </div>
          <div class="space-y-1">
            <h4
              class="text-xs font-bold text-on-surface uppercase tracking-wider"
            >
              Smart Recommendations
            </h4>
            <p
              class="text-[11px] text-on-surface-variant/60 max-w-[200px] mx-auto leading-relaxed"
            >
              We are building a smart system to recommend more themes based on
              your music taste.
            </p>
          </div>
          <div
            class="px-2 py-0.5 rounded-sm text-[8px] font-black uppercase tracking-widest bg-primary/20 text-primary"
          >
            Coming Soon
          </div>
        </div>
      </div>
    </div>
  </div>
  <!-- Comments Section -->
  <div class="space-y-6">
    <h2 class="text-2xl font-bold flex items-center gap-3">
      <MessageCircle size={24} class="text-primary" />
      Comments
    </h2>

    <!-- New Comment Input -->
    <div class="flex gap-4">
      <div class="w-10 h-10 rounded-full bg-white/10 overflow-hidden shrink-0">
        {#if authState.isAuthenticated && authState.user}
          <OptimizedImage
            src={authState.user.avatar_url}
            sources={authState.user.avatar_sources}
            alt="{authState.user.name}'s avatar"
            class="w-full h-full object-cover"
            sizes="40px"
          />
        {:else}
          <UserIcon size={24} class="text-white/40" />
        {/if}
      </div>
      {#if authState.user?.is_softbanned}
        <div
          class="flex-1 bg-red-500/5 border border-red-500/20 rounded-md p-4 flex items-center gap-3"
        >
          <AlertTriangle class="text-red-500" size={20} />
          <p
            class="text-[11px] text-red-500 font-black uppercase tracking-widest leading-relaxed"
          >
            Your account is currently restricted due to low reputation or
            pending reports. You cannot post comments or ratings at this time.
          </p>
        </div>
      {:else}
        <div class="flex-1 flex flex-col gap-2">
          <label for="main-comment-textarea" class="sr-only"
            >Add a comment</label
          >
          <textarea
            id="main-comment-textarea"
            bind:value={newCommentText}
            placeholder={authState.isAuthenticated
              ? "Add a comment..."
              : "Sign in to comment..."}
            class="w-full bg-surface-container border border-outline-variant/20 rounded-md p-3 text-sm text-on-surface focus:outline-none focus:border-primary transition-colors resize-none disabled:opacity-50 shadow-inner"
            rows="2"
            disabled={!authState.isAuthenticated}
          ></textarea>
          <div class="flex justify-end">
            <button
              onclick={postComment}
              disabled={!newCommentText.trim() || !authState.isAuthenticated}
              class="bg-primary hover:bg-primary/80 text-white font-bold text-xs px-4 py-2 rounded-sm transition-colors disabled:opacity-50"
              title="Post comment"
              aria-label="Post comment"
            >
              Comment
            </button>
          </div>
        </div>
      {/if}
    </div>

    <!-- Comments List -->
    <div class="space-y-4 pb-12">
      {#each comments as comment (comment.uuid)}
        <div class="flex gap-4" id="comment-{comment.uuid}">
          <div
            class="w-10 h-10 rounded-full bg-surface-highest overflow-hidden shrink-0 border border-outline-variant/10"
          >
            <OptimizedImage
              src={comment.user?.avatar_url}
              sources={comment.user?.avatar_sources}
              alt={comment.user?.name}
              class="w-full h-full object-cover"
              sizes="40px"
            />
          </div>
          <div class="flex-1 space-y-2">
            <div
              class="bg-surface-low border border-outline-variant/10 rounded-md rounded-tl-sm p-4"
            >
              <div class="flex justify-between items-start mb-1">
                <div class="flex items-center gap-2">
                  <span class="font-bold text-sm text-on-surface"
                    >{comment.user?.name || "Unknown User"}</span
                  >
                  {#if comment.user?.badges}
                    {#each comment.user.badges as badge}
                      <OptimizedImage
                        src={badge.icon_url || badge.image_url}
                        sources={badge.icon_sources}
                        alt={badge.name}
                        class="w-4 h-4"
                        sizes="16px"
                      />
                    {/each}
                  {/if}
                  <span class="text-xs text-on-surface-variant/40"
                    >{comment.created_at
                      ? new Date(comment.created_at).toLocaleDateString()
                      : "Just now"}</span
                  >
                </div>
              </div>
              {#if editingCommentUuid === comment.uuid}
                <div class="space-y-2">
                  <textarea
                    bind:value={editText}
                    class="w-full bg-surface-container border border-primary/30 rounded-sm p-3 text-sm text-on-surface focus:outline-none focus:border-primary transition-colors resize-none"
                    rows="3"
                  ></textarea>
                  <div class="flex justify-end gap-2">
                    <button
                      onclick={cancelEditing}
                      class="text-xs font-bold text-on-surface-variant hover:text-on-surface px-3 py-1.5 transition-colors"
                    >
                      Cancel
                    </button>
                    <button
                      onclick={saveEdit}
                      class="bg-primary text-white text-xs font-bold px-3 py-1.5 rounded-sm hover:bg-primary/80 transition-colors"
                    >
                      Save
                    </button>
                  </div>
                </div>
              {:else}
                <p class="text-sm text-on-surface-variant whitespace-pre-wrap">
                  {comment.content}
                </p>
              {/if}
            </div>
            <div class="flex justify-between text-md">
              <div class="flex gap-2 items-center">
                <button
                  onclick={() => toggleCommentLike(comment.uuid)}
                  class="flex items-center gap-1 transition-colors {comment.is_liked
                    ? 'text-primary'
                    : 'text-on-surface-variant/40 hover:text-on-surface'}"
                  title={comment.is_liked ? "Unlike comment" : "Like comment"}
                  aria-label={comment.is_liked
                    ? "Unlike comment"
                    : "Like comment"}
                >
                  <ThumbsUp
                    size={16}
                    class={comment.is_liked ? "fill-primary" : ""}
                  />
                  {comment.likes_count || 0}
                </button>
                <button
                  onclick={() => toggleCommentDislike(comment.uuid)}
                  class="flex items-center gap-1 transition-colors {comment.is_disliked
                    ? 'text-red-500'
                    : 'text-on-surface-variant/40 hover:text-on-surface'}"
                  title={comment.is_disliked
                    ? "Undislike comment"
                    : "Dislike comment"}
                  aria-label={comment.is_disliked
                    ? "Undislike comment"
                    : "Dislike comment"}
                >
                  <ThumbsDown
                    size={16}
                    class={comment.is_disliked ? "fill-red-500" : ""}
                  />
                  {comment.dislikes_count || 0}
                </button>
                <button
                  onclick={() => {
                    replyingToUuid =
                      replyingToUuid === comment.uuid ? null : comment.uuid;
                    replyText = "";
                  }}
                  class="text-on-surface-variant/60 hover:text-primary tracking-wider transition-colors flex items-center gap-1"
                  title="Reply to comment"
                  aria-label="Reply to comment"
                  >Reply
                  <CornerDownRight size={16} />
                </button>
              </div>

              <div class="relative">
                <button
                  onclick={(e) => {
                    e.stopPropagation();
                    openDropdownUuid =
                      openDropdownUuid === comment.uuid ? null : comment.uuid;
                  }}
                  class="p-1 hover:bg-surface-highest text-on-surface-variant/40 hover:text-on-surface transition-colors rounded-sm"
                  title="More options"
                  aria-label="More options"
                >
                  <MoreVertical size={16} />
                </button>

                {#if openDropdownUuid === comment.uuid}
                  <div
                    class="absolute right-0 top-full mt-1 w-48 bg-surface-container border border-outline-variant/10 rounded-md shadow-xl z-20 py-1 overflow-hidden"
                  >
                    {#if authState.user && authState.user.uuid === comment.user?.uuid}
                      <button
                        onclick={() => startEditing(comment)}
                        class="w-full text-left px-4 py-2 text-xs font-bold text-on-surface-variant hover:bg-surface-low hover:text-primary transition-colors flex items-center gap-2"
                      >
                        <Edit2 size={14} /> Edit Comment
                      </button>
                    {/if}
                    {#if authState.user && (authState.user.uuid === comment.user?.uuid || authState.isAdmin)}
                      <button
                        onclick={() => deleteComment(comment.uuid)}
                        class="w-full text-left px-4 py-2 text-xs font-bold text-red-400 hover:bg-red-500/10 transition-colors flex items-center gap-2"
                      >
                        <Trash2 size={14} /> Delete Comment
                      </button>
                    {/if}
                    <button
                      onclick={() => openCommentReportModal(comment.uuid)}
                      class="w-full text-left px-4 py-2 text-xs font-bold text-on-surface-variant hover:bg-surface-low hover:text-red-400 transition-colors flex items-center gap-2"
                    >
                      <Flag size={14} /> Report Comment
                    </button>
                    <button
                      onclick={() => openUserReportModal(comment.user)}
                      class="w-full text-left px-4 py-2 text-xs font-bold text-on-surface-variant hover:bg-surface-low hover:text-red-400 transition-colors flex items-center gap-2"
                    >
                      <UserIcon size={14} /> Report User
                    </button>
                  </div>
                {/if}
              </div>
            </div>

            <!-- Inline Reply Input -->
            {#if replyingToUuid === comment.uuid}
              <div class="flex gap-3 mt-3">
                <div class="flex-1 flex flex-col gap-2">
                  <label for="reply-textarea-{comment.uuid}" class="sr-only"
                    >Write a reply</label
                  >
                  <div class="flex gap-2 items-start">
                    <textarea
                      id="reply-textarea-{comment.uuid}"
                      bind:value={replyText}
                      placeholder="Write a reply..."
                      class="w-full bg-surface-container border border-outline-variant/10 rounded-sm p-2 text-sm text-on-surface focus:outline-none focus:border-primary transition-colors resize-none"
                      rows="1"
                    ></textarea>
                    <button
                      onclick={() => postReply(comment.uuid)}
                      disabled={!replyText.trim()}
                      class="bg-surface-highest hover:bg-primary hover:text-white text-on-surface font-bold text-xs px-3 py-2 rounded-sm transition-colors disabled:opacity-50 shrink-0"
                      title="Send reply"
                      aria-label="Send reply"
                    >
                      Send
                    </button>
                  </div>
                </div>
              </div>
            {/if}

            <!-- Replies List -->
            {#if comment.replies && comment.replies.length > 0}
              <div
                class="space-y-3 mt-3 border-l-2 border-outline-variant/10 pl-4"
              >
                {#each comment.replies as reply (reply.uuid)}
                  <div class="flex gap-3" id="comment-{reply.uuid}">
                    <div
                      class="w-8 h-8 rounded-full bg-surface-highest overflow-hidden shrink-0 border border-outline-variant/10"
                    >
                      <OptimizedImage
                        src={reply.user?.avatar_url}
                        sources={reply.user?.avatar_sources}
                        alt={reply.user?.name}
                        class="w-full h-full object-cover"
                        sizes="32px"
                      />
                    </div>
                    <div class="flex-1 space-y-1">
                      <div
                        class="bg-surface-container/50 border border-outline-variant/10 rounded-md rounded-tl-sm p-3"
                      >
                        <div class="flex justify-between items-start mb-1">
                          <div class="flex items-center gap-2">
                            <span class="font-bold text-xs text-on-surface"
                              >{reply.user?.name || "Unknown User"}</span
                            >
                            {#if reply.user?.badges}
                              <div class="flex gap-1 ml-1">
                                {#each reply.user.badges as badge}
                                  <OptimizedImage
                                    src={badge.icon_url || badge.image_url}
                                    sources={badge.icon_sources}
                                    alt={badge.name}
                                    class="w-3.5 h-3.5"
                                    sizes="14px"
                                  />
                                {/each}
                              </div>
                            {/if}
                            <span class="text-[10px] text-on-surface-variant/40"
                              >{reply.created_at
                                ? new Date(
                                    reply.created_at,
                                  ).toLocaleDateString()
                                : "Just now"}</span
                            >
                          </div>
                          <div class="relative">
                            <button
                              onclick={(e) => {
                                e.stopPropagation();
                                openDropdownUuid =
                                  openDropdownUuid === reply.uuid
                                    ? null
                                    : reply.uuid;
                              }}
                              class="p-1 hover:bg-surface-highest text-on-surface-variant/40 hover:text-on-surface transition-colors rounded-sm"
                              title="More options"
                              aria-label="More options"
                            >
                              <MoreVertical size={14} />
                            </button>

                            {#if openDropdownUuid === reply.uuid}
                              <div
                                class="absolute right-0 top-full mt-1 w-48 bg-surface-container border border-outline-variant/10 rounded-md shadow-xl z-20 py-1 overflow-hidden"
                              >
                                {#if authState.user && authState.user.uuid === reply.user?.uuid}
                                  <button
                                    onclick={() => startEditing(reply)}
                                    class="w-full text-left px-4 py-2 text-[11px] font-bold text-on-surface-variant hover:bg-surface-low hover:text-primary transition-colors flex items-center gap-2"
                                  >
                                    <Edit2 size={12} /> Edit Reply
                                  </button>
                                {/if}
                                {#if authState.user && (authState.user.uuid === reply.user?.uuid || authState.isAdmin)}
                                  <button
                                    onclick={() =>
                                      deleteComment(reply.uuid, comment.uuid)}
                                    class="w-full text-left px-4 py-2 text-[11px] font-bold text-red-400 hover:bg-red-500/10 transition-colors flex items-center gap-2"
                                  >
                                    <Trash2 size={12} /> Delete Reply
                                  </button>
                                {/if}
                                <button
                                  onclick={() =>
                                    openCommentReportModal(reply.uuid)}
                                  class="w-full text-left px-4 py-2 text-[11px] font-bold text-on-surface-variant hover:bg-surface-low hover:text-red-400 transition-colors flex items-center gap-2"
                                >
                                  <Flag size={12} /> Report Reply
                                </button>
                                <button
                                  onclick={() =>
                                    openUserReportModal(reply.user)}
                                  class="w-full text-left px-4 py-2 text-[11px] font-bold text-on-surface-variant hover:bg-surface-low hover:text-red-400 transition-colors flex items-center gap-2"
                                >
                                  <UserIcon size={12} /> Report User
                                </button>
                              </div>
                            {/if}
                          </div>
                        </div>
                        {#if editingCommentUuid === reply.uuid}
                          <div class="space-y-2 mt-2">
                            <textarea
                              bind:value={editText}
                              class="w-full bg-surface-container border border-primary/30 rounded-sm p-3 text-sm text-on-surface focus:outline-none focus:border-primary transition-colors resize-none"
                              rows="2"
                            ></textarea>
                            <div class="flex justify-end gap-2">
                              <button
                                onclick={cancelEditing}
                                class="text-xs font-bold text-on-surface-variant hover:text-on-surface px-3 py-1.5 transition-colors"
                              >
                                Cancel
                              </button>
                              <button
                                onclick={saveEdit}
                                class="bg-primary text-white text-xs font-bold px-3 py-1.5 rounded-sm hover:bg-primary/80 transition-colors"
                              >
                                Save
                              </button>
                            </div>
                          </div>
                        {:else}
                          <p
                            class="text-[13px] text-on-surface-variant whitespace-pre-wrap mt-1"
                          >
                            {reply.content}
                          </p>
                        {/if}
                        <div class="flex gap-4 items-center mt-2">
                          <button
                            onclick={() =>
                              toggleCommentLike(reply.uuid, comment.uuid)}
                            class="flex items-center gap-1 transition-colors {reply.is_liked
                              ? 'text-primary'
                              : 'text-on-surface-variant/20 hover:text-on-surface'}"
                            title={reply.is_liked
                              ? "Unlike reply"
                              : "Like reply"}
                            aria-label={reply.is_liked
                              ? "Unlike reply"
                              : "Like reply"}
                          >
                            <ThumbsUp
                              size={14}
                              class={reply.is_liked ? "fill-primary" : ""}
                            />
                            <span class="text-xs">{reply.likes_count || 0}</span
                            >
                          </button>
                          <button
                            onclick={() =>
                              toggleCommentDislike(reply.uuid, comment.uuid)}
                            class="flex items-center gap-1 transition-colors {reply.is_disliked
                              ? 'text-red-500'
                              : 'text-on-surface-variant/20 hover:text-on-surface'}"
                          >
                            <ThumbsDown
                              size={14}
                              class={reply.is_disliked ? "fill-red-500" : ""}
                            />
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
          <MessageSquare size={48} class="text-on-surface-variant/10 mb-2" />
          <p class="text-on-surface-variant/40 font-bold text-sm">
            No comments yet
          </p>
          <p class="text-on-surface-variant/20 text-xs mt-1">
            Be the first to share your thoughts on this song!
          </p>
        </div>
      {/each}
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

{#if showCommentReportModal && reportingCommentUuid}
  <CommentReportModal
    show={showCommentReportModal}
    commentId={reportingCommentUuid}
    onClose={() => {
      showCommentReportModal = false;
      reportingCommentUuid = null;
    }}
  />
{/if}

<PlaylistModal
  show={showPlaylistModal}
  song={currentSong}
  onClose={() => (showPlaylistModal = false)}
/>

{#if showUserReportModal && reportingUser}
  <UserReportModal
    show={showUserReportModal}
    reportedUser={reportingUser}
    onClose={() => {
      showUserReportModal = false;
      reportingUser = null;
    }}
  />
{/if}

<svelte:window onclick={() => (openDropdownUuid = null)} />
