import { describe, it, expect, beforeEach } from 'vitest';
import { authState, setUser, logout, setAuthToken, getAuthToken, removeAuthToken } from './auth.svelte';

describe('Auth State (Svelte 5 Runes)', () => {
    beforeEach(() => {
        logout();
        localStorage.clear();
    });

    it('should initialize with null user and loading false after logout', () => {
        expect(authState.user).toBeNull();
        expect(authState.loading).toBe(false);
        expect(authState.isAuthenticated).toBe(false);
    });

    it('should update state when setUser is called', () => {
        const mockUser = {
            id: 1,
            uuid: 'user-uuid',
            name: 'Test User',
            email: 'test@example.com',
            slug: 'test-user'
        };

        setUser(mockUser);

        expect(authState.user).toEqual(mockUser);
        expect(authState.loading).toBe(false);
        expect(authState.isAuthenticated).toBe(true);
    });

    it('should correctly derive isAdmin status', () => {
        setUser({
            id: 1,
            uuid: 'admin-uuid',
            name: 'Admin',
            email: 'admin@example.com',
            slug: 'admin',
            roles: [{ id: 1, name: 'Admin', slug: 'admin' }]
        });

        expect(authState.isAdmin).toBe(true);
        expect(authState.isStaff).toBe(true);
    });

    it('should correctly derive isStaff status for editor', () => {
        setUser({
            id: 1,
            uuid: 'editor-uuid',
            name: 'Editor',
            email: 'editor@example.com',
            slug: 'editor',
            roles: [{ id: 2, name: 'Editor', slug: 'editor' }]
        });

        expect(authState.isAdmin).toBe(false);
        expect(authState.isStaff).toBe(true);
    });

    it('should handle logout', () => {
        setUser({ id: 1, uuid: 'u', name: 'N', email: 'E', slug: 'S' });
        setAuthToken('header.payload.signature');
        
        logout();

        expect(authState.user).toBeNull();
        expect(authState.isAuthenticated).toBe(false);
        expect(getAuthToken()).toBeNull();
    });

    it('should persist and retrieve token in localStorage', () => {
        const token = 'header.payload.signature';
        setAuthToken(token);
        expect(getAuthToken()).toBe(token);
        
        removeAuthToken();
        expect(getAuthToken()).toBeNull();
    });
});
