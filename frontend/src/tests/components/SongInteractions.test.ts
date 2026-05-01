import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import SongPage from '../../routes/(app)/songs/[anime_slug]/[song_slug]/+page.svelte';
import RatingModal from '$lib/components/RatingModal.svelte';
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
            slug: 'test',
            score_format: 'POINT_100'
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
        const svg = favoriteButton.querySelector('svg');
        
        // Initial state: not filled
        expect(svg).not.toHaveClass('fill-pink-500');

        // Click favorite
        await fireEvent.click(favoriteButton);

        // CHECK INSTANTLY (Optimistic update)
        // Even though the API is delayed by 200ms, the UI should change immediately
        expect(svg).toHaveClass('fill-pink-500');
        
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
        const svg = favoriteButton.querySelector('svg');

        // Click favorite
        await fireEvent.click(favoriteButton);
        
        // Optimistically filled
        expect(svg).toHaveClass('fill-pink-500');

        // Wait for reversal after failure
        await waitFor(() => {
            expect(svg).not.toHaveClass('fill-pink-500');
        }, { timeout: 2000 });
    });

    it('should call onRated optimistically in RatingModal', async () => {
        const onRated = vi.fn();
        server.use(
            http.post(`${API_URL}/interactions/ratings`, async () => {
                await delay(100);
                return HttpResponse.json({ success: true, data: { rating: 80, average: 85 } });
            })
        );

        render(RatingModal, { 
            show: true, 
            song: { id: 1, user_rating: 0 }, 
            onClose: vi.fn(), 
            onRated 
        });

        const slider = screen.getByRole('slider');
        await fireEvent.input(slider, { target: { value: '80' } });
        
        const submit = screen.getByRole('button', { name: /submit rating/i });
        await fireEvent.click(submit);

        // Should have been called once with optimistic value
        expect(onRated).toHaveBeenCalledWith(expect.objectContaining({ rating: 80 }));
        
        // Wait for final call
        await waitFor(() => {
            expect(onRated).toHaveBeenCalledWith(expect.objectContaining({ rating: 80, average: 85 }));
        });
    });
});
