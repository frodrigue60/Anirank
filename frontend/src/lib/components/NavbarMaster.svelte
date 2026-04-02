<script lang="ts">
  import { authState, logout } from "$lib/state/auth.svelte";
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import SearchModal from "$lib/components/SearchModal.svelte";
  import RequestModal from "$lib/components/RequestModal.svelte";

  let showUserDropdown = $state(false);
  let showDiscoverDropdown = $state(false);
  let showMobileMenu = $state(false);
  let showSearchModal = $state(false);
  let showRequestModal = $state(false);

  let unreadCount = $state(0);

  async function fetchUnreadCount() {
    if (!authState.isAuthenticated) return;
    try {
      const response = await api.get("/notifications");
      unreadCount = response.data.unread_count || 0;
    } catch (error) {
      console.error("Error fetching notifications:", error);
    }
  }

  $effect(() => {
    if (authState.isAuthenticated) {
      fetchUnreadCount();
      const interval = setInterval(fetchUnreadCount, 60000); // Check every minute
      return () => clearInterval(interval);
    }
  });

  // function to close dropdowns when clicking outside
  function closeDropdowns() {
    showUserDropdown = false;
    showDiscoverDropdown = false;
  }

  function handleKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === "k") {
      e.preventDefault();
      showSearchModal = true;
    }
  }

  async function handleLogout() {
    logout();
    showUserDropdown = false;
    showDiscoverDropdown = false;
    goto("/");
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  onclick={closeDropdowns}
  class="sticky top-0 z-50 bg-surface-dark border-b border-white/5 w-full"
>
  <header
    class="max-w-[1440px] mx-auto px-6 h-16 flex items-center justify-between gap-4"
  >
    <div class="flex items-center gap-10">
      <a
        class="flex items-center gap-2 group"
        href="/"
        title="AniRank Home"
        aria-label="Go to AniRank Home"
      >
        <!-- <div
          class="w-8 h-8 rounded-lg bg-primary flex items-center justify-center text-white shadow-lg shadow-primary/40 group-hover:scale-105 transition-transform"
        >
          <span class="material-symbols-outlined text-[20px]">music_note</span>
        </div> -->
        <span class="text-xl font-bold tracking-tight text-white">AniRank</span>
      </a>
      <nav class="hidden md:flex items-center gap-6">
        <a
          class="text-sm font-medium transition-colors {page.url.pathname.includes(
            '/songs/seasonal',
          )
            ? 'text-primary'
            : 'text-white/60 hover:text-white'}"
          href="/songs/seasonal"
          title="Seasonal Themes"
          aria-label="View seasonal anime themes">Season</a
        >
        <a
          class="text-sm font-medium transition-colors {page.url.pathname.includes(
            '/songs/ranking',
          )
            ? 'text-primary'
            : 'text-white/60 hover:text-white'}"
          href="/songs/ranking"
          title="Song Rankings"
          aria-label="View top rated anime songs">Ranking</a
        >

        <!-- Discover Dropdown -->
        <div class="relative">
          <button
            onclick={(e) => {
              e.stopPropagation();
              showDiscoverDropdown = !showDiscoverDropdown;
              showUserDropdown = false;
            }}
            class="flex items-center gap-1 text-sm font-medium transition-colors {showDiscoverDropdown
              ? 'text-primary'
              : 'text-white/60 hover:text-white'}"
            title="Discover more content"
            aria-label="Toggle discover dropdown menu"
          >
            Discover
            <span
              class="material-symbols-outlined text-[18px] transition-transform {showDiscoverDropdown
                ? 'rotate-180'
                : ''}">expand_more</span
            >
          </button>

          {#if showDiscoverDropdown}
            <div
              class="absolute left-0 top-full mt-3 w-48 overflow-hidden rounded-xl border border-white/10 bg-surface-dark py-1 shadow-2xl"
            >
              <a
                href="/animes"
                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                title="Browse Anime Series"
              >
                <span class="material-symbols-outlined text-[18px]">tv</span> Animes
              </a>
              <a
                href="/artists"
                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                title="Browse Artists"
              >
                <span class="material-symbols-outlined text-[18px]">group</span> Artists
              </a>
              <a
                href="/songs"
                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                title="Browse Themes and Songs"
              >
                <span class="material-symbols-outlined text-[18px]"
                  >music_note</span
                > Songs
              </a>
              <a
                href="/studios"
                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                title="Browse Animation Studios"
              >
                <span class="material-symbols-outlined text-[18px]">movie</span> Studios
              </a>
              <a
                href="/producers"
                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                title="Browse Music Producers"
              >
                <span class="material-symbols-outlined text-[18px]"
                  >theater_comedy</span
                > Producers
              </a>
              <a
                href="/playlists"
                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                title="Browse Playlists"
              >
                <span class="material-symbols-outlined text-[18px]"
                  >queue_music</span
                > Playlists
              </a>
              <a
                href="/tournaments"
                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                title="View Tournaments"
              >
                <span class="material-symbols-outlined text-[18px]">trophy</span
                > Tournaments
              </a>
              <a
                href="/users/ranking"
                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                title="View User Rankings"
              >
                <span class="material-symbols-outlined text-[18px]"
                  >leaderboard</span
                > User Ranking
              </a>
            </div>
          {/if}
        </div>
      </nav>
    </div>

    <div class="flex items-center gap-4">
      <!-- Search Trigger -->
      <button
        onclick={() => (showSearchModal = true)}
        class="relative hidden sm:flex items-center group h-10 w-10 rounded-full bg-white/5 hover:bg-white/10 text-white transition-colors justify-center"
        title="Search (Ctrl + K)"
        aria-label="Open search modal"
      >
        <span class="material-symbols-outlined text-[20px]">search</span>
      </button>

      <!-- Notifications (Hidden on mobile) -->
      {#if authState.isAuthenticated}
        <a
          href="/notifications"
          class="hidden sm:flex h-10 w-10 items-center justify-center rounded-full bg-white/5 hover:bg-white/10 text-white transition-colors relative"
          title="Notifications"
          aria-label="View notifications"
        >
          <span class="material-symbols-outlined text-[20px]"
            >notifications</span
          >

          {#if unreadCount > 0}
            <span
              class="absolute top-2.5 right-2.5 w-2 h-2 bg-primary rounded-full border border-surface-dark"
            ></span>
          {/if}
        </a>
      {/if}

      <div class="h-8 w-px bg-white/10 hidden sm:block"></div>

      <!-- User Toggle -->
      <div class="relative">
        {#if authState.loading}
          <div
            class="h-9 w-9 animate-pulse rounded-full bg-surface-darker"
          ></div>
        {:else}
          <button
            onclick={(e) => {
              e.stopPropagation();
              showUserDropdown = !showUserDropdown;
            }}
            class="flex items-center gap-2 group"
            title="User Menu"
            aria-label="Toggle user dropdown menu"
          >
            <div
              class="w-9 h-9 overflow-hidden rounded-full border-2 border-transparent group-hover:border-primary transition-all bg-white/5 flex items-center justify-center text-white/40 group-hover:text-primary"
            >
              {#if authState.isAuthenticated && authState.user}
                {#if authState.user.avatar_url}
                  <img
                    src={authState.user.avatar_url}
                    alt="{authState.user.name}'s avatar"
                    title="{authState.user.name}'s avatar"
                    class="h-full w-full object-cover"
                  />
                {:else}
                  <img
                    src="/images/placeholders/default.jpg"
                    alt="{authState.user.name}'s default avatar"
                    title="{authState.user.name}'s default avatar"
                    class="h-full w-full object-cover"
                  />
                {/if}
              {:else}
                <span class="material-symbols-outlined text-[20px]">person</span
                >
              {/if}
            </div>
            <span
              class="material-symbols-outlined text-white/40 text-[20px] group-hover:text-white hidden sm:block"
              >expand_more</span
            >
          </button>

          {#if showUserDropdown}
            <div
              class="absolute right-0 top-full mt-3 w-56 overflow-hidden rounded-xl border border-white/10 bg-surface-dark py-1 shadow-2xl"
            >
              {#if authState.isAuthenticated && authState.user}
                <div class="border-b border-white/5 px-4 py-3">
                  <p class="truncate text-sm font-bold text-white">
                    {authState.user.name}
                  </p>
                </div>
                <div class="py-1">
                  {#if authState.isStaff}
                    <a
                      href="/admin"
                      class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                    >
                      <span class="material-symbols-outlined text-[18px]"
                        >dashboard</span
                      > Dashboard
                    </a>
                  {/if}
                  <a
                    href="/users/{authState.user.slug}"
                    class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                  >
                    <span class="material-symbols-outlined text-[18px]"
                      >person_pin</span
                    > My Profile
                  </a>
                  <a
                    href="/users/{authState.user.slug}/playlists"
                    class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                  >
                    <span class="material-symbols-outlined text-[18px]"
                      >list_alt</span
                    > My Playlists
                  </a>
                  <a
                    href="/settings"
                    class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                  >
                    <span class="material-symbols-outlined text-[18px]"
                      >settings</span
                    > Settings
                  </a>
                  <button
                    onclick={() => {
                      showRequestModal = true;
                      showUserDropdown = false;
                    }}
                    class="flex w-full items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                  >
                    <span class="material-symbols-outlined text-[18px]"
                      >add_box</span
                    > Request Song
                  </button>
                </div>
                <div class="border-t border-white/5 py-1">
                  <button
                    onclick={handleLogout}
                    class="flex w-full items-center gap-3 px-4 py-2 text-sm text-red-400 transition-colors hover:bg-red-400/10 hover:text-red-300"
                    title="Sign Out"
                  >
                    <span class="material-symbols-outlined text-[18px]"
                      >logout</span
                    > Logout
                  </button>
                </div>
              {:else}
                <div class="border-b border-white/5 px-4 py-3">
                  <p class="text-sm font-bold text-white">Guest User</p>
                  <p class="text-xs text-white/40">Join the community!</p>
                </div>
                <div class="py-1">
                  <a
                    href="/login"
                    class="flex w-full items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                  >
                    <span class="material-symbols-outlined text-[18px]"
                      >login</span
                    >
                    Sign In
                  </a>
                  <a
                    href="/register"
                    class="flex w-full items-center gap-3 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5 hover:text-primary"
                  >
                    <span class="material-symbols-outlined text-[18px]"
                      >person_add</span
                    >
                    Register
                  </a>
                </div>
              {/if}
            </div>
          {/if}
        {/if}
      </div>

      <!-- Mobile Toggle -->
      <button
        onclick={() => (showMobileMenu = true)}
        class="flex h-10 w-10 items-center justify-center text-white/60 transition-colors hover:text-white md:hidden"
        title="Open Mobile Menu"
        aria-label="Open mobile menu"
      >
        <span class="material-symbols-outlined">menu</span>
      </button>
    </div>
  </header>
</div>

<!-- Mobile Menu Drawer -->
{#if showMobileMenu}
  <div class="fixed inset-0 z-100">
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      onclick={() => (showMobileMenu = false)}
      class="absolute inset-0 bg-black/60"
    ></div>

    <div
      class="absolute right-0 top-0 flex h-full w-[280px] flex-col border-l border-white/5 bg-surface-dark shadow-2xl"
    >
      <div
        class="flex h-16 items-center justify-between border-b border-white/5 px-6"
      >
        <span class="text-lg font-bold tracking-tight text-white">Menu</span>
        <button
          onclick={() => (showMobileMenu = false)}
          class="text-white/40 transition-colors hover:text-white flex items-center justify-center"
        >
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>

      <nav class="flex-1 space-y-2 overflow-y-auto px-4 py-6">
        <a
          href="/songs/seasonal"
          class="flex items-center gap-3 rounded-xl px-4 py-3 text-white/60 transition-all hover:bg-white/5 hover:text-white"
        >
          <span class="material-symbols-outlined text-[18px]"
            >calendar_month</span
          > <span class="text-sm font-medium">Seasonal</span>
        </a>
        <a
          href="/songs/ranking"
          class="flex items-center gap-3 rounded-xl px-4 py-3 text-white/60 transition-all hover:bg-white/5 hover:text-white"
        >
          <span class="material-symbols-outlined text-[18px]">leaderboard</span>
          <span class="text-sm font-medium">Ranking</span>
        </a>
        <a
          href="/animes"
          class="flex items-center gap-3 rounded-xl px-4 py-3 text-white/60 transition-all hover:bg-white/5 hover:text-white"
        >
          <span class="material-symbols-outlined text-[18px]">tv</span>
          <span class="text-sm font-medium">Animes</span>
        </a>
        <a
          href="/users/ranking"
          class="flex items-center gap-3 rounded-xl px-4 py-3 text-white/60 transition-all hover:bg-white/5 hover:text-white"
        >
          <span class="material-symbols-outlined text-[18px]">leaderboard</span>
          <span class="text-sm font-medium">User Ranking</span>
        </a>
        <a
          href="/artists"
          class="flex items-center gap-3 rounded-xl px-4 py-3 text-white/60 transition-all hover:bg-white/5 hover:text-white"
        >
          <span class="material-symbols-outlined text-[18px]">group</span>
          <span class="text-sm font-medium">Artists</span>
        </a>
        <a
          href="/themes"
          class="flex items-center gap-3 rounded-xl px-4 py-3 text-white/60 transition-all hover:bg-white/5 hover:text-white"
        >
          <span class="material-symbols-outlined text-[18px]">music_note</span>
          <span class="text-sm font-medium">Themes</span>
        </a>
        <a
          href="/studios"
          class="flex items-center gap-3 rounded-xl px-4 py-3 text-white/60 transition-all hover:bg-white/5 hover:text-white"
        >
          <span class="material-symbols-outlined text-[18px]">movie</span>
          <span class="text-sm font-medium">Studios</span>
        </a>
        <a
          href="/producers"
          class="flex items-center gap-3 rounded-xl px-4 py-3 text-white/60 transition-all hover:bg-white/5 hover:text-white"
        >
          <span class="material-symbols-outlined text-[18px]"
            >theater_comedy</span
          > <span class="text-sm font-medium">Producers</span>
        </a>
      </nav>
    </div>
  </div>
{/if}

<SearchModal bind:show={showSearchModal} />
<RequestModal
  show={showRequestModal}
  onClose={() => (showRequestModal = false)}
/>
