<script>
  import { authState } from "$lib/state/auth.svelte";
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
        src="https://lh3.googleusercontent.com/aida-public/AB6AXuC4NKMsiVFSxKoMy0QJq1bdQInEoybDwBon5RstY0kyN604vExTI9pu8NcCmoEH1v6KITdZ4av_LtDghv2w-_XaKXQElk4leiUH9GVq1u0IyulYjmeCIqspuNtkNr8PCg2bZhyubZmFlVu3i7A2-Ug2FVNBcK4uF7HEtkwYOOHS-swOGUyTo14YgLkUPJsdetP6SpSqK0fw-RR74Zhd-zOxh-7r1ruGpFQDSftrvuYB_oXOcfl7wzCHPUY-joO8Cn9Nr-6SuEVYS6A"
      />
    </div>

    <div class="relative z-20 mx-auto max-w-4xl px-6 text-center">
      <h1
        class="mb-6 bg-linear-to-b from-white to-white/60 bg-clip-text text-5xl font-black leading-none tracking-tight text-transparent md:text-7xl"
      >
        Your Anime Soundtrack Journey Starts Here
      </h1>
      <div class="mb-10 flex flex-wrap justify-center gap-8 text-lg">
        <div class="flex items-center gap-2 text-white/90">
          <span class="material-symbols-outlined text-primary">bar_chart</span>
          <span>Track favorites</span>
        </div>
        <div class="flex items-center gap-2 text-white/90">
          <span class="material-symbols-outlined text-primary">star</span>
          <span>Influence rankings</span>
        </div>
        <div class="flex items-center gap-2 text-white/90">
          <span class="material-symbols-outlined text-primary">checklist</span>
          <span>Curate playlists</span>
        </div>
      </div>
      <a
        href="/songs"
        class="group relative inline-flex items-center gap-3 rounded-xl bg-primary px-10 py-5 text-xl font-bold transition-all hover:scale-105 hover:bg-primary/90 active:scale-95 shadow-[0_0_20px_rgba(127,19,236,0.4)] text-white"
      >
        <span>Start Exploring</span>
        <span
          class="material-symbols-outlined transition-transform group-hover:translate-x-1"
          >arrow_forward</span
        >
      </a>
    </div>
    <div
      class="absolute bottom-10 left-1/2 z-20 -translate-x-1/2 animate-bounce opacity-50 text-white"
    >
      <span class="material-symbols-outlined text-[32px]"
        >keyboard_double_arrow_down</span
      >
    </div>
  </section>
{/if}
