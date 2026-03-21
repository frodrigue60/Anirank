import satori from 'satori';
import { Resvg } from '@resvg/resvg-js';
import { html } from 'satori-html';
import fs from 'fs';
import path from 'path';

export async function GET({ params, fetch }) {
  // Load the font from static directory
  const fontPath = path.resolve('static/fonts/Inter-Bold.ttf');
  const fontData = fs.readFileSync(fontPath);

  const { slug } = params;

  try {
    const apiUrl = import.meta.env.VITE_API_URL || "http://127.0.0.1:8080/api";
    const res = await fetch(`${apiUrl}/users/${slug}`);
    
    if (!res.ok) {
      return new Response('Not found', { status: 404 });
    }

    const json = await res.json();
    const user = json.user || json.data || json;

    if (!user || !user.id) {
      return new Response('Not found', { status: 404 });
    }

    const bannerStyle = `width: 100%; height: 100%; display: flex; ${
        user.banner_url ? `background-image: url('${user.banner_url}'); background-size: cover; background-position: center; opacity: 0.4;` : 'background-color: #1a1a24;'
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
    display: flex;
    align-items: center;
    width: 100%;
    height: 100%;
    padding: 80px;
    z-index: 10;
    justify-content: center;
    background-image: linear-gradient(to top, rgba(15, 15, 19, 0.9), transparent);
  ">
    
    <!-- User Avatar -->
    <div style="
      display: flex;
      width: 280px;
      height: 280px;
      border-radius: 140px;
      border: 8px solid rgba(255, 255, 255, 0.1);
      box-shadow: 0 20px 40px rgba(0,0,0,0.6);
      overflow: hidden;
      margin-right: 60px;
      background-color: #2a2a35;
      justify-content: center;
      align-items: center;
    ">
      ${user.avatar_url ? `<img src="${user.avatar_url}" style="width: 100%; height: 100%; object-fit: cover;" />` : `<span style="font-size: 100px; color: white;">${user.name.charAt(0).toUpperCase()}</span>`}
    </div>

    <!-- User Info -->
    <div style="display: flex; flex-direction: column;">
      
      <div style="
        display: flex;
        font-size: 32px;
        font-weight: 900;
        color: #ff4e50;
        margin-bottom: 10px;
        text-transform: uppercase;
        letter-spacing: 4px;
      ">
        AniRank User
      </div>

      <h1 style="
        font-size: 84px;
        font-weight: 900;
        margin: 0 0 20px 0;
        line-height: 1.1;
        display: flex;
        color: white;
        text-shadow: 0 4px 12px rgba(0,0,0,0.5);
      ">
        ${user.name}
      </h1>

      <div style="display: flex; align-items: center; margin-top: 10px;">
        <!-- Level Badge -->
        <div style="
          display: flex;
          background-color: rgba(255, 255, 255, 0.1);
          border: 2px solid rgba(255, 255, 255, 0.2);
          border-radius: 12px;
          padding: 10px 20px;
          margin-right: 20px;
        ">
          <div style="font-size: 20px; color: rgba(255,255,255,0.6); text-transform: uppercase; font-weight: bold; margin-right: 15px;">Level</div>
          <div style="font-size: 36px; font-weight: 900; color: white;">${user.level_number || '1'}</div>
        </div>

        <!-- XP Badge -->
        <div style="
          display: flex;
          background-color: rgba(255, 78, 80, 0.1);
          border: 2px solid rgba(255, 78, 80, 0.3);
          border-radius: 12px;
          padding: 10px 20px;
        ">
          <div style="font-size: 20px; color: rgba(255,78,80,0.8); text-transform: uppercase; font-weight: bold; margin-right: 15px;">XP</div>
          <div style="font-size: 36px; font-weight: 900; color: #ff4e50;">${user.total_xp || '0'}</div>
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
    console.error('Error generating user OG image:', error);
    return new Response(error.stack || error.message, { status: 500 });
  }
}
