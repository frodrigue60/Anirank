import type { PageLoad } from './$types';
import api from '$lib/api';
import { logLoadError } from '$lib/logger';

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
                logLoadError('admin/artists/[id] — songs', err);
            }
        }

		return {
			artist,
            songs
		};
	} catch (error) {
		logLoadError('admin/artists/[id]', error);
		return {
			artist: null,
            songs: []
		};
	}
};
