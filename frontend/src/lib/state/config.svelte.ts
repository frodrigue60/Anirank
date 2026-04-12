export interface TaxonomyItem {
    id: number;
    name: string;
    slug: string;
}

export const configState = $state<{
    years: TaxonomyItem[];
    seasons: TaxonomyItem[];
    formats: TaxonomyItem[];
    genres: TaxonomyItem[];
    songTypes: TaxonomyItem[];
    loading: boolean;
}>({
    years: [],
    seasons: [],
    formats: [],
    genres: [],
    songTypes: [],
    loading: true
});

export function setConfig(payload: any) {
    // Handle nested data property if present (common in Go/Fiber responses)
    const data = payload.data || payload;
    
    configState.years = data.years || [];
    configState.seasons = data.seasons || [];
    configState.formats = data.formats || [];
    configState.genres = data.genres || [];
    configState.songTypes = data.song_types || [];
    configState.loading = false;
}
