import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request }) => {
    try {
        const { query, variables } = await request.json();
        
        const response = await fetch('https://graphql.anilist.co', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({ query, variables })
        });

        const data = await response.json();
        return json(data, { status: response.status });
    } catch (err: any) {
        console.error('AniList proxy error:', err);
        return json({ errors: [{ message: err.message || 'Internal proxy error' }] }, { status: 500 });
    }
};
