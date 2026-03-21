import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
	// Roles are usually static, but we'll fetch from an API in a real scenario
	return {
		roles: [
			{
				id: 1,
				name: 'admin',
				description: 'Full access to all system features including configuration and user management.',
				users_count: 2,
				color: 'rose'
			},
			{
				id: 2,
				name: 'editor',
				description: 'Can manage catalog items (Animes, Songs, Artists) and handle taxonomies.',
				users_count: 5,
				color: 'blue'
			},
			{
				id: 3,
				name: 'creator',
				description: 'Can submit new catalog items but requires approval. Can manage tournaments.',
				users_count: 12,
				color: 'emerald'
			},
			{
				id: 4,
				name: 'user',
				description: 'Default role for all registered accounts. Can vote, comment, and create playlists.',
				users_count: 1342,
				color: 'gray'
			}
		]
	};
};
