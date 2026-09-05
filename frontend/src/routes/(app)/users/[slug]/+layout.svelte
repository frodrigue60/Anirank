<script lang="ts">
  import { fade } from "svelte/transition";
  import { page } from "$app/state";
  import { authState, setUser } from "$lib/state/auth.svelte";
  import { goto } from "$app/navigation";
  import { toastState } from "$lib/state/toast.svelte";
  import api from "$lib/api";
  import XPProgressBar from "$lib/components/XPProgressBar.svelte";
  import UserReportModal from "$lib/components/UserReportModal.svelte";
  import Settings from "lucide-svelte/icons/settings";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import UserMinus from "lucide-svelte/icons/user-minus";
  import UserPlus from "lucide-svelte/icons/user-plus";
  import Flag from "lucide-svelte/icons/flag";
  import UserX from "lucide-svelte/icons/user-x";
  import Share2 from "lucide-svelte/icons/share-2";
  import LibraryMusic from "lucide-svelte/icons/library";
  import Star from "lucide-svelte/icons/star";
  import Clapperboard from "lucide-svelte/icons/clapperboard";
  import ListMusic from "lucide-svelte/icons/list-music";
  import Calendar from "lucide-svelte/icons/calendar";
  import Construction from "lucide-svelte/icons/construction";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";

  let { data, children } = $props();

  // svelte-ignore state_referenced_locally
  let isFollowing = $state(data.profile?.is_following || false);
  // svelte-ignore state_referenced_locally
  let followersCount = $state(data.profile?.followers_count || 0);
  // svelte-ignore state_referenced_locally
  let followingCount = $state(data.profile?.following_count || 0);
  let isProcessing = $state(false);
  let showReportModal = $state(false);

  $effect(() => {
    isFollowing = data.profile?.is_following || false;
    followersCount = data.profile?.followers_count || 0;
    followingCount = data.profile?.following_count || 0;

    if (
      data.profile &&
      authState.user &&
      data.profile.uuid === authState.user.uuid
    ) {
      if (
        authState.user.xp !== data.profile.xp ||
        authState.user.level !== data.profile.level
      ) {
        setUser({
          ...authState.user,
          xp: data.profile.xp,
          level: data.profile.level,
        });
      }
    }
  });

  async function handleFollow() {
    if (!authState.user) {
      goto(`/login?redirect=${encodeURIComponent(page.url.pathname)}`);
      return;
    }

    isProcessing = true;
    try {
      const response = isFollowing
        ? await api.delete(`/users/${data.profile.uuid}/follow`)
        : await api.post(`/users/${data.profile.uuid}/follow`);

      if (response.status === 200) {
        isFollowing = !isFollowing;
        followersCount += isFollowing ? 1 : -1;
      }
    } catch (error) {
      console.error("Error following user:", error);
    } finally {
      isProcessing = false;
    }
  }

  async function handleShare() {
    const url = `${page.url.origin}/users/${data.profile.slug}`;
    try {
      if (navigator.share) {
        await navigator.share({
          title: `${data.profile.name} on AniRank`,
          url,
        });
        return;
      }
      await navigator.clipboard.writeText(url);
      toastState.addToast("Profile link copied", "success");
    } catch (e: any) {
      if (e?.name === "AbortError") return;
      try {
        await navigator.clipboard.writeText(url);
        toastState.addToast("Profile link copied", "success");
      } catch {
        toastState.addToast("Could not share profile", "error");
      }
    }
  }

  function memberSinceLabel(createdAt: string | undefined) {
    if (!createdAt) return null;
    const d = new Date(createdAt);
    if (Number.isNaN(d.getTime())) return null;
    return d.toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
    });
  }

  function tabActive(path: string) {
    return page.url.pathname === path;
  }

  function tabClass(active: boolean) {
    return active
      ? "bg-primary text-white"
      : "bg-surface-container text-on-surface-variant/80 hover:bg-surface-highest hover:text-on-surface";
  }

  const isOwnProfile = $derived(
    authState.user && authState.user.uuid === data.profile?.uuid,
  );
  const avatarUrl = $derived(
    isOwnProfile ? authState.user?.avatar_url : data.profile?.avatar_url,
  );
  const avatarSources = $derived(
    isOwnProfile ? authState.user?.avatar_sources : data.profile?.avatar_sources,
  );
  const bannerUrl = $derived(
    isOwnProfile ? authState.user?.banner_url : data.profile?.banner_url,
  );
  const bannerSources = $derived(
    isOwnProfile ? authState.user?.banner_sources : data.profile?.banner_sources,
  );

  const profileXP = $derived(
    isOwnProfile ? (authState.user?.xp ?? 0) : (data.profile?.xp ?? 0),
  );
  const profileLevel = $derived(
    isOwnProfile ? (authState.user?.level ?? 1) : (data.profile?.level ?? 1),
  );
  const accentColor = $derived(
    (isOwnProfile
      ? authState.user?.profile_color
      : data.profile?.profile_color) || "#683bc9",
  );

  const currentLevelMinXP = $derived(
    ((profileLevel * (profileLevel - 1)) / 2) * 1000,
  );
  const nextLevelXP = $derived(
    (((profileLevel + 1) * profileLevel) / 2) * 1000,
  );

  const memberSince = $derived(memberSinceLabel(data.profile?.created_at));
  const ratingsCount = $derived(data.profile?.ratings_count ?? 0);
  const playlistsAndFavorites = $derived(
    (data.playlistsCount ?? 0) + (data.favoritesCount ?? 0),
  );

  const basePath = $derived(`/users/${data.profile?.slug}`);
</script>

{#if data.profile}
  <div class="w-full bg-surface">
    <!-- Banner -->
    <section class="relative w-full overflow-hidden bg-surface-low">
      <div class="relative w-full h-48 sm:h-64 md:h-80">
        {#if bannerUrl}
          <OptimizedImage
            src={bannerUrl}
            sources={bannerSources}
            alt=""
            class="w-full h-full object-cover"
            sizes="100vw"
          />
        {:else}
          <OptimizedImage
            src="/images/placeholders/default-banner.svg"
            alt=""
            class="w-full h-full object-cover"
            sizes="100vw"
          />
        {/if}
        <div
          class="absolute inset-0 bg-linear-to-b from-surface/30 via-surface/50 to-surface"
        ></div>
      </div>

      <div class="max-w-7xl mx-auto px-6 lg:px-12 -mt-20 sm:-mt-24 relative z-10 pb-6">
        <div
          class="flex flex-col lg:flex-row lg:items-end justify-between gap-6 pb-6"
        >
          <div
            class="flex flex-col sm:flex-row items-center sm:items-end gap-5 text-center sm:text-left"
          >
            <div
              class="size-28 sm:size-32 md:size-36 rounded-full overflow-hidden bg-surface-highest shrink-0 ring-4 ring-surface"
            >
              {#if avatarUrl}
                <OptimizedImage
                  src={avatarUrl}
                  sources={avatarSources}
                  alt="{data.profile.name}'s avatar"
                  class="w-full h-full object-cover"
                  sizes="(max-width: 640px) 112px, 144px"
                />
              {:else}
                <OptimizedImage
                  src="/images/placeholders/default.svg"
                  alt="{data.profile.name}'s avatar"
                  class="w-full h-full object-cover"
                  sizes="(max-width: 640px) 112px, 144px"
                />
              {/if}
            </div>

            <div class="flex flex-col gap-2 min-w-0">
              <div
                class="flex flex-wrap items-center justify-center sm:justify-start gap-2"
              >
                <h1
                  class="text-3xl sm:text-4xl font-black tracking-tight text-on-surface"
                >
                  {data.profile.name}
                </h1>
                {#if data.profile.badges?.length}
                  <div class="flex flex-wrap items-center gap-1.5">
                    {#each data.profile.badges as badge}
                      {@const tip = badge.description
                        ? `${badge.name} — ${badge.description}`
                        : badge.name}
                      <span
                        class="relative group inline-flex size-8 items-center justify-center rounded-sm bg-surface-container hover:bg-surface-highest transition-colors"
                        title={tip}
                        tabindex="0"
                      >
                        <OptimizedImage
                          src={badge.icon_url || badge.image_url}
                          sources={badge.icon_sources}
                          alt={badge.name}
                          class="w-5 h-5"
                          sizes="20px"
                        />
                        <span
                          role="tooltip"
                          class="pointer-events-none absolute left-1/2 -translate-x-1/2 top-full mt-1.5 z-30 whitespace-nowrap rounded-sm bg-surface-highest px-2 py-1 text-[10px] font-bold text-on-surface opacity-0 group-hover:opacity-100 group-focus-within:opacity-100 transition-opacity shadow-sm"
                        >
                          {badge.name}
                        </span>
                      </span>
                    {/each}
                  </div>
                {/if}
              </div>

              <p
                class="text-sm text-on-surface-variant/80 flex flex-wrap items-center justify-center sm:justify-start gap-x-2 gap-y-1"
              >
                <span>
                  <span class="font-bold text-on-surface">{followersCount}</span>
                  followers
                </span>
                <span class="text-outline-variant" aria-hidden="true">·</span>
                <span>
                  <span class="font-bold text-on-surface">{followingCount}</span>
                  following
                </span>
                {#if memberSince}
                  <span class="text-outline-variant" aria-hidden="true">·</span>
                  <span class="inline-flex items-center gap-1">
                    <Calendar size={14} class="text-primary shrink-0" aria-hidden="true" />
                    Member since {memberSince}
                  </span>
                {/if}
              </p>

              <div class="mt-1 max-w-md mx-auto sm:mx-0 w-full">
                <XPProgressBar
                  xp={profileXP}
                  level={profileLevel}
                  {nextLevelXP}
                  {currentLevelMinXP}
                  {accentColor}
                />
              </div>
            </div>
          </div>

          <div
            class="flex flex-wrap items-center justify-center sm:justify-end gap-2"
          >
            {#if isOwnProfile}
              <a
                href="/settings"
                class="inline-flex items-center justify-center gap-2 min-h-11 px-5 rounded-md bg-surface-highest text-on-surface font-bold text-sm hover:bg-surface-container transition-colors"
                title="Edit profile"
                aria-label="Edit profile"
              >
                <Settings size={16} aria-hidden="true" />
                Edit Profile
              </a>
            {:else}
              <button
                type="button"
                onclick={handleFollow}
                disabled={isProcessing}
                class="inline-flex items-center justify-center gap-2 min-h-11 px-5 rounded-md font-bold text-sm transition-colors disabled:opacity-50 {isFollowing
                  ? 'bg-surface-highest text-on-surface hover:bg-surface-container'
                  : 'bg-primary text-white hover:bg-primary-container'}"
                title={isFollowing ? "Unfollow" : "Follow"}
                aria-label={isFollowing ? "Unfollow user" : "Follow user"}
              >
                {#if isProcessing}
                  <RefreshCw size={16} class="animate-spin" aria-hidden="true" />
                {:else if isFollowing}
                  <UserMinus size={16} aria-hidden="true" />
                {:else}
                  <UserPlus size={16} aria-hidden="true" />
                {/if}
                {isFollowing ? "Unfollow" : "Follow"}
              </button>
            {/if}

            <button
              type="button"
              onclick={handleShare}
              class="inline-flex items-center justify-center size-11 rounded-md bg-surface-highest text-on-surface-variant hover:text-on-surface hover:bg-surface-container transition-colors"
              title="Share profile"
              aria-label="Share profile"
            >
              <Share2 size={18} aria-hidden="true" />
            </button>

            {#if !isOwnProfile}
              {#if authState.isAuthenticated}
                <button
                  type="button"
                  onclick={() => (showReportModal = true)}
                  class="inline-flex items-center justify-center size-11 rounded-md bg-surface-highest text-on-surface-variant hover:text-red-500 hover:bg-surface-container transition-colors"
                  title="Report user"
                  aria-label="Report user"
                >
                  <Flag size={18} aria-hidden="true" />
                </button>
              {:else}
                <a
                  href="/login?redirect={encodeURIComponent(page.url.pathname)}"
                  class="inline-flex items-center justify-center size-11 rounded-md bg-surface-highest text-on-surface-variant hover:text-red-500 hover:bg-surface-container transition-colors"
                  title="Report user"
                  aria-label="Report user"
                >
                  <Flag size={18} aria-hidden="true" />
                </a>
              {/if}
            {/if}
          </div>
        </div>

        <!-- Stats row -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
          <div
            class="bg-surface-container rounded-md p-4 flex items-center justify-between gap-3"
          >
            <div class="min-w-0">
              <span
                class="text-[10px] font-bold uppercase tracking-widest text-on-surface-variant/80"
              >
                Themes Rated
              </span>
              <div class="text-2xl font-black text-on-surface mt-1 tabular-nums">
                {ratingsCount.toLocaleString()}
              </div>
            </div>
            <div
              class="size-10 rounded-md bg-surface-highest flex items-center justify-center text-primary shrink-0"
              aria-hidden="true"
            >
              <LibraryMusic size={20} />
            </div>
          </div>

          <div
            class="relative bg-surface-low rounded-md p-4 flex items-center justify-between gap-3 border border-dashed border-outline-variant"
          >
            <span
              class="absolute top-2 right-2 inline-flex items-center gap-1 px-1.5 py-0.5 rounded-sm bg-surface-highest text-[9px] font-black uppercase tracking-wider text-on-surface-variant"
            >
              <Construction size={10} aria-hidden="true" />
              WIP
            </span>
            <div class="min-w-0 pr-8">
              <span
                class="text-[10px] font-bold uppercase tracking-widest text-on-surface-variant/80"
              >
                Average Score
              </span>
              <div class="text-2xl font-black text-on-surface-variant/40 mt-1">
                —.—
              </div>
              <p class="text-[11px] text-on-surface-variant/70 mt-0.5">
                Needs score aggregate API
              </p>
            </div>
            <div
              class="size-10 rounded-md bg-surface-highest flex items-center justify-center text-on-surface-variant/50 shrink-0"
              aria-hidden="true"
            >
              <Star size={20} />
            </div>
          </div>

          <div
            class="relative bg-surface-low rounded-md p-4 flex items-center justify-between gap-3 border border-dashed border-outline-variant"
          >
            <span
              class="absolute top-2 right-2 inline-flex items-center gap-1 px-1.5 py-0.5 rounded-sm bg-surface-highest text-[9px] font-black uppercase tracking-wider text-on-surface-variant"
            >
              <Construction size={10} aria-hidden="true" />
              WIP
            </span>
            <div class="min-w-0 pr-8">
              <span
                class="text-[10px] font-bold uppercase tracking-widest text-on-surface-variant/80"
              >
                Animes in List
              </span>
              <div class="text-2xl font-black text-on-surface-variant/40 mt-1">
                —
              </div>
              <p class="text-[11px] text-on-surface-variant/70 mt-0.5">
                {#if data.profile.anilist_id}
                  AniList linked — count API pending
                {:else}
                  Link AniList or add list API
                {/if}
              </p>
            </div>
            <div
              class="size-10 rounded-md bg-surface-highest flex items-center justify-center text-on-surface-variant/50 shrink-0"
              aria-hidden="true"
            >
              <Clapperboard size={20} />
            </div>
          </div>

          <div
            class="bg-surface-container rounded-md p-4 flex items-center justify-between gap-3"
          >
            <div class="min-w-0">
              <span
                class="text-[10px] font-bold uppercase tracking-widest text-on-surface-variant/80"
              >
                Playlists & Favorites
              </span>
              <div class="text-2xl font-black text-on-surface mt-1 tabular-nums">
                {playlistsAndFavorites.toLocaleString()}
              </div>
              <p class="text-[11px] text-on-surface-variant/80 mt-0.5 truncate">
                {(data.playlistsCount ?? 0).toLocaleString()} playlists · {(data.favoritesCount ?? 0).toLocaleString()} favorites
              </p>
            </div>
            <div
              class="size-10 rounded-md bg-surface-highest flex items-center justify-center text-primary shrink-0"
              aria-hidden="true"
            >
              <ListMusic size={20} />
            </div>
          </div>
        </div>
      </div>
    </section>

    <UserReportModal
      show={showReportModal}
      reportedUser={data.profile}
      onClose={() => (showReportModal = false)}
    />

    <!-- Sticky tabs -->
    <nav
      class="sticky top-16 z-40 bg-surface border-b border-outline-variant/20"
      aria-label="Profile sections"
    >
      <div
        class="max-w-7xl mx-auto px-6 lg:px-12 flex items-center gap-2 overflow-x-auto py-3 no-scrollbar"
      >
        <a
          href={basePath}
          data-sveltekit-noscroll
          class="px-4 py-2 rounded-md font-semibold text-xs tracking-tight whitespace-nowrap transition-colors {tabClass(
            tabActive(basePath),
          )}"
        >
          Overview
        </a>
        {#if data.profile.anilist_id}
          <a
            href="{basePath}/anime-list"
            data-sveltekit-noscroll
            class="px-4 py-2 rounded-md font-semibold text-xs tracking-tight whitespace-nowrap transition-colors {tabClass(
              tabActive(`${basePath}/anime-list`),
            )}"
          >
            Anime List
          </a>
        {/if}
        <a
          href="{basePath}/playlists"
          data-sveltekit-noscroll
          class="px-4 py-2 rounded-md font-semibold text-xs tracking-tight whitespace-nowrap transition-colors {tabClass(
            tabActive(`${basePath}/playlists`),
          )}"
        >
          Playlists
        </a>
        <a
          href="{basePath}/favorites"
          data-sveltekit-noscroll
          class="px-4 py-2 rounded-md font-semibold text-xs tracking-tight whitespace-nowrap transition-colors {tabClass(
            tabActive(`${basePath}/favorites`),
          )}"
        >
          Favorites
        </a>
        <a
          href="{basePath}/artists"
          data-sveltekit-noscroll
          class="px-4 py-2 rounded-md font-semibold text-xs tracking-tight whitespace-nowrap transition-colors {tabClass(
            tabActive(`${basePath}/artists`),
          )}"
        >
          Artists
        </a>
        <a
          href="{basePath}/followers"
          data-sveltekit-noscroll
          class="px-4 py-2 rounded-md font-semibold text-xs tracking-tight whitespace-nowrap transition-colors inline-flex items-center gap-2 {tabClass(
            tabActive(`${basePath}/followers`),
          )}"
        >
          Followers
          <span
            class="text-[10px] px-1.5 py-0.5 rounded-sm font-bold {tabActive(
              `${basePath}/followers`,
            )
              ? 'bg-primary-container text-white'
              : 'bg-surface-highest text-on-surface-variant'}"
          >
            {followersCount}
          </span>
        </a>
        <a
          href="{basePath}/following"
          data-sveltekit-noscroll
          class="px-4 py-2 rounded-md font-semibold text-xs tracking-tight whitespace-nowrap transition-colors inline-flex items-center gap-2 {tabClass(
            tabActive(`${basePath}/following`),
          )}"
        >
          Following
          <span
            class="text-[10px] px-1.5 py-0.5 rounded-sm font-bold {tabActive(
              `${basePath}/following`,
            )
              ? 'bg-primary-container text-white'
              : 'bg-surface-highest text-on-surface-variant'}"
          >
            {followingCount}
          </span>
        </a>
      </div>
    </nav>

    <div class="max-w-7xl mx-auto px-6 lg:px-12 py-8">
      {@render children()}
    </div>
  </div>
{:else}
  <div
    class="min-h-[60vh] flex flex-col items-center justify-center text-center p-6"
    in:fade
  >
    <div
      class="w-24 h-24 rounded-full bg-surface-container flex items-center justify-center text-red-500 mb-6"
    >
      <UserX size={48} aria-hidden="true" />
    </div>
    <h1 class="text-3xl font-black text-on-surface mb-2">User Not Found</h1>
    <p class="text-on-surface-variant/80 max-w-md">
      The profile you're looking for doesn't exist or has been removed.
    </p>
    <a
      href="/"
      class="mt-8 bg-primary hover:bg-primary-container text-white px-8 py-3.5 rounded-md font-bold text-sm transition-colors"
    >
      Back to Home
    </a>
  </div>
{/if}
