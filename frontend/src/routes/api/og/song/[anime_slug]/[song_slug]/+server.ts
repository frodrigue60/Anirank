import satori from 'satori';
import { Resvg } from '@resvg/resvg-js';
import { html } from 'satori-html';
import fs from 'fs';
import path from 'path';

export async function GET({ params, fetch }) {
  // Load the font from static directory
  const fontPath = path.resolve('static/fonts/Inter-Bold.ttf');
  const fontData = fs.readFileSync(fontPath);

  const { anime_slug, song_slug } = params;

  try {
    const apiUrl = import.meta.env.VITE_API_URL || "http://127.0.0.1:8080/api";
    const res = await fetch(`${apiUrl}/animes/${anime_slug}/songs/${song_slug}`);
    
    if (!res.ok) {
      return new Response('Not found', { status: 404 });
    }

    const json = await res.json();
    const song = json.song || json.data || json;

    if (!song || !song.id) {
      return new Response('Not found', { status: 404 });
    }

    // Process title
    const songName = song.title || "Unknown Title";
    // Process artists
    const artistNames = song.artists ? song.artists.map((a: any) => a.name).join(', ') : 'Unknown Artist';
    
    const anime = song.anime || {};

    const bannerStyle = `width: 100%; height: 100%; display: flex; ${
        anime.banner_url || anime.cover_image_url ? `background-image: url('${anime.banner_url || anime.cover_image_url}'); background-size: cover; background-position: center; opacity: 0.2; filter: blur(10px);` : 'background-color: #1a1a24;'
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
  <div style="position: absolute; top: -20px; left: -20px; width: 105%; height: 105%; display: flex;">
    <div style="${bannerStyle}"></div>
  </div>

  <!-- Content Container -->
  <div style="
    display: flex;
    flex-direction: column;
    justify-content: center;
    width: 100%;
    height: 100%;
    padding: 80px;
    z-index: 10;
  ">
    <!-- Song Type -->
    <div style="
      display: flex;
      font-size: 32px;
      font-weight: 900;
      color: #ff4e50;
      margin-bottom: 25px;
      text-transform: uppercase;
      letter-spacing: 4px;
    ">
      ${song.type || 'Song'} ${song.theme_num ? song.theme_num : ''}
    </div>

    <!-- Title -->
    <h1 style="
      font-size: 84px;
      font-weight: 900;
      margin: 0 0 20px 0;
      line-height: 1.1;
      display: flex;
      flex-direction: column;
      text-shadow: 0 4px 12px rgba(0,0,0,0.5);
    ">
      ${songName}
    </h1>

    <!-- Artists -->
    <div style="display: flex; font-size: 36px; color: rgba(255, 255, 255, 0.8); font-weight: bold; margin-bottom: auto; text-shadow: 0 2px 8px rgba(0,0,0,0.5);">
      ${artistNames}
    </div>

    <!-- Bottom Bar: Anime Name & Score -->
    <div style="display: flex; justify-content: space-between; align-items: flex-end; width: 100%; margin-top: 60px;">
      <div style="display: flex; flex-direction: column;">
        <div style="font-size: 24px; color: rgba(255, 255, 255, 0.5); font-weight: bold; text-transform: uppercase; letter-spacing: 2px; margin-bottom: 10px;">
          Featured In
        </div>
        <div style="font-size: 42px; font-weight: 800; color: white;">
          ${anime.title || ''}
        </div>
      </div>
      
      <!-- Score -->
      <div style="display: flex; flex-direction: column; align-items: flex-end;">
        <div style="font-size: 24px; color: rgba(255, 255, 255, 0.5); font-weight: bold; text-transform: uppercase; letter-spacing: 2px; margin-bottom: 10px;">
          Community Score
        </div>
        <div style="display: flex; align-items: center; font-size: 64px; font-weight: 900;">
          <span style="color: #FFD700; margin-right: 15px;">★</span>
          <span style="color: white; padding-top: 5px;">${song.average_score ? song.average_score + '%' : 'N/A'}</span>
        </div>
      </div>
    </div>
  </div>

  <!-- Branding Top Right -->
  <div style="
    position: absolute;
    top: 50px;
    right: 50px;
    display: flex;
    font-size: 42px;
    font-weight: 900;
    color: white;
    letter-spacing: 2px;
    z-index: 20;
    text-shadow: 0 4px 12px rgba(0,0,0,0.5);
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
    console.error('Error generating song OG image:', error);
    return new Response(error.stack || error.message, { status: 500 });
  }
}
