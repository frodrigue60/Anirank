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
            api.post(`/users/favorites/themes`, { user_uuid: user.uuid, page: 1 })
                .then(res => res.data.data || [])
                .catch(e => {
                    console.warn("Failed to load favorite songs preview. It may require authentication.", e.message);
                    return [];
                }),
            api.post(`/users/favorites/artists`, { user_uuid: user.uuid, page: 1 })
                .then(res => res.data.data || [])
                .catch(e => {
                    console.warn("Failed to load favorite artists preview. It may require authentication.", e.message);
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
