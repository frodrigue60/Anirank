export interface SavePrefetchCandidate {
	song_uuid: string;
	audio_url?: string;
}

const prefetchedUrls = new Set<string>();

export function prefetchSaveCandidateMedia(candidates: SavePrefetchCandidate[] | undefined) {
	if (typeof document === "undefined" || !candidates?.length) return;

	for (const candidate of candidates) {
		const url = candidate.audio_url;
		if (!url || prefetchedUrls.has(url)) continue;
		prefetchedUrls.add(url);

		const link = document.createElement("link");
		link.rel = "prefetch";
		link.as = "video";
		link.href = url;
		document.head.appendChild(link);
	}
}

export function warmSaveVideoElement(vid: HTMLVideoElement) {
	vid.preload = "auto";
	if (vid.readyState === 0) {
		vid.load();
	}
}

export function warmSaveVideoElements(videos: Record<string, HTMLVideoElement>) {
	for (const vid of Object.values(videos)) {
		warmSaveVideoElement(vid);
	}
}

export function resetSaveMediaPrefetchCache() {
	prefetchedUrls.clear();
}
