<script lang="ts">
  import TaxonomyManager from "$lib/components/admin/TaxonomyManager.svelte";
  let { data } = $props();
</script>

<svelte:head>
  <title>Seasons | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-1">
      Seasons Management
    </h1>
    <p class="text-on-surface-variant/70">
      Manage anime seasons (Winter, Spring, Summer, Fall).
    </p>
  </div>
</div>

<div class="grid grid-cols-1 lg:grid-cols-1 gap-6">
  <div class="lg:col-span-1">
    <TaxonomyManager type="seasons" initialItems={data.seasons} />
  </div>
</div>
