import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import AccountSettings from '../../routes/(app)/settings/account/+page.svelte';
import { authState, setUser } from '$lib/state/auth.svelte';
import { http, HttpResponse } from 'msw';
import { server } from '../setup';
import { page } from '$app/state';

// Constant for API URL base
const API_URL = 'http://localhost:8080/api';

describe('AccountSettings Component', () => {
    beforeEach(() => {
        // Mock current user
        setUser({
            id: 1,
            uuid: 'user-uuid',
            name: 'Test Test',
            email: 'test@example.com',
            slug: 'test-user',
            social_identities: [
                { provider: 'google', provider_username: 'google_user' }
            ]
        });

        // Reset page mock
        vi.mocked(page).url = new URL('http://localhost:5173/settings/account');
    });

    it('should render "Synced" for Google and "Sync account" for others', () => {
        render(AccountSettings);

        // Google should be synced
        const googleButton = screen.getByRole('button', { name: /synced/i });
        expect(googleButton).toBeInTheDocument();

        // Discord should not be synced
        const discordButtons = screen.getAllByRole('button', { name: /sync account/i });
        expect(discordButtons.length).toBeGreaterThan(0);
    });

    it('should handle successful Discord linking callback', async () => {
        // Arrange: Mock URL with code and state
        vi.mocked(page).url = new URL('http://localhost:5173/settings/account?code=abc&state=discord_link');
        
        // Mock API response for callback and profile refresh
        server.use(
            http.post(`${API_URL}/auth/discord/callback`, () => {
                return HttpResponse.json({ success: true });
            }),
            http.get(`${API_URL}/profile`, () => {
                return HttpResponse.json({
                    data: {
                        id: 1, uuid: 'u', name: 'N', email: 'E', slug: 'S',
                        social_identities: [
                            { provider: 'google', provider_username: 'g' },
                            { provider: 'discord', provider_username: 'd' }
                        ]
                    }
                });
            })
        );

        render(AccountSettings);

        // We wait for the Svelte 5 reaction
        await waitFor(() => {
            const discordButton = screen.getByRole('button', { name: /synced/i });
            expect(discordButton).toBeInTheDocument();
        }, { timeout: 2000 });
    });

    it('should handle unlinking a social account', async () => {
        // Mock window.confirm to return true
        const confirmSpy = vi.spyOn(window, 'confirm').mockImplementation(() => true);

        // Mock API delete call
        server.use(
            http.delete(`${API_URL}/auth/google/unlink`, () => {
                return HttpResponse.json({ success: true });
            })
        );

        render(AccountSettings);

        const googleButton = screen.getByRole('button', { name: /synced/i });
        await fireEvent.click(googleButton);

        expect(confirmSpy).toHaveBeenCalled();
        
        // Wait for authState update
        await waitFor(() => {
            expect(screen.queryByRole('button', { name: /synced/i })).not.toBeInTheDocument();
            const syncButtons = screen.getAllByRole('button', { name: /sync account/i });
            // Since we started with 1 linked (Google) and 2 unlinked (Anilist, Discord),
            // after unlinking Google, we should have 3 unlinked.
            expect(syncButtons.length).toBe(3);
        });
        
        confirmSpy.mockRestore();
    });
});
