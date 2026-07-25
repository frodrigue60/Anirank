import { describe, it, expect } from "vitest";
import {
	isSaveGameType,
	saveGameModeLabel,
	getSaveActiveCandidateIndex,
	shouldShowSavedSelection,
} from "$lib/amq/save-mode";

describe("AMQ save mode helpers", () => {
	it("detects save game types", () => {
		expect(isSaveGameType("save-4")).toBe(true);
		expect(isSaveGameType("save-6")).toBe(true);
		expect(isSaveGameType("type-in")).toBe(false);
	});

	it("formats save mode labels", () => {
		expect(saveGameModeLabel("save-4")).toBe("Save 1 of 4");
		expect(saveGameModeLabel("save-6")).toBe("Save 1 of 6");
	});

	it("returns preview index during preview_select", () => {
		const idx = getSaveActiveCandidateIndex(
			{
				candidates: [{ song_uuid: "a" }, { song_uuid: "b" }],
				preview_index: 1,
			},
			"preview_select",
			null
		);
		expect(idx).toBe(1);
	});

	it("returns winner playback index during winner_playback", () => {
		const idx = getSaveActiveCandidateIndex(
			{
				candidates: [{ song_uuid: "a" }, { song_uuid: "b" }],
				winners: ["b"],
				winner_play_index: 0,
			},
			"winner_playback",
			null
		);
		expect(idx).toBe(1);
	});

	it("only shows saved badge during preview or vote when selected", () => {
		expect(shouldShowSavedSelection("preview_select", "a", "a")).toBe(true);
		expect(shouldShowSavedSelection("vote_select", "a", "a")).toBe(true);
		expect(shouldShowSavedSelection("preview_select", "a", "b")).toBe(false);
		expect(shouldShowSavedSelection("winner_playback", "a", "a")).toBe(false);
	});

	it("returns no active preview index during vote_select", () => {
		const idx = getSaveActiveCandidateIndex(
			{ candidates: [{ song_uuid: "a" }, { song_uuid: "b" }], preview_index: 1 },
			"vote_select",
			null
		);
		expect(idx).toBe(-1);
	});
});
