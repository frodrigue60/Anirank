import type { RequestHandler } from './$types';
import { PUBLIC_API_URL } from "$lib/api";

export const prerender = false;

export const GET: RequestHandler = async ({ fetch }) => {
  try {
    const apiBase = PUBLIC_API_URL;
    // Fetch the pre-constructed XML from the backend
    const response = await fetch(`${apiBase}/catalog/sitemap.xml`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch sitemap from backend: ${response.statusText}`);
    }

    const xml = await response.text();

    return new Response(xml, {
      headers: {
        'Content-Type': 'application/xml',
        'Cache-Control': 'public, max-age=3600' // 1 hour
      }
    });
  } catch (error) {
    console.error('Sitemap proxy error:', error);
    return new Response('Error loading sitemap', { status: 500 });
  }
};
