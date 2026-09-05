export type QueueMode = "host_only" | "everyone" | "disabled";
export type RevealMode = "blind" | "live";
export type AutoAdvance = "never" | "all_rated";
export type RateStatus = "lobby" | "waiting" | "rating" | "finished";
export type SourceMode = "manual" | "seasonal_pool";
export type SeasonalPoolThemeType = "all" | "OP" | "ED";
/** Anime format slug from taxonomy, or "all". */
export type SeasonalPoolFormat = string;

/** Host → server: load happens automatically on start_session when source_mode is seasonal_pool. */
export interface SeasonalPoolRequest {
	year: string;
	season: string;
	theme_type: SeasonalPoolThemeType;
	format?: SeasonalPoolFormat;
	pool_limit?: number;
}

export interface RateConfig {
	name: string;
	private: boolean;
	queue_mode: QueueMode;
	queue_limit_per_user: number;
	reveal_mode: RevealMode;
	max_players: number;
	auto_advance: AutoAdvance;
	source_mode: SourceMode;
	pool_year?: string;
	pool_season?: string;
	pool_theme_type?: SeasonalPoolThemeType;
	pool_format?: SeasonalPoolFormat;
	pool_limit?: number;
}

export interface RatePlayer {
	session_id: string;
	user_uuid?: string;
	nickname: string;
	avatar_url?: string | null;
	profile_color?: string | null;
	device_id: string;
	is_host: boolean;
	is_spectator: boolean;
	offline: boolean;
}

export interface RateQueueItem {
	item_id: string;
	song_uuid: string;
	song_name: string;
	anime_title?: string;
	anime_slug?: string;
	theme_label?: string;
	cover_url?: string;
	added_by_session_id: string;
	added_by_user_uuid?: string;
	added_by_nickname: string;
}

export interface RatingEntry {
	rated: boolean;
	score?: number;
}

export interface RatingData {
	song_uuid?: string;
	session_avg?: number | null;
	rated_count: number;
	player_count: number;
	ratings: Record<string, RatingEntry>;
	my_score?: number | null;
	/** Existing global rating for this song (0–100), before submitting in the live session. */
	prior_score?: number | null;
	reveal_mode?: RevealMode;
}

export interface RateRoomState {
	room_id: string;
	status: RateStatus;
	config: RateConfig;
	players: RatePlayer[];
	spectators: RatePlayer[];
	queue: RateQueueItem[];
	current_song?: unknown;
	audio_url?: string;
	songs_rated?: number;
	rating_data?: RatingData;
	my_session_id?: string;
}

export function defaultRateConfig(): RateConfig {
	return {
		name: "Rate Party",
		private: false,
		queue_mode: "host_only",
		queue_limit_per_user: 3,
		reveal_mode: "blind",
		max_players: 16,
		auto_advance: "never",
		source_mode: "manual",
	};
}

export function isSeasonalPool(config: RateConfig | undefined | null): boolean {
	return config?.source_mode === "seasonal_pool";
}

export function canAddToQueue(
	config: RateConfig,
	player: RatePlayer | undefined,
	queue: RateQueueItem[]
): { ok: boolean; reason?: string } {
	if (!player || player.is_spectator || player.offline) {
		return { ok: false, reason: "Not allowed" };
	}
	if (isSeasonalPool(config)) {
		return {
			ok: false,
			reason: "Seasonal pool mode — host must switch to manual in lobby to add themes",
		};
	}
	if (config.queue_mode === "disabled") {
		return { ok: false, reason: "Queue is disabled" };
	}
	if (config.queue_mode === "host_only" && !player.is_host) {
		return { ok: false, reason: "Only the host can add songs" };
	}
	if (config.queue_mode === "everyone" && !player.user_uuid) {
		return { ok: false, reason: "Login required to queue songs" };
	}
	if (queue.length >= 50) {
		return { ok: false, reason: "Queue is full" };
	}
	if (config.queue_mode === "everyone" && player.user_uuid) {
		const count = queue.filter((q) => q.added_by_user_uuid === player.user_uuid).length;
		if (count >= config.queue_limit_per_user) {
			return {
				ok: false,
				reason: `Limit of ${config.queue_limit_per_user} songs per user`,
			};
		}
	}
	return { ok: true };
}

export function applyLobbyStateUpdate(
	prev: RateRoomState | null,
	payload: RateRoomState
): RateRoomState {
	return {
		...(prev || ({} as RateRoomState)),
		...payload,
		players: payload.players || [],
		spectators: payload.spectators || [],
		queue: payload.queue || [],
		rating_data: payload.rating_data || prev?.rating_data,
	};
}

export function applyRatingUpdate(
	prev: RateRoomState | null,
	payload: RatingData
): RateRoomState | null {
	if (!prev) return prev;
	return {
		...prev,
		rating_data: {
			...(prev.rating_data || {
				rated_count: 0,
				player_count: 0,
				ratings: {},
			}),
			...payload,
		},
	};
}

/** Convert UI score format value to canonical 0–100 for submit_rating. */
export function toCanonicalScore(display: number, format: string): number {
	switch (format) {
		case "POINT_100":
			return Math.min(100, Math.max(0, display));
		case "POINT_10":
			return Math.min(100, Math.max(0, Math.round(display) * 10));
		case "POINT_5":
			return Math.min(100, Math.max(0, display * 20));
		case "POINT_10_DECIMAL":
		default:
			return Math.min(100, Math.max(0, display * 10));
	}
}

export function fromCanonicalScore(canonical: number, format: string): number {
	switch (format) {
		case "POINT_100":
			return Math.round(canonical);
		case "POINT_10":
			return Math.round(canonical / 10);
		case "POINT_5":
			return canonical / 20;
		case "POINT_10_DECIMAL":
		default:
			return canonical / 10;
	}
}
