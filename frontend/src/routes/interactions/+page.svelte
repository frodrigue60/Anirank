<script lang="ts">
  import ActivityFeed from "$lib/components/ActivityFeed.svelte";
</script>

<svelte:head>
  <title>Recent Activity - Anirank</title>
</svelte:head>

<main class="max-w-[800px] mx-auto px-6 py-12">
  <div class="flex flex-col gap-8">
    <div class="flex flex-col gap-2">
      <h1
        class="text-3xl font-black text-white uppercase italic tracking-tighter"
      >
        Global <span class="text-primary italic">Interactions</span>
      </h1>
      <p class="text-white/40 text-sm">
        Stay updated with what the community is doing.
      </p>
    </div>

    <ActivityFeed />
  </div>
</main>
