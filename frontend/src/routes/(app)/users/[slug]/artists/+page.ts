import api from '$lib/api';

export const load = async ({ params, parent }: { params: { slug: string }, parent: () => Promise<any> }) => {
    const parentData = await parent();
    const user = parentData.profile;
    
    if (!user) {
        return { artists: { data: [], pagination: { current_page: 1, last_page: 1 } } };
    }

    try {
        const res = await api.post(`/users/favorites/artists`, { user_uuid: user.uuid, page: 1 });
        return {
            artists: res.data.artists || res.data || { data: [], pagination: { current_page: 1, last_page: 1 } }
        };
    } catch (e: any) {
        // Log silently instead of crashing SvelteKit's client-side router
        console.warn("Failed to load favorite artists. It may require authentication.", e.message);
        return {
            artists: { data: [], pagination: { current_page: 1, last_page: 1 } }
        };
    }
};
