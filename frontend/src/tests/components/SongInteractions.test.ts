import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import SongPage from '../../routes/(app)/songs/[anime_slug]/[song_slug]/+page.svelte';
import { authState, setUser } from '$lib/state/auth.svelte';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../setup';

const API_URL = 'http://localhost:8080/api';

describe('Song Interactions (Optimistic UI)', () => {
    const mockSong = {
        id: 123,
        slug: 'test-song',
        title: 'Blue Bird',
        type: 'OP',
        theme_num: 1,
        anime: { title: 'Naruto', slug: 'naruto' },
        views: 1000,
        likes_count: 10,
        dislikes_count: 2,
        is_liked: false,
        is_disliked: false,
        is_favorited: false,
        average_rating: 8.5,
        song_variants: [{ version_number: 1, video: null }]
    };

    beforeEach(() => {
        setUser({
            id: 1,
            uuid: 'user-uuid',
            name: 'Test user',
            email: 'test@example.com',
            slug: 'test'
        });
    });

    it('should update Favorite UI optimistically', async () => {
        // Mock a DELAYED response from the API
        server.use(
            http.post(`${API_URL}/interactions/favorites`, async () => {
                await delay(200); // Wait 200ms
                return HttpResponse.json({
                    success: true,
                    data: { favorited: true }
                });
            })
        );

        render(SongPage, { 
            data: { song: mockSong, comments: [], related: [] } 
        });

        const favoriteButton = screen.getByLabelText(/add to favorites/i);
        const icon = favoriteButton.querySelector('.material-symbols-outlined');
        
        // Initial state: not filled
        expect(icon).not.toHaveClass('filled');

        // Click favorite
        await fireEvent.click(favoriteButton);

        // CHECK INSTANTLY (Optimistic update)
        // Even though the API is delayed by 200ms, the UI should change immediately
        expect(icon).toHaveClass('filled');
        
        // Wait for the API to actually finish (to avoid unhandled promise leaks)
        await waitFor(() => {
            expect(screen.getByLabelText(/remove from favorites/i)).toBeInTheDocument();
        });
    });

    it('should revert state if API call fails', async () => {
         // Mock API FAILURE after delay
         server.use(
            http.post(`${API_URL}/interactions/favorites`, async () => {
                await delay(50);
                return new HttpResponse(null, { status: 500 });
            })
        );

        render(SongPage, { 
            data: { song: mockSong, comments: [], related: [] } 
        });

        const favoriteButton = screen.getByLabelText(/add to favorites/i);
        const icon = favoriteButton.querySelector('.material-symbols-outlined');

        // Click favorite
        await fireEvent.click(favoriteButton);
        
        // Optimistically filled
        expect(icon).toHaveClass('filled');

        // Wait for reversal after failure
        await waitFor(() => {
            expect(icon).not.toHaveClass('filled');
        }, { timeout: 2000 });
    });
});
