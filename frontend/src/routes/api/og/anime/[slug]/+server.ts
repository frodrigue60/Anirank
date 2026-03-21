import satori from 'satori';
import { Resvg } from '@resvg/resvg-js';
import { html } from 'satori-html';
import fs from 'fs';
import path from 'path';

export async function GET({ params, fetch }) {
  // Load the font from static directory
  const fontPath = path.resolve('static/fonts/Inter-Bold.ttf');
  const fontData = fs.readFileSync(fontPath);

  const slug = params.slug;

  try {
    // We use the absolute URL for backend call server-side.
    // In production, this should point to your internal backend service URL.
    const apiUrl = import.meta.env.VITE_API_URL || "http://127.0.0.1:8080/api";
    const res = await fetch(`${apiUrl}/animes/${slug}`);
    
    if (!res.ok) {
      return new Response('Not found', { status: 404 });
    }

    const json = await res.json();
    const anime = json.data || json.anime || json;

    if (!anime || !anime.id) {
      return new Response('Not found', { status: 404 });
    }

    // Process genres
    const genresList = anime.genres ? anime.genres.map((g: any) => g.name).join(' • ') : '';

    const bannerStyle = `width: 100%; height: 100%; display: flex; ${
        anime.banner_url || anime.cover_image_url ? `background-image: url('${anime.banner_url || anime.cover_image_url}'); background-size: cover; background-position: center; opacity: 0.3;` : 'background-color: #1a1a24;'
    }`;

    // HTML template optimized for Satori
    const htmlString = `
<div style="
  display: flex;
  height: 100%;
  width: 100%;
  background-color: #0f0f13;
  color: white;
  font-family: 'Inter';
  position: relative;
  overflow: hidden;
">
  <!-- Background Banner -->
  <div style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; display: flex;">
    <div style="${bannerStyle}"></div>
  </div>

  <!-- Content Container -->
  <div style="
    position: absolute;
    top: 0;
    left: 0;
    width: 65%;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: 60px;
    background-image: linear-gradient(to right, #0f0f13 65%, transparent);
  ">
    <!-- Meta info -->
    <div style="
      display: flex;
      font-size: 28px;
      font-weight: bold;
      color: rgba(255, 255, 255, 0.7);
      margin-bottom: 25px;
      text-transform: uppercase;
      letter-spacing: 2px;
    ">
      ${anime.season ? anime.season : ''} ${anime.year ? anime.year : ''} ${anime.episodes ? `• ${anime.episodes} EP` : ''}
    </div>

    <!-- Title -->
    <h1 style="
      font-size: 72px;
      font-weight: 800;
      margin: 0 0 30px 0;
      line-height: 1.1;
      display: flex;
      flex-direction: column;
    ">
      ${anime.title}
    </h1>

    <!-- Score & Genres -->
    <div style="display: flex; flex-direction: column; margin-top: auto;">
      <div style="display: flex; align-items: center; font-size: 54px; font-weight: bold; margin-bottom: 15px;">
        <span style="color: #FFD700; margin-right: 15px;">★</span>
        <span style="color: white;">${anime.average_score ? anime.average_score + '%' : 'N/A'}</span>
      </div>
      <div style="display: flex; font-size: 26px; color: rgba(255, 255, 255, 0.6); font-weight: bold;">
        ${genresList}
      </div>
    </div>
  </div>

  <!-- Cover Image (Right Side) -->
  <div style="
    position: absolute;
    right: 70px;
    top: 50%;
    transform: translateY(-50%);
    width: 360px;
    height: 520px;
    display: flex;
    flex-direction: column;
    border-radius: 20px;
    background-color: #1a1a24;
    box-shadow: 0 20px 40px rgba(0,0,0,0.5);
    overflow: hidden;
    border: 4px solid rgba(255, 255, 255, 0.1);
  ">
    ${anime.cover_image_url ? `<img src="${anime.cover_image_url}" style="width: 100%; height: 100%; object-fit: cover;" />` : ''}
  </div>

  <!-- Branding -->
  <div style="
    position: absolute;
    bottom: 40px;
    right: 70px;
    display: flex;
    font-size: 36px;
    font-weight: 900;
    color: white;
    letter-spacing: 2px;
  ">
    <span style="color: #ff4e50;">ANI</span>RANK
  </div>
</div>
    `;

    const element = html(htmlString);

    const svg = await satori(element, {
      width: 1200,
      height: 630,
      fonts: [
        {
          name: 'Inter',
          data: fontData,
          weight: 700,
          style: 'normal',
        },
      ],
    });

    const resvg = new Resvg(svg, {
      fitTo: { mode: 'width', value: 1200 },
    });

    const pngData = resvg.render();
    const pngBuffer = pngData.asPng();

    return new Response(new Uint8Array(pngBuffer), {
      headers: {
        'Content-Type': 'image/png',
        'Cache-Control': 'public, max-age=86400, s-maxage=86400',
      },
    });
  } catch (error: any) {
    console.error('Error generating anime OG image:', error);
    return new Response(error.stack || error.message, { status: 500 });
  }
}
