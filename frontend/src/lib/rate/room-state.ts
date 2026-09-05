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
	/** When true, majority of online players can vote to skip the current song. */
	vote_skip: boolean;
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

export interface SkipVoteData {
	enabled: boolean;
	count: number;
	needed: number;
	my_voted: boolean;
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
	skip_vote?: SkipVoteData;
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
		vote_skip: false,
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

/** Session statuses where playback / queue / rating controls apply. */
export function isLiveSessionStatus(status: RateStatus | string | undefined): boolean {
	return status === "waiting" || status === "rating";
}

/** Stable round identity for resetting client locks between songs. */
export function roundIdentity(
	status: RateStatus | string | undefined,
	songUuid: string | undefined | null
): string {
	return `${status || ""}:${songUuid || ""}`;
}

export function resolveSongUuid(
	song: unknown,
	ratingSongUuid?: string | null
): string {
	if (ratingSongUuid) return ratingSongUuid;
	if (!song || typeof song !== "object") return "";
	const s = song as { id?: string; uuid?: string };
	return s.id || s.uuid || "";
}

export type RateControlState = {
	/** Show the control in the UI. */
	visible: boolean;
	/** Accept clicks (connected + role + round state). */
	enabled: boolean;
	reason?: string;
};

export type RateControlContext = {
	status: RateStatus | string;
	config: RateConfig;
	me?: RatePlayer | null;
	/** WebSocket open and seat online. */
	connected: boolean;
	queue: RateQueueItem[];
	skipVote?: SkipVoteData | null;
	authenticated: boolean;
	draftScore: number;
	alreadyRated: boolean;
	/** Short-lived client lock after sending an action. */
	busy?: boolean;
};

function offlinePlayerAsOnline(player: RatePlayer): RatePlayer {
	return player.offline ? { ...player, offline: false } : player;
}

/** Search anime (open modal) — host always, others when queue mode allows. */
export function searchAnimeControl(ctx: RateControlContext): RateControlState {
	if (!isLiveSessionStatus(ctx.status)) {
		return { visible: false, enabled: false, reason: "Session not active" };
	}
	if (isSeasonalPool(ctx.config)) {
		return { visible: false, enabled: false, reason: "Seasonal pool mode" };
	}
	const me = ctx.me;
	if (!me || me.is_spectator) {
		return { visible: false, enabled: false, reason: "Spectators cannot search" };
	}
	if (me.is_host) {
		if (!ctx.connected) {
			return { visible: true, enabled: false, reason: "Reconnecting…" };
		}
		if (ctx.busy) {
			return { visible: true, enabled: false, reason: "Please wait…" };
		}
		return { visible: true, enabled: true };
	}
	if (ctx.config.queue_mode === "disabled") {
		return { visible: false, enabled: false, reason: "Queue is disabled" };
	}
	const perm = canAddToQueue(ctx.config, offlinePlayerAsOnline(me), ctx.queue);
	if (!perm.ok) {
		return { visible: false, enabled: false, reason: perm.reason };
	}
	if (!ctx.connected || me.offline) {
		return { visible: true, enabled: false, reason: "Reconnecting…" };
	}
	if (ctx.busy) {
		return { visible: true, enabled: false, reason: "Please wait…" };
	}
	return { visible: true, enabled: true };
}

/** Host Next / Next from pool. */
export function hostNextControl(ctx: RateControlContext): RateControlState {
	const me = ctx.me;
	if (!me?.is_host || me.is_spectator) {
		return { visible: false, enabled: false };
	}
	if (!isLiveSessionStatus(ctx.status)) {
		return { visible: false, enabled: false, reason: "Session not active" };
	}
	if (!ctx.connected || me.offline) {
		return { visible: true, enabled: false, reason: "Reconnecting…" };
	}
	if (ctx.busy) {
		return { visible: true, enabled: false, reason: "Advancing…" };
	}
	return { visible: true, enabled: true };
}

/** Vote skip during the current rating round. */
export function voteSkipControl(ctx: RateControlContext): RateControlState {
	if (!ctx.config.vote_skip) {
		return { visible: false, enabled: false };
	}
	const me = ctx.me;
	if (!me || me.is_spectator) {
		return { visible: false, enabled: false };
	}
	if (ctx.status !== "rating") {
		return { visible: false, enabled: false, reason: "No song playing" };
	}
	if (!ctx.connected || me.offline) {
		return { visible: true, enabled: false, reason: "Reconnecting…" };
	}
	if (ctx.skipVote?.my_voted) {
		return { visible: true, enabled: false, reason: "Already voted" };
	}
	if (ctx.busy) {
		return { visible: true, enabled: false, reason: "Please wait…" };
	}
	return { visible: true, enabled: true };
}

/** Host Play now (set_song) from search modal. */
export function playNowControl(ctx: RateControlContext): RateControlState {
	const me = ctx.me;
	if (!me?.is_host || me.is_spectator) {
		return { visible: false, enabled: false };
	}
	if (!isLiveSessionStatus(ctx.status)) {
		return { visible: false, enabled: false, reason: "Session not active" };
	}
	if (isSeasonalPool(ctx.config)) {
		return { visible: false, enabled: false, reason: "Use Next in seasonal pool" };
	}
	if (!ctx.connected || me.offline) {
		return { visible: true, enabled: false, reason: "Reconnecting…" };
	}
	if (ctx.busy) {
		return { visible: true, enabled: false, reason: "Please wait…" };
	}
	return { visible: true, enabled: true };
}

/** + Queue from search modal. */
export function queueAddControl(ctx: RateControlContext): RateControlState {
	if (!isLiveSessionStatus(ctx.status)) {
		return { visible: false, enabled: false, reason: "Session not active" };
	}
	if (isSeasonalPool(ctx.config) || ctx.config.queue_mode === "disabled") {
		return { visible: false, enabled: false, reason: "Queue is disabled" };
	}
	const me = ctx.me;
	if (!me || me.is_spectator) {
		return { visible: false, enabled: false };
	}
	if (me.is_host) {
		if (!ctx.connected || me.offline) {
			return { visible: true, enabled: false, reason: "Reconnecting…" };
		}
		if (ctx.busy) {
			return { visible: true, enabled: false, reason: "Please wait…" };
		}
		return { visible: true, enabled: true };
	}
	const perm = canAddToQueue(ctx.config, offlinePlayerAsOnline(me), ctx.queue);
	if (!perm.ok) {
		return { visible: false, enabled: false, reason: perm.reason };
	}
	if (!ctx.connected || me.offline) {
		return { visible: true, enabled: false, reason: "Reconnecting…" };
	}
	if (ctx.busy) {
		return { visible: true, enabled: false, reason: "Please wait…" };
	}
	return { visible: true, enabled: true };
}

/** Submit rating form for the current song. */
export function submitRatingControl(ctx: RateControlContext): RateControlState {
	if (ctx.status !== "rating") {
		return { visible: false, enabled: false, reason: "No song is being rated" };
	}
	const me = ctx.me;
	if (!me || me.is_spectator) {
		return { visible: false, enabled: false };
	}
	if (!ctx.authenticated) {
		return { visible: true, enabled: false, reason: "Login required to rate" };
	}
	if (ctx.alreadyRated) {
		return { visible: true, enabled: false, reason: "Already rated" };
	}
	if (!ctx.connected || me.offline) {
		return { visible: true, enabled: false, reason: "Reconnecting…" };
	}
	if (ctx.busy) {
		return { visible: true, enabled: false, reason: "Saving…" };
	}
	const score = Number(ctx.draftScore);
	if (!Number.isFinite(score) || score <= 0) {
		return { visible: true, enabled: false, reason: "Pick a score to submit" };
	}
	return { visible: true, enabled: true };
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
		// Prefer incoming snapshot so a new round never keeps stale ratings/skip votes.
		rating_data:
			payload.rating_data !== undefined ? payload.rating_data : prev?.rating_data,
		skip_vote: payload.skip_vote !== undefined ? payload.skip_vote : prev?.skip_vote,
		// Keep identity if a partial/stale payload omits it (reconnect races).
		my_session_id: payload.my_session_id || prev?.my_session_id,
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
