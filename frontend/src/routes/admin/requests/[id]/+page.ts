import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ params }) => {
	try {
		const res = await api.get(`/admin/user-requests/${params.id}`);
		return { request: res.data.data };
	} catch (error) {
		console.error("Error loading user request:", error);
		return { request: null };
	}
};
