import type { PageLoad } from './$types';

import api from '$lib/api';

export const load: PageLoad = async ({ url }) => {
	const page = url.searchParams.get('page') || '1';
	const search = url.searchParams.get('search') || '';
	const anime = url.searchParams.get('anime') || '';
	const status = url.searchParams.get('status') || '';
	
	try {
		const res = await api.get(`/admin/songs?page=${page}&limit=20&search=${search}&anime=${anime}&status=${status}`);
		return {
			songs: res.data.data || [],
			meta: {
				...(res.data.meta || { current_page: 1, total_pages: 1 }),
				anime,
				status
			}
		};
	} catch (error) {
		console.error("Error loading admin songs:", error);
		return {
			songs: [],
			meta: { current_page: 1, total_pages: 1 }
		};
	}
};
