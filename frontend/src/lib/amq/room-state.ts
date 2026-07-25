import { DEFAULT_SAVE_VOTE_SECONDS, isSaveGameType } from "$lib/amq/save-mode";

export interface SaveRoundResults {
	votes: Record<string, number>;
	winners: string[];
}

export interface SaveRoundData {
	game_type?: string;
	round_phase?: string;
	preview_seconds?: number;
	vote_seconds?: number;
	preview_index?: number;
	candidates?: Array<{
		song_uuid: string;
		vote_count?: number;
		is_winner?: boolean;
		[key: string]: unknown;
	}>;
	winners?: string[];
	[key: string]: unknown;
}

export interface SaveRoundHistoryEntry {
	round_number: number;
	theme_label: string;
	round_theme_type?: string;
	is_fallback?: boolean;
	winners?: string[];
	vote_counts?: Record<string, number>;
	candidates?: Array<{
		song_uuid: string;
		anime_title?: string;
		theme_label?: string;
		vote_count?: number;
		is_winner?: boolean;
	}>;
}

export function isSaveRoundData(roundData: SaveRoundData | null | undefined): boolean {
	return isSaveGameType(roundData?.game_type);
}

export function resolveSelectedCandidateOnReconnect(
	roundPhase: string | undefined,
	selectedSongUuid: string | undefined
): string {
	if (roundPhase === "preview_select" || roundPhase === "vote_select") {
		return selectedSongUuid || "";
	}
	return "";
}

export function applySaveRoundStart(payload: SaveRoundData): {
	currentRoundData: SaveRoundData;
	localTimer: number;
} {
	return {
		currentRoundData: payload,
		localTimer: payload.preview_seconds ?? 12,
	};
}

export function applySavePhaseChange(
	currentRoundData: SaveRoundData | null,
	payload: Partial<SaveRoundData> & { votes?: Record<string, number>; winners?: string[] },
	fallbackPreviewSeconds = 12
): {
	currentRoundData: SaveRoundData | null;
	saveRoundResults: SaveRoundResults | null;
	localTimer: number;
} {
	const merged = currentRoundData ? { ...currentRoundData, ...payload } : { ...payload };
	let saveRoundResults: SaveRoundResults | null = null;
	if (payload.votes) {
		saveRoundResults = {
			votes: payload.votes,
			winners: payload.winners || [],
		};
	}
	return {
		currentRoundData: merged,
		saveRoundResults,
		localTimer:
			merged.round_phase === "vote_select"
				? (payload.vote_seconds ?? merged.vote_seconds ?? DEFAULT_SAVE_VOTE_SECONDS)
				: (merged.preview_seconds ?? fallbackPreviewSeconds),
	};
}

export function applySaveRoundResults(
	currentRoundData: SaveRoundData | null,
	payload: SaveRoundResults
): {
	currentRoundData: SaveRoundData | null;
	saveRoundResults: SaveRoundResults;
} {
	if (!currentRoundData?.candidates?.length) {
		return { currentRoundData, saveRoundResults: payload };
	}

	const candidates = currentRoundData.candidates.map((c) => ({
		...c,
		vote_count: payload.votes?.[c.song_uuid] ?? 0,
		is_winner: payload.winners?.includes(c.song_uuid),
	}));

	if (!payload.winners?.length) {
		return {
			currentRoundData: {
				...currentRoundData,
				winners: [],
				candidates,
			},
			saveRoundResults: payload,
		};
	}

	return {
		currentRoundData: {
			...currentRoundData,
			round_phase: "winner_playback",
			winners: payload.winners,
			candidates,
		},
		saveRoundResults: payload,
	};
}

export function applySaveLobbyStateUpdate(
	roomState: {
		status?: string;
		round_data?: SaveRoundData;
		players?: Array<{ user_uuid?: string; device_id?: string; selected_song_uuid?: string }>;
		spectators?: Array<{ user_uuid?: string; device_id?: string; selected_song_uuid?: string }>;
	},
	opts: { isAuthenticated: boolean; userUuid?: string; deviceId: string }
): {
	currentRoundData: SaveRoundData | null;
	selectedCandidate: string;
} {
	if (roomState.status !== "playing" || !isSaveRoundData(roomState.round_data)) {
		return { currentRoundData: null, selectedCandidate: "" };
	}

	const allPlayers = [...(roomState.players || []), ...(roomState.spectators || [])];
	const me = allPlayers.find((p) =>
		opts.isAuthenticated ? p.user_uuid === opts.userUuid : p.device_id === opts.deviceId
	);

	return {
		currentRoundData: roomState.round_data ?? null,
		selectedCandidate: resolveSelectedCandidateOnReconnect(
			roomState.round_data?.round_phase,
			me?.selected_song_uuid
		),
	};
}
