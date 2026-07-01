import { describe, expect, it } from "vitest";
import { getVideoTagText, hasMeaningfulVideoTag } from "./song-utils";

describe("getVideoTagText", () => {
  it("formats source, resolution, and quality tags", () => {
    expect(
      getVideoTagText({
        source: "DVD",
        resolution: 480,
        is_nc: true,
      }, 0),
    ).toBe("DVD 480p NC");
  });

  it("formats source and resolution without quality tags", () => {
    expect(
      getVideoTagText({ source: "DVD", resolution: 480 }, 0),
    ).toBe("DVD 480p");
  });

  it("formats TV with resolution", () => {
    expect(getVideoTagText({ source: "TV", resolution: 480 }, 0)).toBe(
      "TV 480p",
    );
  });

  it("formats source-only labels", () => {
    expect(getVideoTagText({ source: "TV" }, 0)).toBe("TV");
  });

  it("falls back to variant source and resolution", () => {
    expect(getVideoTagText({}, 1, { source: "DVD", resolution: 480 })).toBe(
      "DVD 480p",
    );
  });

  it("avoids duplicating BD when source is already BD", () => {
    expect(
      getVideoTagText({
        source: "BD",
        resolution: 1080,
        is_nc: true,
        is_bd: true,
      }, 0),
    ).toBe("BD 1080p NC");
  });

  it("shows BD quality tag when source is not BD", () => {
    expect(
      getVideoTagText({
        source: "TV",
        resolution: 1080,
        is_nc: true,
        is_bd: true,
      }, 0),
    ).toBe("TV 1080p NC BD");
  });

  it("ignores empty or none sources", () => {
    expect(
      getVideoTagText({ source: "None", resolution: 720 }, 0),
    ).toBe("720p");
  });

  it("falls back to Video N when no metadata exists", () => {
    expect(getVideoTagText({}, 2)).toBe("Video 3");
  });

  it("detects meaningful video tags", () => {
    expect(hasMeaningfulVideoTag({ source: "TV", resolution: 720 }, 0)).toBe(true);
    expect(hasMeaningfulVideoTag({}, 0)).toBe(false);
  });
});
