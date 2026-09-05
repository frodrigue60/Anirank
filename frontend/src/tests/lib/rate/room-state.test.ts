import { describe, expect, it } from "vitest";
import {
	applyLobbyStateUpdate,
	applyRatingUpdate,
	canAddToQueue,
	fromCanonicalScore,
	hostNextControl,
	playNowControl,
	queueAddControl,
	roundIdentity,
	searchAnimeControl,
	submitRatingControl,
	toCanonicalScore,
	voteSkipControl,
	type RateConfig,
	type RatePlayer,
	type RateQueueItem,
	type RateRoomState,
} from "$lib/rate/room-state";

const baseConfig: RateConfig = {
	name: "Test",
	private: false,
	queue_mode: "everyone",
	queue_limit_per_user: 2,
	reveal_mode: "blind",
	max_players: 16,
	auto_advance: "never",
	vote_skip: false,
	source_mode: "manual",
};

const authPlayer: RatePlayer = {
	session_id: "s1",
	user_uuid: "u1",
	nickname: "Luis",
	device_id: "d1",
	is_host: false,
	is_spectator: false,
	offline: false,
};

describe("canAddToQueue", () => {
	it("enforces per-user limit in everyone mode", () => {
		const queue: RateQueueItem[] = [
			{
				item_id: "1",
				song_uuid: "a",
				song_name: "A",
				added_by_session_id: "s1",
				added_by_user_uuid: "u1",
				added_by_nickname: "Luis",
			},
			{
				item_id: "2",
				song_uuid: "b",
				song_name: "B",
				added_by_session_id: "s1",
				added_by_user_uuid: "u1",
				added_by_nickname: "Luis",
			},
		];
		const result = canAddToQueue(baseConfig, authPlayer, queue);
		expect(result.ok).toBe(false);
		expect(result.reason).toContain("Limit");
	});

	it("allows host in host_only mode", () => {
		const host = { ...authPlayer, is_host: true };
		const result = canAddToQueue(
			{ ...baseConfig, queue_mode: "host_only" },
			host,
			[]
		);
		expect(result.ok).toBe(true);
	});

	it("blocks adds in seasonal pool mode", () => {
		const host = { ...authPlayer, is_host: true };
		const result = canAddToQueue(
			{
				...baseConfig,
				source_mode: "seasonal_pool",
				pool_year: "2026",
				pool_season: "summer",
				queue_mode: "disabled",
			},
			host,
			[]
		);
		expect(result.ok).toBe(false);
		expect(result.reason).toMatch(/seasonal/i);
	});
});

describe("session control gates", () => {
	const host: RatePlayer = {
		...authPlayer,
		is_host: true,
	};

	const liveCtx = {
		status: "rating" as const,
		config: baseConfig,
		me: host,
		connected: true,
		queue: [] as RateQueueItem[],
		skipVote: { enabled: true, count: 0, needed: 1, my_voted: false },
		authenticated: true,
		draftScore: 70,
		alreadyRated: false,
		busy: false,
	};

	it("enables search/play/queue/next for connected host", () => {
		expect(searchAnimeControl(liveCtx).enabled).toBe(true);
		expect(hostNextControl(liveCtx).enabled).toBe(true);
		expect(playNowControl(liveCtx).enabled).toBe(true);
		expect(
			queueAddControl({ ...liveCtx, config: { ...baseConfig, queue_mode: "host_only" } }).enabled
		).toBe(true);
		expect(submitRatingControl(liveCtx).enabled).toBe(true);
		expect(
			voteSkipControl({ ...liveCtx, config: { ...baseConfig, vote_skip: true } }).enabled
		).toBe(true);
	});

	it("keeps controls visible but disabled while reconnecting", () => {
		const offline = { ...liveCtx, connected: false };
		expect(searchAnimeControl(offline)).toMatchObject({ visible: true, enabled: false });
		expect(hostNextControl(offline)).toMatchObject({ visible: true, enabled: false });
		expect(submitRatingControl(offline)).toMatchObject({ visible: true, enabled: false });
	});

	it("hides vote skip outside rating and after voting", () => {
		const cfg = { ...baseConfig, vote_skip: true };
		expect(voteSkipControl({ ...liveCtx, config: cfg, status: "waiting" }).visible).toBe(false);
		expect(
			voteSkipControl({
				...liveCtx,
				config: cfg,
				skipVote: { enabled: true, count: 1, needed: 1, my_voted: true },
			})
		).toMatchObject({ visible: true, enabled: false });
	});

	it("blocks submit when already rated or score is empty", () => {
		expect(submitRatingControl({ ...liveCtx, alreadyRated: true }).enabled).toBe(false);
		expect(submitRatingControl({ ...liveCtx, draftScore: 0 }).enabled).toBe(false);
	});

	it("roundIdentity changes between songs", () => {
		expect(roundIdentity("rating", "a")).not.toBe(roundIdentity("rating", "b"));
		expect(roundIdentity("rating", "a")).not.toBe(roundIdentity("waiting", "a"));
	});
});

describe("score conversion", () => {
	it("converts POINT_10_DECIMAL to canonical", () => {
		expect(toCanonicalScore(7.5, "POINT_10_DECIMAL")).toBe(75);
		expect(fromCanonicalScore(75, "POINT_10_DECIMAL")).toBe(7.5);
	});

	it("converts POINT_100", () => {
		expect(toCanonicalScore(88, "POINT_100")).toBe(88);
	});
});

describe("room state reducers", () => {
	it("applies lobby and rating updates", () => {
		const lobby = applyLobbyStateUpdate(null, {
			room_id: "ABC",
			status: "rating",
			config: baseConfig,
			players: [authPlayer],
			spectators: [],
			queue: [],
			my_session_id: "s1",
			rating_data: {
				rated_count: 0,
				player_count: 1,
				ratings: {},
				session_avg: null,
			},
		} as RateRoomState);

		expect(lobby.room_id).toBe("ABC");

		const next = applyRatingUpdate(lobby, {
			rated_count: 1,
			player_count: 1,
			ratings: { s1: { rated: true, score: 80 } },
			session_avg: 80,
			my_score: 80,
			reveal_mode: "blind",
		});

		expect(next?.rating_data?.session_avg).toBe(80);
		expect(next?.rating_data?.my_score).toBe(80);
	});

	it("preserves my_session_id when omitted from a later payload", () => {
		const first = applyLobbyStateUpdate(null, {
			room_id: "ABC",
			status: "waiting",
			config: baseConfig,
			players: [authPlayer],
			spectators: [],
			queue: [],
			my_session_id: "s1",
		} as RateRoomState);

		const second = applyLobbyStateUpdate(first, {
			room_id: "ABC",
			status: "waiting",
			config: baseConfig,
			players: [authPlayer],
			spectators: [],
			queue: [],
		} as RateRoomState);

		expect(second.my_session_id).toBe("s1");
	});
});
