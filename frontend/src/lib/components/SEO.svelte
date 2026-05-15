<script lang="ts">
  import { page } from "$app/stores";
  //import logo from "$lib/assets/favicon.svg";

  let {
    title = "AniRank - The Ultimate Anime Music Ranking Platform",
    description = "Discover, rank, and listen to the best anime openings and endings. Create playlists, share with friends, and find your next favorite anime song.",
    image = "https://anirank.work/images/default-banner.png",
    logo = "/favicon.png",
    url = null,
    type = "website",
    author = "AniRank",
    keywords = "anime, music, ranking, openings, endings, ost, leaderboard",
    jsonLd = null,
  } = $props();

  let canonicalUrl = $derived(
    url || `https://anirank.work${$page.url.pathname === '/' ? '/' : $page.url.pathname.replace(/\/$/, '')}`
  );

  // Strip HTML from description and truncate to 160 characters
  let cleanDescription = $derived(
    description
      .replace(/<[^>]*>?/gm, "") // Remove HTML tags
      .replace(/\s+/g, " ") // Normalize spaces
      .trim()
      .substring(0, 160) + (description.length > 160 ? "..." : ""),
  );

  let siteName = "AniRank";
  let fullTitle = $derived(
    title.includes(siteName) ? title : `${title} - ${siteName}`,
  );

  let absoluteImage = $derived(
    image.startsWith("http") ? image : `${$page.url.origin}${image}`,
  );

  let absoluteLogo = $derived(
    logo.startsWith("http") ? logo : `${$page.url.origin}${logo}`,
  );
</script>

<svelte:head>
  <!-- Primary Meta Tags -->
  <title>{fullTitle}</title>
  <meta name="title" content={fullTitle} />
  <meta name="description" content={cleanDescription} />
  <meta name="author" content={author} />
  <link rel="canonical" href={canonicalUrl} />

  <!-- Open Graph / Facebook -->
  <meta property="og:type" content={type} />
  <meta property="og:url" content={canonicalUrl} />
  <meta property="og:title" content={fullTitle} />
  <meta property="og:description" content={cleanDescription} />
  <meta property="og:image" content={absoluteImage} />
  <meta property="og:logo" content={absoluteLogo} />
  <meta property="og:site_name" content={siteName} />

  <!-- Twitter -->
  <meta property="twitter:card" content="summary_large_image" />
  <meta property="twitter:url" content={canonicalUrl} />
  <meta property="twitter:title" content={fullTitle} />
  <meta property="twitter:description" content={cleanDescription} />
  <meta property="twitter:image" content={absoluteImage} />

  {#if keywords}
    <meta name="keywords" content={keywords} />
  {/if}

  {#if jsonLd}
    {@html `<script type="application/ld+json">${JSON.stringify(jsonLd)}</script>`}
  {/if}
</svelte:head>
