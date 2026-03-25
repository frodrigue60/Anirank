import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ url }) => {
	const page = url.searchParams.get('page') || '1';
	const search = url.searchParams.get('search') || '';
	const year = url.searchParams.get('year') || '';
	const season = url.searchParams.get('season') || '';
	const format = url.searchParams.get('format') || '';
	const status = url.searchParams.get('status') || '';
	
	try {
		const [res, years, seasons, formats] = await Promise.all([
			api.get(`/admin/animes?page=${page}&limit=20&search=${search}&year=${year}&season=${season}&format=${format}&status=${status}`),
			api.get('/admin/years'),
			api.get('/admin/seasons'),
			api.get('/admin/formats')
		]);

		return {
			animes: res.data.data || [],
			pagination: res.data.pagination || { current_page: 1, last_page: 1 },
			filters: {
				search,
				year,
				season,
				format,
				status
			},
			years: years.data.data || [],
			seasons: seasons.data.data || [],
			formats: formats.data.data || []
		};
	} catch (e) {
		console.error("Error loading admin animes:", e);
		return {
			animes: [],
			pagination: { current_page: 1, last_page: 1 },
			filters: { search: '', year: '', season: '', format: '', status: '' },
			years: [],
			seasons: [],
			formats: []
		};
	}
};
