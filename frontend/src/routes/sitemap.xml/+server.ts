import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ fetch }) => {
  try {
    // PUBLIC_API_URL should be available in SvelteKit
    const apiBase = import.meta.env?.PUBLIC_API_URL || 'http://localhost:8080/api';
    const siteUrl = import.meta.env?.APP_URL || 'https://anirank.work';

    const response = await fetch(`${apiBase}/catalog/sitemap`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch sitemap data: ${response.statusText}`);
    }

    const { data } = await response.json();

    const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <!-- Static Pages -->
  <url>
    <loc>${siteUrl}/</loc>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>${siteUrl}/animes</loc>
    <changefreq>weekly</changefreq>
    <priority>0.9</priority>
  </url>
  <url>
    <loc>${siteUrl}/songs</loc>
    <changefreq>weekly</changefreq>
    <priority>0.9</priority>
  </url>
  <url>
    <loc>${siteUrl}/artists</loc>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>

  <!-- Dynamic Content -->
  ${data.map((item: any) => `
  <url>
    <loc>${siteUrl}${item.loc}</loc>
    <lastmod>${new Date(item.lastmod).toISOString()}</lastmod>
    <changefreq>${item.changefreq || 'monthly'}</changefreq>
    <priority>${item.priority || 0.5}</priority>
  </url>`).join('')}
</urlset>`;

    return new Response(sitemap, {
      headers: {
        'Content-Type': 'application/xml',
        'Cache-Control': 'max-age=0, s-maxage=3600' // Cache for 1 hour on CDN
      }
    });
  } catch (error) {
    console.error('Sitemap generation error:', error);
    return new Response('Error generating sitemap', { status: 500 });
  }
};
