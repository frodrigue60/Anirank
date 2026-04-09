<script lang="ts">
  import { getFormattedScore } from "$lib/song-utils";
  import { authState } from "$lib/state/auth.svelte";
  import { Plus, Check, Play } from "lucide-svelte";

  let { anime, view = "grid" } = $props<{
    anime: any;
    view?: "grid" | "card" | "list";
  }>();

  function formatScore(score: number | string | null | undefined) {
    return getFormattedScore(score as any, authState.user?.score_format);
  }

  const statusColors = {
    FINISHED: "bg-green-500",
    RELEASING: "bg-blue-500",
    NOT_YET_RELEASED: "bg-gray-500",
    CANCELLED: "bg-red-500",
    HIATUS: "bg-yellow-500",
  };
</script>

{#if view === "grid"}
  <!-- --- GRID VIEW (Poster Minimal) --- -->
  <div class="group flex flex-col gap-3 cursor-pointer">
    <a
      href="/animes/{anime.slug}"
      class="relative aspect-2/3 rounded-md shadow-sm overflow-hidden card-shadow group block"
      title="View details for {anime.title}"
    >
      <img
        alt={anime.title}
        title={anime.title}
        class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
        src={anime.cover_url ?? "/images/placeholders/default.jpg"}
      />
      <div
        class="absolute inset-0 bg-linear-to-t from-black/80 via-black/20 to-transparent opacity-100 transition-opacity"
      ></div>
      <div
        class="absolute bottom-3 left-3 right-3 flex flex-wrap gap-2 pointer-events-none"
      >
        {#if anime.format?.name}
          <span
            class="bg-black/60 text-white px-2 py-1 rounded-sm text-[10px] font-bold shadow-sm border border-white/10"
          >
            {anime.format.name}
          </span>
        {/if}
        <span
          class="bg-primary/40 text-white px-2 py-1 rounded-sm text-[10px] font-bold shadow-sm flex items-center gap-1 border border-primary/20"
        >
          {anime.songs_count || 0} Themes
        </span>
      </div>
      <!-- Status Indicator -->
      {#if anime.status}
        <div class="absolute top-3 left-3">
          <div
            class="w-3 h-3 rounded-sm border-2 border-surface-darker {statusColors[
              anime.status as keyof typeof statusColors
            ] || 'bg-gray-400'}"
          ></div>
        </div>
      {/if}
    </a>
    <div>
      <h3
        class="font-bold text-on-surface group-hover:text-primary transition-colors line-clamp-1"
      >
        {anime.title}
      </h3>
      <p class="text-xs text-on-surface-variant">
        {anime.season?.name || "Season"}
        {anime.year?.name || "Year"}
      </p>
    </div>
  </div>
{:else if view === "card"}
  <!-- --- CARD VIEW (Detailed Card) --- -->
  <a
    href="/animes/{anime.slug}"
    class="group block bg-surface-container shadow-sm hover:bg-surface-highest rounded-md overflow-hidden transition-all duration-300 h-full"
  >
    <div class="flex flex-row h-full">
      <!-- Media Section -->
      <div class="relative w-28 sm:w-40 aspect-2/3 shrink-0 overflow-hidden">
        <img
          src={anime.cover_url ?? "/images/placeholders/default.jpg"}
          alt={anime.title}
          class="w-full h-full object-cover transition-transform duration-700 group-hover:scale-110"
        />
        <div
          class="absolute inset-0 bg-linear-to-t from-black/90 via-black/20 to-transparent"
        ></div>
        <div class="absolute bottom-0 left-0 right-0 p-2">
          <h4
            class="text-[10px] sm:text-xs font-bold text-white/80 transition-colors uppercase tracking-wider mb-1 line-clamp-1"
          >
            {anime.studios?.[0]?.name || "Generic Studio"}
          </h4>
        </div>
      </div>

      <!-- Info Section -->
      <div class="flex-1 p-3 sm:p-6 flex flex-col justify-between min-w-0">
        <div class="space-y-4">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2 mb-1 font-bold">
                <span
                  class="text-[10px] sm:text-xs text-on-surface-variant uppercase tracking-widest"
                >
                  {anime.season?.name || ""}
                  {anime.year?.name || ""}
                </span>
                {#if anime.format?.name}
                  <span class="text-[10px] text-on-surface-variant">•</span>
                  <span class="text-[10px] text-on-surface-variant uppercase">
                    {anime.format.name}
                  </span>
                {/if}
                <span
                  class="text-[10px] sm:text-xs text-on-surface-variant flex items-center gap-1.5"
                >
                  <span class="text-[10px] text-on-surface-variant">•</span>
                  {anime.songs_count || 0} Themes
                </span>
              </div>
              <h3
                class="text-base sm:text-xl font-bold text-primary group-hover:text-primary-container transition-all line-clamp-2 leading-tight"
              >
                {anime.title}
              </h3>
            </div>

            {#if anime.average_score || anime.rating}
              <div class="flex flex-col items-end shrink-0">
                <div class="flex items-center gap-1.5 text-primary">
                  <span
                    class="material-symbols-outlined text-xs sm:text-sm filled"
                    >sentiment_satisfied</span
                  >
                  <span class="font-bold text-sm sm:text-lg"
                    >{formatScore(anime.average_score || anime.rating)}%</span
                  >
                </div>
                <span
                  class="text-[8px] sm:text-[10px] text-white/30 uppercase font-black tracking-tighter"
                >
                  {anime.users_count?.toLocaleString() || 0} users
                </span>
              </div>
            {/if}
          </div>

          {#if anime.description}
            <p
              class="text-on-surface-variant text-[10px] sm:text-sm line-clamp-2 sm:line-clamp-4 leading-relaxed font-medium"
            >
              {@html anime.description}
            </p>
          {:else}
            <p
              class="text-on-surface-variant text-[10px] sm:text-sm line-clamp-2 sm:line-clamp-4 leading-relaxed font-medium"
            >
              No description available
            </p>
          {/if}

          <div class="flex flex-wrap gap-2"></div>
        </div>

        <!-- Card Footer -->
        <div
          class="mt-auto flex items-center justify-between pt-3 sm:pt-4 border-t border-white/5"
        >
          <div class="flex flex-wrap gap-1 sm:gap-2">
            {#each anime.genres?.slice(0, 3) || [] as genre}
              <span
                class="px-2 py-0.5 rounded-sm bg-surface-highest/60 border border-on-surface-variant/5 text-on-surface-variant text-[8px] sm:text-[10px] font-bold uppercase tracking-wider"
              >
                {genre.name}
              </span>
            {/each}
          </div>
        </div>
      </div>
    </div>
  </a>
{:else}
  <!-- --- LIST VIEW (Horizontal Row) --- -->
  <a
    href="/animes/{anime.slug}"
    class="group block bg-surface-container hover:bg-surface-highest rounded-md shadow-sm overflow-hidden transition-all duration-300"
  >
    <div class="flex items-center h-20 sm:h-24 pr-4 sm:pr-8">
      <!-- Media Section -->
      <div class="h-full aspect-2/3 shrink-0 overflow-hidden">
        <img
          src={anime.cover_url ?? "/images/placeholders/default.jpg"}
          alt={anime.title}
          class="w-full h-full object-cover transition-transform duration-700 group-hover:scale-110"
        />
      </div>

      <!-- Content Row -->
      <div
        class="flex-1 flex items-center justify-between min-w-0 pl-4 sm:pl-6 gap-6"
      >
        <!-- Title & Genres -->
        <div class="flex-1 min-w-0 space-y-2">
          <h3
            class="text-base sm:text-lg font-bold text-on-surface group-hover:text-primary transition-colors truncate"
          >
            {anime.title}
          </h3>
          <div class="flex items-center gap-1.5 flex-wrap px-1">
            {#if anime.studios?.length > 0}
              {#each anime.studios.slice(0, 3) as studio, i}
                <span
                  class="text-[10px] italic text-on-surface-variant font-medium uppercase tracking-wider"
                >
                  {studio.name}
                </span>
                {#if i < Math.min(anime.studios.length, 3) - 1}
                  <span class="text-[10px] text-on-surface-variant">•</span>
                {/if}
              {/each}
            {/if}
          </div>
          <div class="flex items-center gap-2 flex-wrap mt-1 mb-1">
            {#each anime.genres?.slice(0, 3) || [] as genre}
              <span
                class="px-2 py-0.5 rounded-sm bg-surface-highest/60 text-[10px] font-bold text-on-surface-variant uppercase tracking-wider border border-on-surface-variant/5"
              >
                {genre.name}
              </span>
            {/each}
          </div>
        </div>

        <!-- Score Section (Image 1 central) -->
        <div
          class="hidden md:flex flex-col items-center shrink-0 w-32 border-x border-white/5 px-4 h-full py-4"
        >
          {#if anime.average_score || anime.rating}
            <div class="flex items-center gap-2 text-primary">
              <span class="material-symbols-outlined text-base filled"
                >sentiment_satisfied</span
              >
              <span class="font-bold text-lg"
                >{formatScore(anime.average_score || anime.rating)}%</span
              >
            </div>
            <span
              class="text-[9px] text-white/30 uppercase font-bold tracking-tight"
            >
              {anime.users_count?.toLocaleString() || 0} users
            </span>
          {/if}
          <span
            class="text-[10px] font-medium sm:text-xs text-on-surface-variant flex items-center gap-1.5"
          >
            {anime.songs_count || 0} Themes
          </span>
        </div>

        <!-- Format & Season (Image 1 Right) -->
        <div
          class="flex flex-col items-end shrink-0 gap-0.5 text-right w-32 sm:w-48"
        >
          <div class="flex items-center gap-2">
            {#if anime.format?.name}
              <span class="text-xs font-bold text-on-surface-variant uppercase"
                >{anime.format.name}</span
              >
              {#if anime.episodes}
                <span class="text-[10px] text-on-surface-variant">•</span>
                <span class="text-xs text-on-surface-variant"
                  >{anime.episodes} episodes</span
                >
              {/if}
            {/if}
          </div>
          <div class="flex items-center gap-2">
            <span
              class="text-xs font-black text-on-surface-variant uppercase tracking-wider"
            >
              {anime.season?.name || ""}
              {anime.year?.name || ""}
            </span>
            {#if anime.status === "FINISHED"}
              <span
                class="w-1.5 h-1.5 rounded-sm bg-green-500 shadow-sm shadow-green-500/50"
              ></span>
            {:else if anime.status === "RELEASING"}
              <span
                class="w-1.5 h-1.5 rounded-sm bg-blue-500 shadow-sm shadow-blue-500/50"
              ></span>
            {/if}
          </div>
        </div>
      </div>
    </div>
  </a>
{/if}
