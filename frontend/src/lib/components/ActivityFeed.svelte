<script lang="ts">
  import { onMount } from "svelte";
  import api from "$lib/api";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import MessagesSquare from "lucide-svelte/icons/messages-square";
  import ArrowRight from "lucide-svelte/icons/arrow-right";
  import Inbox from "lucide-svelte/icons/inbox";
  import Sparkles from "lucide-svelte/icons/sparkles";
  import ChevronRight from "lucide-svelte/icons/chevron-right";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";

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

<div class="flex flex-col gap-8">
  <div class="flex items-center justify-between px-2">
    <h2
      class="flex items-center gap-3 text-2xl font-black tracking-tight text-on-surface"
    >
      <div class="w-10 h-10 flex items-center justify-center text-primary">
        <MessagesSquare size={20} />
      </div>
      What's Happening
    </h2>
    {#if recentOnly}
      <a
        href="/interactions"
        class="text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant transition-all hover:text-primary hover:translate-x-1"
        >Explore Feed <ArrowRight class="inline-block ml-1" size={12} /></a
      >
    {/if}
  </div>

  <div
    class="divide-y divide-outline-variant/5 overflow-hidden rounded-md shadow-sm"
  >
    {#if loading && activities.length === 0}
      {#each Array(recentOnly ? 4 : 8) as _}
        <div
          class="flex animate-pulse items-start gap-4 p-6 bg-surface-container"
        >
          <div
            class="h-16 w-12 shrink-0 rounded-md bg-surface-highest/50"
          ></div>
          <div class="flex flex-1 flex-col gap-3 justify-center">
            <div class="h-3 w-1/3 rounded-full bg-surface-highest/50"></div>
            <div class="h-5 w-2/3 rounded-full bg-surface-highest/50"></div>
          </div>
        </div>
      {/each}
    {:else if activities.length === 0}
      <div class="bg-surface-container">
        <EmptyState
          title="No activities to show"
          message="There's no recent activity to display right now. Check back later!"
          icon={Inbox}
        />
      </div>
    {:else}
      {#each activities as activity}
        {@const target = activity.target}
        {@const type = activity.type}
        {@const isSong = activity.target_type === "song"}
        {@const isArtist = activity.target_type === "artist"}
        {@const isLevel = activity.target_type === "level"}
        {@const isBadge = activity.target_type === "badge"}
        {@const isUser = type === "follow"}

        <div
          class="group flex cursor-pointer items-start gap-5 p-6 transition-all bg-surface-container hover:bg-primary/10"
        >
          <div class="relative shrink-0">
            <!-- Target Cover -->
            {#if isSong && target}
              <OptimizedImage
                src={target.anime?.cover_url}
                sources={target.anime?.cover_sources}
                alt={target.anime?.title}
                class="h-20 w-14 rounded-sm object-cover border border-outline-variant/10 shadow-lg"
                sizes="56px"
              />
            {:else if isArtist && target}
              <OptimizedImage
                src={target.avatar_url}
                sources={target.avatar_sources}
                alt={target.name}
                class="h-20 w-14 rounded-sm object-cover border border-outline-variant/10 shadow-lg"
                sizes="56px"
              />
            {:else if isLevel}
              <div
                class="flex h-20 w-14 flex-col items-center justify-center rounded-sm border border-amber-500/30 bg-linear-to-br from-amber-400/90 to-orange-600 shadow-lg shadow-amber-500/20"
              >
                <Sparkles size={24} class="text-white/40 mb-1" />
                <span class="text-2xl font-black text-white leading-none"
                  >{activity.target_id}</span
                >
              </div>
            {:else if isBadge && activity.badge}
              <div
                class="flex h-20 w-14 items-center justify-center rounded-sm border border-primary/20 bg-surface-highest/50 p-2 shadow-lg shadow-primary/10"
              >
                <OptimizedImage
                  src={activity.badge.icon_url}
                  sources={activity.badge.icon_sources}
                  alt={activity.badge.name}
                  class="h-full w-full object-contain"
                  sizes="56px"
                />
              </div>
            {:else if isUser && target}
              <OptimizedImage
                src={target.avatar_url}
                sources={target.avatar_sources}
                alt={target.name}
                class="w-14 h-14 rounded-sm object-cover border border-outline-variant/10 shadow-lg"
                sizes="56px"
              />
            {:else}
              <div
                class="flex h-14 w-14 items-center justify-center rounded-sm border border-outline-variant/10 bg-surface-highest/30 shadow-inner"
              >
                <Sparkles size={30} />
              </div>
            {/if}

            <!-- User Avatar (Overlapping) -->
            <div
              class="absolute -right-2 -bottom-2 h-8 w-8 overflow-hidden rounded-full border-4 border-surface shadow-xl shadow-black/10 transition-transform group-hover:scale-110"
            >
              <OptimizedImage
                src={activity.user?.avatar_url}
                sources={activity.user?.avatar_sources}
                alt={activity.user?.name}
                class="w-full h-full object-cover"
                sizes="32px"
              />
            </div>
          </div>

          <div
            class="flex-1 min-w-0 flex flex-col justify-center h-full self-center"
          >
            <div class="flex flex-col">
              <span
                class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant/90 leading-tight mb-1"
              >
                <span class="text-on-surface">{activity.user?.name}</span>
                {#if type === "rate"}
                  rated <span class="text-rating-star font-black"
                    >★ {activity.value}</span
                  >
                {:else if type === "favorite"}
                  added to <span class="text-red-400">Favorites</span>
                {:else if type === "comment"}
                  commented on
                {:else if type === "reply"}
                  replied on
                {:else if type === "follow"}
                  started following
                {:else if type === "level_up"}
                  reached <span class="text-amber-500 font-black">New Level {activity.target_id}</span>
                {:else if type === "badge_award"}
                  earned a <span class="text-primary font-black">New Badge</span>
                {:else}
                  interacted
                {/if}
              </span>

              {#if isSong && target}
                <a
                  href="/songs/{target.anime?.slug}/{target.slug}"
                  class="flex flex-col truncate text-lg font-bold tracking-tight text-on-surface group-hover:text-primary transition-colors"
                >
                  <span class="truncate">{target.anime?.title}</span>
                  <span class="text-xs text-on-surface-variant/85 leading-none"
                    >{target.song_romaji || target.name}</span
                  >
                </a>
              {:else if isArtist && target}
                <a
                  href="/artists/{target.slug}"
                  class="truncate text-lg font-bold tracking-tight text-on-surface group-hover:text-primary transition-colors"
                >
                  <span>{target.name}</span>
                </a>
              {:else if isUser && target}
                <a
                  href="/profile/{target.slug}"
                  class="truncate text-lg font-bold tracking-tight text-on-surface group-hover:text-primary transition-colors"
                >
                  <span>{target.name}</span>
                </a>
              {:else if isLevel}
                <div class="flex flex-col truncate">
                  <span class="text-lg font-bold tracking-tight text-on-surface"
                    >{activity.value ? 'Level ' + activity.value : 'Unknown Level'}</span
                  >
                  <span class="text-xs text-on-surface-variant/85 leading-none"
                    >Keep going to reach the next milestone</span
                  >
                </div>
              {:else if isBadge && activity.badge}
                <div class="flex flex-col truncate">
                  <span
                    class="text-lg font-bold tracking-tight text-on-surface group-hover:text-primary transition-colors"
                    >{activity.badge.name ?? 'Unknown Badge'}</span
                  >
                  <span class="text-xs text-on-surface-variant/85 leading-none"
                    >New Achievement Unlocked</span
                  >
                </div>
              {:else}
                <span class="text-on-surface-variant/20 italic"
                  >Item unavailable</span
                >
              {/if}
            </div>
          </div>
          <div class="flex flex-col items-end gap-2 shrink-0 self-center">
            <span class="text-[10px] text-on-surface-variant/80 font-bold"
              >{getTimeAgo(activity.created_at)}</span
            >
            <ChevronRight
              size={18}
              class="transition-transform group-hover:translate-x-1 group-hover:text-primary/30"
            />
          </div>
        </div>
      {/each}
    {/if}
  </div>

  {#if !recentOnly && activities.length > 0}
    <div class="flex items-center justify-between gap-4 px-4 pt-4">
      <button
        class="text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant transition-all hover:text-on-surface hover:-translate-x-1 disabled:opacity-20 flex items-center gap-2"
        onclick={() => goToPage(page - 1)}
        disabled={page <= 1 || loading}
      >
        <ArrowLeft size={14} />
        Previous
      </button>
      <div class="flex items-center gap-2">
        <span class="w-8 h-px bg-outline-variant/20"></span>
        <span
          class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant/40 whitespace-nowrap"
          >Page {page}</span
        >
        <span class="w-8 h-px bg-outline-variant/20"></span>
      </div>
      <button
        class="text-[10px] font-black uppercase tracking-[0.2em] text-primary transition-all hover:text-primary-container hover:translate-x-1 disabled:opacity-20 flex items-center gap-2"
        onclick={() => goToPage(page + 1)}
        disabled={!hasMore || loading}
      >
        Next
        <ArrowRight size={14} />
      </button>
    </div>
  {/if}
</div>
