import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ url }) => {
	const page = url.searchParams.get('page') || '1';
	const search = url.searchParams.get('search') || '';
	
	try {
		const res = await api.get(`/admin/artists?page=${page}&limit=24&search=${search}`);
		
		return {
			artists: res.data.data || [],
			meta: res.data.meta || { current_page: 1, total_pages: 1 }
		};
	} catch (error) {
		console.error("Error loading admin artists:", error);
		return {
			artists: [],
			meta: { current_page: 1, total_pages: 1 }
		};
	}
};
