import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import TagsInput from '../../../lib/components/admin/TagsInput.svelte';
import { http, HttpResponse } from 'msw';
import { server } from '../../setup';

const API_URL = 'http://localhost:8080/api';

describe('TagsInput Component', () => {
    it('should add tag on Enter', async () => {
        // Mock the checkExists call
        server.use(
            http.get(`${API_URL}/tags`, () => HttpResponse.json({ data: [] }))
        );

        render(TagsInput, { value: '', endpoint: '/tags', entityName: 'Tag' });

        const input = screen.getByPlaceholderText(/add tags/i);
        await fireEvent.input(input, { target: { value: 'NewTag' } });
        await fireEvent.keyDown(input, { key: 'Enter' });

        await waitFor(() => {
            expect(screen.getByText('NewTag')).toBeInTheDocument();
        });
    });

    it('should remove tag when clicking remove button', async () => {
        render(TagsInput, { value: 'Tag1, Tag2', endpoint: '/tags' });

        expect(screen.getByText('Tag1')).toBeInTheDocument();
        expect(screen.getByText('Tag2')).toBeInTheDocument();
        
        const removeButtons = screen.getAllByLabelText(/remove tag/i);
        await fireEvent.click(removeButtons[0]);

        expect(screen.queryByText('Tag1')).not.toBeInTheDocument();
        expect(screen.getByText('Tag2')).toBeInTheDocument();
    });

    it('should show suggestions when typing', async () => {
        server.use(
            http.get(`${API_URL}/tags`, ({ request }) => {
                const url = new URL(request.url);
                if (url.searchParams.get('search') === 'Sug') {
                    return HttpResponse.json({ data: [{ id: 1, name: 'SugTag' }] });
                }
                return HttpResponse.json({ data: [] });
            })
        );

        render(TagsInput, { value: '', endpoint: '/tags' });

        const input = screen.getByPlaceholderText(/add tags/i);
        await fireEvent.input(input, { target: { value: 'Sug' } });

        // Wait for debounce (300ms) and API response
        const suggestion = await screen.findByText('SugTag', {}, { timeout: 2000 });
        expect(suggestion).toBeInTheDocument();

        await fireEvent.click(suggestion);
        
        await waitFor(() => {
            expect(screen.getByText('SugTag')).toBeInTheDocument();
        });
    });

    it('should handle comma as a separator', async () => {
        server.use(
            http.get(`${API_URL}/tags`, () => HttpResponse.json({ data: [] }))
        );

        render(TagsInput, { value: '', endpoint: '/tags' });

        const input = screen.getByPlaceholderText(/add tags/i);
        await fireEvent.input(input, { target: { value: 'TagA' } });
        await fireEvent.keyDown(input, { key: ',' });

        await waitFor(() => {
            expect(screen.getByText('TagA')).toBeInTheDocument();
        });
    });
});
