import api from '$lib/api';
import { setConfig } from '$lib/state/config.svelte';
import { setUser } from '$lib/state/auth.svelte';

// Desactivamos Server-Side Rendering (SSR)
// Esto convierte a SvelteKit en una Single Page Application (SPA) pura
// Simplifica el manejo de cookies (Sanctum) porque todas las peticiones a la API salen del navegador del cliente
export const ssr = false;
export const prerender = false; 

export const load = async () => {
    try {
        // Obtenemos los años, temporadas, géneros y formatos para toda la app
        const configRes = await api.get('/init');
        setConfig(configRes.data);

        // La hidratación de la sesión del usuario ahora se maneja en +layout.svelte onMount
        // para evitar llamadas 401 a la API si no existe el Token en el localStorage.
    } catch (e) {
        console.error("Failed to initialize Anirank config", e);
    }
    
    return {};
};
