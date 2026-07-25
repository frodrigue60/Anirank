import { warmSaveVideoElement } from "$lib/amq/save-video-prefetch";

export type SaveVideoPlaybackMode = "initial" | "realign";

export interface SaveVideoPlaybackContext {
	startPercent: number;
	mode: SaveVideoPlaybackMode;
	stepSeconds?: number;
	elapsedSeconds?: number;
}

export function computeSaveVideoSeekTime(
	duration: number,
	ctx: SaveVideoPlaybackContext
): number | null {
	if (!Number.isFinite(duration) || duration <= 0) return null;

	let targetTime = duration * ctx.startPercent;
	if (ctx.mode === "realign" && ctx.stepSeconds != null && ctx.elapsedSeconds != null) {
		targetTime += Math.max(0, ctx.elapsedSeconds);
	}

	return Math.min(Math.max(0, targetTime), duration - 0.05);
}

export function attemptSaveVideoPlayback(
	vid: HTMLVideoElement,
	ctx: SaveVideoPlaybackContext
): boolean {
	warmSaveVideoElement(vid);
	if (vid.readyState < 1) return false;

	const target = computeSaveVideoSeekTime(vid.duration, ctx);
	if (target == null) return false;

	if (ctx.mode === "initial" || Math.abs(vid.currentTime - target) > 0.25) {
		vid.currentTime = target;
	}

	vid.play().catch(() => {});
	return true;
}

const SAVE_VIDEO_READY_EVENTS = ["loadedmetadata", "loadeddata", "canplay", "canplaythrough"] as const;

export function watchSaveVideoPlayback(
	vid: HTMLVideoElement,
	getContext: () => SaveVideoPlaybackContext | null,
	onStarted: () => void,
	options?: { maxAttempts?: number; intervalMs?: number }
): () => void {
	let cancelled = false;
	let started = false;
	const maxAttempts = options?.maxAttempts ?? 30;
	const intervalMs = options?.intervalMs ?? 200;
	let attempts = 0;

	const cleanup = () => {
		if (cancelled) return;
		cancelled = true;
		clearInterval(timer);
		for (const ev of SAVE_VIDEO_READY_EVENTS) {
			vid.removeEventListener(ev, onReadyEvent);
		}
	};

	const tryOnce = (): boolean => {
		if (cancelled || started) return true;

		const ctx = getContext();
		if (!ctx) {
			cleanup();
			return true;
		}

		if (!attemptSaveVideoPlayback(vid, ctx)) return false;

		started = true;
		onStarted();
		cleanup();
		return true;
	};

	const onReadyEvent = () => {
		tryOnce();
	};

	for (const ev of SAVE_VIDEO_READY_EVENTS) {
		vid.addEventListener(ev, onReadyEvent);
	}

	const timer = setInterval(() => {
		attempts++;
		if (tryOnce() || attempts >= maxAttempts) {
			cleanup();
		}
	}, intervalMs);

	tryOnce();

	return cleanup;
}
