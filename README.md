# AniRank — Anime Music Discovery & Ranking Platform

AniRank is a comprehensive web application designed for exploring, rating, and ranking anime music themes (openings, endings, and inserts). The platform enables anime enthusiasts to discover music through detailed metadata, participate in community voting, and manage personalized profiles with gamification elements.

## Core Features

- **Content Discovery:** Advanced search for songs by title, artist, format, season, and year.
- **Community Interaction:** Real-time rating, reaction, and comment system.
- **Dynamic Rankings:** Weekly and seasonal rankings generated based on user activity.
- **Gamification:** User profiles featuring an experience point (XP) system, levels, and automatic badges.
- **Tournament Systems:** Public voting brackets to determine the community's favorite songs.
- **OAuth2 Integration:** Account linking with AniList, Google, and Discord.
- **Administrative Panel:** Comprehensive tools for content management, moderation, and auditing.

## Technology Stack

### Frontend
- **Framework:** SvelteKit (Svelte 5 utilizing Runes for advanced reactivity).
- **Styling:** TailwindCSS v4 with a unified editorial design system.
- **State Management:** Reactive state singletons via Svelte 5 Runes.
- **API Communication:** Axios client with interceptors for authentication management.
- **Testing:** Vitest with MSW (Mock Service Worker) for API simulation.

### Backend
- **Language:** Go 1.25+.
- **Web Framework:** GoFiber v2 for high performance.
- **Database:** PostgreSQL utilizing `sqlx` and raw SQL queries (no ORM for maximum optimization).
- **Cache & Rate Limiting:** Redis (with fallback to in-memory cache).
- **Storage:** S3-compatible (MinIO for development, AWS S3 for production).
- **Security:** JWT-based authentication (HS256) and sensitive data encryption.

## Architecture & Security

### Clean Architecture
The backend strictly adheres to **Clean Architecture** principles, separating responsibilities into layers:
- **Domain:** Pure business definitions and interfaces.
- **Usecase:** Application logic and orchestration.
- **Delivery:** HTTP handlers and DTO mapping.
- **Repository/Infrastructure:** Concrete database implementations and external services.

### Security Model
- **Identity Protection:** UUIDs are employed for all public API references, concealing internal sequential database IDs to prevent enumeration attacks.
- **Access Control (RBAC):** Granular role and permission system managed through specialized middleware.
- **Data Security:** External credentials and third-party tokens are stored encrypted in the database.

## Project Structure

```text
├── backend/            # Server logic (Go)
│   ├── cmd/api/        # Application entry point
│   ├── internal/       # Private code (Clean Architecture)
│   └── database/       # Migrations and database scripts
├── frontend/           # User interface (SvelteKit)
│   ├── src/lib/        # Components and shared logic
│   └── src/routes/     # Routing and pages
├── docker-compose.yml  # Container orchestration
└── .env.example        # Centralized configuration template
```

## Setup & Deployment

The platform utilizes a centralized configuration system through a `.env` file located at the repository root.

1.  **Requirements:** Docker and Docker Compose installed.
2.  **Preparation:** Copy the `.env.example` file to `.env` and complete the required variables.
3.  **Execution:**
    ```bash
    docker-compose up --build
    ```

## License

This project is licensed under the MIT License. See the license file for further details.
