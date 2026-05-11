import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	const response = await resolve(event);

	// Security Headers
	const csp = [
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://*.google.com https://*.anilist.co",
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
		"img-src 'self' data: blob: http://localhost:8080 http://localhost:9000 https://*.anirank.work https://*.googleusercontent.com https://*.anilist.co https://*.discordapp.com https://*.media-amazon.com https://*.r2.dev",
		"font-src 'self' https://fonts.gstatic.com",
		"connect-src 'self' http://localhost:8080 https://api.anirank.work https://*.google.com https://*.anilist.co",
		"frame-src 'self' https://www.youtube-nocookie.com https://www.youtube.com https://*.google.com",
		"object-src 'none'",
		"base-uri 'self'",
		"form-action 'self'",
		"frame-ancestors 'none'",
		"upgrade-insecure-requests",
		"trusted-types svelte-trusted-html 'allow-duplicates'",
		"require-trusted-types-for 'script'"
	].join('; ');

	response.headers.set('Content-Security-Policy', csp);
	response.headers.set('Strict-Transport-Security', 'max-age=63072000; includeSubDomains; preload');
	response.headers.set('X-Content-Type-Options', 'nosniff');
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
	response.headers.set('Permissions-Policy', 'geolocation=(), microphone=(), camera=()');
	response.headers.set('Cross-Origin-Opener-Policy', 'same-origin');
	response.headers.set('Cross-Origin-Resource-Policy', 'cross-origin');

	return response;
};
