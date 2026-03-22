<script lang="ts">
  let { anime } = $props();
</script>

<div class="group flex flex-col gap-3 cursor-pointer">
  <a
    href="/animes/{anime.slug}"
    class="relative aspect-2/3 rounded-2xl overflow-hidden card-shadow group block"
    title="View details for {anime.title}"
  >
    <img
      alt={anime.title}
      class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
      src={anime.cover_url ||
        "https://placehold.co/400x600/2a2136/white?text=No+Image"}
    />
    <div
      class="absolute inset-0 bg-linear-to-t from-black/80 via-black/20 to-transparent opacity-100 transition-opacity"
    ></div>
    <div
      class="absolute bottom-3 left-3 right-3 flex flex-wrap gap-2 pointer-events-none"
    >
      {#if anime.format?.name}
        <span
          class="bg-black/60 backdrop-blur-md text-white px-2 py-1 rounded-md text-[10px] font-bold shadow-sm border border-white/10"
        >
          {anime.format.name}
        </span>
      {/if}
      <span
        class="bg-primary/40 backdrop-blur-md text-white px-2 py-1 rounded-md text-[10px] font-bold shadow-sm flex items-center gap-1 border border-primary/20"
      >
        {anime.songs_count || 0} Themes
      </span>
    </div>
  </a>
  <div>
    <h3
      class="font-bold text-white group-hover:text-primary transition-colors line-clamp-1"
    >
      {anime.title}
    </h3>
    <p class="text-xs text-white/40">
      {anime.season?.name || "Season"}
      {anime.year?.name || "Year"}
    </p>
  </div>
</div>
