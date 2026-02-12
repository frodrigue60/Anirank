@extends('layouts.app')

@section('meta')
    <title>
        {{ $post->title }} {{ $song->slug != null ? $song->slug : $song->type }}
    </title>
    <meta name="title" content="{{ $post->title }} {{ $song->slug != null ? $song->slug : $song->type }}">

    @php
        $thumbnail_url = $post->thumbnail_url ?? asset('/storage/thumbnails/' . $post->thumbnail);
        $artists_string = $song->artists->pluck('name')->join(', ');
        $desc = 'Song: ' . $song->getNameAttribute() . ' - Artist: ' . $artists_string;
    @endphp

    <meta name="description" content="{{ $desc }}">
    <meta name="robots" content="index, follow, max-image-preview:standard">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta property="og:type" content="article">
    <meta property="og:title" content="{{ $post->title }} {{ $song->slug != null ? $song->slug : $song->type }}">
    <meta name="og:description" content="{{ $desc }}">
    <meta property="og:url" content="{{ url()->current() }}">
    <meta property="og:image" content="{{ $thumbnail_url }}" alt="{{ $post->title }}">
    <meta property="og:image:secure_url" content="{{ $thumbnail_url }}" alt="{{ $post->title }}">

    {{-- Plyr CSS --}}
    <link rel="stylesheet" href="https://cdn.plyr.io/3.7.8/plyr.css" />
@endsection

@section('content')
    <livewire:song-detail :song="$song" :post="$post" />
@endsection

@section('script')
    <script src="https://cdn.plyr.io/3.7.8/plyr.js"></script>
    {{-- Any other global scripts if needed --}}
@endsection
