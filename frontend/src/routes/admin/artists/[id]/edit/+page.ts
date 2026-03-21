import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ params }) => {
	try {
		const res = await api.get(`/admin/artists/${params.id}`);
		return {
			artist: res.data.data
		};
	} catch (error) {
		console.error("Error loading artist for edit:", error);
		return {
			artist: null
		};
	}
};
