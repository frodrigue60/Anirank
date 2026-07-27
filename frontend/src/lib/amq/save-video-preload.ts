import { warmSaveVideoElement } from "$lib/amq/save-video-prefetch";

export interface SavePreloadCandidate {
	song_uuid: string;
	audio_url?: string;
}

const warmedUrls = new Set<string>();
const SAVE_VIDEO_READY_EVENTS = ["loadedmetadata", "loadeddata", "canplay", "canplaythrough"] as const;

export function resetSaveVideoWarmCache() {
	warmedUrls.clear();
}

export function isSaveMediaUrlWarmed(url: string | undefined): boolean {
	return !!url && warmedUrls.has(url);
}

export function buildSavePreloadOrder(length: number, priorityIndex?: number): number[] {
	if (length <= 0) return [];
	const order: number[] = [];
	if (priorityIndex != null && priorityIndex >= 0 && priorityIndex < length) {
		order.push(priorityIndex);
	}
	for (let i = 0; i < length; i++) {
		if (i !== priorityIndex) order.push(i);
	}
	return order;
}

export function warmVideoUntilReady(
	vid: HTMLVideoElement,
	minReadyState = 1,
	timeoutMs = 8000
): Promise<boolean> {
	warmSaveVideoElement(vid);
	if (vid.readyState >= minReadyState) return Promise.resolve(true);

	return new Promise((resolve) => {
		let settled = false;
		const finish = (ok: boolean) => {
			if (settled) return;
			settled = true;
			clearTimeout(timer);
			for (const ev of SAVE_VIDEO_READY_EVENTS) {
				vid.removeEventListener(ev, onReady);
			}
			resolve(ok);
		};

		const onReady = () => {
			if (vid.readyState >= minReadyState) finish(true);
		};

		for (const ev of SAVE_VIDEO_READY_EVENTS) {
			vid.addEventListener(ev, onReady);
		}

		const timer = setTimeout(() => finish(vid.readyState >= minReadyState), timeoutMs);
		onReady();
	});
}

export async function preloadSaveMediaUrl(url: string, timeoutMs = 8000): Promise<boolean> {
	if (!url || typeof document === "undefined") return false;
	if (warmedUrls.has(url)) return true;

	const vid = document.createElement("video");
	vid.preload = "auto";
	vid.muted = true;
	vid.playsInline = true;
	vid.src = url;

	const ok = await warmVideoUntilReady(vid, 1, timeoutMs);
	if (ok) warmedUrls.add(url);

	vid.removeAttribute("src");
	vid.load();
	return ok;
}

export async function preloadSaveMediaUrls(urls: string[], maxConcurrent = 2): Promise<number> {
	const unique = [...new Set(urls.filter(Boolean))];
	let ready = 0;
	let cursor = 0;

	async function worker() {
		while (cursor < unique.length) {
			const idx = cursor++;
			if (await preloadSaveMediaUrl(unique[idx])) ready++;
		}
	}

	const workers = Array.from({ length: Math.min(maxConcurrent, unique.length) }, () => worker());
	await Promise.all(workers);
	return ready;
}

export class SaveMediaPreloadController {
	private cancelled = false;

	cancel() {
		this.cancelled = true;
	}

	async preloadRoundSlots(
		candidates: SavePreloadCandidate[],
		videos: Record<string, HTMLVideoElement>,
		order: number[],
		onProgress?: (ready: number, total: number, slotIndex: number) => void
	): Promise<{ ready: number; slotZeroReady: boolean }> {
		this.cancelled = false;
		let ready = 0;
		let slotZeroReady = false;
		const total = order.length;

		for (const slotIndex of order) {
			if (this.cancelled) break;
			const candidate = candidates[slotIndex];
			if (!candidate?.song_uuid) continue;

			const vid = videos[candidate.song_uuid];
			if (!vid) continue;

			if (candidate.audio_url && isSaveMediaUrlWarmed(candidate.audio_url)) {
				ready++;
				if (slotIndex === 0) slotZeroReady = true;
				onProgress?.(ready, total, slotIndex);
				continue;
			}

			const ok = await warmVideoUntilReady(vid);
			if (ok && candidate.audio_url) warmedUrls.add(candidate.audio_url);
			if (ok) {
				ready++;
				if (slotIndex === 0) slotZeroReady = true;
			}
			onProgress?.(ready, total, slotIndex);
		}

		return { ready, slotZeroReady };
	}

	async preloadLookaheadSlot(
		candidate: SavePreloadCandidate | undefined,
		videos: Record<string, HTMLVideoElement>
	): Promise<boolean> {
		if (!candidate?.song_uuid) return false;
		const vid = videos[candidate.song_uuid];
		if (!vid) {
			if (candidate.audio_url) return preloadSaveMediaUrl(candidate.audio_url);
			return false;
		}
		const ok = await warmVideoUntilReady(vid);
		if (ok && candidate.audio_url) warmedUrls.add(candidate.audio_url);
		return ok;
	}
}
