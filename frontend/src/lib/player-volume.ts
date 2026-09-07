export const PLAYER_VOLUME_STORAGE_KEY = "anirank_volume";
export const DEFAULT_PLAYER_VOLUME = 0.5;

export function normalizePlayerVolume(
  value: unknown,
  fallback = DEFAULT_PLAYER_VOLUME,
): number {
  const parsed = typeof value === "number" ? value : Number(value);
  if (!Number.isFinite(parsed)) return fallback;
  return Math.min(1, Math.max(0, parsed));
}

export function readPlayerVolume(
  storage: Pick<Storage, "getItem">,
  fallback = DEFAULT_PLAYER_VOLUME,
): number {
  const stored = storage.getItem(PLAYER_VOLUME_STORAGE_KEY);
  return stored === null ? fallback : normalizePlayerVolume(stored, fallback);
}

export function writePlayerVolume(
  storage: Pick<Storage, "setItem">,
  volume: number,
): void {
  storage.setItem(
    PLAYER_VOLUME_STORAGE_KEY,
    normalizePlayerVolume(volume).toString(),
  );
}
