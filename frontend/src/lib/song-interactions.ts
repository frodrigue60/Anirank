export interface InteractiveSong {
  is_liked?: boolean;
  is_disliked?: boolean;
  is_favorited?: boolean;
  likes_count?: number;
  dislikes_count?: number;
}

export type SongInteractionSnapshot = Required<InteractiveSong>;
export type ReactionKind = "like" | "dislike";

export function snapshotSongInteractions(
  song: InteractiveSong,
): SongInteractionSnapshot {
  return {
    is_liked: Boolean(song.is_liked),
    is_disliked: Boolean(song.is_disliked),
    is_favorited: Boolean(song.is_favorited),
    likes_count: song.likes_count ?? 0,
    dislikes_count: song.dislikes_count ?? 0,
  };
}

export function restoreSongInteractions(
  song: InteractiveSong,
  snapshot: SongInteractionSnapshot,
): void {
  Object.assign(song, snapshot);
}

export function applyOptimisticReaction(
  song: InteractiveSong,
  kind: ReactionKind,
): void {
  const previous = snapshotSongInteractions(song);
  const nextLiked = kind === "like" ? !previous.is_liked : false;
  const nextDisliked = kind === "dislike" ? !previous.is_disliked : false;

  song.is_liked = nextLiked;
  song.is_disliked = nextDisliked;
  song.likes_count = Math.max(
    0,
    previous.likes_count + Number(nextLiked) - Number(previous.is_liked),
  );
  song.dislikes_count = Math.max(
    0,
    previous.dislikes_count +
      Number(nextDisliked) -
      Number(previous.is_disliked),
  );
}

export function applyOptimisticFavorite(song: InteractiveSong): void {
  song.is_favorited = !Boolean(song.is_favorited);
}

export function interactionResponseData(
  response: unknown,
): Record<string, unknown> {
  const axiosData = (response as { data?: unknown } | null)?.data;
  if (!axiosData || typeof axiosData !== "object") return {};
  const nested = (axiosData as { data?: unknown }).data;
  return nested && typeof nested === "object"
    ? (nested as Record<string, unknown>)
    : (axiosData as Record<string, unknown>);
}

export function syncReactionCounts(
  song: InteractiveSong,
  response: unknown,
): void {
  const data = interactionResponseData(response);
  if (typeof data.likesCount === "number") song.likes_count = data.likesCount;
  if (typeof data.dislikesCount === "number") {
    song.dislikes_count = data.dislikesCount;
  }
}

export function syncFavoriteState(
  song: InteractiveSong,
  response: unknown,
): void {
  const data = interactionResponseData(response);
  if (typeof data.favorited === "boolean") {
    song.is_favorited = data.favorited;
  } else if (typeof data.favorite === "boolean") {
    song.is_favorited = data.favorite;
  }
}
