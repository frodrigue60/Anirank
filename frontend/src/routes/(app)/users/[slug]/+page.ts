import api from '$lib/api';

export const load = async ({ params, parent }: { params: { slug: string }, parent: () => Promise<any> }) => {
    // Await the parent layout data (user profile)
    const parentData = await parent();
    const user = parentData.profile;

    if (!user) {
        return {
            initialSongs: [],
            artists: []
        };
    }

    try {
        // Fetch only a small preview of data for the Overview tab
        const [songsData, artistsData] = await Promise.all([
            api.get(`/users/${params.slug}/favorites/themes`, { params: { page: 1 } })
                .then(res => res.data.data || [])
                .catch(e => {
                    console.warn("Failed to load favorite songs preview.", e.message);
                    return [];
                }),
            api.get(`/users/${params.slug}/favorites/artists`, { params: { page: 1 } })
                .then(res => res.data.data || [])
                .catch(e => {
                    console.warn("Failed to load favorite artists preview.", e.message);
                    return [];
                })
        ]);

        return {
            initialSongs: songsData,
            artists: artistsData
        };
    } catch (e: any) {
        console.error("Failed to load overview data", e);
        return {
            initialSongs: [],
            artists: []
        };
    }
};
