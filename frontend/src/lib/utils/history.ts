const STORAGE_KEY = 'anirank_search_history_v2';
const MAX_ITEMS = 10;

export interface HistoryItem {
    id: string; // Unique key (e.g., "anime:slug" or "query:text")
    type: 'anime' | 'song' | 'artist' | 'user' | 'query';
    label: string;
    description?: string; 
    slug: string;
    image?: string;
    animeSlug?: string; // Required for song URLs
}

export function getSearchHistory(): HistoryItem[] {
    if (typeof window === 'undefined') return [];
    try {
        const stored = localStorage.getItem(STORAGE_KEY);
        return stored ? JSON.parse(stored) : [];
    } catch (e) {
        console.error('Failed to parse search history', e);
        return [];
    }
}

export function saveToSearchHistory(item: HistoryItem): HistoryItem[] {
    if (!item || !item.label) return getSearchHistory();
    
    let history = getSearchHistory();
    
    // Remove if already exists (to move to top)
    history = history.filter(h => h.id !== item.id);
    
    // Add to top
    history.unshift(item);
    
    // Limit
    if (history.length > MAX_ITEMS) {
        history = history.slice(0, MAX_ITEMS);
    }
    
    localStorage.setItem(STORAGE_KEY, JSON.stringify(history));
    return history;
}

export function removeFromSearchHistory(id: string): HistoryItem[] {
    let history = getSearchHistory();
    history = history.filter(h => h.id !== id);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(history));
    return history;
}

export function clearSearchHistory(): void {
    localStorage.removeItem(STORAGE_KEY);
}
