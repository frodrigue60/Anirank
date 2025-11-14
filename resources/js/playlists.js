// public/js/playlists.js
import { API, csrfToken, token } from '@/app.js';

let headersData = null;
let bodyData = null;
const currentSong = document.querySelector('#add-to-playlist').dataset.songId;

class PlaylistManager {
    constructor() {
        this.initModals();
        this.bindEvents();
    }

    initModals() {
        this.createPlaylistModal = document.getElementById('createPlaylistModal');
        this.addToPlaylistModal = document.getElementById('addToPlaylistModal');
    }

    bindEvents() {
        document.querySelectorAll('#add-to-playlist').forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.openAddToPlaylistModal();
            });
        });

        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('add-to-playlist-btn')) {
                const playlistId = e.target.dataset.playlistId;
                this.addPostToPlaylist(playlistId, e.target);
            }
        });

        document.getElementById('createPlaylistForm')?.addEventListener('submit', (e) => {
            e.preventDefault();
            this.createPlaylist();
        });

        document.getElementById('createNewPlaylistBtn')?.addEventListener('click', () => {
            this.closeAddToPlaylistModal();
            this.openCreatePlaylistModal();
        });

        document.querySelectorAll('.modal .close, .close-modal').forEach(closeBtn => {
            closeBtn.addEventListener('click', () => {
                this.closeAllModals();
            });
        });

        window.addEventListener('click', (e) => {
            if (e.target === this.createPlaylistModal) {
                this.closeAllModals();
            }
            if (e.target === this.addToPlaylistModal) {
                this.closeAllModals();
            }
        });
    }

    openAddToPlaylistModal() {
        this.addToPlaylistModal.style.display = 'block';
    }

    updatePlaylistButtonsState() {
    }

    openCreatePlaylistModal() {
        this.createPlaylistModal.style.display = 'block';
    }

    closeAllModals() {
        this.createPlaylistModal.style.display = 'none';
        this.addToPlaylistModal.style.display = 'none';
    }

    closeAddToPlaylistModal() {
        this.addToPlaylistModal.style.display = 'none';
    }

    closeCreatePlaylistModal() {
        this.createPlaylistModal.style.display = 'none';
    }

    async addPostToPlaylist(playlistId, button) {
        try {
            const originalText = button.textContent;
            button.textContent = 'Adding...';
            button.disabled = true;

            headersData = {
                'Content-Type': 'application/json',
                'Accept': 'application/json, text/html;q=0.9',
                'X-CSRF-TOKEN': csrfToken,
                'Authorization': 'Bearer ' + token,
            }
            bodyData = JSON.stringify({
                song_id: currentSong
            });

            const response = await API.post(API.PLAYLISTS.TOGGLE_SONG(playlistId), headersData, bodyData);
            console.log(response);

            if (response.success === true) {
                button.textContent = 'added';
                button.classList.add('btn-success');
                this.showNotification(response.message, 'success');

                setTimeout(() => {
                    this.closeAllModals();
                }, 1000);

                let inPlaylist = response.data.in_playlist;
                let action = response.data.action;

                updateUI(playlistId, inPlaylist, action, button);
            } else {
                throw new Error("No se pudo añadir a la playlist");
            }
        } catch (error) {
            console.log(error);

            button.textContent = 'Error';
            this.showNotification('error', 'error');
            setTimeout(() => {
                this.closeAllModals();
                button.textContent = 'Add';
                button.disabled = false;
            }, 1000);
        }
    }

    async createPlaylist() {
        const form = document.getElementById('createPlaylistForm');
        const submitButton = form.querySelector('button[type="submit"]');
        const originalText = submitButton.textContent;

        try {
            submitButton.textContent = 'Creando...';
            submitButton.disabled = true;

            headersData = {
                'X-CSRF-TOKEN': csrfToken,
                'Authorization': 'Bearer ' + token,
                'Accept': 'application/json, text/html;q=0.9',
                'Content-Type': 'application/json',
            };

            const formData = new FormData(form);
            const json = Object.fromEntries(formData.entries());
            bodyData = JSON.stringify(json)

            const response = await API.post(API.PLAYLISTS.BASE, headersData, bodyData);

            if (response.status === 201) {
                this.showNotification('Playlist created', 'success');
                form.reset();
                submitButton.textContent = originalText;
                submitButton.disabled = false;
                this.closeAllModals();
                console.log(response);
                renderNewPlaylistElement(response.playlist, '#playlist-list');
                this.openAddToPlaylistModal();

            } else {
                throw new Error(data.message || 'Error al crear playlist (front side)');
            }
        } catch (error) {
            console.error('Error:', error);
            this.showNotification('Error al crear playlist', 'error');
            submitButton.textContent = originalText;
            submitButton.disabled = false;
        }
    }

    showNotification(message, type) {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 15px 20px;
            border-radius: 5px;
            color: white;
            z-index: 10000;
            background: ${type === 'success' ? '#28a745' : '#dc3545'};
        `;

        document.body.appendChild(notification);

        setTimeout(() => {
            notification.remove();
        }, 3000);
    }
}



function createPlaylistElement(playlist) {
    const isInPlaylist = playlist.is_in_playlist || false;

    const buttonClass = isInPlaylist ? 'btn-success' : 'btn-primary';
    const buttonText = isInPlaylist ? 'Added' : 'Add';

    return `
        <div class="playlist-item" data-playlist-id="${playlist.id}">
            <div class="playlist-info">
                <strong>${escapeHtml(playlist.name)}</strong>
                <small id="counter-${playlist.id}">
                    ${playlist.songs_count || 0} songs
                </small>
            </div>

            <button class="btn btn-sm ${buttonClass} add-to-playlist-btn"
                    data-playlist-id="${playlist.id}"
                    data-song-id="${playlist.song_id || window.currentSongId}">
                ${buttonText}
            </button>
        </div>
    `;
}

// Función segura para escapar HTML
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function renderNewPlaylistElement(playlist, containerSelector) {

    const container = document.querySelector(containerSelector);
    if (!container) return;
    let btnClass = '';
    let btnText = '';
    if (playlist.isInPlaylist) {
        btnClass = 'btn-success';
        btnText = 'Added';
    } else {
        btnClass = 'btn-primary';
        btnText = 'Add';
    }
    container.innerHTML += `
        <div class="playlist-item" data-playlist-id="${playlist.id}">
            <div class="playlist-info">
                <strong>${playlist.name}</strong>
                <small>${playlist.songs_count || 0} posts</small>
            </div>
            <button class="btn btn-sm  ${btnClass} add-to-playlist-btn" 
                    data-playlist-id="${playlist.id}">
                ${btnText}
            </button>
        </div>
    `;
}


function renderPlaylistsQuick(playlists, containerSelector, currentSong) {
    const container = document.querySelector(containerSelector);
    if (!container) return;

    if (playlists.length === 0) {
        container.innerHTML = '<p class="no-playlists">No tienes playlists creadas</p>';
        return;
    }

    container.innerHTML = playlists.map(playlist =>
        createPlaylistElement(playlist, currentSong)
    ).join('');
}

const updateUI = (playlistId, inPlaylist, action, button) => {

    if (inPlaylist) {
        button.classList.remove('btn-primary');
        button.classList.add('btn-success');
    } else {
        button.classList.remove('btn-success');
        button.classList.add('btn-primary');
    }

    button.disabled = false;

    const counter = document.querySelector('#counter-' + playlistId);

    if (counter) {
        let count = parseInt(counter.textContent);
        if (action === 'added') {
            count++;
            button.textContent = 'Added';
        } else if (action === 'removed') {
            count--;
            button.textContent = 'Add';
        }
        counter.textContent = count;
    }

};

async function getPlaylists() {
    headersData = {
        'X-CSRF-TOKEN': csrfToken,
        'Accept': 'application/json',
        'Authorization': 'Bearer ' + token,
        'Accept': 'application/json, text/html;q=0.9',
    };

    let params = {
        song_id: currentSong
    };

    const response = await API.get(API.PLAYLISTS.BASE, headersData, params);

    if (response.playlists.length === 0) {
        return;
    }

    renderPlaylistsQuick(response.playlists, '#playlist-list', currentSong);
}

document.addEventListener('DOMContentLoaded', () => {
    new PlaylistManager();
    getPlaylists();
});