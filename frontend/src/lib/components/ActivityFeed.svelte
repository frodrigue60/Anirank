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
      if (!recentOnly && response.data.pagination) {
        hasMore = response.data.pagination.has_more;
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
        {@const target = activity.target}
        {@const type = activity.type}
        {@const isSong = activity.target_type === "song"}
        {@const isArtist = activity.target_type === "artist"}
        {@const isUser = type === "follow"}

        <div
          class="p-4 flex items-start gap-4 hover:bg-white/5 transition-colors group cursor-pointer"
        >
          <div class="relative shrink-0">
            <!-- Target Cover -->
            {#if isSong && target}
              <img
                src={target.anime?.cover_url ||
                  "/images/placeholders/default.jpg"}
                alt={target.anime?.title}
                class="w-12 h-18 rounded-lg object-cover border border-white/10"
              />
            {:else if isArtist && target}
              <img
                src={target.avatar_url || "/images/placeholders/default.jpg"}
                alt={target.name}
                class="w-12 h-18 rounded-lg object-cover border border-white/10"
              />
            {:else if isUser && target}
              <img
                src={target.avatar_url || "/images/placeholders/default.jpg"}
                alt={target.name}
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
                {#if type === "rate"}
                  rated <span class="text-yellow-400">{activity.value}</span>
                {:else if type === "favorite"}
                  added favorite
                {:else if type === "comment"}
                  commented on
                {:else if type === "reply"}
                  replied on
                {:else if type === "follow"}
                  started following
                {:else}
                  interacted
                {/if}
              </span>

              {#if isSong && target}
                <a
                  href="/songs/{target.anime?.slug}/{target.slug}"
                  class="text-base font-bold text-white group-hover:text-primary transition-colors truncate flex flex-col"
                >
                  <span class="text-primary">{target.anime?.title}</span>
                  <span class="italic text-white/60 font-normal"
                    >{target.song_romaji || target.name}</span
                  >
                </a>
              {:else if isArtist && target}
                <a
                  href="/artists/{target.slug}"
                  class="text-base font-bold text-white group-hover:text-primary transition-colors truncate"
                >
                  <span class="text-primary">{target.name}</span>
                </a>
              {:else if isUser && target}
                <a
                  href="/profile/{target.slug}"
                  class="text-base font-bold text-white group-hover:text-primary transition-colors truncate"
                >
                  <span class="text-primary">{target.name}</span>
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
