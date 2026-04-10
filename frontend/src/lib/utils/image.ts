export interface ImageSource {
	url: string;
	width: number;
}

/**
 * Generates a srcset string from an array of image sources.
 * @param sources Array of image sources (URL and width)
 * @returns A string in the format "url1 width1w, url2 width2w, ..."
 */
export function getSrcset(sources: ImageSource[] | undefined | null): string | undefined {
	if (!sources || sources.length === 0) return undefined;
	return sources
		.map((s) => `${s.url} ${s.width}w`)
		.join(', ');
}
