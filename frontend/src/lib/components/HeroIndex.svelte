<script>
  import { authState } from "$lib/state/auth.svelte";
  import BarChart3 from "lucide-svelte/icons/bar-chart-3";
  import Star from "lucide-svelte/icons/star";
  import ListTodo from "lucide-svelte/icons/list-todo";
  import ArrowRight from "lucide-svelte/icons/arrow-right";
  import ChevronsDown from "lucide-svelte/icons/chevrons-down";
</script>

{#if !authState.isAuthenticated && !authState.loading}
  <section
    class="relative flex h-screen items-center justify-center overflow-hidden"
  >
    <div class="absolute inset-0 z-0">
      <div
        class="from-background-dark via-background-dark/60 absolute inset-0 z-10 bg-linear-to-t to-transparent"
      ></div>
      <div
        class="from-background-dark/80 to-background-dark/80 absolute inset-0 z-10 bg-linear-to-r via-transparent"
      ></div>
      <img
        class="h-full w-full scale-105 object-cover"
        alt="Cinematic wide shot of anime cityscape at night"
        title="Cinematic wide shot of anime cityscape at night"
        loading="eager"
        fetchpriority="high"
        src="https://lh3.googleusercontent.com/aida-public/AB6AXuC4NKMsiVFSxKoMy0QJq1bdQInEoybDwBon5RstY0kyN604vExTI9pu8NcCmoEH1v6KITdZ4av_LtDghv2w-_XaKXQElk4leiUH9GVq1u0IyulYjmeCIqspuNtkNr8PCg2bZhyubZmFlVu3i7A2-Ug2FVNBcK4uF7HEtkwYOOHS-swOGUyTo14YgLkUPJsdetP6SpSqK0fw-RR74Zhd-zOxh-7r1ruGpFQDSftrvuYB_oXOcfl7wzCHPUY-joO8Cn9Nr-6SuEVYS6A"
      />
    </div>

    <div class="relative z-20 mx-auto max-w-4xl px-6 text-center">
      <h1
        class="mb-6 bg-linear-to-b from-white to-white/85 bg-clip-text text-5xl font-black leading-none tracking-tight text-transparent md:text-7xl"
      >
        Your Anime Soundtrack Journey Starts Here
      </h1>
      <div class="mb-10 flex flex-wrap justify-center gap-8 text-lg">
        <div class="flex items-center gap-2 text-white/90">
          <BarChart3 class="text-primary" size={20} />
          <span>Track favorites</span>
        </div>
        <div class="flex items-center gap-2 text-white/90">
          <Star class="text-primary fill-primary" size={20} />
          <span>Influence rankings</span>
        </div>
        <div class="flex items-center gap-2 text-white/90">
          <ListTodo class="text-primary" size={20} />
          <span>Curate playlists</span>
        </div>
      </div>
      <a
        href="/songs"
        class="group relative inline-flex items-center gap-3 rounded-xl bg-primary px-10 py-5 text-xl font-bold transition-all hover:scale-105 hover:bg-primary/90 active:scale-95 shadow-[0_0_20px_rgba(127,19,236,0.4)] text-white"
      >
        <span>Start Exploring</span>
          <ArrowRight class="transition-transform group-hover:translate-x-1" />

      </a>
    </div>
    <div
      class="absolute bottom-10 left-1/2 z-20 -translate-x-1/2 animate-bounce opacity-80 text-white"
    >
      <ChevronsDown size={32} />

    </div>
  </section>
{/if}
