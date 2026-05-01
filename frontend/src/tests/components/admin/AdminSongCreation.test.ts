import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import AdminSongCreation from '../../../routes/admin/songs/create/+page.svelte';
import { setConfig } from '$lib/state/config.svelte';
import { http, HttpResponse } from 'msw';
import { server } from '../../setup';
import { goto } from '$app/navigation';

const API_URL = 'http://localhost:8080/api';

describe('AdminSongCreation Component', () => {
    beforeEach(() => {
        // Setup config state
        setConfig({
            song_types: [
                { id: '1', name: 'Opening', slug: 'OP' },
                { id: '2', name: 'Ending', slug: 'ED' }
            ],
            years: [{ id: 1, name: '2024', slug: '2024' }],
            seasons: [{ id: 1, name: 'Winter', slug: 'winter' }]
        });
    });

    it('should disable submit button when form is invalid', () => {
        render(AdminSongCreation);
        
        const submitButton = screen.getByRole('button', { name: /create song/i });
        expect(submitButton).toBeDisabled();
    });

    it('should handle anime search and selection', async () => {
        const mockAnime = { id: 123, title: 'Boku no Hero Academia', season_id: 1, year_id: 1 };
        
        server.use(
            http.get(`${API_URL}/admin/animes`, ({ request }) => {
                const url = new URL(request.url);
                if (url.searchParams.get('search') === 'Boku') {
                    return HttpResponse.json({ data: [mockAnime] });
                }
                return HttpResponse.json({ data: [] });
            }),
            http.get(`${API_URL}/admin/songs/latest-number`, () => {
                return HttpResponse.json({ number: 7 });
            })
        );

        render(AdminSongCreation);

        const searchInput = screen.getByPlaceholderText(/search anime to link/i);
        await fireEvent.input(searchInput, { target: { value: 'Boku' } });

        // Wait for search results
        const resultButton = await screen.findByText('Boku no Hero Academia');
        await fireEvent.click(resultButton);

        // Verify selection
        expect(screen.getByText('Boku no Hero Academia')).toBeInTheDocument();
        
        // Verify auto-fill theme number
        await waitFor(() => {
            const themeNumInput = screen.getByLabelText(/theme number/i) as HTMLInputElement;
            expect(themeNumInput.value).toBe('7');
        });
    });

    it('should submit successfully when all required fields are filled', async () => {
        server.use(
            http.post(`${API_URL}/admin/songs`, async ({ request }) => {
                const body = await request.json() as any;
                if (body.song_romaji === 'Gurenge') {
                    return HttpResponse.json({ success: true, message: 'Created!' }, { status: 201 });
                }
                return HttpResponse.json({ success: false }, { status: 400 });
            })
        );

        render(AdminSongCreation);

        // Manually trigger anime selection (bypassing search for speed)
        // In a real test we'd do the search, but here we just want to test submission
        // Since anime_id is $state, we'd need to trigger it via UI.
        
        // Let's do the full flow
        server.use(
             http.get(`${API_URL}/admin/animes`, () => HttpResponse.json({ data: [{ id: 1, title: 'Anime 1' }] })),
             http.get(`${API_URL}/admin/songs/latest-number`, () => HttpResponse.json({ number: 1 }))
        );

        const searchInput = screen.getByPlaceholderText(/search anime to link/i);
        await fireEvent.input(searchInput, { target: { value: 'Anime' } });
        const result = await screen.findByText('Anime 1');
        await fireEvent.click(result);

        const titleInput = screen.getByLabelText(/title \(romaji\)/i);
        await fireEvent.input(titleInput, { target: { value: 'Gurenge' } });

        const typeSelect = screen.getByTitle(/song type/i);
        await fireEvent.change(typeSelect, { target: { value: '1' } });

        const submitButton = screen.getByRole('button', { name: /create song/i });
        expect(submitButton).not.toBeDisabled();
        await fireEvent.click(submitButton);

        await waitFor(() => {
            expect(goto).toHaveBeenCalledWith('/admin/songs');
        });
    });
});
