{{-- resources/views/components/create-playlist-modal.blade.php --}}
<div id="createPlaylistModal" class="modal" style="display: none;">
    <div class="modal-content">
        <span class="close">&times;</span>
        <h3>Crear Nueva Playlist</h3>
        <form id="createPlaylistForm">
            @csrf
            <div class="form-group">
                <label for="playlistName">Nombre:</label>
                <input type="text" id="playlistName" name="name" required>
            </div>
            <div class="form-group">
                <label for="playlistDescription">Descripción:</label>
                <textarea id="playlistDescription" name="description"></textarea>
            </div>
            <button type="submit" class="btn btn-sm btn-primary">Crear Playlist</button>
        </form>
    </div>
</div>