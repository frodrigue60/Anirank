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
