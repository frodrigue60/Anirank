import api from '$lib/api';
import { setConfig } from '$lib/state/config.svelte';
import { setUser } from '$lib/state/auth.svelte';

// Habilitamos Server-Side Rendering (SSR) para que los bots de Discord/Twitter puedan leer las etiquetas OG
export const ssr = true;
export const prerender = false;
export const trailingSlash = 'always';

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
