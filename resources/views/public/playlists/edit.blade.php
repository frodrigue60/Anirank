@extends ('layouts.app')

@section('title', 'Edit Playlist')

@section('content')

<div class="container">
        <form action="{{ route('playlists.update', $playlist->id) }}" method="post">
            @csrf
            @method('PUT')
            <div class="mb-3">
                <label for="name" class="form-label">Name</label>
                <input type="text" class="form-control" id="name" name="name" required value="{{ $playlist->name }}">
            </div>
            <div class="mb-3">
                <label for="description" class="form-label">Description</label>
                <textarea class="form-control" id="description" name="description" rows="3" maxlength="255" >{{$playlist->description}}</textarea>
            </div>
            <button type="submit" class="btn btn-primary">Update Playlist</button>
        </form>
    </div>

@endsection