<script lang="ts">
  import TaxonomyManager from "$lib/components/admin/TaxonomyManager.svelte";
  let { data } = $props();
</script>

<svelte:head>
  <title>Formats | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-1">
      Formats Management
    </h1>
    <p class="text-on-surface-variant/70">Manage release formats (TV, Movie, OVA, etc.)</p>
  </div>
</div>

<div class="grid grid-cols-1 lg:grid-cols-1 gap-6">
  <div class="lg:col-span-1">
    <TaxonomyManager type="formats" initialItems={data.formats} />
  </div>
</div>
