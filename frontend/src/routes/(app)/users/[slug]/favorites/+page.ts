import api from '$lib/api';

export const load = async ({ params, parent }: { params: { slug: string }, parent: () => Promise<any> }) => {
    const parentData = await parent();
    const user = parentData.profile;
    
    if (!user) {
        return { songs: { data: [], pagination: { current_page: 1, last_page: 1 } } };
    }

    try {
        const res = await api.get(`/users/${params.slug}/favorites/themes`, { params: { page: 1 } });
        return {
            songs: res.data || { data: [], pagination: { current_page: 1, last_page: 1 } }
        };
    } catch (e: any) {
        // Log silently instead of crashing SvelteKit's client-side router
        console.warn("Failed to load favorite songs.", e.message);
        return {
            songs: { data: [], pagination: { current_page: 1, last_page: 1 } }
        };
    }
};
