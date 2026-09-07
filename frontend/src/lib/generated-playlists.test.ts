import { describe, expect, it } from "vitest";
import { generatedPlaylistApiPath } from "./generated-playlists";

describe("generatedPlaylistApiPath", () => {
  it("maps owner playlists to protected endpoints", () => {
    expect(generatedPlaylistApiPath("user", "rated")).toBe(
      "/me/playlists/generated/rated",
    );
    expect(generatedPlaylistApiPath("user", "liked")).toBe(
      "/me/playlists/generated/liked",
    );
  });

  it("maps year and seasonal playlists", () => {
    expect(generatedPlaylistApiPath("year", "2026")).toBe(
      "/playlists/generated/year/2026",
    );
    expect(generatedPlaylistApiPath("season", "2026/summer")).toBe(
      "/playlists/generated/season/2026/summer",
    );
  });

  it("rejects unsupported or incomplete routes", () => {
    expect(generatedPlaylistApiPath("user", "disliked")).toBeNull();
    expect(generatedPlaylistApiPath("season", "summer")).toBeNull();
    expect(generatedPlaylistApiPath("unknown", "2026")).toBeNull();
  });
});
