export function generatedPlaylistApiPath(
  kind: string,
  key: string,
): string | null {
  const parts = key.split("/").filter(Boolean).map(encodeURIComponent);

  if (
    kind === "user" &&
    parts.length === 1 &&
    ["rated", "liked", "favorited"].includes(parts[0])
  ) {
    return `/me/playlists/generated/${parts[0]}`;
  }
  if (kind === "year" && parts.length === 1) {
    return `/playlists/generated/year/${parts[0]}`;
  }
  if (kind === "season" && parts.length === 2) {
    return `/playlists/generated/season/${parts[0]}/${parts[1]}`;
  }

  return null;
}
