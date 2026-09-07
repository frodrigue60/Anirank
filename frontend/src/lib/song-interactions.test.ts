import { describe, expect, it } from "vitest";
import {
  applyOptimisticFavorite,
  applyOptimisticReaction,
  restoreSongInteractions,
  snapshotSongInteractions,
  syncFavoriteState,
  syncReactionCounts,
} from "./song-interactions";

describe("song interaction state", () => {
  it("applies a reaction optimistically and syncs server counters", () => {
    const song = {
      is_liked: false,
      is_disliked: true,
      likes_count: 4,
      dislikes_count: 2,
    };

    applyOptimisticReaction(song, "like");
    expect(song).toMatchObject({
      is_liked: true,
      is_disliked: false,
      likes_count: 5,
      dislikes_count: 1,
    });

    syncReactionCounts(song, {
      data: { data: { likesCount: 8, dislikesCount: 3 } },
    });
    expect(song.likes_count).toBe(8);
    expect(song.dislikes_count).toBe(3);
  });

  it("restores the previous reaction state after an error", () => {
    const song = {
      is_liked: true,
      is_disliked: false,
      is_favorited: false,
      likes_count: 9,
      dislikes_count: 1,
    };
    const previous = snapshotSongInteractions(song);

    applyOptimisticReaction(song, "like");
    restoreSongInteractions(song, previous);

    expect(song).toEqual(previous);
  });

  it("handles an explicit false favorite response", () => {
    const song = { is_favorited: true };
    const previous = snapshotSongInteractions(song);

    applyOptimisticFavorite(song);
    syncFavoriteState(song, { data: { data: { favorited: false } } });
    expect(song.is_favorited).toBe(false);

    restoreSongInteractions(song, previous);
    expect(song.is_favorited).toBe(true);
  });
});
