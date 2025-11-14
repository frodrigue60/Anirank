@extends ('layouts.app')

@section('title', 'Playlists {{ $playlist->name }}')

@section('content')
    <div class="container">
        <h1>{{ $playlist->name }}</h1>
        <p>{{ $playlist->description != null ? $playlist->description : 'description' }}</p>

        @foreach ($playlist->songs as $item)
            <a href="{{ $item->url }}">{{ $item->name }}</a>
        @endforeach
    </div>
@endsection
