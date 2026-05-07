import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// Load repo-root .env (APP_URL, API_URL, PUBLIC_API_URL, etc.)
export default defineConfig({
	envDir: path.resolve(__dirname, '..'),
	envPrefix: ['VITE_', 'PUBLIC_'],
	plugins: [tailwindcss(), sveltekit()],
	server: {
		port: 5173,
		host: true
	},
	optimizeDeps: {
		exclude: ['lucide-svelte']
	}
});
