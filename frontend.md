# Requerimientos para Frontend Reactivo en Svelte (Anirank SPA)

Este documento detalla la arquitectura, las rutas y los componentes necesarios para construir el portal público (SPA) utilizando **Svelte / SvelteKit**, basando sus funcionalidades en las rutas existentes del antiguo frontend monolítico en Blade (`routes/web.php`).

---

## 🏗️ 1. Arquitectura de Rutas SvelteKit (`src/routes`)

La estructura de carpetas en SvelteKit se mapea usando los mismos slugs amigables actuales.

| Ruta en SvelteKit (Frontend)         | Equivalencia en `web.php`       | Propósito                                                                                               |
| :----------------------------------- | :------------------------------ | :------------------------------------------------------------------------------------------------------ |
| ✅ `/`                               | `/`                             | Página de inicio (Home). Muestra OPs/EDs populares, recientes, rankings semanales y el buscador global. |
| ✅ `/animes`                         | `/animes`                       | Catálogo de Animes. Grid interactivo con barra flotante premium para buscar y filtrar.                  |
| ✅ `/animes/[slug]`                  | `/anime/{post:slug}`            | Ficha técnica de un Anime. Muestra los openings, endings insert songs y metadata.                       |
| ✅ `/songs`                          | `/songs`                        | Directorio de Canciones de Anime. (Index general con filtros premium).                                  |
| ❌ `/songs/seasonal`                 | `/songs/seasonal`               | Canciones de Animes de la temporada actual. (Pendiente de estandarización premium).                     |
| ❌ `/songs/ranking`                  | `/songs/ranking`                | Top 100 de las mejores canciones de la historia. (Pendiente de estandarización premium).                |
| ❌ `/songs/[anime_slug]/[song_slug]` | `/song/{post:slug}/{song:slug}` | Ficha de una Canción específica (Modo Teatro). (Pendiente de estandarización de interfaz interactiva).  |
| ✅ `/playlists`                      | `/playlists`                    | Directorio de Playlists de la comunidad. (Estandarizado con filtros premium).                           |
| ❌ `/users/[user_slug]`              | `/users/{user:slug}`            | Perfil público de un usuario. Muestra badges, playlists públicas y actividad.                           |
| ❌ `/settings`                       | `/settings`                     | Panel de Cuenta (Privado). Cambiar el sistema de calificación, avatares, banners, etc.                  |
| ❌ `/users/[user_slug]/favorites`    | `/{user:slug}/favorites`        | Lista de canciones y animes favoritos del usuario.                                                      |
| ✅ `/artists`                        | `/artists`                      | Directorio indexado de todos los cantantes o bandas. (Estandarizado).                                   |
| ✅ `/artists/[slug]`                 | `/artists/{artist:slug}`        | Ficha de un Artista y todos los temas que ha interpretado. (Estandarizado).                             |
| ✅ `/studios`                        | `/studios`                      | Directorio de estudios de animación. (Estandarizado).                                                   |
| ✅ `/studios/[slug]`                 | `/studios/{studio:slug}`        | Ficha del estudio de animación (ej: MAPPA, Ufotable) y sus obras. (Estandarizado).                      |
| ✅ `/producers`                      | `/producers`                    | Directorio de productoras discográficas o de anime. (Estandarizado).                                    |
| ✅ `/producers/[slug]`               | `/producers/{producer:slug}`    | Ficha de productoras discográficas o de anime. (Estandarizado).                                         |

---

## ⚛️ 2. Componentes Fundamentales a Desarrollar

Para garantizar la usabilidad y que la interfaz se sienta "Premium", se deben construir los siguientes componentes de UI en Svelte interactivos:

### Navigación y Global

- ✅ `Navbar.svelte`: Barra opaca/Glassmorphism. Incluye: Buscador Predictivo y Dropdown de Usuario. _(Implementado como `NavbarMaster.svelte`)_
- ✅ `AuthModal.svelte`: Modal de Login/Registro para no perder el contexto de la página actual si el usuario quiere guardar a Favoritos.
- ✅ `Footer.svelte`: Footer simple con SEO links de Anirank. _(Implementado como `FooterMaster.svelte`)_

### Componentes de Dominio (Content Cards)

- ✅ `AnimeCard.svelte`: Tarjeta para Animes, incluye hover dinámico que muestra resumen, estudios, año y temporada.
- ✅ `SongCard.svelte`/`VariantCard.svelte`: Tarjeta para Canciones o versiones. Con hover dinámico que activa preview del thumbnail o botones directos de "Me gusta" o "Añadir a Playlist".
- ❌ `ArtistCard.svelte`: Tarjeta circular (generalmente) para artistas musicales.
- ✅ `CustomSelect.svelte`: Selector premium personalizado que emula un popover para permitir una estilización extrema (backgrounds, hovers, scrollbars) imposible en selects nativos.

### Reproductores e Interactividad Crítica

- ⏳ `TheatrePlayer.svelte`: Reproductor de video encapsulado (YouTube Iframe API v3 o Video.js para videos propios). _(Integrado funcionalmente en el `+page.svelte` de la canción, sin extraer como componente global independiente aún)_
- ✅ `RatingStars.svelte`: Componente interactivo que transforma los puntos 1-100 o 1-5 dependiendo del ajuste global de perfil del usuario. _(Implementado vía `RatingModal.svelte`)_
- ⏳ `InteractionBar.svelte`: Para interactuar con las canciones. Debe agrupar botones de: Like, Dislike, Agregar a Playlist, Score, Comentarios. _(Integrado funcionalmente en la vista, falta aislar como componente independiente)_
- ✅ `PlaylistsSelectorModal.svelte`: Modal global que se dispara desde cualquier `InteractionBar` para decidir a qué lista añadir la canción actual. _(Implementado usando `PlaylistModal.svelte` y `CreatePlaylistModal.svelte`)_

### Sistema de Comentarios

- ✅ `CommentSection.svelte`: Input principal (Editor enriquecido tipo markdown ligero). _(Lógica y UI integrada)_
- ✅ `CommentThread.svelte`: Soporte para cargar respuestas recursivas (Replies). _(Lógica y UI integrada)_

### Listados Mágicos (Filters & Paginators)

- ❌ `DynamicGrid.svelte`: Grid CSS/Tailwind virtualizado o con Infinite Scroll (Intersection Observer de SvelteKit) para la página `/animes`.

---

## 🔌 3. Integración con la API Central

### Inicialización (`routes/+layout.svelte` en SvelteKit)

Todo flujo en la app debe arrancar pidiendo la base al nuevo endpoint que hemos diseñado en la fase de refactoring:

- ✅ **`GET /api/init`**: Obténer de golpe en el stores local (ej. `$globalConfig`): Años, Temporadas, Formatos y Géneros (Crucial para construir menús desplegables globales sin cargas asíncronas).
- ✅ **`GET /api/auth/me`**: Ejecutar el handshake con Sanctum para validar la sesión y guardar estatus del estado (`$user`).

### Endpoints Cruciales a Consumir por Página

- ✅ **Página de Inicio**: Petición a `/api/home`. _(Integrado)_
- ✅ **Animes Explorer**: Petición a `/api/animes` reaccionando a las _stores URL params_ (`?season_id=X&genre_id=X`). _(Integrado)_
- ✅ **Ver Anime**: Petición a `/api/posts/[slug]`. _(Integrado)_
- ✅ **Comentarios Polimórficos**: Tanto canciones como posts pueden llamar a `/api/comments` enviando su `TYPE` y el `ID` para poblar el componente `CommentSection.svelte`. _(Integrado)_

---

## 🛡️ 4. Seguridad (Sanctum SPA Authentication)

**Laravel Sanctum Cookie-Based Authentication Flow** en SvelteKit:

1. El servidor de Node.js (SSR SvelteKit) y el cliente deben mantener sus credenciales CORS encendidas (`credentials: 'include'`).
2. SvelteKit debe solicitar el CSRF-Cookie a Laravel (`GET /sanctum/csrf-cookie`) en el _mount_ inicial de su componente `App` **solo una vez**.
3. SvelteKit hace POST a `/api/auth/login`. Sanctum establecerá un token HTTPOnly en forma de cookie al Frontend.
4. **NO** se manejará ningún Token en `localStorage`, la solicitud viajará de regreso siempre que se use una instancia pre-configurada de `axios` o un interceptor seguro para todas las peticiones a la API.

---

## 💅 5. Herramientas y Framework de Diseño Recomendado

- **Framework**: `Svelte 5` con `SvelteKit`.
- **Styling**: `Tailwind CSS v4`.
- **State Management**: `Svelte Runes` (el nuevo estándar para stores más rápidos y ligeros en lugar de context API obsoleto).
- **Motion**: `svelte/transition` para microinteracciones y page-transitions o un plugin ligero como `framer-motion` / `Motion`.
- **HTTP Client**: `axios` con interceptores predeterminados o `ky`.
- **Iconografía**: `Phosphor Icons` o `Material Symbols` (para unificar con el rediseño anterior del Admin backend Laravel).

---

## 🎨 6. Guía de Diseño de Formularios y Filtros (Premium UI)

Para mantener la consistencia visual "Premium" en toda la SPA, los formularios y secciones de filtrado deben seguir este patrón estandarizado (basado en la implementación de `/animes`):

### Contenedor de Filtros (Glassmorphism & Layering)

```html
<!-- La sección debe tener un z-index elevado (z-40) para que los floats no se corten -->
<section
  class="relative z-40 flex flex-col gap-4 bg-surface-dark/30 p-4 rounded-3xl border border-white/5 backdrop-blur-md shadow-2xl"
>
  <!-- Layout en grid 12 (Estructura recomendada: 3-2-2-2-3) -->
  <div class="grid grid-cols-1 lg:grid-cols-12 gap-4 items-end">...</div>
</section>
```

### Convención de Opción "Any" (Data Clearing)

Para permitir que el usuario limpie filtros individuales sin necesidad de un reset global:

- **Valor**: El valor `"any"` está reservado para indicar "Sin filtro".
- **Lógica**: La función de actualización de filtros debe detectar este valor y realizar un `url.searchParams.delete()` en lugar de enviar el parámetro.
- **Implementación**:
  ```javascript
  const setParam = (key, val) => {
    if (val && val !== "any") url.searchParams.set(key, val);
    else url.searchParams.delete(key);
  };
  ```

### Inputs y Selects Estándar (CustomSelect)

- **Uso**: Se debe usar `CustomSelect.svelte` para menús desplegables premium.
- **Manejo de Capas**: El componente gestiona su propio `z-index` dinámico (`z-50` cuando está abierto) para sobreponerse a otros elementos del formulario.
- **Ventajas**:
  - Fondos con `backdrop-blur` y opacidad personalizada.
  - Hover states con colores de marca (`primary`).
  - Scrollbars estandarizadas.
- **Configuración**: Recibe un array de `options` (objetos `{value, label}`) y soporta `bind:value`.

### Botones de Acción (Ej: Reset/Trash)

- **Estilo**: Fondo sutil con color semántico (`red-500/10`), borde del mismo tono (`border-red-500/20`) y hover interactivo.
- **Clases**: `bg-red-500/10 hover:bg-red-500/20 text-red-500 px-4 rounded-xl font-bold transition-all shadow-lg shadow-red-500/5 flex items-center justify-center min-w-[48px]`
