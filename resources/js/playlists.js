import { API, csrfToken, token, bootstrap } from '@/app.js';

const currentSong = document.querySelector('#add-to-playlist').dataset.songId;
let headersData = {};
let bodyData = {};

class PlaylistManager {
    constructor() {
        this.cache = new Map(); // ← caché: songId → { playlists, timestamp }
        this.currentSong = null;
        this.addToPlaylistModal = null;
        this.createPlaylistModal = null;
        this.init();
    }

    init() {
        this.currentSong = document.querySelector('#add-to-playlist')?.dataset.songId;
        this.initModals();
        this.bindEvents();
    }

    initModals() {
        const addEl = document.getElementById('addToPlaylistModal');
        const createEl = document.getElementById('createPlaylistModal');
        this.addToPlaylistModal = addEl ? new bootstrap.Modal(addEl) : null;
        this.createPlaylistModal = createEl ? new bootstrap.Modal(createEl) : null;
    }

    bindEvents() {
        // Abrir modal principal
        document.querySelectorAll('#add-to-playlist').forEach(btn => {
            btn.addEventListener('click', () => {
                this.openAddToPlaylistModal();
            });
        });

        // Añadir/quitar canción
        document.addEventListener('click', e => {
            if (e.target.matches('.add-to-playlist-btn')) {
                const playlistId = e.target.dataset.playlistId;
                this.addPostToPlaylist(playlistId, e.target);
            }
        });

        // Crear playlist
        document.getElementById('createPlaylistForm')?.addEventListener('submit', e => {
            e.preventDefault();
            this.createPlaylist();
        });

        // Botón "New Playlist"
        document.getElementById('createNewPlaylistBtn')?.addEventListener('click', () => {
            this.addToPlaylistModal?.hide();
            this.createPlaylistModal?.show();
        });
    }

    async openAddToPlaylistModal() {
        // ← Solo carga si NO está en caché o si es una canción diferente
        if (!this.cache.has(this.currentSong)) {
            await this.loadPlaylists();
        }

        renderPlaylistsQuick(this.cache.get(this.currentSong).playlists, '#playlist-list');
        this.addToPlaylistModal?.show();
    }

    async loadPlaylists() {
        try {
            const response = await API.get(API.PLAYLISTS.BASE, {
                'X-CSRF-TOKEN': csrfToken,
                'Authorization': `Bearer ${token}`,
                'Accept': 'application/json'
            }, { song_id: this.currentSong });

            this.cache.set(this.currentSong, {
                playlists: response.playlists,
                timestamp: Date.now()
            });
        } catch (error) {
            console.error('Error cargando playlists:', error);
            this.showNotification('Error al cargar playlists', 'error');
        }
    }

    async addPostToPlaylist(playlistId, button) {
        const originalText = button.textContent;
        button.textContent = 'Adding...';
        button.disabled = true;

        headersData = {
            'Content-Type': 'application/json',
            'X-CSRF-TOKEN': csrfToken,
            'Authorization': `Bearer ${token}`,
            'Accept': 'application/json'
        };
        bodyData = JSON.stringify({ song_id: currentSong });
        
        try {
            const response = await API.post(API.PLAYLISTS.TOGGLE_SONG(playlistId), headersData, bodyData);

            if (response.success) {
                const { in_playlist, action } = response.data;
                updateUI(playlistId, in_playlist, action, button);

                // ← ACTUALIZAR CACHÉ
                const cached = this.cache.get(this.currentSong);
                if (cached) {
                    const playlist = cached.playlists.find(p => p.id == playlistId);
                    if (playlist) {
                        playlist.is_in_playlist = in_playlist;
                        playlist.songs_count += in_playlist ? 1 : -1;
                    }
                }

                this.showNotification(response.message, 'success');
                setTimeout(() => this.addToPlaylistModal?.hide(), 800);
            } else {
                throw new Error(response.message);
            }
        } catch (error) {
            button.textContent = 'Error';
            this.showNotification('Error al modificar playlist', 'error');
            setTimeout(() => {
                button.textContent = originalText;
                button.disabled = false;
            }, 1500);
        }
    }

    async createPlaylist() {
        const form = document.getElementById('createPlaylistForm');
        const submitBtn = form.querySelector('button[type="submit"]');
        const originalText = submitBtn.textContent;

        try {
            submitBtn.textContent = 'Creando...';
            submitBtn.disabled = true;

            const formData = new FormData(form);
            const data = Object.fromEntries(formData.entries());
            data.song_id = currentSong; // Añadir canción automáticamente
            
            const response = await API.post(
                API.PLAYLISTS.BASE,
                {
                    'Content-Type': 'application/json',
                    'X-CSRF-TOKEN': csrfToken,
                    'Authorization': `Bearer ${token}`,
                    'Accept': 'application/json'
                },
                JSON.stringify(data)
            );

            if (response.status === 201 || response.playlist) {
                form.reset();
                this.createPlaylistModal?.hide();
                this.showNotification('Playlist creada', 'success');

                // ← AÑADIR AL DOM (sin request)
                renderNewPlaylistElement(response.playlist);

                // ← ACTUALIZAR CACHÉ
                const cached = this.cache.get(this.currentSong) || { playlists: [] };
                cached.playlists.unshift({
                    ...response.playlist,
                    is_in_playlist: true,
                    songs_count: (response.playlist.songs_count || 0) + 1
                });
                this.cache.set(this.currentSong, cached);

                this.addToPlaylistModal?.show();
            } else {
                throw new Error(response.message);
            }
        } catch (error) {
            console.error(error);
            this.showNotification('Error al crear playlist', 'error');
            submitBtn.textContent = originalText;
            submitBtn.disabled = false;
        }
    }

    showNotification(message, type) {
        const alert = document.createElement('div');
        alert.className = `alert alert-${type === 'success' ? 'success' : 'danger'} position-fixed`;
        alert.style.cssText = 'top: 20px; right: 20px; z-index: 1055; min-width: 300px;';
        alert.innerHTML = `
            <div class="d-flex align-items-center">
                <span>${message}</span>
                <button type="button" class="btn-close ms-auto" data-bs-dismiss="alert"></button>
            </div>
        `;
        document.body.appendChild(alert);
        setTimeout(() => alert.remove(), 3000);
    }
}

// === Plantilla de item ===
function createPlaylistElement(playlist) {
    const isInPlaylist = playlist.is_in_playlist || false;
    const btnClass = isInPlaylist ? 'btn-success' : 'btn-primary';
    const btnText = isInPlaylist ? 'Added' : 'Add';

    return `
        <div class="d-flex justify-content-between align-items-center p-2 border-bottom playlist-item" data-playlist-id="${playlist.id}">
            <div class="playlist-info">
                <strong>${escapeHtml(playlist.name)}</strong>
                <small class="text-muted" id="counter-${playlist.id}">
                    ${playlist.songs_count || 0} songs
                </small>
            </div>
            <button class="btn btn-sm ${btnClass} add-to-playlist-btn"
                    data-playlist-id="${playlist.id}">
                ${btnText}
            </button>
        </div>
    `;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// === Render nuevo elemento ===
function renderNewPlaylistElement(playlist) {
    const container = document.getElementById('playlist-list');
    const noMsg = document.getElementById('no-playlists-msg');
    if (!container) return;

    if (noMsg) noMsg.style.display = 'none';

    const html = createPlaylistElement({
        ...playlist,
        is_in_playlist: false // porque acabamos de añadir la canción
    });

    container.insertAdjacentHTML('afterend', html);
}

// === Renderizado rápido ===
function renderPlaylistsQuick(playlists, containerSelector) {
    const container = document.querySelector(containerSelector);
    const noMsg = document.getElementById('no-playlists-msg');
    if (!container) return;

    if (playlists.length === 0) {
        container.innerHTML = '';
        if (noMsg) noMsg.style.display = 'block';
        return;
    }

    if (noMsg) noMsg.style.display = 'none';
    container.innerHTML = playlists.map(createPlaylistElement).join('');
}

// === Actualizar UI ===
const updateUI = (playlistId, inPlaylist, action, button) => {
    button.classList.toggle('btn-success', inPlaylist);
    button.classList.toggle('btn-primary', !inPlaylist);
    button.textContent = inPlaylist ? 'Added' : 'Add';
    button.disabled = false;

    const counter = document.getElementById(`counter-${playlistId}`);
    if (counter) {
        let count = parseInt(counter.textContent) || 0;
        counter.textContent = (inPlaylist ? count + 1 : count - 1) + ' songs';
    }
};

// === Inicializar ===
document.addEventListener('DOMContentLoaded', () => {
    new PlaylistManager();
});