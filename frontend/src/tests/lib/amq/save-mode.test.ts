import { describe, it, expect } from "vitest";
import {
	isSaveGameType,
	saveGameModeLabel,
	formatSaveVoteSeconds,
	getSaveActiveCandidateIndex,
	shouldShowSavedSelection,
	shouldRevealSaveCandidateLabels,
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

	it("formats vote seconds for display", () => {
		expect(formatSaveVoteSeconds(10)).toBe("10s");
		expect(formatSaveVoteSeconds(0)).toBe("Instant");
		expect(formatSaveVoteSeconds(undefined)).toBe("10s");
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

	it("hides candidate labels until winner_playback reveals all options", () => {
		expect(shouldRevealSaveCandidateLabels("media_buffer")).toBe(false);
		expect(shouldRevealSaveCandidateLabels("preview_select")).toBe(false);
		expect(shouldRevealSaveCandidateLabels("vote_select")).toBe(false);
		expect(shouldRevealSaveCandidateLabels("winner_playback")).toBe(true);
	});
});
