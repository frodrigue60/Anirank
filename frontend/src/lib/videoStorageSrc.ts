/** True when video_src is an object-storage key (S3/R2), not a legacy http(s) URL. */
export function isStorageVideoSrc(src?: string | null): boolean {
  if (!src?.trim()) return false;
  const lower = src.trim().toLowerCase();
  return !lower.startsWith("http://") && !lower.startsWith("https://");
}

/** Playback URL for storage-backed videos (resolved local_url / video_url). */
export function storagePlaybackUrl(video?: {
  video_src?: string | null;
  local_url?: string | null;
  video_url?: string | null;
} | null): string | undefined {
  if (!video || !isStorageVideoSrc(video.video_src)) return undefined;
  return video.local_url || video.video_url || undefined;
}
