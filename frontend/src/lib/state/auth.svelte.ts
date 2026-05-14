import type { ImageSource } from './types/media';

export interface Role {
    id: number;
    name: string;
    slug: string;
    description?: string;
}

export interface UserSocialIdentity {
    provider: string;
    provider_username: string;
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
    avatar_sources?: ImageSource[];
    banner_url?: string;
    banner_sources?: ImageSource[];
    xp?: number;
    level?: number;
    social_identities?: UserSocialIdentity[];
    profile_color?: string;
    about?: string;
    is_softbanned?: boolean;
    is_shadowbanned?: boolean;
    truth_score?: number;
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

    canPublish = $derived.by(() => {
        if (!this.user?.roles) return false;
        const roles = this.user.roles as (string | Role)[];
        const slugs = roles.map(r => typeof r === 'string' ? r.toLowerCase() : r.slug.toLowerCase());
        return slugs.some(s => ['admin', 'owner', 'editor'].includes(s));
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
        if (!token || typeof token !== 'string') {
            console.error("Attempted to set an invalid token type:", typeof token);
            return;
        }
        localStorage.setItem('auth_token', token);
    }
}

export function getAuthToken(): string | null {
    if (typeof window !== 'undefined') {
        const token = localStorage.getItem('auth_token');
        
        // Basic filtering for empty or stringified null/undefined
        if (!token || token === 'null' || token === 'undefined') {
            return null;
        }

        // Structural validation for JWT (header.payload.signature)
        if (token.split('.').length !== 3) {
            console.warn("Malformed token found in localStorage. Clearing state.");
            localStorage.removeItem('auth_token');
            return null;
        }

        return token;
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
