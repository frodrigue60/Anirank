import { describe, expect, it } from "vitest";
import {
  DEFAULT_PLAYER_VOLUME,
  normalizePlayerVolume,
  readPlayerVolume,
  writePlayerVolume,
} from "./player-volume";

describe("player volume preferences", () => {
  it("normalizes invalid and out-of-range values", () => {
    expect(normalizePlayerVolume("invalid")).toBe(DEFAULT_PLAYER_VOLUME);
    expect(normalizePlayerVolume(2)).toBe(1);
    expect(normalizePlayerVolume(-1)).toBe(0);
  });

  it("reads a persisted volume", () => {
    const storage = { getItem: () => "0.35" };
    expect(readPlayerVolume(storage)).toBe(0.35);
  });

  it("persists a normalized volume", () => {
    let stored = "";
    const storage = { setItem: (_key: string, value: string) => (stored = value) };
    writePlayerVolume(storage, 0.42);
    expect(stored).toBe("0.42");
  });
});
