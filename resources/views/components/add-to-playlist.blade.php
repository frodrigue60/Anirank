{{-- resources/views/components/add-to-playlist.blade.php --}}
<div id="addToPlaylistModal" class="modal" style="display: none;">
    <div class="modal-content">
        <span class="close">&times;</span>
        <h3 class="modal-title">Add to playlist</h3>
        <input type="hidden" id="currentPostId">
        
        <div class="playlist-list"  id="playlist-list">
            
        </div>
        
        <div class="modal-actions">
            <button id="createNewPlaylistBtn" class="btn btn-sm btn-primary">New Playlist</button>
            <button class="btn btn-sm btn-secondary close-modal">Close</button>
        </div>
    </div>
</div>