<script lang="ts">
  import ActivityFeed from "$lib/components/ActivityFeed.svelte";
</script>

<svelte:head>
  <title>Recent Activity - Anirank</title>
</svelte:head>

<main class="max-w-[800px] mx-auto px-6 py-16">
  <div class="flex flex-col gap-10">
    <div class="flex flex-col gap-2">
      <h1
        class="text-4xl font-black text-on-surface uppercase tracking-tighter leading-none"
      >
        Global <span class="text-primary">Interactions</span>
      </h1>
      <p class="text-on-surface-variant/60 text-sm font-medium">
        Stay updated with what the community is doing in real-time.
      </p>
    </div>

    <ActivityFeed />
  </div>
</main>
