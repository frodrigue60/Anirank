@extends ('layouts.app')

@section('title', 'Create Playlists')

@section('content')
    <div class="container">
        <form action="{{ route('playlists.store') }}" method="post">
            @csrf
            @method('POST')
            <div class="mb-3">
                <label for="name" class="form-label">Name</label>
                <input type="text" class="form-control" id="name" name="name" required>
            </div>
            <div class="mb-3">
                <label for="description" class="form-label">Description</label>
                <textarea class="form-control" id="description" name="description" rows="3" maxlength="255"></textarea>
            </div>
            <button type="submit" class="btn btn-primary">Create Playlist</button>
        </form>
    </div>
@endsection
