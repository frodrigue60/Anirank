{{-- resources/views/components/create-playlist-modal.blade.php --}}
{{-- <div id="createPlaylistModal" class="modal" style="display: none;">
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
</div> --}}
<div class="modal fade" id="createPlaylistModal" tabindex="-1" aria-labelledby="createPlaylistModalLabel" aria-hidden="true">
    <div class="modal-dialog">
        <div class="modal-content">
            <div class="modal-header">
                <h5 class="modal-title" id="createPlaylistModalLabel">Crear Nueva Playlist</h5>
                <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
            </div>

            <div class="modal-body">
                <form id="createPlaylistForm">
                    @csrf
                    <div class="mb-3">
                        <label for="playlistName" class="form-label">Nombre</label>
                        <input type="text" class="form-control" id="playlistName" name="name" required>
                    </div>

                    <div class="mb-3">
                        <label for="playlistDescription" class="form-label">Descripción</label>
                        <textarea class="form-control" id="playlistDescription" name="description" rows="3"></textarea>
                    </div>

                    <div class="d-grid">
                        <button type="submit" class="btn btn-sm btn-success">Crear Playlist</button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</div>