{{-- @foreach ($playlists as $playlist)
    @php
        $buttonClass = '';
        $buttonText = '';

        if ($playlist->songs->contains($song->id)) {
            $buttonClass = 'btn-success';
            $buttonText = 'Added';
        } else {
            $buttonClass = 'btn-primary';
            $buttonText = 'Add';
        }

    @endphp
    <div class="playlist-item" data-playlist-id="{{ $playlist->id }}">
        <div class="playlist-info">
            <strong>{{ $playlist->name }}</strong>
            <small id="counter-{{ $playlist->id }}">{{ $playlist->songs->count() ?? 0 }} posts</small>
        </div>
        <button class="btn btn-sm {{ $buttonClass }} add-to-playlist-btn" data-playlist-id="{{ $playlist->id }}">
            {{ $buttonText }}
        </button>
    </div>
@endforeach

@if (auth()->user()->playlists->count() === 0)
    <p class="no-playlists">No tienes playlists creadas</p>
@endif --}}

@foreach ($playlists as $playlist)
    @php
        $isAdded = in_array($playlist->id, $playlistsWithSong);
        $buttonClass = $isAdded ? 'btn-success' : 'btn-primary';
        $buttonText = $isAdded ? 'Added' : 'Add';
    @endphp

    <div class="playlist-item" data-playlist-id="{{ $playlist->id }}">
        <div class="playlist-info">
            <strong>{{ $playlist->name }}</strong>
            <small id="counter-{{ $playlist->id }}">
                {{ $playlist->songs_count }}
            </small> <span>songs</span>
        </div>

        <button class="btn btn-sm {{ $buttonClass }} add-to-playlist-btn" data-playlist-id="{{ $playlist->id }}"
            data-song-id="{{ $songId }}">
            {{ $buttonText }}
        </button>
    </div>
@endforeach

@if ($playlists->count() === 0)
    <p class="no-playlists">No playlists</p>
@endif
