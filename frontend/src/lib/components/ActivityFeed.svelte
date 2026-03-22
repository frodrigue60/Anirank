<script lang="ts">
  import { onMount } from "svelte";
  import api from "$lib/api";

  let { recentOnly = false } = $props();

  let activities = $state<any[]>([]);
  let loading = $state(true);

  let page = $state(1);
  let hasMore = $state(true);
  const perPage = 20;

  async function fetchActivities(pageToFetch = 1) {
    if (recentOnly) pageToFetch = 1;
    try {
      loading = true;
      const response = await api.get(
        recentOnly ? "/activities/recent" : `/activities?page=${pageToFetch}`,
      );
      activities = response.data.data || [];
      if (!recentOnly) {
        hasMore = activities.length === perPage;
      }
    } catch (error) {
      console.error("Failed to fetch activities:", error);
    } finally {
      loading = false;
    }
  }

  function goToPage(newPage: number) {
    if (recentOnly) return;
    if (newPage < 1) return;
    if (newPage === page) return;
    page = newPage;
    void fetchActivities(newPage);
  }

  function getTimeAgo(dateString: string) {
    const date = new Date(dateString);
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diffInSeconds < 60) return "just now";
    const minutes = Math.floor(diffInSeconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
  }

  onMount(() => {
    void fetchActivities(1);
  });
</script>

<div
  class="bg-surface-darker rounded-2xl p-6 border border-white/5 flex flex-col gap-6"
>
  <div class="flex items-center justify-between">
    <h3 class="font-bold text-white text-lg flex items-center gap-2">
      <span class="material-symbols-outlined text-primary">history</span>
      Recent Activity
    </h3>
    {#if recentOnly}
      <a
        href="/interactions"
        class="text-primary text-xs font-bold uppercase tracking-wider hover:underline"
        title="View full activity feed"
        aria-label="View all interaction activities"
        >View All</a
      >
    {/if}
  </div>

  {#if loading}
    <div class="flex flex-col gap-4">
      {#each Array(3) as _}
        <div class="flex gap-3 animate-pulse">
          <div class="w-10 h-10 rounded-full bg-white/5"></div>
          <div class="flex-1 flex flex-col gap-2">
            <div class="h-3 w-3/4 bg-white/5 rounded"></div>
            <div class="h-2 w-1/2 bg-white/5 rounded"></div>
          </div>
        </div>
      {/each}
    </div>
  {:else if activities.length === 0}
    <div class="text-sm text-white/40 text-center py-4 italic">
      No recent activity found.
    </div>
  {:else}
    <div class="flex flex-col gap-5">
      {#each activities as activity}
        <div class="flex gap-3 items-start group">
          <div class="shrink-0 pt-1">
            {#if activity.user?.avatar_url}
              <img
                src={activity.user.avatar_url}
                alt="{activity.user.name}'s avatar"
                title="{activity.user.name}'s avatar"
                class="w-8 h-8 rounded-full object-cover border border-white/10"
              />
            {:else}
              <div
                class="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-primary text-xs font-bold border border-primary/20"
              >
                {activity.user?.name?.charAt(0) || "U"}
              </div>
            {/if}
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-xs text-white/90 leading-relaxed">
              <span
                class="font-bold text-white uppercase text-[10px] tracking-wider"
                >{activity.user?.name}</span
              >
              <span class="text-white/60">
                {#if activity.action === "favorite"}
                  added to favorites
                {:else if activity.action === "rate"}
                  rated {#if activity.action_value} <span class="text-white font-bold">{activity.action_value}</span> {/if}
                {:else if activity.action === "comment"}
                  commented on
                {:else if activity.action === "reply"}
                  replied on
                {:else if activity.action === "follow"}
                  started following
                {:else}
                  interacted with
                {/if}
              </span>

              {#if activity.target_type === "song" && activity.song}
                <a
                  href="/songs/{activity.song.anime?.slug}/{activity.song.slug}"
                  class="text-primary hover:underline font-medium"
                  title="View song: {activity.song.name}"
                >
                  {activity.song.name}
                </a>
                <span class="text-white/40 text-[10px] block mt-0.5"
                  >{activity.song.anime?.title}</span
                >
              {:else if activity.target_type === "artist" && activity.artist}
                <a
                  href="/artists/{activity.artist.slug}"
                  class="text-primary hover:underline font-medium"
                  title="View artist: {activity.artist.name}"
                >
                  {activity.artist.name}
                </a>
              {:else}
                <span class="text-white/40">something</span>
              {/if}
            </p>
            <span
              class="text-[9px] font-bold text-white/20 uppercase tracking-widest mt-1 block"
            >
              {getTimeAgo(activity.created_at)}
            </span>
          </div>
        </div>
      {/each}
    </div>

    {#if !recentOnly && activities.length > 0}
      <div class="flex items-center justify-between gap-4 pt-2">
        <button
          class="text-white/40 text-xs font-bold uppercase tracking-wider hover:underline disabled:text-white/20"
          onclick={() => goToPage(page - 1)}
          disabled={page <= 1 || loading}
          title="Previous page"
          aria-label="Go to previous page of activities"
        >
          Prev
        </button>
        <span class="text-white/40 text-xs">Page {page}</span>
        <button
          class="text-primary text-xs font-bold uppercase tracking-wider hover:underline disabled:text-white/20"
          onclick={() => goToPage(page + 1)}
          disabled={!hasMore || loading}
          title="Next page"
          aria-label="Go to next page of activities"
        >
          Next
        </button>
      </div>
    {/if}
  {/if}
</div>
