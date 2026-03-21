import type { PageLoad } from './$types';

import api from '$lib/api';

export const load: PageLoad = async () => {
	try {
		const res = await api.get('/admin/tournaments');
		return { tournaments: res.data.data || [] };
	} catch (error) {
		console.error("Error loading admin tournaments:", error);
		return { tournaments: [] };
	}
};
