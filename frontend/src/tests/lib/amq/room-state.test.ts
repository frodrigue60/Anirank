import { describe, it, expect } from "vitest";
import {
	applySaveLobbyStateUpdate,
	applySavePhaseChange,
	applySaveRoundResults,
	applySaveRoundStart,
	resolveSelectedCandidateOnReconnect,
} from "$lib/amq/room-state";

describe("AMQ save room-state reducer", () => {
	it("applies round_start with preview timer", () => {
		const result = applySaveRoundStart({
			game_type: "save-4",
			preview_seconds: 14,
			candidates: [{ song_uuid: "a" }],
		});
		expect(result.localTimer).toBe(14);
		expect(result.currentRoundData.game_type).toBe("save-4");
	});

	it("merges phase_change and preserves vote snapshot", () => {
		const current = { game_type: "save-4", preview_seconds: 12, candidates: [{ song_uuid: "a" }] };
		const result = applySavePhaseChange(
			current,
			{ round_phase: "winner_playback", winner_play_index: 0, votes: { a: 2 }, winners: ["a"] },
			12
		);
		expect(result.currentRoundData?.round_phase).toBe("winner_playback");
		expect(result.saveRoundResults?.votes.a).toBe(2);
		expect(result.localTimer).toBe(12);
	});

	it("maps round_results onto candidates", () => {
		const current = {
			game_type: "save-4",
			candidates: [{ song_uuid: "a" }, { song_uuid: "b" }],
		};
		const result = applySaveRoundResults(current, { votes: { a: 3, b: 1 }, winners: ["a"] });
		expect(result.currentRoundData?.round_phase).toBe("winner_playback");
		expect(result.currentRoundData?.candidates?.[0].vote_count).toBe(3);
		expect(result.currentRoundData?.candidates?.[0].is_winner).toBe(true);
	});

	it("restores selection on reconnect during preview", () => {
		const result = applySaveLobbyStateUpdate(
			{
				status: "playing",
				round_data: { game_type: "save-6", round_phase: "preview_select" },
				players: [{ device_id: "dev-1", selected_song_uuid: "song-b" }],
			},
			{ isAuthenticated: false, deviceId: "dev-1" }
		);
		expect(result.selectedCandidate).toBe("song-b");
		expect(result.currentRoundData?.game_type).toBe("save-6");
	});

	it("clears selection on winner_playback reconnect", () => {
		expect(resolveSelectedCandidateOnReconnect("winner_playback", "song-a")).toBe("");
	});
});
