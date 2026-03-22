<script lang="ts">
  import { authState } from "$lib/state/auth.svelte";
  import {
    User,
    Settings,
    Tv,
    Bell,
    Download,
    Smartphone,
    Code,
    Check,
  } from "lucide-svelte";
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";

  let { children } = $props();

  const sections = [
    {
      title: "Settings",
      items: [
        { id: "profile", label: "Profile", icon: User, path: "/settings/profile" },
        { id: "account", label: "Account", icon: Settings, path: "/settings/account" },
        { id: "anime-list", label: "Anime List", icon: Tv, path: "/settings/anime-list" },
        { id: "notifications", label: "Notifications", icon: Bell, path: "/settings/notifications" },
        { id: "import-lists", label: "Import Lists", icon: Download, path: "/settings/import-lists" },
      ],
    },
    {
      title: "Apps",
      items: [
        { id: "apps", label: "Apps", icon: Smartphone, path: "/settings/apps" },
        { id: "developer", label: "Developer", icon: Code, path: "/settings/developer" },
      ],
    },
  ];

  onMount(() => {
    if (!authState.isAuthenticated && !authState.loading) {
      goto("/");
    }
  });

  let activePath = $derived(page.url.pathname);
</script>

<div class="min-h-screen bg-background-dark pt-12 pb-24 text-white">
  <div class="max-w-6xl mx-auto px-6 flex flex-col md:flex-row gap-12">
    <!-- Sidebar -->
    <aside class="w-full md:w-64 shrink-0">
      <div class="space-y-8 sticky top-12">
        {#each sections as section}
          <div>
            <h3
              class="text-[11px] font-black uppercase tracking-[0.2em] text-white/20 mb-4 px-4"
            >
              {section.title}
            </h3>
            <nav class="space-y-1">
              {#each section.items as item}
                <a
                  href={item.path}
                  class="w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 group {activePath ===
                  item.path
                    ? 'bg-primary/5 text-primary border border-primary/10 shadow-[0_0_20px_rgba(127,19,236,0.1)]'
                    : 'text-white/40 hover:text-white/70 hover:bg-white/2'}"
                >
                  <item.icon
                    size={18}
                    class={activePath === item.path
                      ? "text-primary"
                      : "text-white/10 group-hover:text-white/30"}
                  />
                  <span class="text-sm font-bold tracking-tight"
                    >{item.label}</span
                  >
                </a>
              {/each}
            </nav>
          </div>
        {/each}
      </div>
    </aside>

    <!-- Content Area -->
    <main class="flex-1 min-w-0">
      {@render children()}
    </main>
  </div>
</div>
