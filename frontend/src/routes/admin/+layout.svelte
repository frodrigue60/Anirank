<script lang="ts">
  import { authState } from "$lib/state/auth.svelte";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";
  import { goto } from "$app/navigation";
  import Menu from "lucide-svelte/icons/menu";
  import LayoutDashboard from "lucide-svelte/icons/layout-dashboard";
  import Film from "lucide-svelte/icons/film";
  import Zap from "lucide-svelte/icons/zap";
  import Music from "lucide-svelte/icons/music";
  import User from "lucide-svelte/icons/user";
  import Megaphone from "lucide-svelte/icons/megaphone";
  import Trophy from "lucide-svelte/icons/trophy";
  import AlertTriangle from "lucide-svelte/icons/alert-triangle";
  import MessageSquare from "lucide-svelte/icons/message-square";
  import Flag from "lucide-svelte/icons/flag";
  import Mail from "lucide-svelte/icons/mail";
  import History from "lucide-svelte/icons/history";
  import ShieldCheck from "lucide-svelte/icons/shield-check";
  import Award from "lucide-svelte/icons/award";
  import Tag from "lucide-svelte/icons/tag";
  import Tv from "lucide-svelte/icons/tv";
  import Sun from "lucide-svelte/icons/sun";
  import Calendar from "lucide-svelte/icons/calendar";
  import LogOut from "lucide-svelte/icons/log-out";
  import Webhook from "lucide-svelte/icons/webhook";
  import Globe from "lucide-svelte/icons/globe";

  let { children } = $props();

  let isSidebarOpen = $state(true);

  function toggleSidebar() {
    isSidebarOpen = !isSidebarOpen;
  }

  $effect(() => {
    if (!authState.loading) {
      if (!authState.isAuthenticated) {
        goto("/login");
      } else if (!authState.isStaff) {
        goto("/"); // Redirect non-staff authenticated users to homepage
      }
    }
  });
</script>

<div class="flex h-screen bg-surface text-on-surface overflow-hidden font-sans">
  <!-- Sidebar -->
  <aside
    class="bg-surface-container border-r border-gray-500 shrink-0 transition-all duration-300 ease-in-out z-20"
    class:w-64={isSidebarOpen}
    class:w-20={!isSidebarOpen}
  >
    <div
      class="h-16 flex items-center justify-between px-4 border-b border-gray-500"
    >
      {#if isSidebarOpen}
        <a href="/admin" class="text-xl font-bold"> Admin Panel </a>
      {/if}
      <button
        onclick={toggleSidebar}
        class="p-2 hover:bg-surface-highest rounded-lg text-on-surface"
      >
        <Menu size={20} />
      </button>
    </div>

    <nav class="p-4 space-y-2 overflow-y-auto h-[calc(100vh-4rem)]">
      <ul class="space-y-1">
        <li>
          <a
            href="/admin"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <LayoutDashboard size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Dashboard</span
              >{/if}
          </a>
        </li>
        <!-- Content Group -->
        <li class="pt-4 pb-2">
          {#if isSidebarOpen}<span
              class="px-3 text-xs font-semibold text-on-surface-variant/40 uppercase tracking-wider"
              >Catalog</span
            >{/if}
        </li>
        <!-- animes -->
        <li>
          <a
            href="/admin/animes"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Film size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Animes</span>{/if}
          </a>
        </li>
        <!-- animes autogen -->
        <li>
          <a
            href="/admin/animes-autogen"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Zap size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Autogen</span
              >{/if}
          </a>
        </li>
        <li>
          <!-- songs -->
          <a
            href="/admin/songs"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Music size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Songs</span>{/if}
          </a>
        </li>
        <li>
          <!-- artists -->
          <a
            href="/admin/artists"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <User size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Artists</span
              >{/if}
          </a>
        </li>

        <!-- Events & Comm -->
        <li class="pt-4 pb-2">
          {#if isSidebarOpen}<span
              class="px-3 text-xs font-semibold text-on-surface-variant/40 uppercase tracking-wider"
              >Events</span
            >{/if}
        </li>
        <li>
          <!-- announcements -->
          <a
            href="/admin/announcements"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Megaphone size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium"
                >Announcements</span
              >{/if}
          </a>
        </li>

        <li>
          <!-- tournaments -->
          <a
            href="/admin/tournaments"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Trophy size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Tournaments</span
              >{/if}
          </a>
        </li>

        <li>
          <!-- webhooks -->
          <a
            href="/admin/webhooks"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Webhook size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Webhooks</span>{/if}
          </a>
        </li>

        <li>
          <!-- partners -->
          <a
            href="/admin/partners"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Globe size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Partners</span>{/if}
          </a>
        </li>

        <li class="pt-4 pb-2">
          {#if isSidebarOpen}<span
              class="px-3 text-xs font-semibold text-on-surface-variant/40 uppercase tracking-wider"
              >Moderation</span
            >{/if}
        </li>

        <li>
          <!-- songs reports -->
          <a
            href="/admin/reports/songs"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <AlertTriangle size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium"
                >Songs Reports</span
              >{/if}
          </a>
        </li>
        <li>
          <!-- comments reports -->
          <a
            href="/admin/reports/comments"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <MessageSquare size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium"
                >Comment Reports</span
              >{/if}
          </a>
        </li>
        <li>
          <!-- user reports -->
          <a
            href="/admin/reports/users"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Flag size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">User Reports</span
              >{/if}
          </a>
        </li>
        <li>
          <!-- requests -->
          <a
            href="/admin/requests"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Mail size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium"
                >User Requests</span
              >{/if}
          </a>
        </li>
        <!-- Platform -->
        <li class="pt-4 pb-2">
          {#if isSidebarOpen}<span
              class="px-3 text-xs font-semibold text-on-surface-variant/40 uppercase tracking-wider"
              >Platform</span
            >{/if}
        </li>
        <li>
          <!-- users -->
          <a
            href="/admin/users"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <User size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Users</span>{/if}
          </a>
        </li>
        <li>
          <!-- audit logs -->
          <a
            href="/admin/audit-logs"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <History size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Audit Logs</span
              >{/if}
          </a>
        </li>
        <li>
          <!-- roles -->
          <a
            href="/admin/roles"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <ShieldCheck size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Roles</span>{/if}
          </a>
        </li>
        <li>
          <!-- badges -->
          <a
            href="/admin/badges"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Award size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Badges</span>{/if}
          </a>
        </li>

        <!-- Metadata -->
        <li class="pt-4 pb-2">
          {#if isSidebarOpen}<span
              class="px-3 text-xs font-semibold text-on-surface-variant/40 uppercase tracking-wider"
              >Metadata</span
            >{/if}
        </li>

        <li>
          <!-- genres -->
          <a
            href="/admin/genres"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Tag size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Genres</span>{/if}
          </a>
        </li>

        <li>
          <!-- formats -->
          <a
            href="/admin/formats"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Tv size={20} />

            {#if isSidebarOpen}<span class="ml-3 font-medium"
                >Anime Formats</span
              >{/if}
          </a>
        </li>
        <li>
          <!-- seasons -->
          <a
            href="/admin/seasons"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Sun size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Seasons</span
              >{/if}
          </a>
        </li>
        <li>
          <!-- years -->
          <a
            href="/admin/years"
            class="flex items-center p-3 rounded-xl hover:bg-primary-container hover:text-primary transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <Calendar size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Years</span>{/if}
          </a>
        </li>

        <li>
          <!-- home -->
          <a
            href="/"
            class="flex items-center p-3 rounded-xl hover:bg-surface-highest text-on-surface-variant/70 hover:text-on-surface mt-8 transition-colors {isSidebarOpen
              ? ''
              : 'justify-center'}"
          >
            <LogOut size={20} />
            {#if isSidebarOpen}<span class="ml-3 font-medium">Back to Site</span
              >{/if}
          </a>
        </li>
      </ul>
    </nav>
  </aside>

  <!-- Main Content Area -->
  <div class="flex-1 flex flex-col min-w-0 bg-surface">
    <!-- Top Navbar -->
    <header
      class="h-16 bg-surface-container border-b border-gray-500 flex items-center justify-between px-6 z-10"
    >
      <div class="flex items-center gap-4">
        <!-- Contextual Breadcrumbs could go here -->
      </div>

      <div class="flex items-center gap-4">
        <div class="flex items-center gap-3 pl-4 border-l border-gray-500">
          <OptimizedImage
            src={authState.user?.avatar_url}
            sources={authState.user?.avatar_sources}
            alt={authState.user?.name || "Admin"}
            class="w-8 h-8 rounded-full object-cover ring-2 ring-white/10"
            sizes="32px"
          />
          <div class="hidden md:block">
            <p class="text-sm font-medium text-on-surface">
              {authState.user?.name}
            </p>
            <p class="text-xs text-on-surface-variant capitalize">
              {authState.isAdmin ? "Administrator" : "Staff Member"}
            </p>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Render Area -->
    <main class="flex-1 overflow-y-auto p-4 md:p-6">
      <div class="max-w-7xl mx-auto space-y-6">
        {@render children()}
      </div>
    </main>
  </div>
</div>
