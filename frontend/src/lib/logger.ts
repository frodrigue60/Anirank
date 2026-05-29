/**
 * Centralized SSR logger.
 *
 * Axios dumps the entire ClientRequest object to stderr when a request fails,
 * which produces hundreds of lines of noise for every 404 hit by bots/crawlers
 * that probe numeric-ID routes (e.g. /artists/10/songs, /producers/7).
 *
 * Rules:
 *  - HTTP 404  → warn with a single compact line (expected: bots, bad links)
 *  - HTTP 4xx  → warn with status + message
 *  - HTTP 5xx+ → full error (real problem, needs attention)
 *  - Non-Axios → full error
 */
export function logLoadError(context: string, err: unknown): void {
	if (!isAxiosError(err)) {
		console.error(`[${context}]`, err);
		return;
	}

	const status = err.response?.status;
	const message = err.response?.data?.message ?? err.message;
	const url = err.config?.url ?? '(unknown)';

	if (status === 404) {
		// Suppress full dump — bots hitting non-existent numeric-ID routes
		console.warn(`[${context}] 404 Not Found — ${url}`);
		return;
	}

	if (status && status >= 400 && status < 500) {
		console.warn(`[${context}] ${status} ${message} — ${url}`);
		return;
	}

	// 5xx or network errors: keep full detail
	console.error(`[${context}] ${status ?? 'ERR'} ${message} — ${url}`, err);
}

// Minimal Axios error type guard (avoids importing axios in the logger)
interface AxiosLike {
	isAxiosError: boolean;
	response?: { status: number; data?: { message?: string } };
	config?: { url?: string };
	message: string;
}

function isAxiosError(err: unknown): err is AxiosLike {
	return typeof err === 'object' && err !== null && (err as AxiosLike).isAxiosError === true;
}
