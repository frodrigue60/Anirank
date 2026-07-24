export interface SaveCandidate {
	song_uuid: string;
}

export interface SaveRoundView {
	candidates?: SaveCandidate[];
	preview_index?: number;
	winner_play_index?: number;
	winners?: string[];
}

export function isSaveGameType(gameType: string | undefined): boolean {
	return gameType === "save-4" || gameType === "save-6";
}

export function saveGameModeLabel(type: string): string {
	if (type === "save-6") return "Save 1 of 6";
	if (type === "save-4") return "Save 1 of 4";
	return type;
}

export function getSaveActiveCandidateIndex(
	activeRound: SaveRoundView | null | undefined,
	savePhase: string,
	saveRoundResults: { winners?: string[] } | null | undefined
): number {
	if (!activeRound?.candidates?.length) return 0;

	if (savePhase === "winner_playback") {
		const winners = activeRound.winners || saveRoundResults?.winners || [];
		const uuid = winners[activeRound.winner_play_index ?? 0];
		const idx = activeRound.candidates.findIndex((c) => c.song_uuid === uuid);
		return idx >= 0 ? idx : 0;
	}

	return activeRound.preview_index ?? 0;
}

export function shouldShowSavedSelection(savePhase: string, selectedUuid: string, candidateUuid: string): boolean {
	return savePhase === "preview_select" && selectedUuid === candidateUuid;
}
