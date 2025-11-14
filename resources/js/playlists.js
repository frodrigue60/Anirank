// public/js/playlists.js
import { API, csrfToken, token } from '@/app.js';

let headersData = null;
let bodyData = null;
const currentSong = document.querySelector('#add-to-playlist').dataset.songId;
//console.log(currentSong);

class PlaylistManager {
    constructor() {
        this.initModals();
        this.bindEvents();
        //this.currentSong = null;
    }

    initModals() {
        this.createPlaylistModal = document.getElementById('createPlaylistModal');
        this.addToPlaylistModal = document.getElementById('addToPlaylistModal');
    }

    bindEvents() {
        // Abrir modal de añadir a playlist
        document.querySelectorAll('#add-to-playlist').forEach(btn => {
            btn.addEventListener('click', (e) => {
                //this.currentSong = e.target.dataset.postId;
                this.openAddToPlaylistModal();
            });
        });

        // Añadir post a playlist existente
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('add-to-playlist-btn')) {
                const playlistId = e.target.dataset.playlistId;
                this.addPostToPlaylist(playlistId, e.target);
            }
        });

        // Crear nueva playlist desde modal
        document.getElementById('createPlaylistForm')?.addEventListener('submit', (e) => {
            e.preventDefault();
            this.createPlaylist();
        });

        // Botón para crear nueva playlist
        document.getElementById('createNewPlaylistBtn')?.addEventListener('click', () => {
            this.closeAddToPlaylistModal();
            this.openCreatePlaylistModal();
        });

        // Cerrar modales
        document.querySelectorAll('.modal .close, .close-modal').forEach(closeBtn => {
            closeBtn.addEventListener('click', () => {
                this.closeAllModals();
            });
        });

        // Cerrar modal al hacer click fuera
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
        //document.getElementById('currentSong').value = this.currentSong;
        this.addToPlaylistModal.style.display = 'block';
        //this.updatePlaylistButtonsState();
    }

    updatePlaylistButtonsState() {
        // Aquí podrías hacer una petición para verificar el estado actual
        // y actualizar los botones (añadir/remover)
    }

    openCreatePlaylistModal() {
        this.createPlaylistModal.style.display = 'block';
    }

    closeAllModals() {
        this.createPlaylistModal.style.display = 'none';
        this.addToPlaylistModal.style.display = 'none';
        //this.currentSong = null;
    }

    closeAddToPlaylistModal() {
        this.addToPlaylistModal.style.display = 'none';
        //this.currentSong = null;
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

                // Cerrar modal después de un tiempo
                setTimeout(() => {
                    this.closeAllModals();
                }, 1000);

                updateUI(currentSong, playlistId, response.data.in_playlist, response.data.action, button);
                //getPlaylists();


            } else {
                //throw new Error(response.message || 'Error al añadir a la playlist');
                throw new Error("No se pudo añadir a la playlist");
            }
        } catch (error) {
            console.log(error);

            button.textContent = 'Error';
            this.showNotification('error', 'error');
            // Restaurar botón después de un tiempo
            setTimeout(() => {
                this.closeAllModals();
                button.textContent = 'Add';
                button.disabled = false;
            }, 1000);
            //getPlaylists();
        }
    }

    async createPlaylist() {
        const form = document.getElementById('createPlaylistForm');
        const submitButton = form.querySelector('button[type="submit"]');
        const originalText = submitButton.textContent;

        try {
            submitButton.textContent = 'Creando...';
            submitButton.disabled = true;

            const formData = new FormData(form);

            headersData = {
                'X-CSRF-TOKEN': csrfToken,
                'Accept': 'application/json',
                'Authorization': 'Bearer ' + token
            };
            bodyData = formData;

            const response = await API.post(API.PLAYLISTS.BASE, headersData, bodyData);

            let data = response || {};
            console.log(response);

            if (response.status === 201) {
                this.showNotification('Playlist created', 'success');
                form.reset();

                this.closeCreatePlaylistModal();
                getPlaylists();
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
        // Puedes implementar un sistema de notificaciones más elegante
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


// Función simple para generar un elemento de playlist
function createPlaylistElement(playlist, currentSong) {
    const isInPlaylist = currentSong &&
        playlist.posts &&
        playlist.posts.some(post => post.id == currentSong);

    return `
        <div class="playlist-item" data-playlist-id="${playlist.id}">
            <div class="playlist-info">
                <strong>${playlist.name}</strong>
                <small>${playlist.posts_count || 0} posts</small>
            </div>
            <button class="add-to-playlist-btn" 
                    data-playlist-id="${playlist.id}"
                    ${isInPlaylist ? 'style="background: #28a745;"' : ''}>
                ${isInPlaylist ? '✓ En playlist' : 'Añadir'}
            </button>
        </div>
    `;
}

// Función para renderizar playlists rápidamente
function renderPlaylistsQuick(playlists, containerSelector, currentSong) {
    const container = document.querySelector(containerSelector);
    container.innerHTML = '';
    if (!container) return;

    if (playlists.length === 0) {
        container.innerHTML = '<p class="no-playlists">No tienes playlists creadas</p>';
        return;
    }

    container.innerHTML = playlists.map(playlist =>
        createPlaylistElement(playlist, currentSong)
    ).join('');
}

function renderSsrPlaylists(html, containerSelector) {
    const container = document.querySelector(containerSelector);
    container.innerHTML = '';
    container.innerHTML = html;
    if (!container) return;
}

// Función para actualizar la UI
const updateUI = (currentSong, playlistId, inPlaylist, action, button) => {
    // Ejemplo: cambiar icono/botón según el estado

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
        'Authorization': 'Bearer ' + token
    };

    let params = {
        song_id: currentSong
    };

    const response = await API.get(API.PLAYLISTS.BASE, headersData, params);

    //let data = response || {};
    console.log(response);

    if (response.playlists.length === 0) {
        return;
    }

    //renderPlaylistsQuick(response.playlists, '#playlist-list', currentSong);
    renderSsrPlaylists(response.html, '#playlist-list');
}

// Inicializar cuando el DOM esté listo
document.addEventListener('DOMContentLoaded', () => {
    new PlaylistManager();
    getPlaylists();
});