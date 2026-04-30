<script lang="ts">
  import { page } from "$app/state";
  import Construction from "lucide-svelte/icons/construction";

  const titleMap: Record<string, string> = {
    "/settings/anime-list": "Anime List",
    "/settings/notifications": "Notifications",
    "/settings/import-lists": "Import Lists",
    "/settings/apps": "Apps",
    "/settings/developer": "Developer",
  };

  let title = $derived(titleMap[page.url.pathname] || "Settings");
</script>

<div
  class="flex flex-col items-center justify-center py-32 text-center animate-in fade-in zoom-in-95 duration-500"
>
  <div
    class="w-20 h-20 rounded-full bg-white/5 flex items-center justify-center mb-6 border border-white/5"
  >
    <Construction size={40} class="text-white/10" />
  </div>
  <h1 class="text-3xl font-black text-white tracking-tighter mb-2">
    {title}
  </h1>
  <p class="text-white/40 max-w-sm font-medium">
    This section is currently under development. Stay tuned for updates!
  </p>
  <div
    class="mt-8 px-4 py-1.5 rounded-full bg-primary/10 border border-primary/20 text-primary text-[10px] font-black uppercase tracking-widest animate-pulse"
  >
    Work In Progress
  </div>
</div>
