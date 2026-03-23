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

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between px-1">
    <h2
      class="text-xl font-bold tracking-tight flex items-center gap-2 text-white"
    >
      <span class="material-symbols-outlined text-primary">forum</span>
      Community Activity
    </h2>
    {#if recentOnly}
      <a
        href="/interactions"
        class="text-white/40 hover:text-white text-[10px] font-bold uppercase tracking-wider transition-colors"
        >View All</a
      >
    {/if}
  </div>

  <div
    class="bg-surface-darker/50 border border-white/5 rounded-2xl overflow-hidden divide-y divide-white/5"
  >
    {#if loading && activities.length === 0}
      {#each Array(recentOnly ? 4 : 8) as _}
        <div class="p-4 flex items-start gap-4 animate-pulse">
          <div class="w-12 h-12 rounded-lg bg-white/5 shrink-0"></div>
          <div class="flex-1 flex flex-col gap-2">
            <div class="h-3 w-1/2 bg-white/5 rounded"></div>
            <div class="h-4 w-3/4 bg-white/5 rounded"></div>
          </div>
        </div>
      {/each}
    {:else if activities.length === 0}
      <div class="p-8 text-center text-sm text-white/30 italic">
        No recent activity.
      </div>
    {:else}
      {#each activities as activity}
        <div
          class="p-4 flex items-start gap-4 hover:bg-white/5 transition-colors group cursor-pointer"
        >
          <div class="relative shrink-0">
            <!-- Target Cover -->
            {#if activity.target_type === "song" && activity.song}
              <img
                src={activity.song.anime?.cover_url ||
                  "/images/placeholders/default.jpg"}
                alt={activity.song.anime?.title}
                class="w-12 h-18 rounded-lg object-cover border border-white/10"
              />
            {:else if activity.target_type === "artist" && activity.artist}
              <img
                src={activity.artist.avatar_url ||
                  "/images/placeholders/default.jpg"}
                alt={activity.artist.name}
                class="w-12 h-12 rounded-lg object-cover border border-white/10"
              />
            {:else if activity.target_type === "user" && activity.user_target}
              <img
                src={activity.user_target.avatar_url ||
                  "/images/placeholders/default.jpg"}
                alt={activity.user_target.name}
                class="w-12 h-12 rounded-lg object-cover border border-white/10"
              />
            {:else}
              <div
                class="w-12 h-12 rounded-lg bg-surface-light flex items-center justify-center border border-white/10"
              >
                <span class="material-symbols-outlined text-white/20"
                  >interests</span
                >
              </div>
            {/if}

            <!-- User Avatar (Overlapping) -->
            <div
              class="absolute -bottom-1 -right-1 w-6 h-6 rounded-full border-2 border-[#0a0a0c] overflow-hidden bg-surface-darker"
            >
              {#if activity.user?.avatar_url}
                <img
                  src={activity.user.avatar_url}
                  alt={activity.user.name}
                  class="w-full h-full object-cover"
                />
              {:else}
                <div
                  class="w-full h-full flex items-center justify-center bg-primary/20 text-[10px] font-bold text-primary"
                >
                  {activity.user?.name?.charAt(0) || "U"}
                </div>
              {/if}
            </div>
          </div>

          <div class="flex-1 min-w-0">
            <div class="flex flex-col">
              <span
                class="text-[11px] text-white/50 uppercase font-bold tracking-wider"
              >
                {activity.user?.name}
                {#if activity.action === "rate"}
                  rated <span class="text-yellow-400"
                    >{activity.action_value}</span
                  >
                {:else if activity.action === "favorite"}
                  added favorite
                {:else if activity.action === "comment"}
                  commented on
                {:else if activity.action === "reply"}
                  replied on
                {:else if activity.action === "follow"}
                  started following
                {:else}
                  interacted
                {/if}
              </span>

              {#if activity.target_type === "song" && activity.song}
                <a
                  href="/songs/{activity.song.anime?.slug}/{activity.song.slug}"
                  class="text-base font-bold text-white group-hover:text-primary transition-colors truncate flex flex-col"
                >
                  <span class="text-primary"
                    >{activity.song.anime?.title} {activity.song.slug}</span
                  >
                  <span class="italic text-white/60 font-normal"
                    >{activity.song.name}</span
                  >
                </a>
              {:else if activity.target_type === "artist" && activity.artist}
                <a
                  href="/artists/{activity.artist.slug}"
                  class="text-base font-bold text-white group-hover:text-primary transition-colors truncate"
                >
                  <span class="text-primary">{activity.artist.name}</span>
                </a>
              {:else if activity.target_type === "user" && activity.user_target}
                <a
                  href="/profile/{activity.user_target.slug}"
                  class="text-base font-bold text-white group-hover:text-primary transition-colors truncate"
                >
                  <span class="text-primary">{activity.user_target.name}</span>
                </a>
              {:else}
                <span class="text-white/40">Item Deleted</span>
              {/if}
            </div>
            <span class="text-[10px] text-white/30 mt-0.5 block"
              >{getTimeAgo(activity.created_at)}</span
            >
          </div>
        </div>
      {/each}
    {/if}
  </div>

  {#if !recentOnly && activities.length > 0}
    <div class="flex items-center justify-between gap-4 pt-2">
      <button
        class="text-white/40 text-xs font-bold uppercase tracking-wider hover:underline disabled:text-white/20 transition-colors"
        onclick={() => goToPage(page - 1)}
        disabled={page <= 1 || loading}
      >
        Prev
      </button>
      <span
        class="text-white/40 text-[10px] font-bold uppercase tracking-widest"
        >Page {page}</span
      >
      <button
        class="text-primary text-xs font-bold uppercase tracking-wider hover:underline disabled:text-white/20 transition-colors"
        onclick={() => goToPage(page + 1)}
        disabled={!hasMore || loading}
      >
        Next
      </button>
    </div>
  {/if}
</div>
