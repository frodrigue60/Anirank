export interface User {
    id: number;
    name: string;
    slug: string | null;
    email: string;
    score_format?: string;
    score_format_id?: number;
    last_login_at?: string;
    roles?: { id: number; name: string }[];
    avatar_url?: string;
    banner_url?: string;
    xp?: number;
    level?: number;
    anilist_id?: number | null;
    anilist_username?: string | null;
    google_id?: string | null;
    google_email?: string | null;
    profile_color?: string;
    about?: string;
}

export const authState = $state<{
    user: User | null;
    isAuthenticated: boolean;
    loading: boolean;
    isAdmin: boolean;
    isStaff: boolean;
}>({
    user: null,
    isAuthenticated: false,
    loading: true,
    isAdmin: false,
    isStaff: false
});

export function setUser(user: User | null) {
    authState.user = user;
    authState.isAuthenticated = !!user;
    authState.loading = false;
    
    // Evaluate roles
    if (user && user.roles) {
        authState.isAdmin = user.roles.some((r: any) => 
            (r.name && r.name.toLowerCase() === 'admin') || 
            (r.slug && r.slug.toLowerCase() === 'admin')
        );
        authState.isStaff = user.roles.some((r: any) => {
            const n = r.name?.toLowerCase();
            const s = r.slug?.toLowerCase();
            return ['admin', 'editor', 'creator'].includes(n) || ['admin', 'editor', 'creator'].includes(s);
        });
    } else {
        authState.isAdmin = false;
        authState.isStaff = false;
    }
}


// Persist and Retrieve token
export function setAuthToken(token: string) {
    if (typeof window !== 'undefined') {
        localStorage.setItem('auth_token', token);
    }
}

export function getAuthToken(): string | null {
    if (typeof window !== 'undefined') {
        return localStorage.getItem('auth_token');
    }
    return null;
}

export function removeAuthToken() {
    if (typeof window !== 'undefined') {
        localStorage.removeItem('auth_token');
    }
}

export function logout() {
    removeAuthToken();
    setUser(null);
}
