export const ssr = false;

import api from '$lib/api';

export const load = async ({ params }: { params: { slug: string } }) => {
    const slug = params.slug;

    try {
        const userResponse = await api.get(`/users/${slug}`);
        const user = userResponse.data.data;

        if (!user) {
            return {
                profile: null
            };
        }

        return {
            profile: user
        };
    } catch (e: any) {
        console.error("Failed to load user profile layout", e);
        return {
            profile: null
        };
    }
};
