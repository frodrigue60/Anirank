<script lang="ts">
  import type { PageData } from "./$types";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import AnimeCard from "$lib/components/AnimeCard.svelte";
  import CustomSelect from "$lib/components/CustomSelect.svelte";
  import { Search, SortDesc } from "lucide-svelte";
  import SEO from "$lib/components/SEO.svelte";

  let { data }: { data: PageData } = $props();
  let studio = $derived(data.studio);
  let animes = $derived(data.animes);

  let searchTerm = $state("");
  let selectedSort = $state("");

  $effect(() => {
    searchTerm = data.params?.name || "";
    selectedSort = data.params?.sort || "";
  });

  function updateFilters() {
    const url = new URL(page.url);

    const setParam = (key: string, val: string) => {
      if (val && val !== "any") url.searchParams.set(key, val);
      else url.searchParams.delete(key);
    };

    if (searchTerm) url.searchParams.set("name", searchTerm);
    else url.searchParams.delete("name");

    setParam("sort", selectedSort);

    url.searchParams.set("page", "1");
    goto(url.toString(), { keepFocus: true, noScroll: true });
  }

  let searchTimeout: any;
  function handleInput() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      if (searchTerm.length === 0 || searchTerm.length >= 2) {
        updateFilters();
      }
    }, 300);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") {
      updateFilters();
    }
  }

  const sortOptions = [
    { value: "any", label: "Alphabetical (A-Z)" },
    { value: "name_desc", label: "Alphabetical (Z-A)" },
    { value: "most_themes", label: "Most Themes" },
    { value: "least_themes", label: "Least Themes" },
  ];

  let queryString = $derived(
    new URLSearchParams(
      Object.fromEntries(
        Object.entries(data.params || {}).filter(([k, v]) => k !== "page" && v),
      ) as Record<string, string>,
    ).toString(),
  );
</script>

<SEO 
  title="{studio.name} Anime Production" 
  description="Explore the catalog of anime produced by {studio.name} and discover their theme songs on AniRank." 
  type="profile"
/>

<main class="w-full max-w-[1440px] mx-auto px-6 py-8">
  <div class="mb-10 text-center">
    <h1 class="text-4xl md:text-5xl font-black text-white mb-4">
      {studio.name}
    </h1>
    <p class="text-white/60">Animes produced by {studio.name}</p>
  </div>

  <!-- Filter Bar -->
  <section
    class="relative z-40 bg-surface-dark/30 p-4 rounded-3xl border border-white/5 backdrop-blur-md shadow-2xl mb-12 max-w-4xl mx-auto"
  >
    <div class="flex flex-col gap-6">
      <div class="grid grid-cols-1 md:grid-cols-12 gap-4 items-end">
        <!-- Search -->
        <div class="md:col-span-8 relative group">
          <label
            for="anime-search"
            class="block text-[10px] uppercase font-black text-white/40 mb-2 ml-1 tracking-widest"
          >
            Search Anime
          </label>
          <div class="relative">
            <span
              class="absolute left-4 top-1/2 -translate-y-1/2 text-white/20 group-focus-within:text-primary transition-colors"
            >
              <Search size={20} />
            </span>
            <input
              id="anime-search"
              bind:value={searchTerm}
              oninput={handleInput}
              onkeydown={handleKeydown}
              class="w-full h-12 bg-surface-darker/50 border border-white/10 rounded-xl pl-12 pr-6 text-sm text-white focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-white/20 transition-all"
              placeholder="Search animes by title..."
              type="text"
            />
          </div>
        </div>

        <!-- Sort -->
        <div class="md:col-span-4">
          <CustomSelect
            label="Sort"
            bind:value={selectedSort}
            options={sortOptions}
            placeholder="Any"
            icon={SortDesc}
            onchange={updateFilters}
          />
        </div>
      </div>
    </div>
  </section>

  {#if animes.data && animes.data.length > 0}
    <div
      class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-6"
    >
      {#each animes.data as anime}
        <AnimeCard {anime} />
      {/each}
    </div>

    <div class="mt-12 flex justify-center gap-2">
      {#if animes.prev_page_url}
        <a
          href="?page={animes.current_page - 1}{queryString
            ? `&${queryString}`
            : ''}"
          class="px-4 py-2 bg-surface-dark text-white rounded-lg hover:bg-primary transition-colors font-bold text-sm border border-white/5"
          >Previous</a
        >
      {/if}
      {#if animes.next_page_url}
        <a
          href="?page={animes.current_page + 1}{queryString
            ? `&${queryString}`
            : ''}"
          class="px-4 py-2 bg-surface-dark text-white rounded-lg hover:bg-primary transition-colors font-bold text-sm border border-white/5"
          >Next</a
        >
      {/if}
    </div>
  {:else}
    <div
      class="text-center py-20 bg-surface-dark rounded-2xl border border-white/5"
    >
      <span class="material-symbols-outlined text-6xl text-white/10 mb-4 block"
        >movie</span
      >
      <h3 class="text-xl font-bold text-white/50 mb-2">No animes found</h3>
      <p class="text-white/30 text-sm">
        There are no series associated with this studio.
      </p>
    </div>
  {/if}
</main>
