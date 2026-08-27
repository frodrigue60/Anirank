import { describe, expect, it } from "vitest";
import {
  isStorageVideoSrc,
  storagePlaybackUrl,
  variantStoragePlaybackUrl,
} from "$lib/videoStorageSrc";

describe("videoStorageSrc", () => {
  it("detects storage keys vs http urls", () => {
    expect(isStorageVideoSrc("videos/foo.webm")).toBe(true);
    expect(isStorageVideoSrc("https://r2.anirank.work/videos/foo.webm")).toBe(
      false,
    );
    expect(isStorageVideoSrc(undefined)).toBe(false);
  });

  it("returns resolved playback url for flat variant fields", () => {
    expect(
      variantStoragePlaybackUrl({
        video_src: "videos/foo.webm",
        local_url: "https://r2.anirank.work/videos/foo.webm",
      }),
    ).toBe("https://r2.anirank.work/videos/foo.webm");
  });

  it("prefers videos[] entries when present", () => {
    expect(
      variantStoragePlaybackUrl({
        video_src: "videos/legacy.webm",
        videos: [
          {
            video_src: "videos/active.webm",
            video_url: "https://r2.anirank.work/videos/active.webm",
          },
        ],
      }),
    ).toBe("https://r2.anirank.work/videos/active.webm");
  });

  it("rejects http-only legacy video_src", () => {
    expect(
      storagePlaybackUrl({
        video_src: "https://youtube.com/embed/abc",
        local_url: "https://youtube.com/embed/abc",
      }),
    ).toBeUndefined();
  });
});
