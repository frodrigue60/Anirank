import api from '$lib/api';

export const load = async ({ params, parent }: { params: { slug: string }, parent: () => Promise<any> }) => {
    const parentData = await parent();
    if (!parentData.profile) {
        return { playlists: [] };
    }

    try {
        const res = await api.get(`/users/${params.slug}/playlists`);
        return {
            playlists: res.data.playlists || []
        };
    } catch (e: any) {
        console.error("Failed to load user playlists", e);
        return {
            playlists: []
        };
    }
};
