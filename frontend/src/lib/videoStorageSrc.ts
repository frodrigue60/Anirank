/** True when video_src is an object-storage key (S3/R2), not a legacy http(s) URL. */
export function isStorageVideoSrc(src?: string | null): boolean {
  if (!src?.trim()) return false;
  const lower = src.trim().toLowerCase();
  return !lower.startsWith("http://") && !lower.startsWith("https://");
}

type VideoSourceShape = {
  video_src?: string | null;
  local_url?: string | null;
  video_url?: string | null;
  video?: {
    video_src?: string | null;
    local_url?: string | null;
    video_url?: string | null;
  } | null;
};

/** Playback URL for storage-backed videos (resolved local_url / video_url). */
export function storagePlaybackUrl(source?: VideoSourceShape | null): string | undefined {
  if (!source) return undefined;

  const nested = source.video;
  const videoSrc = source.video_src?.trim() || nested?.video_src?.trim();
  if (!videoSrc || !isStorageVideoSrc(videoSrc)) return undefined;

  return (
    source.local_url ||
    source.video_url ||
    nested?.local_url ||
    nested?.video_url ||
    undefined
  );
}

/** Resolve playback URL from a song variant DTO (flat fields + videos[]). */
export function variantStoragePlaybackUrl(variant?: {
  video_src?: string | null;
  local_url?: string | null;
  video_url?: string | null;
  video?: VideoSourceShape["video"];
  videos?: VideoSourceShape[] | null;
} | null): string | undefined {
  if (!variant) return undefined;

  if (variant.videos?.length) {
    for (const vid of variant.videos) {
      const url = storagePlaybackUrl(vid);
      if (url) return url;
    }
  }

  return storagePlaybackUrl(variant);
}
