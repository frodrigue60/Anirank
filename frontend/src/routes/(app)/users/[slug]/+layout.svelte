<script lang="ts">
  import { fade } from "svelte/transition";
  import { page } from "$app/state";
  import { authState, setUser } from "$lib/state/auth.svelte";
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import XPProgressBar from "$lib/components/XPProgressBar.svelte";
  import UserReportModal from "$lib/components/UserReportModal.svelte";

  let { data, children } = $props();

  // svelte-ignore state_referenced_locally
  let isFollowing = $state(data.profile?.is_following || false);
  // svelte-ignore state_referenced_locally
  let followersCount = $state(data.profile?.followers_count || 0);
  // svelte-ignore state_referenced_locally
  let followingCount = $state(data.profile?.following_count || 0);
  let isProcessing = $state(false);
  let showReportModal = $state(false);

  // Sync state if data changes
  $effect(() => {
    isFollowing = data.profile?.is_following || false;
    followersCount = data.profile?.followers_count || 0;
    followingCount = data.profile?.following_count || 0;

    // Sync authState if this is the logged-in user's profile
    // This handles cases like daily login XP awarded during profile retrieval
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

  // Reactive URLs that favor authState when viewing own profile
  const isOwnProfile = $derived(
    authState.user && authState.user.uuid === data.profile?.uuid,
  );
  const avatarUrl = $derived(
    isOwnProfile ? authState.user?.avatar_url : data.profile?.avatar_url,
  );
  const bannerUrl = $derived(
    isOwnProfile ? authState.user?.banner_url : data.profile?.banner_url,
  );

  // Quadratic XP formula: Min XP = Level * (Level - 1) / 2 * 1000
  const profileXP = $derived(
    isOwnProfile ? (authState.user?.xp ?? 0) : (data.profile?.xp ?? 0),
  );
  const profileLevel = $derived(
    isOwnProfile ? (authState.user?.level ?? 1) : (data.profile?.level ?? 1),
  );
  const accentColor = $derived(
    (isOwnProfile
      ? authState.user?.profile_color
      : data.profile?.profile_color) || "#3db4f2",
  );

  const currentLevelMinXP = $derived(
    ((profileLevel * (profileLevel - 1)) / 2) * 1000,
  );
  const nextLevelXP = $derived(
    (((profileLevel + 1) * profileLevel) / 2) * 1000,
  );
</script>

{#if data.profile}
  <div>
    <div class="relative h-[200px] md:h-[300px] w-full overflow-hidden">
      <div
        class="absolute inset-0 bg-linear-to-t from-surface-container via-transparent to-transparent z-10"
      ></div>
      <!-- svelte-ignore a11y_img_redundant_alt -->
      {#if bannerUrl}
        <img
          alt="Cover Image"
          class="w-full h-full object-cover"
          data-alt="User banner image"
          src={bannerUrl}
        />
      {:else}
        <img
          alt="Cover Image"
          class="w-full h-full object-cover"
          data-alt="User banner image"
          src="/images/placeholders/default-banner.jpg"
        />
      {/if}
      <div
        class="absolute bottom-0 left-0 w-full px-6 md:px-20 pb-8 z-20 flex flex-col md:flex-row items-end gap-6"
      >
        <div class="flex w-full gap-6">
          <div class="relative group">
            <div class="size-32 md:size-44 rounded-full overflow-hidden">
              {#if avatarUrl}
                <img
                  alt="Profile"
                  class="w-full h-full object-cover"
                  data-alt="User avatar image"
                  src={avatarUrl}
                />
              {:else}
                <img
                  alt="Profile"
                  class="w-full h-full object-cover"
                  data-alt="User avatar image"
                  src="/images/placeholders/default.jpg"
                />
              {/if}
            </div>
          </div>

          <div class="flex-1 mb-1">
            <h1
              class="text-3xl md:text-5xl font-black text-on-surface tracking-tighter"
            >
              {data.profile.name}
            </h1>
            <p class="text-on-surface-variant font-medium text-lg mt-1">
              followers {followersCount} | following {followingCount}
            </p>
            <div class="mt-4">
              <XPProgressBar
                xp={profileXP}
                level={profileLevel}
                {nextLevelXP}
                {currentLevelMinXP}
                {accentColor}
              />
            </div>

            <div class="flex flex-col text-sm text-on-surface-variant mt-4">
              {#if data.profile.badges}
                <div class="flex gap-1 items-center">
                  {#each data.profile.badges as badge}
                    <img
                      src={badge.icon_url}
                      alt={badge.name}
                      class="w-5 h-5"
                      title={badge.name}
                    />
                  {/each}
                </div>
              {/if}
            </div>
          </div>
        </div>
        <div class="hidden md:flex gap-3 mb-2">
          {#if isOwnProfile}
            <a
              href="/settings"
              class="flex items-center justify-center gap-2 px-6 h-11 rounded-sm bg-surface-highest/50 hover:bg-surface-highest text-on-surface border border-on-surface-variant/10 transition-all font-bold text-sm"
            >
              <span class="material-symbols-outlined text-sm">settings</span>
              Edit Profile
            </a>
          {:else}
            <button
              onclick={handleFollow}
              disabled={isProcessing}
              class="flex items-center justify-center gap-2 px-8 h-11 rounded-sm font-bold text-sm transition-all shadow-lg {isFollowing
                ? 'bg-surface-highest text-on-surface hover:bg-surface-highest/80'
                : 'text-white hover:opacity-90 shadow-lg'}"
              style={!isFollowing
                ? `background-color: ${accentColor}; box-shadow: 0 10px 20px ${accentColor}33`
                : ""}
            >
              {#if isProcessing}
                <span class="animate-spin material-symbols-outlined text-sm"
                  >sync</span
                >
              {:else}
                <span class="material-symbols-outlined text-sm"
                  >{isFollowing ? "person_remove" : "person_add"}</span
                >
              {/if}
              {isFollowing ? "Unfollow" : "Follow"}
            </button>
            {#if authState.isAuthenticated}
              <button
                onclick={() => (showReportModal = true)}
                class="flex items-center justify-center size-11 rounded-sm bg-red-500/50 hover:bg-red-500 border border-on-surface-variant/10 text-on-surface transition-all"
                title="Report User"
              >
                <span class="material-symbols-outlined">report</span>
              </button>
            {:else}
              <a
                href="/login"
                class="flex items-center justify-center size-11 rounded-sm bg-red-500/50 hover:bg-red-500 border border-on-surface-variant/10 text-on-surface transition-all"
              >
                <span class="material-symbols-outlined text-sm">report</span>
              </a>
            {/if}
          {/if}
        </div>
      </div>
    </div>

    <!-- User Report Modal -->
    <UserReportModal
      show={showReportModal}
      reportedUser={data.profile}
      onClose={() => (showReportModal = false)}
    />

    <!-- content -->
    <div class="max-w-7xl mx-auto px-6 py-6">
      <div class="py-6 border-b border-primary/10 mb-8">
        <div class="flex gap-8 overflow-x-auto no-scrollbar">
          <a
            href={`/users/${data.profile.slug}`}
            class="pb-4 font-bold transition-all border-b-2 {page.url
              .pathname === `/users/${data.profile.slug}`
              ? ''
              : 'border-transparent text-on-surface-variant hover:text-on-surface'}"
            style={page.url.pathname === `/users/${data.profile.slug}`
              ? `border-color: ${accentColor}; color: ${accentColor}`
              : ""}>Overview</a
          >
          {#if data.profile.anilist_id}
            <a
              href={`/users/${data.profile.slug}/anime-list`}
              class="pb-4 font-bold transition-all border-b-2 {page.url
                .pathname === `/users/${data.profile.slug}/anime-list`
                ? ''
                : 'border-transparent text-on-surface-variant hover:text-on-surface'}"
              style={page.url.pathname ===
              `/users/${data.profile.slug}/anime-list`
                ? `border-color: ${accentColor}; color: ${accentColor}`
                : ""}>Anime List</a
            >
          {/if}
          <a
            href={`/users/${data.profile.slug}/playlists`}
            class="pb-4 font-bold transition-all border-b-2 {page.url
              .pathname === `/users/${data.profile.slug}/playlists`
              ? ''
              : 'border-transparent text-on-surface-variant hover:text-on-surface'}"
            style={page.url.pathname === `/users/${data.profile.slug}/playlists`
              ? `border-color: ${accentColor}; color: ${accentColor}`
              : ""}>Playlists</a
          >
          <a
            href={`/users/${data.profile.slug}/favorites`}
            class="pb-4 font-bold transition-all border-b-2 {page.url
              .pathname === `/users/${data.profile.slug}/favorites`
              ? ''
              : 'border-transparent text-on-surface-variant hover:text-on-surface'}"
            style={page.url.pathname === `/users/${data.profile.slug}/favorites`
              ? `border-color: ${accentColor}; color: ${accentColor}`
              : ""}>Favorites</a
          >
          <a
            href={`/users/${data.profile.slug}/artists`}
            class="pb-4 font-bold transition-all border-b-2 {page.url
              .pathname === `/users/${data.profile.slug}/artists`
              ? ''
              : 'border-transparent text-on-surface-variant hover:text-on-surface'}"
            style={page.url.pathname === `/users/${data.profile.slug}/artists`
              ? `border-color: ${accentColor}; color: ${accentColor}`
              : ""}>Artists</a
          >
          <a
            href={`/users/${data.profile.slug}/followers`}
            class="pb-4 font-bold transition-all border-b-2 {page.url
              .pathname === `/users/${data.profile.slug}/followers`
              ? ''
              : 'border-transparent text-on-surface-variant hover:text-on-surface'} flex gap-2 items-center"
            style={page.url.pathname === `/users/${data.profile.slug}/followers`
              ? `border-color: ${accentColor}; color: ${accentColor}`
              : ""}
          >
            Followers
            <span
              class="text-[10px] bg-surface-highest px-2 py-0.5 rounded-full font-bold text-on-surface-variant"
              >{followersCount}</span
            >
          </a>
          <a
            href={`/users/${data.profile.slug}/following`}
            class="pb-4 font-bold transition-all border-b-2 {page.url
              .pathname === `/users/${data.profile.slug}/following`
              ? ''
              : 'border-transparent text-on-surface-variant hover:text-on-surface'} flex gap-2 items-center"
            style={page.url.pathname === `/users/${data.profile.slug}/following`
              ? `border-color: ${accentColor}; color: ${accentColor}`
              : ""}
          >
            Following
            <span
              class="text-[10px] bg-surface-highest px-2 py-0.5 rounded-full font-bold text-on-surface-variant"
              >{followingCount}</span
            >
          </a>
        </div>
      </div>

      <!-- Render the active sub-route here -->
      {@render children()}
    </div>
  </div>
{:else}
  <div
    class="min-h-[60vh] flex flex-col items-center justify-center text-center p-6"
    in:fade
  >
    <div
      class="w-24 h-24 rounded-full bg-red-500/10 flex items-center justify-center text-red-500 mb-6"
    >
      <span class="material-symbols-outlined text-[48px]">person_off</span>
    </div>
    <h1 class="text-3xl font-black text-on-surface mb-2">User Not Found</h1>
    <p class="text-on-surface-variant max-w-md">
      The profile you're looking for doesn't exist or has been removed.
    </p>
    <a
      href="/"
      class="mt-8 bg-primary hover:bg-primary/80 text-white px-8 py-3.5 rounded-xl font-bold text-sm transition-all shadow-lg shadow-primary/20"
    >
      Back to Home
    </a>
  </div>
{/if}
