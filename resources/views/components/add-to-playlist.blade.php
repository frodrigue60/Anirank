{{-- resources/views/components/add-to-playlist.blade.php --}}
<div class="modal fade" id="addToPlaylistModal" tabindex="-1" aria-labelledby="addToPlaylistModalLabel" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered modal-dialog-scrollable">
        <div class="modal-content">
            <div class="modal-header">
                <h5 class="modal-title" id="addToPlaylistModalLabel">Add to playlist</h5>
                <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
            </div>

            <div class="modal-body">
                <input type="hidden" id="currentPostId">

                <div class="playlist-list" id="playlist-list">
                    <!-- Items se insertan aquí -->
                </div>

                <p class="text-center text-muted mt-3" id="no-playlists-msg" style="display: none;">
                    No tienes playlists creadas
                </p>
            </div>

            <div class="modal-footer">
                <button type="button" id="createNewPlaylistBtn" class="btn btn-sm btn-primary">
                    New Playlist
                </button>
                <button type="button" class="btn btn-sm btn-secondary" data-bs-dismiss="modal">
                    Close
                </button>
            </div>
        </div>
    </div>
</div>