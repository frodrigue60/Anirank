<script lang="ts">
  import OptimizedImage from "./OptimizedImage.svelte";
  let { artist } = $props();
</script>

<div class="group flex flex-col items-center gap-4 cursor-pointer">
  <a
    href="/artists/{artist.slug}"
    class="relative w-full aspect-square rounded-full overflow-hidden card-shadow ring-4 ring-transparent group-hover:ring-primary/50 transition-all duration-300"
    title="View artist profile: {artist.name}"
  >
    <OptimizedImage
      src={artist.avatar_url}
      sources={artist.avatar_sources}
      alt="Avatar for {artist.name}"
      class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
      sizes="(max-width: 640px) 120px, 180px"
    />
    <div
      class="absolute inset-0 bg-linear-to-t from-primary/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"
    ></div>
  </a>
  <div class="text-center">
    <a
      href="/artists/{artist.slug}"
      class="font-bold text-on-surface group-hover:text-primary transition-colors text-lg line-clamp-1"
      title="View artist profile: {artist.name}">{artist.name}</a
    >
    <div
      class="mt-1.5 inline-flex items-center px-3 py-1 rounded-sm bg-primary/20 text-primary text-[11px] font-bold border border-primary/20"
    >
      {artist.enabled_songs || 0} Themes
    </div>
  </div>
</div>
