export function getSongName(song: any): string {
  if (!song) return "N/A";
  return song.name || song.song_romaji || song.song_en || song.song_jp || "N/A";
}

export function getSongArtistNames(
  artists: Array<{ name?: string; status?: boolean }> | undefined | null,
): string {
  if (!artists || artists.length === 0) return "N/A";

  const names = artists.map((artist) =>
    artist?.status === false ? "N/A" : artist?.name || "N/A",
  );
  return names.join(", ");
}

/** Minimal variant shape for URL/index resolution. */
export type VariantVersionRef = { version_number?: number };

/**
 * Maps ?v= query param to a variant array index.
 * Primary: version_number (V1 → ?v=1). Fallback: legacy 0-based array index.
 */
export function resolveVariantIndex(
  variants: VariantVersionRef[] | undefined | null,
  param: string | null,
): number {
  if (!variants?.length) return 0;
  if (param === null || param.trim() === "") return 0;

  const parsed = Number.parseInt(param, 10);
  if (Number.isNaN(parsed)) return 0;

  const byVersionNumber = variants.findIndex(
    (variant) => (variant.version_number ?? 0) === parsed,
  );
  if (byVersionNumber >= 0) return byVersionNumber;

  if (parsed >= 0 && parsed < variants.length) return parsed;

  return 0;
}

export function buildSongPlayHref(
  animeSlug: string,
  songSlug: string,
  versionNumber?: number,
): string {
  const base = `/animes/${animeSlug}/${songSlug}`;
  if (!versionNumber || versionNumber < 1) return base;
  return `${base}?v=${versionNumber}`;
}

export function getFormattedScore(
  score: string | number | undefined,
  format: string = "POINT_10_DECIMAL",
): string {
  if (!score || score === "0" || score === 0) return "0.0";

  const raw = typeof score === "string" ? parseFloat(score) : score;
  if (isNaN(raw)) return "0.0";

  switch (format) {
    case "POINT_100":
      return raw.toFixed(1);
    case "POINT_10":
      return (raw / 10).toFixed(1);
    case "POINT_10_DECIMAL":
      return (raw / 10).toFixed(1);
    case "POINT_5":
      return (raw / 20).toFixed(1);
    default:
      return (raw / 10).toFixed(1);
  }
}
