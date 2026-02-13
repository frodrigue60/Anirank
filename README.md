# Anirank

Anirank is a modern, premium web application for discovering, exploring, and ranking anime themes (Openings and Endings). It features a sleek glassmorphic UI, deep AniList integration, and a sophisticated data model that distinguishes between animation studios and production committees.

## Key Features

- **🚀 AniList Synchronization**: Automatically fetch anime metadata, themes, and studio/producer relationships directly from AniList.
- **🎬 Studio vs. Producer Distinction**: Unlike many platforms, Anirank separates animation studios (creative) from production companies/committees (business).
- **📊 Seasonal Charts**: Browse the latest themes organized by year and season.
- **🏆 Global Rankings**: Real-time leaderboards for top-rated and most-viewed openings and endings.
- **🎵 Immersive Player**: A dedicated cinema-mode player for seamless theme playback.
- **📁 User Collections**: Create public/private playlists and maintain a personalized favorites list.
- **🛠️ Robust Admin Panel**: Full CRUD control over all entities, including a bulk-sync tool for data consistency.

## Tech Stack

- **Backend**: Laravel 9.x, PHP 8.2
- **Frontend**: Livewire 2.x, Tailwind CSS, Vite, Vanilla JS
- **Database**: MySQL 8.0
- **icons**: Material Symbols (Google)

## Getting Started

### Prerequisites

- PHP 8.2+
- Composer
- Node.js & NPM
- MySQL 8.0

### Installation

1. **Clone the repository**:

    ```bash
    git clone https://github.com/your-repo/Anirank.git
    cd Anirank
    ```

2. **Install dependencies**:

    ```bash
    composer install
    npm install
    ```

3. **Environment Setup**:

    ```bash
    cp .env.example .env
    php artisan key:generate
    ```

    _Configure your database settings in the `.env` file._

4. **Migrations & Seeding**:

    ```bash
    php artisan migrate --seed
    ```

5. **Start Development Server**:
    ```bash
    php artisan serve
    npm run dev
    ```

## Development

The project uses **Livewire** for most reactive components. You can find these in `app/Http/Livewire` and their corresponding views in `resources/views/livewire`.

For styling, we use **Tailwind CSS** with a custom purple-centric dark theme. See `tailwind.config.js` and `resources/css/app.css` for design tokens.

---

_Built with ❤️ for the anime community._
