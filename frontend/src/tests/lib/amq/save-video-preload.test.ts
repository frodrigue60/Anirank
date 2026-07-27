import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
	buildSavePreloadOrder,
	resetSaveVideoWarmCache,
	SaveMediaPreloadController,
} from "$lib/amq/save-video-preload";

function mockVideo(overrides: Partial<HTMLVideoElement> = {}): HTMLVideoElement {
	return {
		readyState: 0,
		preload: "metadata",
		load: vi.fn(),
		play: vi.fn().mockResolvedValue(undefined),
		pause: vi.fn(),
		addEventListener: vi.fn(),
		removeEventListener: vi.fn(),
		...overrides,
	} as unknown as HTMLVideoElement;
}

describe("save video preload", () => {
	beforeEach(() => {
		resetSaveVideoWarmCache();
	});

	it("orders priority slot first", () => {
		expect(buildSavePreloadOrder(4, 2)).toEqual([2, 0, 1, 3]);
		expect(buildSavePreloadOrder(4)).toEqual([0, 1, 2, 3]);
	});

	it("preloads round slots sequentially and reports slot zero readiness", async () => {
		const controller = new SaveMediaPreloadController();
		const listeners: Record<string, () => void> = {};
		const v0 = mockVideo({
			readyState: 1,
			addEventListener: vi.fn((ev: string, fn: () => void) => {
				listeners[ev] = fn;
			}),
		});
		const v1 = mockVideo({ readyState: 1 });

		const result = await controller.preloadRoundSlots(
			[
				{ song_uuid: "a", audio_url: "/v/a.mp4" },
				{ song_uuid: "b", audio_url: "/v/b.mp4" },
			],
			{ a: v0, b: v1 },
			[0, 1]
		);

		expect(result.slotZeroReady).toBe(true);
		expect(result.ready).toBe(2);
	});

	afterEach(() => {
		vi.useRealTimers();
	});
});
