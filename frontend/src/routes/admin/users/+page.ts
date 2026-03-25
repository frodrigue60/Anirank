import type { PageLoad } from './$types';

import api from '$lib/api';

export const load: PageLoad = async ({ url }) => {
	const page = url.searchParams.get('page') || '1';
	const search = url.searchParams.get('search') || '';
	
	try {
		const res = await api.get(`/admin/users?page=${page}&limit=20&search=${search}`);
		return {
			users: res.data.data || [],
			pagination: res.data.pagination || { current_page: 1, last_page: 1 }
		};
	} catch (error) {
		console.error("Error loading admin users:", error);
		return {
			users: [],
			pagination: { current_page: 1, last_page: 1 }
		};
	}
};
