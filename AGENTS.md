# Anirank - Agent Guidelines

## Overview

**Anirank** is a web application for discovering, playing, and rating anime openings (OP) and endings (ED). Users can explore anime themes, rate them, add them to favorites, and create playlists. The project is built with **Laravel 9** and uses **MySQL** as its database.

## Core Features

1.  **Anime Theme Database**: A comprehensive catalog of anime openings and endings, fetched and synchronized from external APIs (like AniList).
2.  **Rating System**: Users can rate songs on a 5-star scale. Ratings are stored per-user and an average is calculated for each song variant.
3.  **Likes/Dislikes & Favorites**: A like/dislike reaction system and favorites list for registered users.
4.  **Playlists**: Users can create public or private playlists to organize their favorite themes.
5.  **Comments**: A polymorphic comment system allows users to comment on any song or variant.
6.  **Seasonal Browsing**: Browse anime themes by season (e.g., Winter 2024) and year.
7.  **Rankings**: View top-rated and most-viewed openings and endings.

## Tech Stack

- **Backend**: PHP 8.2, Laravel 9.x
- **Frontend**: Blade templates, Livewire 2.x, Tailwind CSS, Vite, Vanilla JS
- **Database**: MySQL 8.0
- **Environment**: Laragon / Docker

---

## Design System

The application uses a **Modern Dark Theme** with a deep purple aesthetic.

### Typography

- **Display**: `Spline Sans` (Headings, UI labels)
- **Body**: `Noto Sans` (Descriptions, content)
- **Icons**: Google Material Symbols (Outlined)

### Color Palette

- **Primary**: `#7f13ec` (Purple accent)
- **Background**: `#191022` (Deep dark purple)
- **Surface**: `#2a2136` (Lighter purple for cards/panels)
- **Surface Darker**: `#22192e` (Used for sidebar and dropdowns)

### Custom Components (SCSS/Tailwind)

- `.glass-panel`: Frosted glass effect (`backdrop-filter: blur(12px)`) with transparent background.
- `.hero-glow`: Purple shadow glow (`box-shadow`) for high-impact sections.
- `.filled`: Utility class for Material Symbols to use the "FILL" variation.

---

## Model Reference

This section provides detailed documentation for all Eloquent models in `app/Models`.

---

### `Post`

Represents an **anime series**.

| Field           | Type    | Description                              |
| --------------- | ------- | ---------------------------------------- |
| `title`         | string  | The name of the anime.                   |
| `slug`          | string  | URL-friendly unique identifier.          |
| `description`   | text    | Synopsis/description (nullable).         |
| `anilist_id`    | bigint  | ID from the AniList API (nullable).      |
| `status`        | boolean | Published status (0=draft, 1=published). |
| `thumbnail`     | string  | Local path to the cover image.           |
| `thumbnail_src` | string  | External URL for the cover image.        |
| `banner`        | string  | Local path to the banner image.          |
| `banner_src`    | string  | External URL for the banner image.       |
| `year_id`       | FK      | References `years.id`.                   |
| `season_id`     | FK      | References `seasons.id`.                 |
| `format_id`     | FK      | References `formats.id`.                 |

**Relationships:**

- `hasMany` → `Song`, `Report`
- `belongsTo` → `Year`, `Season`, `Format`
- `belongsToMany` → `Studio`, `Producer`, `ExternalLink`
- Custom: `openings()`, `endings()`

---

### `Song`

Represents a **theme song** (Opening or Ending) associated with a Post.

| Field         | Type   | Description                                |
| ------------- | ------ | ------------------------------------------ |
| `song_romaji` | string | Romanized song title (nullable).           |
| `song_jp`     | string | Japanese song title (nullable).            |
| `song_en`     | string | English song title (nullable).             |
| `theme_num`   | string | Theme number (e.g., "1" for OP1).          |
| `type`        | enum   | `OP`, `ED`, `INS` (Insert), `OTH` (Other). |
| `slug`        | string | URL-friendly identifier.                   |
| `views`       | bigint | View counter.                              |
| `post_id`     | FK     | References `posts.id`.                     |
| `year_id`     | FK     | References `years.id`.                     |
| `season_id`   | FK     | References `seasons.id`.                   |

**Relationships:**

- `belongsTo` → `Post`, `Year`, `Season`
- `hasMany` → `SongVariant`
- `belongsToMany` → `Artist`, `Playlist`
- Polymorphic: `morphMany` → `Comment`, `Reaction`, `Favorite`

**Trait:** Uses `Rateable` for star ratings.

**Key Methods:**

- `getName()` → Returns the best available song name.
- `getUrl()` → Generates the public URL.
- `incrementViews()` → Session-aware view counter.
- `liked()`, `disliked()`, `isFavorited()` → Check user interaction state.

---

### `SongVariant`

Represents a **specific version** of a song (e.g., V1, V2, Creditless, Spoiler).

| Field            | Type    | Description                             |
| ---------------- | ------- | --------------------------------------- |
| `version_number` | bigint  | The version number (1, 2, etc.).        |
| `song_id`        | FK      | References `songs.id`.                  |
| `views`          | bigint  | View counter for this specific variant. |
| `slug`           | string  | URL-friendly identifier.                |
| `spoiler`        | boolean | Flag if the variant contains spoilers.  |
| `year_id`        | FK      | References `years.id`.                  |
| `season_id`      | FK      | References `seasons.id`.                |

**Relationships:**

- `belongsTo` → `Song`, `Year`, `Season`
- `hasOne` → `Video`
- Polymorphic: `morphMany` → `Comment`, `Reaction`, `Favorite`, `Report`
- `morphOne` → `ReactionCounter`

**Trait:** Uses `Rateable` for star ratings.

---

### `Video`

Represents the **actual video file** for a SongVariant.

| Field             | Type   | Description                                    |
| ----------------- | ------ | ---------------------------------------------- |
| `video_src`       | string | Path to the local video file (storage).        |
| `embed_code`      | string | Embed URL for external videos (e.g., YouTube). |
| `type`            | string | `file` for local, `embed` for external.        |
| `song_variant_id` | FK     | References `song_variants.id`.                 |

**Relationships:**

- `belongsTo` → `SongVariant`, `Song`

**Key Methods:**

- `isEmbed()`, `isLocal()` → Check video type.
- `getLocalUrlAttribute()` → Returns the full public URL.

---

### `Artist`

Represents a **musician or band**.

| Field           | Type   | Description                      |
| --------------- | ------ | -------------------------------- |
| `name`          | string | Artist name (romanized/English). |
| `name_jp`       | string | Japanese name (nullable).        |
| `slug`          | string | URL-friendly identifier.         |
| `thumbnail`     | string | Local path to artist image.      |
| `thumbnail_src` | string | External URL for artist image.   |

**Relationships:**

- `belongsToMany` → `Song`

---

### `User`

Standard Laravel user model with extensions.

| Field          | Type   | Description                                 |
| -------------- | ------ | ------------------------------------------- |
| `name`         | string | User's display name.                        |
| `email`        | string | User's email address.                       |
| `password`     | string | Hashed password.                            |
| `type`         | string | Role: `admin`, `editor`, `creator`, `user`. |
| `image`        | string | Path to profile picture.                    |
| `banner`       | string | Path to profile banner.                     |
| `score_format` | string | User's preferred rating display format.     |
| `slug`         | string | URL-friendly identifier.                    |

**Relationships:**

- `hasMany` → `UserRequest`, `Comment`, `Favorite`, `Reaction`, `Playlist`

**Key Methods:**

- `isStaff()`, `isAdmin()`, `isEditor()`, `isCreator()` → Role checks.
- `generateSlug()` → Creates a unique slug from the name.
- `canViewPlaylist(Playlist $playlist)` → Permission check.

---

### `Comment`

Polymorphic comment system with nested replies.

| Field              | Type   | Description                              |
| ------------------ | ------ | ---------------------------------------- |
| `content`          | text   | The comment text.                        |
| `user_id`          | FK     | References `users.id`.                   |
| `commentable_id`   | bigint | Polymorphic ID (Song or SongVariant).    |
| `commentable_type` | string | Polymorphic type.                        |
| `parent_id`        | FK     | Self-referencing for replies (nullable). |

**Relationships:**

- `belongsTo` → `User`
- `morphTo` → `commentable` (Song or SongVariant)
- `hasMany` → `replies` (self-referencing)
- Polymorphic: `morphMany` → `Reaction`

---

### `Playlist`

User-created playlists.

| Field         | Type    | Description             |
| ------------- | ------- | ----------------------- |
| `name`        | string  | Playlist name.          |
| `description` | text    | Optional description.   |
| `user_id`     | FK      | References `users.id`.  |
| `is_public`   | boolean | Public visibility flag. |

**Relationships:**

- `belongsTo` → `User`
- `belongsToMany` → `Song` (with `position` pivot)

---

### Polymorphic Models

#### `Rating` (via `Rateable` Trait)

Stores user star ratings for Songs and SongVariants.

- `rateable_id`, `rateable_type` → Polymorphic target.
- `user_id` → The user who rated.
- `rating` → The numeric score.

#### `Reaction`

Stores likes (`type = 1`) and dislikes (`type = -1`).

- `reactable_id`, `reactable_type` → Polymorphic target (Song, SongVariant, Comment).
- `user_id` → The user who reacted.

#### `ReactionCounter`

Caches the total `likes_count` and `dislikes_count` for a reactable entity to avoid N+1 queries.

#### `Favorite`

Marks an entity as favorited by a user.

- `favoritable_id`, `favoritable_type` → Polymorphic target (Song, SongVariant).
- `user_id` → The user who favorited.

---

### Supporting Models

| Model          | Purpose                                                |
| -------------- | ------------------------------------------------------ |
| `Year`         | Represents a year (e.g., 2024). Has many Posts, Songs. |
| `Season`       | Represents a season (Winter, Spring, Summer, Fall).    |
| `Studio`       | Animation studio (creative). Many-to-many with Post.   |
| `Producer`     | Production company/committee. Many-to-many with Post.  |
| `Format`       | Anime format (TV, Movie, OVA). Has many Posts.         |
| `ExternalLink` | External links (MAL, AniList). Many-to-many with Post. |
| `Report`       | User-submitted reports for SongVariants.               |
| `UserRequest`  | User-submitted requests (e.g., "add this anime").      |

---

## Key Relationships Diagram

```
┌────────────────────────────────────────────────────────────────────┐
│                              User                                  │
│  ├─► Playlist ───► Song (M:M)                                      │
│  ├─► Favorite ───► Song | SongVariant (Polymorphic)                │
│  ├─► Reaction ───► Song | SongVariant | Comment (Polymorphic)      │
│  ├─► Rating   ───► Song | SongVariant (Polymorphic via Rateable)   │
│  └─► Comment  ───► Song | SongVariant (Polymorphic)                │
└────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│                              Post (Anime)                          │
│  ├─► Song (1:N)                                                    │
│  │     └─► SongVariant (1:N)                                       │
│  │           └─► Video (1:1)                                       │
│  ├─► Year, Season, Format (N:1)                                    │
│  └─► Studio, ExternalLink (M:M)                                    │
└────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│                              Song                                  │
│  ├─► Artist (M:M)                                                  │
│  └─► Playlist (M:M)                                                │
└────────────────────────────────────────────────────────────────────┘
```

---

## Controller Reference (Public)

This section provides detailed documentation for all public controllers in `app/Http/Controllers`.

---

### `PostController`

Handles the homepage and anime detail pages.

| Method        | Route               | Description                                        |
| ------------- | ------------------- | -------------------------------------------------- |
| `index()`     | `GET /`             | Homepage: Top openings, top endings, recent songs. |
| `show($slug)` | `GET /anime/{slug}` | Anime detail page with openings and endings list.  |
| `animes()`    | `GET /animes`       | Filter/paginate anime series.                      |
| `themes()`    | `GET /themes`       | Browse all openings and endings.                   |

**Key Helper Methods:**

- `setScoreSongs($songs, $user)` → Attaches rating data to a collection of songs.
- `setScoreOnlyVariants($variants, $user)` → Attaches rating data to variants.
- `sortSongs($sort, $songs)` → Sorts songs by various criteria (recent, title, score, views, likes).
- `paginate($songs)` → Custom pagination for collections.

---

### `SongController`

Handles song detail pages and seasonal/ranking views.

| Method                        | Route                            | Description                              |
| ----------------------------- | -------------------------------- | ---------------------------------------- |
| `show($animeSlug, $songSlug)` | `GET /anime/{anime}/song/{song}` | Song detail page with variants list.     |
| `seasonal()`                  | `GET /seasonal`                  | Browse songs by current/selected season. |
| `ranking()`                   | `GET /ranking`                   | Top-rated songs overall.                 |

**Key Helper Methods:**

- `setScoreSong($song, $user)` → Attaches rating data to a single song.
- `getUserRating($song_id, $user_id)` → Gets the user's rating for a song.
- `formatScoreString($score, $format, $denominator)` → Formats rating for display.

---

### `SongVariantController`

Handles the video player page and user interactions (rating, likes, favorites).

| Method                          | Route                                              | Description                                  |
| ------------------------------- | -------------------------------------------------- | -------------------------------------------- |
| `show($anime, $song, $variant)` | `GET /anime/{anime}/song/{song}/variant/{variant}` | Video player page with comments and ratings. |
| `rate(Request, $variant_id)`    | `POST /variants/{id}/rate`                         | Submit or update a star rating.              |
| `like($id)`                     | `POST /variants/{id}/like`                         | Toggle like on a variant.                    |
| `dislike($id)`                  | `POST /variants/{id}/dislike`                      | Toggle dislike on a variant.                 |
| `toggleFavorite($id)`           | `POST /variants/{id}/favorite`                     | Toggle favorite status.                      |
| `ranking()`                     | `GET /ranking/variants`                            | Top-rated variants.                          |
| `seasonal(Request)`             | `GET /seasonal/variants`                           | Seasonal variants view.                      |

**Key Helper Methods:**

- `handleReaction($songVariant, $type)` → Manages like/dislike logic (toggle or switch).
- `setScoreOnlyOneVariant($variant, $user)` → Attaches full rating data to one variant.

---

### `UserController`

Handles user profile, favorites, and settings.

| Method                | Route                         | Description                                   |
| --------------------- | ----------------------------- | --------------------------------------------- |
| `index()`             | `GET /profile`                | Current user dashboard (uses `UserSettings`). |
| `show($slug)`         | `GET /users/{slug}`           | Public user profile.                          |
| `favorites(Request)`  | `GET /favorites`              | Favorites list (uses `FavoritesTable`).       |
| `userList($slug)`     | `GET /user/{slug}/list`       | A specific user's rated themes list.          |
| `uploadProfilePic()`  | `POST /profile/upload-pic`    | (Legacy) Upload profile picture.              |
| `uploadBannerPic()`   | `POST /profile/upload-banner` | (Legacy) Upload profile banner.               |
| `changeScoreFormat()` | `POST /profile/score-format`  | (Legacy) Change score display format.         |
| `welcome()`           | `GET /welcome`                | Welcome/onboarding page.                      |

**Key Helper Methods:**

- `setScore($songs, $score_format)` → Formats scores based on user preference (5-star, 10-point, 100-point).
- `sortPosts($sort, $songs)` → Sorts favorites by various criteria.
- `setScoreOnlyVariants($variants, $user)` → Attaches rating data to variants.

---

### `ArtistController`

Handles artist listing and detail pages.

| Method        | Route                 | Description                         |
| ------------- | --------------------- | ----------------------------------- |
| `index()`     | `GET /artists`        | Filter/browse all artists.          |
| `show($slug)` | `GET /artists/{slug}` | Artist detail with all their songs. |

---

### `StudioController`

Handles animation studio pages.

| Method        | Route                 | Description                         |
| ------------- | --------------------- | ----------------------------------- |
| `index()`     | `GET /studios`        | Filter/browse all studios.          |
| `show($slug)` | `GET /studios/{slug}` | Studio detail with all their anime. |

---

### `ProducerController`

Handles production company and committee pages.

| Method        | Route                   | Description                           |
| ------------- | ----------------------- | ------------------------------------- |
| `index()`     | `GET /producers`        | Filter/browse all producers.          |
| `show($slug)` | `GET /producers/{slug}` | Producer detail with all their anime. |

---

### `CommentController`

Handles commenting on songs and variants.

| Method                     | Route                         | Description                              |
| -------------------------- | ----------------------------- | ---------------------------------------- |
| `store(Request)`           | `POST /comments`              | Create a new comment on a SongVariant.   |
| `update(Request, Comment)` | `PUT /comments/{comment}`     | Update an existing comment (owner only). |
| `destroy(Comment)`         | `DELETE /comments/{comment}`  | Delete a comment (owner or staff).       |
| `like($id)`                | `POST /comments/{id}/like`    | Toggle like on a comment.                |
| `dislike($id)`             | `POST /comments/{id}/dislike` | Toggle dislike on a comment.             |
| `reply(Request, Comment)`  | `POST /comments/{id}/reply`   | Reply to an existing comment (nested).   |

---

### `PlaylistController`

Handles user-created playlists and the immersive player.

| Method                      | Route                       | Description                               |
| --------------------------- | --------------------------- | ----------------------------------------- |
| `index()`                   | `GET /playlists`            | Premium grid of user's playlists.         |
| `create()`                  | `GET /playlists/create`     | Glassmorphic creation form.               |
| `store(Request)`            | `POST /playlists`           | Create a new playlist.                    |
| `show(Playlist)`            | `GET /playlists/{playlist}` | **Cinema Mode** playback page with queue. |
| `edit(Playlist)`            | `GET /playlists/{id}/edit`  | Glassmorphic edit form.                   |
| `update(Request, Playlist)` | `PUT /playlists/{id}`       | Update details (Fixed persistence bug).   |
| `destroy(Playlist)`         | `DELETE /playlists/{id}`    | Delete a playlist.                        |

---

### `FavoriteController`

Handles adding/removing favorites.

| Method        | Route                         | Description                                 |
| ------------- | ----------------------------- | ------------------------------------------- |
| `toggle($id)` | `POST /favorites/{id}/toggle` | Add or remove a SongVariant from favorites. |

---

### `ReactionController`

Handles like/dislike reactions via API.

| Method               | Route                       | Description                          |
| -------------------- | --------------------------- | ------------------------------------ |
| `react(SongVariant)` | `POST /reactions/{variant}` | Toggle like/dislike (JSON response). |

---

### Other Controllers

| Controller               | Purpose                                           |
| ------------------------ | ------------------------------------------------- |
| `YearController`         | Filter content by year.                           |
| `SeasonController`       | Filter content by season.                         |
| `FormatController`       | Filter content by anime format (TV, Movie, etc.). |
| `ReportController`       | Submit user reports for variants with issues.     |
| `UserRequestController`  | Submit requests to add new anime/songs.           |
| `ExternalLinkController` | Manage external links (MAL, AniList, YouTube).    |

---

## Controller Reference (Admin)

All admin controllers are located in `app/Http/Controllers/Admin/` and are protected by the `staff` middleware. They handle CRUD operations for all entities, synchronization with AniList, and moderation tasks.

---

### `Admin\PostController`

**The primary admin controller** – manages anime series (Posts) and AniList synchronization.

| Method                 | Route                            | Description                                    |
| ---------------------- | -------------------------------- | ---------------------------------------------- |
| `index()`              | `GET admin/posts`                | List all anime with search and pagination.     |
| `create()`             | `GET admin/posts/create`         | Form to add new anime (manual or via AniList). |
| `store(Request)`       | `POST admin/posts`               | Create a new anime record.                     |
| `show($id)`            | `GET admin/posts/{id}`           | View anime details.                            |
| `edit($id)`            | `GET admin/posts/{id}/edit`      | Edit anime form.                               |
| `update(Request, $id)` | `PUT admin/posts/{id}`           | Update anime details.                          |
| `destroy($id)`         | `DELETE admin/posts/{id}`        | Delete anime and all associated media.         |
| `search(Request)`      | `GET admin/posts/search`         | Search posts in admin panel.                   |
| `toggleStatus(Post)`   | `POST admin/posts/{post}/toggle` | Toggle publish/draft status.                   |
| `songs($post_id)`      | `GET admin/posts/{id}/songs`     | View songs belonging to this anime.            |
| `addSong($post_id)`    | `GET admin/posts/{id}/songs/add` | Form to add a new song to this anime.          |
| `dashboard()`          | `GET admin/dashboard`            | Admin dashboard with statistics.               |

**AniList Integration Methods:**

- `searchInAnilist(Request)` → Search AniList API for anime by title.
- `getById($anilist_id)` → Fetch anime details from AniList by ID.
- `getSeasonalAnimes(Request)` → Fetch seasonal anime list from AniList.
- `generateMassive($data)` → Bulk import anime from AniList seasonal data.
- `forceUpdate($id)` → Re-sync anime data from AniList.
- `syncAllFromAnilist()` → **Bulk Synchronizer**: Re-evaluates all posts and re-links them to Studios or Producers based on AniList flags.
- `wipePosts()` → Delete all posts (dangerous, admin-only).

**Image Handling:**

- `saveAnimeThumbnail($item, $post)` → Download and save cover image.
- `saveAnimeBanner($item, $post)` → Download and save banner image.
- `storePostImages($post, $request)` → Handle image uploads from form.

---

### `Admin\SongController`

Manages songs (OPs/EDs) for anime.

| Method                 | Route                               | Description                          |
| ---------------------- | ----------------------------------- | ------------------------------------ |
| `index()`              | `GET admin/songs`                   | List all songs.                      |
| `store(Request)`       | `POST admin/songs`                  | Create a new song for an anime.      |
| `edit($id)`            | `GET admin/songs/{id}/edit`         | Edit song form.                      |
| `update(Request, $id)` | `PUT admin/songs/{id}`              | Update song details.                 |
| `destroy($id)`         | `DELETE admin/songs/{id}`           | Delete song and all variants/videos. |
| `variants($song_id)`   | `GET admin/songs/{id}/variants`     | View variants for this song.         |
| `addVariant($song_id)` | `GET admin/songs/{id}/variants/add` | Form to add a new variant.           |

**Key Helpers:**

- `parseName($rawName)` → Parse song names from AniList format.
- `decodeUnicodeIfNeeded($string)` → Handle Unicode in song titles.

---

### `Admin\SongVariantController`

Manages song variants (different versions like V1, V2, Creditless).

| Method                   | Route                                | Description                          |
| ------------------------ | ------------------------------------ | ------------------------------------ |
| `store(Request)`         | `POST admin/variants`                | Create a new variant for a song.     |
| `edit($id)`              | `GET admin/variants/{id}/edit`       | Edit variant form.                   |
| `update(Request, $id)`   | `PUT admin/variants/{id}`            | Update variant details.              |
| `destroy($id)`           | `DELETE admin/variants/{id}`         | Delete variant and associated video. |
| `videos($variant_id)`    | `GET admin/variants/{id}/videos`     | View video for this variant.         |
| `addVideos($variant_id)` | `GET admin/variants/{id}/videos/add` | Form to add video to variant.        |

---

### `Admin\VideoController`

Manages video files (local uploads or embed URLs).

| Method                 | Route                        | Description                                  |
| ---------------------- | ---------------------------- | -------------------------------------------- |
| `index($song_id)`      | `GET admin/videos`           | List videos for a song.                      |
| `store(Request)`       | `POST admin/videos`          | Upload local video or save embed URL.        |
| `show($id)`            | `GET admin/videos/{id}`      | View video details.                          |
| `edit($id)`            | `GET admin/videos/{id}/edit` | Edit video form.                             |
| `update(Request, $id)` | `PUT admin/videos/{id}`      | Update video (replace file or change embed). |
| `destroy($id)`         | `DELETE admin/videos/{id}`   | Delete video file from storage.              |

**Key Features:**

- Supports both local file uploads and embed URLs (YouTube, etc.).
- Handles video MIME types and extensions.
- Deletes old video files when replacing.

---

### `Admin\ArtistController`

Manages artist records.

| Method                  | Route                         | Description                          |
| ----------------------- | ----------------------------- | ------------------------------------ |
| `index()`               | `GET admin/artists`           | List all artists with pagination.    |
| `create()`              | `GET admin/artists/create`    | Create artist form.                  |
| `store(Request)`        | `POST admin/artists`          | Create new artist.                   |
| `edit($id)`             | `GET admin/artists/{id}/edit` | Edit artist form.                    |
| `update(Request, $id)`  | `PUT admin/artists/{id}`      | Update artist details.               |
| `destroy($id)`          | `DELETE admin/artists/{id}`   | Delete artist (detaches from songs). |
| `searchArtist(Request)` | `GET admin/artists/search`    | Search artists by name.              |

---

### `Admin\UserController`

Manages user accounts and roles.

| Method                 | Route                       | Description                             |
| ---------------------- | --------------------------- | --------------------------------------- |
| `index()`              | `GET admin/users`           | List all users with pagination.         |
| `create()`             | `GET admin/users/create`    | Create user form (with role selection). |
| `store(Request)`       | `POST admin/users`          | Create new user account.                |
| `edit($id)`            | `GET admin/users/{id}/edit` | Edit user form.                         |
| `update(Request, $id)` | `PUT admin/users/{id}`      | Update user details and role.           |
| `destroy($id)`         | `DELETE admin/users/{id}`   | Delete user and their ratings.          |
| `searchUser(Request)`  | `GET admin/users/search`    | Search users by name.                   |

**Roles:** `user`, `admin`, `editor`, `creator`

---

### `Admin\ReportController`

Manages user-submitted reports for problematic content.

| Method              | Route                            | Description                           |
| ------------------- | -------------------------------- | ------------------------------------- |
| `index()`           | `GET admin/reports`              | List all reports (sorted by date).    |
| `show($id)`         | `GET admin/reports/{id}`         | View report details.                  |
| `destroy($id)`      | `DELETE admin/reports/{id}`      | Delete a report.                      |
| `toggleStatus($id)` | `POST admin/reports/{id}/toggle` | Toggle report status (pending/fixed). |

---

### `Admin\UserRequestController`

Manages user-submitted requests (e.g., "add this anime").

| Method         | Route                        | Description                        |
| -------------- | ---------------------------- | ---------------------------------- |
| `index()`      | `GET admin/requests`         | List all user requests.            |
| `show($id)`    | `GET admin/requests/{id}`    | View and mark request as attended. |
| `destroy($id)` | `DELETE admin/requests/{id}` | Delete a request.                  |

---

### `Admin\YearController`

Manages year records for filtering.

| Method                 | Route                          | Description          |
| ---------------------- | ------------------------------ | -------------------- |
| `index()`              | `GET admin/years`              | List all years.      |
| `create()`             | `GET admin/years/create`       | Create year form.    |
| `store(Request)`       | `POST admin/years`             | Create new year.     |
| `edit($id)`            | `GET admin/years/{id}/edit`    | Edit year form.      |
| `update(Request, $id)` | `PUT admin/years/{id}`         | Update year.         |
| `destroy($id)`         | `DELETE admin/years/{id}`      | Delete year.         |
| `toggle(Year)`         | `POST admin/years/{id}/toggle` | Set as current year. |

---

### `Admin\SeasonController`

Manages season records (Winter, Spring, Summer, Fall).

| Method                 | Route                            | Description            |
| ---------------------- | -------------------------------- | ---------------------- |
| `index()`              | `GET admin/seasons`              | List all seasons.      |
| `create()`             | `GET admin/seasons/create`       | Create season form.    |
| `store(Request)`       | `POST admin/seasons`             | Create new season.     |
| `edit($id)`            | `GET admin/seasons/{id}/edit`    | Edit season form.      |
| `update(Request, $id)` | `PUT admin/seasons/{id}`         | Update season.         |
| `destroy($id)`         | `DELETE admin/seasons/{id}`      | Delete season.         |
| `toggle(Season)`       | `POST admin/seasons/{id}/toggle` | Set as current season. |

---

### `Admin\TagController`

Manages seasonal tags (e.g., "Winter 2024") for categorization.

| Method                 | Route                        | Description                            |
| ---------------------- | ---------------------------- | -------------------------------------- |
| `index()`              | `GET admin/tags`             | List all tags.                         |
| `create()`             | `GET admin/tags/create`      | Create tag form (season + year combo). |
| `store(Request)`       | `POST admin/tags`            | Create new tag.                        |
| `edit($id)`            | `GET admin/tags/{id}/edit`   | Edit tag form.                         |
| `update(Request, $id)` | `PUT admin/tags/{id}`        | Update tag.                            |
| `destroy($id)`         | `DELETE admin/tags/{id}`     | Delete tag and untag all posts.        |
| `search(Request)`      | `GET admin/tags/search`      | Search tags by name.                   |
| `set($id)`             | `POST admin/tags/{id}/set`   | Set tag as active (flag=1).            |
| `unset($id)`           | `POST admin/tags/{id}/unset` | Unset tag (flag=0).                    |

---

### `Admin\StudioController`

Manages animation studio records.

| Method                 | Route                         | Description                          |
| ---------------------- | ----------------------------- | ------------------------------------ |
| `index()`              | `GET admin/studios`           | List all studios with search.        |
| `create()`             | `GET admin/studios/create`    | Create studio form.                  |
| `store(Request)`       | `POST admin/studios`          | Create new studio.                   |
| `edit($id)`            | `GET admin/studios/{id}/edit` | Edit studio form.                    |
| `update(Request, $id)` | `PUT admin/studios/{id}`      | Update studio details.               |
| `destroy($id)`         | `DELETE admin/studios/{id}`   | Delete studio (detaches from posts). |

---

### `Admin\ProducerController`

Manages production company/committee records.

| Method                 | Route                           | Description                            |
| ---------------------- | ------------------------------- | -------------------------------------- |
| `index()`              | `GET admin/producers`           | List all producers with search.        |
| `create()`             | `GET admin/producers/create`    | Create producer form.                  |
| `store(Request)`       | `POST admin/producers`          | Create new producer.                   |
| `edit($id)`            | `GET admin/producers/{id}/edit` | Edit producer form.                    |
| `update(Request, $id)` | `PUT admin/producers/{id}`      | Update producer details.               |
| `destroy($id)`         | `DELETE admin/producers/{id}`   | Delete producer (detaches from posts). |

---

### Other Admin Controllers

| Controller        | Purpose                                               |
| ----------------- | ----------------------------------------------------- |
| `CommentControl`  | View and moderate comments.                           |
| `FormatControl`   | CRUD for anime formats (TV, Movie, OVA, etc.).        |
| `ExternalLinkCon` | CRUD for external links (MAL, AniList, YouTube URLs). |

---

## Database Schema

All migrations are located in `database/migrations/`. The database uses MySQL 8.0.

---

### Core Tables

#### `users`

User accounts with roles and preferences.

| Column              | Type                                                        | Description                 |
| ------------------- | ----------------------------------------------------------- | --------------------------- |
| `id`                | `bigint` (PK)                                               | Primary key                 |
| `name`              | `string`                                                    | Display name                |
| `slug`              | `string` (nullable)                                         | URL-friendly username       |
| `email`             | `string` (unique)                                           | Email address               |
| `email_verified_at` | `timestamp` (nullable)                                      | Email verification date     |
| `password`          | `string`                                                    | Hashed password             |
| `type`              | `enum('admin','editor','creator','user')`                   | User role (default: `user`) |
| `image`             | `string` (nullable)                                         | Profile picture path        |
| `banner`            | `string` (nullable)                                         | Profile banner path         |
| `score_format`      | `enum('POINT_100','POINT_10_DECIMAL','POINT_10','POINT_5')` | Rating display preference   |
| `remember_token`    | `string`                                                    | Laravel remember token      |
| `timestamps`        | `datetime`                                                  | Created/updated at          |

---

#### `posts`

Anime series (the core content entity).

| Column          | Type                | Description                       |
| --------------- | ------------------- | --------------------------------- |
| `id`            | `bigint` (PK)       | Primary key                       |
| `title`         | `string`            | Anime title                       |
| `slug`          | `string` (unique)   | URL-friendly title                |
| `description`   | `text` (nullable)   | Synopsis                          |
| `anilist_id`    | `bigint` (nullable) | AniList API ID for sync           |
| `status`        | `boolean`           | Published status (default: false) |
| `thumbnail`     | `string` (nullable) | Cover image path (local)          |
| `thumbnail_src` | `string` (nullable) | Cover image source URL            |
| `banner`        | `string` (nullable) | Banner image path (local)         |
| `banner_src`    | `string` (nullable) | Banner image source URL           |
| `year_id`       | `FK → years`        | Release year                      |
| `season_id`     | `FK → seasons`      | Release season                    |
| `format_id`     | `FK → formats`      | Anime format (TV, Movie, etc.)    |
| `timestamps`    | `datetime`          | Created/updated at                |

---

#### `songs`

Opening/Ending theme songs.

| Column        | Type                          | Description                   |
| ------------- | ----------------------------- | ----------------------------- |
| `id`          | `bigint` (PK)                 | Primary key                   |
| `song_romaji` | `string` (nullable)           | Song title (romanized)        |
| `song_jp`     | `string` (nullable)           | Song title (Japanese)         |
| `song_en`     | `string` (nullable)           | Song title (English)          |
| `theme_num`   | `string`                      | Theme number (e.g., "1", "2") |
| `type`        | `enum('OP','ED','INS','OTH')` | Theme type                    |
| `slug`        | `string`                      | URL-friendly identifier       |
| `post_id`     | `FK → posts` (cascade)        | Parent anime                  |
| `season_id`   | `FK → seasons` (cascade)      | Airing season                 |
| `year_id`     | `FK → years` (cascade)        | Airing year                   |
| `views`       | `bigint`                      | View count (default: 0)       |
| `timestamps`  | `datetime`                    | Created/updated at            |

**Song Types:**

- `OP` – Opening theme
- `ED` – Ending theme
- `INS` – Insert song
- `OTH` – Other

---

#### `song_variants`

Different versions of a song (V1, V2, Creditless, etc.).

| Column           | Type                     | Description                        |
| ---------------- | ------------------------ | ---------------------------------- |
| `id`             | `bigint` (PK)            | Primary key                        |
| `version_number` | `bigint`                 | Version number (1, 2, 3...)        |
| `song_id`        | `FK → songs` (cascade)   | Parent song                        |
| `views`          | `bigint`                 | View count (default: 0)            |
| `slug`           | `string`                 | URL identifier (e.g., "v1")        |
| `season_id`      | `FK → seasons` (cascade) | Airing season                      |
| `year_id`        | `FK → years` (cascade)   | Airing year                        |
| `spoiler`        | `boolean`                | Contains spoilers (default: false) |
| `timestamps`     | `datetime`               | Created/updated at                 |

---

#### `videos`

Video files associated with song variants.

| Column            | Type                           | Description           |
| ----------------- | ------------------------------ | --------------------- |
| `id`              | `bigint` (PK)                  | Primary key           |
| `embed_code`      | `text` (nullable)              | External embed URL    |
| `video_src`       | `text` (nullable)              | Local video file path |
| `type`            | `enum('embed','file')`         | Video source type     |
| `song_variant_id` | `FK → song_variants` (cascade) | Parent variant        |
| `timestamps`      | `datetime`                     | Created/updated at    |

---

#### `artists`

Musicians and bands.

| Column          | Type                | Description             |
| --------------- | ------------------- | ----------------------- |
| `id`            | `bigint` (PK)       | Primary key             |
| `name`          | `string`            | Artist name (romanized) |
| `name_jp`       | `string` (nullable) | Artist name (Japanese)  |
| `slug`          | `string` (unique)   | URL-friendly name       |
| `thumbnail`     | `string` (nullable) | Artist image path       |
| `thumbnail_src` | `string` (nullable) | Artist image source URL |
| `timestamps`    | `datetime`          | Created/updated at      |

---

### Supporting Tables

#### `years`

Year records for filtering.

| Column       | Type          | Description             |
| ------------ | ------------- | ----------------------- |
| `id`         | `bigint` (PK) | Primary key             |
| `name`       | `smallint`    | Year value (e.g., 2024) |
| `current`    | `boolean`     | Currently active year   |
| `timestamps` | `datetime`    | Created/updated at      |

#### `seasons`

Season records (Winter, Spring, Summer, Fall).

| Column       | Type          | Description             |
| ------------ | ------------- | ----------------------- |
| `id`         | `bigint` (PK) | Primary key             |
| `name`       | `string`      | Season name             |
| `current`    | `boolean`     | Currently active season |
| `timestamps` | `datetime`    | Created/updated at      |

#### `formats`

Anime formats (TV, Movie, OVA, etc.).

| Column       | Type          | Description        |
| ------------ | ------------- | ------------------ |
| `id`         | `bigint` (PK) | Primary key        |
| `name`       | `string`      | Format name        |
| `slug`       | `string`      | URL-friendly name  |
| `timestamps` | `datetime`    | Created/updated at |

#### `studios`

Animation studios (creative entities).

| Column       | Type          | Description        |
| ------------ | ------------- | ------------------ |
| `id`         | `bigint` (PK) | Primary key        |
| `name`       | `string`      | Studio name        |
| `slug`       | `string`      | URL-friendly name  |
| `timestamps` | `datetime`    | Created/updated at |

#### `producers`

Production companies and committees (business entities).

| Column       | Type          | Description        |
| ------------ | ------------- | ------------------ |
| `id`         | `bigint` (PK) | Primary key        |
| `name`       | `string`      | Producer name      |
| `slug`       | `string`      | URL-friendly name  |
| `timestamps` | `datetime`    | Created/updated at |

#### `external_links`

External links (MAL, AniList, YouTube).

| Column       | Type                | Description                    |
| ------------ | ------------------- | ------------------------------ |
| `id`         | `bigint` (PK)       | Primary key                    |
| `icon`       | `string` (nullable) | Icon identifier                |
| `type`       | `string`            | Link type (mal, anilist, etc.) |
| `name`       | `string`            | Display name                   |
| `url`        | `string`            | Full URL                       |
| `timestamps` | `datetime`          | Created/updated at             |

---

### User Interaction Tables

#### `ratings`

User ratings (polymorphic).

| Column          | Type         | Description             |
| --------------- | ------------ | ----------------------- |
| `id`            | `int` (PK)   | Primary key             |
| `rating`        | `int`        | Rating score (1-100)    |
| `rateable_id`   | `bigint`     | Polymorphic target ID   |
| `rateable_type` | `string`     | Polymorphic target type |
| `user_id`       | `FK → users` | User who rated          |
| `timestamps`    | `datetime`   | Created/updated at      |

**Rateable Types:** `Song`, `SongVariant`

#### `reactions`

Like/dislike reactions (polymorphic).

| Column           | Type          | Description             |
| ---------------- | ------------- | ----------------------- |
| `id`             | `bigint` (PK) | Primary key             |
| `user_id`        | `FK → users`  | User who reacted        |
| `reactable_id`   | `bigint`      | Polymorphic target ID   |
| `reactable_type` | `string`      | Polymorphic target type |
| `type`           | `tinyint`     | 1 = like, -1 = dislike  |
| `timestamps`     | `datetime`    | Created/updated at      |

**Reactable Types:** `SongVariant`, `Comment`

#### `reaction_counters`

Cached reaction counts (polymorphic).

| Column           | Type          | Description                 |
| ---------------- | ------------- | --------------------------- |
| `id`             | `bigint` (PK) | Primary key                 |
| `reactable_id`   | `bigint`      | Polymorphic target ID       |
| `reactable_type` | `string`      | Polymorphic target type     |
| `likes_count`    | `bigint`      | Total likes (default: 0)    |
| `dislikes_count` | `bigint`      | Total dislikes (default: 0) |
| `timestamps`     | `datetime`    | Created/updated at          |

#### `favorites`

User favorites (polymorphic).

| Column             | Type          | Description             |
| ------------------ | ------------- | ----------------------- |
| `id`               | `bigint` (PK) | Primary key             |
| `user_id`          | `FK → users`  | User who favorited      |
| `favoritable_id`   | `bigint`      | Polymorphic target ID   |
| `favoritable_type` | `string`      | Polymorphic target type |
| `timestamps`       | `datetime`    | Created/updated at      |

**Favoritable Types:** `Song`, `SongVariant`

#### `comments`

User comments (polymorphic, with nested replies).

| Column             | Type                       | Description                |
| ------------------ | -------------------------- | -------------------------- |
| `id`               | `bigint` (PK)              | Primary key                |
| `parent_id`        | `FK → comments` (nullable) | Parent comment for replies |
| `commentable_id`   | `bigint`                   | Polymorphic target ID      |
| `commentable_type` | `string`                   | Polymorphic target type    |
| `user_id`          | `FK → users`               | Comment author             |
| `content`          | `text`                     | Comment text               |
| `timestamps`       | `datetime`                 | Created/updated at         |

**Commentable Types:** `SongVariant`

#### `playlists`

User-created playlists.

| Column        | Type                | Description                        |
| ------------- | ------------------- | ---------------------------------- |
| `id`          | `bigint` (PK)       | Primary key                        |
| `name`        | `string`            | Playlist name                      |
| `description` | `string` (nullable) | Playlist description               |
| `user_id`     | `FK → users`        | Playlist owner                     |
| `is_public`   | `boolean`           | Public visibility (default: false) |
| `timestamps`  | `datetime`          | Created/updated at                 |

---

### Moderation Tables

#### `reports`

User-submitted content reports.

| Column       | Type                      | Description        |
| ------------ | ------------------------- | ------------------ |
| `id`         | `bigint` (PK)             | Primary key        |
| `song_id`    | `FK → songs`              | Reported song      |
| `user_id`    | `FK → users`              | Reporter           |
| `source`     | `string`                  | Report source      |
| `title`      | `string`                  | Report title       |
| `content`    | `text`                    | Report description |
| `status`     | `enum('fixed','pending')` | Resolution status  |
| `timestamps` | `datetime`                | Created/updated at |

#### `user_requests`

User requests (add anime, fix issues).

| Column        | Type                         | Description        |
| ------------- | ---------------------------- | ------------------ |
| `id`          | `bigint` (PK)                | Primary key        |
| `content`     | `text`                       | Request content    |
| `user_id`     | `FK → users`                 | Requester          |
| `attended_by` | `FK → users` (nullable)      | Admin who handled  |
| `status`      | `enum('pending','attended')` | Processing status  |
| `timestamps`  | `datetime`                   | Created/updated at |

---

### Pivot Tables (Many-to-Many)

| Table                | Columns                              | Relationships                  |
| -------------------- | ------------------------------------ | ------------------------------ |
| `artist_song`        | `artist_id`, `song_id`               | Artists ↔ Songs                |
| `post_studio`        | `post_id`, `studio_id`               | Posts ↔ Studios                |
| `post_producer`      | `post_id`, `producer_id`             | Posts ↔ Producers              |
| `external_link_post` | `post_id`, `external_link_id`        | Posts ↔ External Links         |
| `playlist_song`      | `playlist_id`, `song_id`, `position` | Playlists ↔ Songs (with order) |

---

### System Tables

| Table                    | Purpose                      |
| ------------------------ | ---------------------------- |
| `password_resets`        | Password reset tokens        |
| `personal_access_tokens` | API tokens (Laravel Sanctum) |
| `failed_jobs`            | Failed queue jobs            |
| `cache`                  | Application cache storage    |

---

## Routes

Routes are defined in `routes/web.php` and `routes/api.php`.

---

### Web Routes (`routes/web.php`)

#### Public Routes

| Route                   | Method | Controller / Action        | Name              | Description               |
| ----------------------- | ------ | -------------------------- | ----------------- | ------------------------- |
| `/`                     | GET    | `PostController@index`     | `/`               | Homepage with top OPs/EDs |
| `/themes`               | GET    | `PostController@themes`    | `themes`          | Browse all themes         |
| `/anime/{slug}`         | GET    | `PostController@show`      | `post.show`       | Anime detail page         |
| `/animes`               | GET    | `PostController@animes`    | `animes`          | Browse all anime          |
| `/anime/{anime}/{song}` | GET    | `SongController@show`      | `songs.show`      | Song detail page          |
| `/seasonal`             | GET    | `SongController@seasonal`  | `seasonal`        | Seasonal songs view       |
| `/ranking`              | GET    | `SongController@ranking`   | `ranking`         | Song rankings             |
| `/welcome`              | GET    | `UserController@welcome`   | `welcome`         | Welcome/onboarding page   |
| `/users/{slug}`         | GET    | `UserController@userList`  | `user.list`       | User's rated themes list  |
| `/profile`              | GET    | `UserController@index`     | `profile`         | Current user's profile    |
| `/favorites`            | GET    | `UserController@favorites` | `favorites`       | Current user's favorites  |
| `/producers`            | GET    | `ProducerController@index` | `producers.index` | Browse all producers      |
| `/offline`              | GET    | (view)                     | —                 | Offline fallback page     |

#### Resource Routes (Public)

| Resource    | Controller           | Actions         |
| ----------- | -------------------- | --------------- |
| `artists`   | `ArtistController`   | `index`, `show` |
| `years`     | `YearController`     | Full resource   |
| `seasons`   | `SeasonController`   | Full resource   |
| `studios`   | `StudioController`   | Full resource   |
| `producers` | `ProducerController` | Full resource   |
| `playlists` | `PlaylistController` | Full resource   |

#### User Interaction Routes

| Route                             | Method | Controller / Action                    | Name                       | Auth? |
| --------------------------------- | ------ | -------------------------------------- | -------------------------- | ----- |
| `/change-score-format`            | POST   | `UserController@changeScoreFormat`     | `change.score.format`      | Yes   |
| `/upload-profile-pic`             | POST   | `UserController@uploadProfilePic`      | `upload.profile.pic`       | Yes   |
| `/upload-banner-pic`              | POST   | `UserController@uploadBannerPic`       | `upload.banner.pic`        | Yes   |
| `/variant/{variant}/rate`         | POST   | `SongVariantController@rate`           | `variant.rate`             | Yes   |
| `/variants/{variant}/like`        | POST   | `SongVariantController@like`           | `variants.like`            | Yes   |
| `/variants/{variant}/dislike`     | POST   | `SongVariantController@dislike`        | `variants.dislike`         | Yes   |
| `/variants/{variant}/favorite`    | POST   | `SongVariantController@toggleFavorite` | `variants.toggle.favorite` | Yes   |
| `/comments/{comment}/like`        | POST   | `CommentController@like`               | `comments.like`            | Yes   |
| `/comments/{comment}/dislike`     | POST   | `CommentController@dislike`            | `comments.dislike`         | Yes   |
| `/comments/{parentComment}/reply` | POST   | `CommentController@reply`              | `comments.reply`           | Yes   |
| `/reports/store`                  | POST   | `ReportController@store`               | `reports.store`            | Yes   |

#### Admin Routes (Prefix: `/admin`, Middleware: `staff`)

All admin routes use the `admin.` prefix for naming.

**Dashboard:**
| Route | Method | Controller / Action | Name |
|--------------------|--------|--------------------------------|-------------------|
| `/admin/dashboard` | GET | `Admin\PostController@dashboard`| `admin.dashboard` |

**Posts Management:**
| Route | Method | Action | Name |
|-------------------------------------|--------|------------------|-------------------------------|
| `/admin/posts` | GET | `index` | `admin.posts.index` |
| `/admin/posts/search` | POST | `search` | `admin.posts.search` |
| `/admin/posts/{post}/toggle-status` | POST | `toggleStatus` | `admin.posts.toggle.status` |
| `/admin/posts/{post}/songs` | GET | `songs` | `admin.posts.songs` |
| `/admin/posts/{post}/songs/add` | GET | `addSong` | `admin.posts.songs.add` |
| `/admin/posts/search-animes` | POST | `searchInAnilist`| `admin.posts.search.animes` |
| `/admin/posts/get-by-id/{id}` | GET | `getById` | `admin.posts.get.by.id` |
| `/admin/posts/get-seasonal-animes` | POST | `getSeasonalAnimes`| `admin.posts.get.seasonal.animes` |
| `/admin/posts/{id}/force-update` | GET | `forceUpdate` | `admin.posts.force.update` |
| `/admin/posts/wipe` | GET | `wipePosts` | `admin.posts.wipe` |

**Songs & Variants:**
| Route | Method | Action | Name |
|----------------------------------------|--------|--------------|------------------------------|
| `/admin/songs/{song}/variants` | GET | `variants` | `admin.songs.variants` |
| `/admin/songs/{song}/variants/add` | GET | `addVariant` | `admin.songs.variants.add` |
| `/admin/variants/{variant}/videos` | GET | `videos` | `admin.variants.videos` |
| `/admin/variants/{variant}/videos/add` | GET | `addVideos` | `admin.variants.videos.add` |

**Resource Controllers (Admin):**
| Resource | Controller | Notes |
|------------|------------------------------|----------------------|
| `songs` | `Admin\SongController` | Full CRUD |
| `variants` | `Admin\SongVariantController`| Full CRUD |
| `videos` | `Admin\VideoController` | Full CRUD |
| `posts` | `Admin\PostController` | Full CRUD |
| `artists` | `Admin\ArtistController` | Full CRUD + search |
| `users` | `Admin\UserController` | Full CRUD + search |
| `reports` | `Admin\ReportController` | Full CRUD + toggle |
| `requests` | `Admin\UserRequestController`| Full CRUD |
| `comments` | `Admin\CommentController` | Full CRUD |
| `studios` | `Admin\StudioController` | Full CRUD |
| `producers`| `Admin\ProducerController`| Full CRUD |
| `years` | `Admin\YearController` | Full CRUD + toggle |
| `seasons` | `Admin\SeasonController` | Full CRUD + toggle |

---

### API Routes (`routes/api.php`)

Base URL: `/api`

#### Public Endpoints

**Search & Browse:**
| Route | Method | Controller / Action | Name | Description |
|--------------------------|--------|----------------------------------|------------------------|--------------------------------|
| `/api/search/{q}` | GET | `Api\PostController@search` | `api.posts.search` | Search anime by title |
| `/api/animes` | GET | `Api\PostController@animes` | `api.posts.animes` | Get paginated anime list |

**Songs:**
| Route | Method | Controller / Action | Name | Description |
|----------------------------|--------|----------------------------------|------------------------|--------------------------------|
| `/api/songs/seasonal` | GET | `Api\SongController@seasonal` | `api.songs.seasonal` | Get seasonal songs |
| `/api/songs/ranking` | GET | `Api\SongController@ranking` | `api.songs.ranking` | Get song rankings |
| `/api/songs/filter` | GET | `Api\SongController@filter` | `api.songs.filter` | Filter songs with params |
| `/api/songs/{song}/comments`| GET | `Api\SongController@comments` | `api.songs.comments` | Get comments for a song |

**Variants:**
| Route | Method | Controller / Action | Name | Description |
|------------------------------------|--------|----------------------------------------|---------------------------|-------------------------------|
| `/api/variants/seasonal` | POST | `Api\SongVariantController@seasonal` | `api.variants.seasonal` | Get seasonal variants |
| `/api/variants/{variant}/comments` | GET | `Api\SongVariantController@comments` | `api.variants.comments` | Get comments for a variant |
| `/api/variants/{variant}/get-videos`| GET | `Api\SongVariantController@getVideos` | `api.variants.get-video` | Get video data for a variant |

**Artists:**
| Route | Method | Controller / Action | Name | Description |
|--------------------------------|--------|--------------------------------------|----------------------------|-------------------------------|
| `/api/artists/filter` | GET | `Api\ArtistController@artistsFilter` | `api.artists.filter` | Filter/search artists |
| `/api/artists/{artist}/filter` | GET | `Api\ArtistController@songsFilter` | `api.artists.songs.filter` | Get songs by artist |

**Studios:**
| Route | Method | Controller / Action | Name | Description |
|--------------------------------|--------|--------------------------------------|------------------------|-------------------------------|
| `/api/studios/filter` | GET | `Api\StudioController@filter` | `api.studios.filter` | Filter/search studios |
| `/api/studios/{studio}/songs` | GET | `Api\StudioController@songsFilter` | `api.studios.songs` | Get songs by studio |
| `/api/studios/{studio}/animes` | GET | `Api\StudioController@postsFilter` | `api.studios.posts` | Get anime by studio |

**Users:**
| Route | Method | Controller / Action | Name | Description |
|--------------------------|--------|-------------------------------|--------------------|-------------------------------|
| `/api/users` | GET | `Api\UserController@index` | `api.users` | List all users |
| `/api/users/{id}` | GET | `Api\UserController@show` | `api.users.show` | Get user profile |
| `/api/users/{id}/list` | GET | `Api\UserController@userList` | `api.users.list` | Get user's rated themes |

#### Authenticated Endpoints (Middleware: `auth:sanctum`)

**Playlists:**
| Route | Method | Controller / Action | Name |
|----------------------------------------|--------|--------------------------------------|--------------------------|
| `/api/playlists` | GET | `Api\PlaylistController@index` | `playlists.index` |
| `/api/playlists` | POST | `Api\PlaylistController@store` | `playlists.store` |
| `/api/playlists/{playlist}` | GET | `Api\PlaylistController@show` | `playlists.show` |
| `/api/playlists/{playlist}` | PUT | `Api\PlaylistController@update` | `playlists.update` |
| `/api/playlists/{playlist}` | DELETE | `Api\PlaylistController@destroy` | `playlists.destroy` |
| `/api/playlists/{playlist}/toggle-song`| POST | `Api\PlaylistController@toggleSong` | `playlists.toggle.song` |

**Variant Interactions:**
| Route | Method | Controller / Action | Name |
|----------------------------------|--------|--------------------------------------------|-----------------------------|
| `/api/variants/{variant}/like` | POST | `Api\SongVariantController@like` | `api.variants.like` |
| `/api/variants/{variant}/dislike`| POST | `Api\SongVariantController@dislike` | `api.variants.dislike` |
| `/api/variants/{variant}/favorite`| POST | `Api\SongVariantController@toggleFavorite` | `api.variants.toggle.favorite`|
| `/api/variants/{variant}/rate` | POST | `Api\SongVariantController@rate` | `api.variants.rate` |

**Song Interactions:**
| Route | Method | Controller / Action | Name |
|------------------------------|--------|------------------------------------------|-----------------------------|
| `/api/songs/{song}/like` | GET | `Api\SongController@like` | `api.songs.like` |
| `/api/songs/{song}/dislike` | GET | `Api\SongController@dislike` | `api.songs.dislike` |
| `/api/songs/{song}/favorite` | GET | `Api\SongController@toggleFavorite` | `api.songs.toggle.favorite` |
| `/api/songs/{song}/rate` | POST | `Api\SongController@rate` | `api.songs.rate` |
| `/api/songs/comments` | POST | `Api\SongController@storeComment` | `api.songs.store.comment` |
| `/api/songs/reports` | POST | `Api\SongController@storeReport` | `api.songs.reports` |

**Comment Interactions:**
| Route | Method | Controller / Action | Name |
|------------------------------------|--------|-----------------------------------|------------------------|
| `/api/comments/{id}/like` | GET | `Api\CommentController@like` | `api.comments.like` |
| `/api/comments/{id}/dislike` | GET | `Api\CommentController@dislike` | `api.comments.dislike` |
| `/api/comments/{parentComment}/reply`| POST | `Api\CommentController@reply` | `comments.reply` |
| `/api/comments` | POST | `Api\CommentController@store` | `api.comments.store` |
| `/api/comments/{comment}` | PUT | `Api\CommentController@update` | `api.comments.update` |
| `/api/comments/{comment}` | DELETE | `Api\CommentController@destroy` | `api.comments.destroy` |

**User Settings:**
| Route | Method | Controller / Action | Name |
|--------------------------|--------|--------------------------------------|--------------------------|
| `/api/users/avatar` | POST | `Api\UserController@uploadAvatar` | `api.users.upload.avatar`|
| `/api/users/banner` | POST | `Api\UserController@uploadBanner` | `api.users.upload.banner`|
| `/api/users/score-format`| POST | `Api\UserController@setScoreFormat` | `api.users.score.format` |
| `/api/users/favorites` | POST | `Api\UserController@favorites` | `api.users.favorites` |

**User Requests:**
| Route | Method | Controller / Action | Name |
|--------------------|--------|-----------------------------------------|-----------------------|
| `/api/requests` | GET | `Api\UserRequestController@index` | `api.requests.index` |
| `/api/requests` | POST | `Api\UserRequestController@store` | `api.requests.store` |
| `/api/requests/{id}`| DELETE| `Api\UserRequestController@destroy` | `api.requests.destroy`|

---

## Livewire Components

The application heavily utilizes **Livewire** for reactive UI components, especially for filtering, searching, and user settings.

### User & Navigation

- **`UserSettings`**: Handles avatar and banner uploads with reactive previews. Manages scoring system preferences.
- **`FavoritesTable`**: Powers the "My Favorites" page with infinite scroll and multi-criteria filtering.
- **`AnimesTable`**: Dynamic anime search and grid with view mode switching (List/Grid).
- **`ArtistsTable`**: Artist listing with real-time search and A-Z sorting.
- **`ArtistThemesTable`**: Specialized song listing for artist profile pages.
- **`StudiosTable`**: Studio discovery with infinite scroll and dynamic counting.
- **`StudioAnimesTable`**: Catalog of anime produced by a specific studio with advanced filtering.
- **`ProducersTable`**: Producer discovery with series counting and alphabetical sorting.
- **`ProducerAnimesTable`**: specialized grid for series produced by a specific company.

### Discovery & Ranking

- **`RankingTable`**: Global and seasonal leaderboards for songs.
- **`SeasonalTable`**: Browse themes by season with OP/ED toggles and infinite scroll.
- **`ThemesTable`**: The primary discovery engine for all themes across the platform.

### Implementation Patterns

- **Infinite Scroll**: Uses Alpine.js with an `Intersection Observer` to trigger `loadMore()` on components.
- **Event Listeners**: Components like `UserSettings` emit events (e.g., `avatarUpdated`) that other parts of the UI listen for to update shared headers/banners without a refresh.
- **Glassmorphic Design**: All component views use `surface-dark` backgrounds and `backdrop-blur-md` for a consistent premium look.

---

## Conventions

1.  **Middleware**:
    - `auth`: For logged-in users.
    - `staff`: For admin/editor/creator roles.
2.  **Asset Management**: Use `@vite` for main assets. Use `@push('scripts')` / `@push('styles')` for page-specific resources.
3.  **Configuration**: Always use `config()` instead of `env()` in views and application code.
4.  **Routes**: Grouped by context (`public`, `admin.`) with resource controllers where applicable.
