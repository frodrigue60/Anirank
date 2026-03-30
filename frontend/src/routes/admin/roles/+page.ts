import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async () => {
	try {
		// Fetch roles with their current permissions AND the full list of available permissions
		const [rolesRes, permsRes] = await Promise.all([
			api.get('/admin/roles/permissions'),
			api.get('/admin/permissions')
		]);

		return {
			roles: rolesRes.data.data || [],
			allPermissions: permsRes.data.data || []
		};
	} catch (error) {
		console.error('Error loading roles/permissions:', error);
		return {
			roles: [],
			allPermissions: []
		};
	}
};
