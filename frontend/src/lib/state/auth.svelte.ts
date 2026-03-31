export interface Role {
    id: number;
    name: string;
    slug: string;
    description?: string;
}

export interface User {
    id: number;
    uuid: string;
    name: string;
    slug: string | null;
    email: string;
    score_format?: string;
    score_format_id?: number;
    last_login_at?: string;
    roles?: Role[];
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

class Auth {
    user = $state<User | null>(null);
    loading = $state(true);

    isAuthenticated = $derived(this.user !== null);

    isAdmin = $derived.by(() => {
        if (!this.user?.roles) return false;
        const roles = this.user.roles as (string | Role)[];
        const slugs = roles.map(r => typeof r === 'string' ? r.toLowerCase() : r.slug.toLowerCase());
        return slugs.some(s => s === 'admin' || s === 'owner');
    });

    isStaff = $derived.by(() => {
        if (!this.user?.roles) return false;
        const roles = this.user.roles as (string | Role)[];
        const slugs = roles.map(r => typeof r === 'string' ? r.toLowerCase() : r.slug.toLowerCase());
        return slugs.some(s => 
            ['admin', 'owner', 'editor', 'creator', 'staff', 'moderator'].includes(s)
        );
    });
}

export const authState = new Auth();

export function setUser(user: User | null) {
    authState.user = user;
    authState.loading = false;
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
