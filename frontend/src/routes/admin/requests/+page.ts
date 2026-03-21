import type { PageLoad } from './$types';

import api from '$lib/api';

export const load: PageLoad = async () => {
	try {
		const res = await api.get('/admin/user-requests');
		return { requests: res.data.data || [] };
	} catch (error) {
		console.error("Error loading admin requests:", error);
		// Return dummy data for now as fallback if endpoint is not fully ready
		return {
			requests: [
				{
					id: 1,
					title: "Missing anime in catalog",
					content: "Jujutsu Kaisen Season 1 is missing ending 2 from the list.",
					user: { name: "luis-rodz-1", id: 1 },
					status: "pending",
					created_at: new Date().toISOString()
				}
			]
		};
	}
};
