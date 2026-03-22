import { readFileSync } from 'fs';
import { resolve } from 'path';
import satori from 'satori';
import { html } from 'satori-html';
import { Resvg } from '@resvg/resvg-js';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ fetch, url }) => {
  try {
    // 1. Fetch Stats from Backend
    const apiBase = import.meta.env?.PUBLIC_API_URL || 'http://localhost:8080/api';
    const statsResponse = await fetch(`${apiBase}/site-statistics`);
    const stats = await statsResponse.json();
    
    const overviews = stats?.overviews || { 
      total_users: 0, 
      total_animes: 0, 
      total_songs: 0 
    };

    // 2. Load Fonts
    const fontBold = readFileSync(resolve('static/fonts/Inter-Bold.ttf'));
    const fontRegular = readFileSync(resolve('static/fonts/Inter-Regular.ttf'));

    // 3. Define the HTML for Satori
    const markup = html`
      <div style="
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        width: 1200px;
        height: 630px;
        background: linear-gradient(135deg, #0f172a 0%, #1e1b4b 50%, #581c87 100%);
        color: white;
        font-family: 'Inter';
        padding: 60px;
      ">
        <!-- Logo/Branding -->
        <div style="display: flex; align-items: center; margin-bottom: 20px;">
          <div style="
            width: 80px;
            height: 80px;
            background-color: #7f13ec;
            border-radius: 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-right: 25px;
            box-shadow: 0 0 40px rgba(127, 19, 236, 0.4);
          ">
            <span style="font-size: 50px; font-weight: 900;">A</span>
          </div>
          <h1 style="font-size: 90px; margin: 0; font-weight: 800; letter-spacing: -2px;">
            Ani<span style="color: #a855f7;">Rank</span>
          </h1>
        </div>

        <p style="font-size: 32px; color: #94a3b8; margin-bottom: 60px; text-align: center;">
          The Ultimate Anime Music Ranking Platform
        </p>

        <!-- Stats Grid -->
        <div style="display: flex; justify-content: space-between; width: 100%; max-width: 900px;">
          <div style="display: flex; flex-direction: column; align-items: center; background: rgba(255,255,255,0.05); padding: 30px; border-radius: 24px; border: 1px solid rgba(255,255,255,0.1); flex: 1; margin: 0 15px;">
            <span style="font-size: 20px; color: #a855f7; font-weight: 700; margin-bottom: 10px; text-transform: uppercase; letter-spacing: 2px;">Songs</span>
            <span style="font-size: 56px; font-weight: 800;">${overviews.total_songs.toLocaleString()}</span>
          </div>
          <div style="display: flex; flex-direction: column; align-items: center; background: rgba(255,255,255,0.05); padding: 30px; border-radius: 24px; border: 1px solid rgba(255,255,255,0.1); flex: 1; margin: 0 15px;">
            <span style="font-size: 20px; color: #a855f7; font-weight: 700; margin-bottom: 10px; text-transform: uppercase; letter-spacing: 2px;">Animes</span>
            <span style="font-size: 56px; font-weight: 800;">${overviews.total_animes.toLocaleString()}</span>
          </div>
          <div style="display: flex; flex-direction: column; align-items: center; background: rgba(255,255,255,0.05); padding: 30px; border-radius: 24px; border: 1px solid rgba(255,255,255,0.1); flex: 1; margin: 0 15px;">
            <span style="font-size: 20px; color: #a855f7; font-weight: 700; margin-bottom: 10px; text-transform: uppercase; letter-spacing: 2px;">Users</span>
            <span style="font-size: 56px; font-weight: 800;">${overviews.total_users.toLocaleString()}</span>
          </div>
        </div>

        <div style="position: absolute; bottom: 40px; right: 60px; display: flex; align-items: center; color: rgba(255,255,255,0.3);">
          <span style="font-size: 18px; font-weight: 500;">anirank.work</span>
        </div>
      </div>
    `;

    // 4. Generate SVG
    const svg = await satori(markup, {
      width: 1200,
      height: 630,
      fonts: [
        {
          name: 'Inter',
          data: fontRegular,
          weight: 400,
          style: 'normal',
        },
        {
          name: 'Inter',
          data: fontBold,
          weight: 700,
          style: 'normal',
        },
      ],
    });

    // 5. Convert SVG to PNG
    const resvg = new Resvg(svg, {
        fitTo: {
            mode: 'width',
            value: 1200,
        },
    });
    const pngData = resvg.render();
    const pngBuffer = pngData.asPng();

    return new Response(new Uint8Array(pngBuffer), {
      headers: {
        'Content-Type': 'image/png',
        'Cache-Control': 'max-age=0, s-maxage=3600', // Cache for 1 hour on CDN
      },
    });
  } catch (error) {
    console.error('OG Image generation error:', error);
    return new Response('Error generating image', { status: 500 });
  }
};
