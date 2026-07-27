export interface SaveCandidate {
	song_uuid: string;
}

export interface SaveRoundView {
	candidates?: SaveCandidate[];
	preview_index?: number;
	winner_play_index?: number;
	winners?: string[];
	vote_seconds?: number;
}

export const DEFAULT_SAVE_VOTE_SECONDS = 10;

export function formatSaveVoteSeconds(seconds: number | undefined | null): string {
	if (seconds === 0) return "Instant";
	return `${seconds ?? DEFAULT_SAVE_VOTE_SECONDS}s`;
}

export function isSaveGameType(gameType: string | undefined): boolean {
	return gameType === "save-4" || gameType === "save-6";
}

export function saveGameModeLabel(type: string): string {
	if (type === "save-6") return "Save 1 of 6";
	if (type === "save-4") return "Save 1 of 4";
	return type;
}

export function canSelectSaveCandidate(savePhase: string): boolean {
	return savePhase === "preview_select" || savePhase === "vote_select";
}

export function isSaveMediaBufferPhase(savePhase: string): boolean {
	return savePhase === "media_buffer";
}

export function savePhaseTimerLabel(savePhase: string): string {
	if (savePhase === "media_buffer") return "Preparing";
	if (savePhase === "winner_playback") return "Winner";
	if (savePhase === "vote_select") return "Vote";
	return "Preview";
}

export function getSaveActiveCandidateIndex(
	activeRound: SaveRoundView | null | undefined,
	savePhase: string,
	saveRoundResults: { winners?: string[] } | null | undefined
): number {
	if (!activeRound?.candidates?.length) return -1;
	if (savePhase === "vote_select") return -1;
	if (savePhase === "media_buffer") return -1;

	if (savePhase === "winner_playback") {
		const winners = activeRound.winners || saveRoundResults?.winners || [];
		const uuid = winners[activeRound.winner_play_index ?? 0];
		const idx = activeRound.candidates.findIndex((c) => c.song_uuid === uuid);
		return idx >= 0 ? idx : 0;
	}

	return activeRound.preview_index ?? 0;
}

export function shouldShowSavedSelection(savePhase: string, selectedUuid: string, candidateUuid: string): boolean {
	return canSelectSaveCandidate(savePhase) && selectedUuid === candidateUuid;
}

/** Candidate song/anime labels stay hidden until winner_playback reveals all options. */
export function shouldRevealSaveCandidateLabels(savePhase: string): boolean {
	return savePhase === "winner_playback";
}
