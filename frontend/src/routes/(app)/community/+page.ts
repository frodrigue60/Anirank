import type { PageLoad } from './$types';
import { getActivePartners } from '$lib/api';

export const load: PageLoad = async () => {
    try {
        const partners = await getActivePartners();
        return {
            partners
        };
    } catch (e) {
        console.error("Failed to load communities", e);
        return {
            partners: []
        };
    }
};
