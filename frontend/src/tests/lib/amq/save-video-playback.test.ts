import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
	attemptSaveVideoPlayback,
	computeSaveVideoSeekTime,
	watchSaveVideoPlayback,
} from "$lib/amq/save-video-playback";

function mockVideo(overrides: Partial<HTMLVideoElement> = {}): HTMLVideoElement {
	return {
		duration: 100,
		currentTime: 0,
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

describe("save video playback helpers", () => {
	it("computes initial seek from start_percent", () => {
		expect(
			computeSaveVideoSeekTime(100, { startPercent: 0.25, mode: "initial" })
		).toBe(25);
	});

	it("computes realign seek with elapsed preview time", () => {
		expect(
			computeSaveVideoSeekTime(100, {
				startPercent: 0.1,
				mode: "realign",
				stepSeconds: 12,
				elapsedSeconds: 4,
			})
		).toBe(14);
	});

	it("returns null when duration is invalid", () => {
		expect(
			computeSaveVideoSeekTime(NaN, { startPercent: 0.2, mode: "initial" })
		).toBeNull();
	});

	it("attempts playback only when metadata is ready", () => {
		const notReady = mockVideo({ readyState: 0 });
		expect(
			attemptSaveVideoPlayback(notReady, { startPercent: 0.2, mode: "initial" })
		).toBe(false);
		expect(notReady.play).not.toHaveBeenCalled();

		const ready = mockVideo({ readyState: 1, duration: 50 });
		expect(
			attemptSaveVideoPlayback(ready, { startPercent: 0.2, mode: "initial" })
		).toBe(true);
		expect(ready.currentTime).toBe(10);
		expect(ready.play).toHaveBeenCalled();
	});

	it("retries playback when video becomes ready later", () => {
		vi.useFakeTimers();
		const vid = mockVideo({ readyState: 0, duration: 40 });
		const onStarted = vi.fn();

		watchSaveVideoPlayback(
			vid,
			() => ({ startPercent: 0.5, mode: "initial" }),
			onStarted,
			{ intervalMs: 100, maxAttempts: 10 }
		);

		expect(onStarted).not.toHaveBeenCalled();

		vid.readyState = 1;
		vi.advanceTimersByTime(100);

		expect(onStarted).toHaveBeenCalledOnce();
		expect(vid.currentTime).toBe(20);
		expect(vid.play).toHaveBeenCalled();
	});

	beforeEach(() => {
		vi.useRealTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
	});
});
