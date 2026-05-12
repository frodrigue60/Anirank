<script lang="ts">
  import { authState, logout } from "$lib/state/auth.svelte";
  import { notificationState } from "$lib/state/notifications.svelte";
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import SearchModal from "$lib/components/SearchModal.svelte";
  import RequestModal from "$lib/components/RequestModal.svelte";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";
  import ChevronDown from "lucide-svelte/icons/chevron-down";
  import Tv from "lucide-svelte/icons/tv";
  import Users from "lucide-svelte/icons/users";
  import Music from "lucide-svelte/icons/music";
  import Film from "lucide-svelte/icons/film";
  import Drama from "lucide-svelte/icons/drama";
  import ListMusic from "lucide-svelte/icons/list-music";
  import Trophy from "lucide-svelte/icons/trophy";
  import BarChart3 from "lucide-svelte/icons/bar-chart-3";
  import Search from "lucide-svelte/icons/search";
  import Bell from "lucide-svelte/icons/bell";
  import User from "lucide-svelte/icons/user";
  import LayoutDashboard from "lucide-svelte/icons/layout-dashboard";
  import UserRound from "lucide-svelte/icons/user-round";
  import ListTodo from "lucide-svelte/icons/list-todo";
  import Settings from "lucide-svelte/icons/settings";
  import MessageSquare from "lucide-svelte/icons/message-square";
  import LogOut from "lucide-svelte/icons/log-out";
  import LogIn from "lucide-svelte/icons/log-in";
  import UserPlus from "lucide-svelte/icons/user-plus";
  import Menu from "lucide-svelte/icons/menu";
  import X from "lucide-svelte/icons/x";
  import CalendarDays from "lucide-svelte/icons/calendar-days";

  let showUserDropdown = $state(false);
  let showDiscoverDropdown = $state(false);
  let showMobileMenu = $state(false);
  let showSearchModal = $state(false);
  let showRequestModal = $state(false);

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
  class="sticky top-0 z-50 bg-surface border-b border-outline-variant/10 w-full shadow-sm shadow-black/10"
>
  <header
    class="max-w-[1440px] mx-auto px-6 h-16 flex items-center justify-between gap-4 text-on-surface"
  >
    <div class="flex items-center gap-10">
      <a
        class="flex items-center gap-2 group"
        href="/"
        title="AniRank Home"
        aria-label="Go to AniRank Home"
      >
        <span class="text-xl font-bold tracking-tight text-on-surface"
          >AniRank</span
        >
      </a>
      <nav class="hidden md:flex items-center gap-6">
        <a
          class="text-sm font-medium transition-colors {page.url.pathname.includes(
            '/songs/seasonal',
          )
            ? 'text-primary'
            : 'text-on-surface-variant hover:text-on-surface'}"
          href="/songs/seasonal"
          title="Seasonal Themes"
          aria-label="Season - View seasonal anime themes">Season</a
        >
        <a
          class="text-sm font-medium transition-colors {page.url.pathname.includes(
            '/songs/ranking',
          )
            ? 'text-primary'
            : 'text-on-surface-variant hover:text-on-surface'}"
          href="/songs/ranking"
          title="Song Rankings"
          aria-label="Ranking - View top rated anime songs">Ranking</a
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
              : 'text-on-surface-variant hover:text-on-surface'}"
            title="Discover more content"
            aria-label="Toggle discover dropdown menu"
          >
            Discover
            <ChevronDown
              size={18}
              class="transition-transform {showDiscoverDropdown
                ? 'rotate-180'
                : ''}"
            />
          </button>

          {#if showDiscoverDropdown}
            <div
              class="absolute left-0 top-full mt-3 w-48 overflow-hidden rounded-md bg-surface-container py-1 shadow-md shadow-black ring-1 ring-outline-variant/5"
            >
              <a
                href="/animes"
                class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                title="Browse Anime Series"
                aria-label="Browse all anime series"
              >
                <Tv size={18} /> Animes
              </a>
              <a
                href="/artists"
                class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                title="Browse Artists"
              >
                <Users size={18} /> Artists
              </a>
              <a
                href="/songs"
                class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                title="Browse Themes and Songs"
                aria-label="Browse all songs and themes"
              >
                <Music size={18} /> Songs
              </a>
              <a
                href="/studios"
                class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                title="Browse Animation Studios"
              >
                <Film size={18} /> Studios
              </a>
              <a
                href="/producers"
                class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                title="Browse Music Producers"
              >
                <Drama size={18} /> Producers
              </a>
              <a
                href="/playlists"
                class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                title="Browse Playlists"
                aria-label="Browse community playlists"
              >
                <ListMusic size={18} /> Playlists
              </a>
              <a
                href="/tournaments"
                class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                title="View Tournaments"
              >
                <Trophy size={18} /> Tournaments
              </a>
              <a
                href="/users/ranking"
                class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                title="View User Rankings"
              >
                <BarChart3 size={18} /> User Ranking
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
        class="relative hidden sm:flex items-center group h-10 w-10 rounded-full bg-surface-low hover:bg-surface-container text-on-surface transition-colors justify-center"
        title="Search (Ctrl + K)"
        aria-label="Open search modal"
      >
        <Search size={20} />
      </button>

      <!-- Notifications (Hidden on mobile) -->
      {#if authState.isAuthenticated}
        <a
          href="/notifications"
          class="hidden sm:flex h-10 w-10 items-center justify-center rounded-full bg-surface-low hover:bg-surface-container text-on-surface transition-colors relative"
          title="Notifications"
          aria-label="View notifications"
        >
          <Bell size={20} />


          {#if notificationState.unreadCount > 0}
            <span
              class="absolute top-2.5 right-2.5 w-2 h-2 bg-primary rounded-full border border-surface"
            ></span>
          {/if}
        </a>
      {/if}

      <div class="h-8 w-px bg-outline-variant/20 hidden sm:block"></div>

      <!-- User Toggle -->
      <div class="relative">
        {#if authState.loading}
          <div
            class="h-9 w-9 animate-pulse rounded-full bg-surface-container"
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
              class="w-9 h-9 overflow-hidden rounded-full border-2 border-transparent group-hover:border-primary transition-all bg-surface-low flex items-center justify-center text-on-surface-variant group-hover:text-primary"
            >
              {#if authState.isAuthenticated && authState.user}
                <OptimizedImage
                  src={authState.user.avatar_url}
                  sources={authState.user.avatar_sources}
                  alt={authState.user.name}
                  class="h-full w-full object-cover"
                  sizes="36px"
                />
              {:else}
                <User size={20} />
              {/if}
            </div>
            <ChevronDown size={20} class="text-on-surface-variant group-hover:text-on-surface hidden sm:block" />

          </button>

          {#if showUserDropdown}
            <div
              class="absolute right-0 top-full mt-3 w-56 overflow-hidden rounded-md bg-surface-container py-1 shadow-md shadow-black ring-1 ring-outline-variant/5"
            >
              {#if authState.isAuthenticated && authState.user}
                <div class="border-b border-outline-variant/10 px-4 py-3">
                  <p class="truncate text-sm font-bold text-on-surface">
                    {authState.user.name}
                  </p>
                </div>
                <div class="py-1">
                  {#if authState.isStaff}
                    <a
                      href="/admin"
                      class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                    >
                      <LayoutDashboard size={18} /> Dashboard
                    </a>
                  {/if}
                  <a
                    href="/users/{authState.user.slug}"
                    class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                  >
                    <UserRound size={18} /> My Profile
                  </a>
                  <a
                    href="/users/{authState.user.slug}/playlists"
                    class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                  >
                    <ListTodo size={18} /> My Playlists
                  </a>
                  <a
                    href="/settings"
                    class="flex items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                  >
                    <Settings size={18} /> Settings
                  </a>
                  <button
                    onclick={() => {
                      showRequestModal = true;
                      showUserDropdown = false;
                    }}
                    class="flex w-full items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                  >
                    <MessageSquare size={18} /> Request Song
                  </button>
                </div>
                <div class="border-t border-outline-variant/10 py-1">
                  <button
                    onclick={handleLogout}
                    class="flex w-full items-center gap-3 px-4 py-2 text-sm text-red-500 transition-colors hover:bg-red-500/10"
                    title="Sign Out"
                  >
                    <LogOut size={18} /> Logout
                  </button>
                </div>
              {:else}
                <div class="border-b border-outline-variant/10 px-4 py-3">
                  <p class="text-sm font-bold text-on-surface">Guest User</p>
                  <p class="text-xs text-on-surface-variant">
                    Join the community!
                  </p>
                </div>
                <div class="py-1">
                  <a
                    href="/login"
                    class="flex w-full items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                  >
                    <LogIn size={18} />

                    Sign In
                  </a>
                  <a
                    href="/register"
                    class="flex w-full items-center gap-3 px-4 py-2 text-sm text-on-surface-variant transition-colors hover:bg-surface-low hover:text-primary"
                  >
                    <UserPlus size={18} />

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
        class="flex h-10 w-10 items-center justify-center text-on-surface-variant transition-colors hover:text-on-surface md:hidden"
        title="Open Mobile Menu"
        aria-label="Open mobile menu"
      >
        <Menu size={24} />
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
      class="absolute inset-0 bg-black/40 backdrop-blur-sm"
    ></div>

    <div
      class="absolute right-0 top-0 flex h-full w-[280px] flex-col bg-surface shadow-2xl"
    >
      <div
        class="flex h-16 items-center justify-between border-b border-outline-variant/10 px-6"
      >
        <span class="text-lg font-bold tracking-tight text-on-surface"
          >Menu</span
        >
        <button
          onclick={() => (showMobileMenu = false)}
          class="text-on-surface-variant transition-colors hover:text-on-surface flex items-center justify-center"
        >
          <X />
        </button>
      </div>

      <nav class="flex-1 space-y-2 overflow-y-auto px-4 py-6">
        <a
          href="/songs/seasonal"
          class="flex items-center gap-3 rounded-sm px-4 py-3 text-on-surface-variant transition-all hover:bg-surface-low hover:text-on-surface"
        >
          <CalendarDays size={18} /> <span class="text-sm font-medium">Seasonal</span>
        </a>
        <a
          href="/songs/ranking"
          class="flex items-center gap-3 rounded-sm px-4 py-3 text-on-surface-variant transition-all hover:bg-surface-low hover:text-on-surface"
        >
          <BarChart3 size={18} />
          <span class="text-sm font-medium">Ranking</span>
        </a>
        <a
          href="/animes"
          class="flex items-center gap-3 rounded-sm px-4 py-3 text-on-surface-variant transition-all hover:bg-surface-low hover:text-on-surface"
        >
          <Tv size={18} />
          <span class="text-sm font-medium">Animes</span>
        </a>
        <a
          href="/users/ranking"
          class="flex items-center gap-3 rounded-sm px-4 py-3 text-on-surface-variant transition-all hover:bg-surface-low hover:text-on-surface"
        >
          <BarChart3 size={18} />
          <span class="text-sm font-medium">User Ranking</span>
        </a>
        <a
          href="/artists"
          class="flex items-center gap-3 rounded-sm px-4 py-3 text-on-surface-variant transition-all hover:bg-surface-low hover:text-on-surface"
        >
          <Users size={18} />
          <span class="text-sm font-medium">Artists</span>
        </a>
        <a
          href="/themes"
          class="flex items-center gap-3 rounded-sm px-4 py-3 text-on-surface-variant transition-all hover:bg-surface-low hover:text-on-surface"
        >
          <Music size={18} />
          <span class="text-sm font-medium">Themes</span>
        </a>
        <a
          href="/studios"
          class="flex items-center gap-3 rounded-sm px-4 py-3 text-on-surface-variant transition-all hover:bg-surface-low hover:text-on-surface"
        >
          <Film size={18} />
          <span class="text-sm font-medium">Studios</span>
        </a>
        <a
          href="/producers"
          class="flex items-center gap-3 rounded-sm px-4 py-3 text-on-surface-variant transition-all hover:bg-surface-low hover:text-on-surface"
        >
          <Drama size={18} />
 <span class="text-sm font-medium">Producers</span>
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
