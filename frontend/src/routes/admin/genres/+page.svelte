<script lang="ts">
  import TaxonomyManager from "$lib/components/admin/TaxonomyManager.svelte";
  let { data } = $props();
</script>

<svelte:head>
  <title>Genres | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-white mb-1">
      Genres Management
    </h1>
    <p class="text-gray-400">
      Manage song and anime genres across the platform.
    </p>
  </div>
</div>

<div class="grid grid-cols-1 lg:grid-cols-1 gap-6">
  <div class="lg:col-span-1">
    <TaxonomyManager type="genres" initialItems={data.genres} />
  </div>
</div>
