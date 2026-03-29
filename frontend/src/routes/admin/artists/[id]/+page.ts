import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ params }) => {
	try {
        // Fetch Artist
		const artistRes = await api.get(`/admin/artists/${params.id}`);
        const artist = artistRes.data.data;

        let songs = [];
        if (artist?.slug) {
            try {
                // We use the catalog endpoint to fetch their latest songs
                const songsRes = await api.get(`/artists/${artist.slug}/songs?limit=10&sort=recent`);
                songs = songsRes.data.data || [];
            } catch (err) {
                console.error("Failed to load artist songs:", err);
            }
        }

		return {
			artist,
            songs
		};
	} catch (error) {
		console.error("Error loading artist details:", error);
		return {
			artist: null,
            songs: []
		};
	}
};
