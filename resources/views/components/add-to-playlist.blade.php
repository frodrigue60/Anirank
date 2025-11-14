{{-- resources/views/components/add-to-playlist.blade.php --}}
<div id="addToPlaylistModal" class="modal" style="display: none;">
    <div class="modal-content">
        <span class="close">&times;</span>
        <h3 class="modal-title">Add to playlist</h3>
        <input type="hidden" id="currentPostId">
        
        <div class="playlist-list"  id="playlist-list">
            {{-- @foreach(auth()->user()->playlists as $playlist)
                <div class="playlist-item" data-playlist-id="{{ $playlist->id }}">
                    <div class="playlist-info">
                        <strong>{{ $playlist->name }}</strong>
                        <small>{{ $playlist->posts_count ?? 0 }} posts</small>
                    </div>
                    <button class="add-to-playlist-btn" 
                            data-playlist-id="{{ $playlist->id }}">
                        {{ $playlist->posts->contains($post->id ?? 0) ? '✓ En playlist' : 'Añadir' }}
                    </button>
                </div>
            @endforeach --}}
            
           
        </div>
        
        <div class="modal-actions">
            <button id="createNewPlaylistBtn" class="btn btn-sm btn-primary">New Playlist</button>
            <button class="btn btn-sm btn-secondary close-modal">Close</button>
        </div>
    </div>
</div>