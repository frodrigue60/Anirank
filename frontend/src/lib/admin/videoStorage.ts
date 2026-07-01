import api from "$lib/api";

export function normalizeStoragePath(path: string): string {
  return path.trim().replace(/^\/+/, "");
}

export async function fetchSongVideoStorageCheck(
  songId: number | string,
): Promise<Record<string, boolean>> {
  const res = await api.get(`/admin/songs/${songId}/video-storage-check`);
  return res.data?.data ?? {};
}

export async function fetchVariantVideoStorageCheck(
  variantId: number | string,
): Promise<Record<string, boolean>> {
  const res = await api.get(`/admin/variants/${variantId}/video-storage-check`);
  return res.data?.data ?? {};
}
