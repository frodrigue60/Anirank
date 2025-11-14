@extends ('layouts.app')

@section('title', 'Index Playlists')

@section('content')
    <div class="container">
        <div class="d-flex justify-content-between align-items-center mb-3">
            <h1>Playlists</h1>

            <a href="{{ route('playlists.create') }}" class="btn btn-primary">
                <i class="fa-solid fa-plus"></i> Create New Playlist
            </a>
        </div>
        <table class="table">
            <thead>
                <tr>
                    <th scope="col">ID</th>
                    <th scope="col">Name</th>
                    <th scope="col">Description</th>
                    <th scope="col">Items</th>
                    <th scope="col">Actions</th>
                </tr>
            </thead>
            <tbody>
                @foreach ($playlists as $item)
                    <tr>
                        <th scope="row">{{ $item->id }}</th>
                        <td>{{ $item->name }}</td>
                        <td>{{ $item->description }}</td>
                        <td>{{ $item->songs_count }}</td>
                        <td class="d-flex gap-1">
                            <a href="{{ route('playlists.edit', $item->id) }}" class="btn btn-sm btn-success"><i
                                    class="fa-solid fa-pencil"></i></a>
                            <a href="{{ route('playlists.show', $item->id) }}" class="btn btn-sm btn-primary"><i
                                    class="fa-solid fa-eye"></i></a>
                            <form action="{{ route('playlists.destroy', $item->id) }}" method="POST">
                                @csrf
                                @method('DELETE')
                                <button type="submit" class="btn btn-sm btn-danger"
                                    onclick="return confirm('Are you sure you want to delete this playlist?')"><i
                                        class="fa-solid fa-trash"></i></button>
                            </form>
                        </td>
                    </tr>
                @endforeach


            </tbody>
        </table>
    </div>

@endsection
