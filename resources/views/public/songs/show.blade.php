@extends('layouts.app')

{{-- @php

    $title = $song_variant->song->post->title;
    $suffix =
        $song_variant->song->slug != null
            ? $song_variant->song->slug
            : $song_variant->song->type . ' v' . $song_variant->version_number;
    $artist_names = [];
    $artists_string = null;

    if (isset($song_variant->song->artists) && $song_variant->song->artists->count() != 0) {
        foreach ($song_variant->song->artists as $artist) {
            $artist_names[] = $artist->name;
            $artists_string = implode(', ', $artist_names);
        }
    } else {
        $artists_string = 'N/A';
    }

    $currentUrl = url()->current();
    $thumbnailUrl = asset('/storage/thumbnails/' . $song_variant->song->post->thumbnail);
@endphp --}}

{{-- @section('meta')
    <title>{{ $title }} {{ $suffix }}</title>
    <meta name="title" content="{{ $title }} {{ $suffix }}">
    <meta name="description" content="Song: {{ $song->name }}  - Artists: {{ $artists_string }}">
    <meta name="robots" content="index, follow, max-image-preview:standard">
    <link rel="canonical" href="{{ $currentUrl }}">
    <meta property="article:section" content="{{ $song_variant->song->type == 'OP' ? 'Opening' : 'Ending' }}">

    <meta property="og:type" content="article">
    <meta property="og:title" content="{{ $title }} {{ $suffix }}">
    <meta name="og:description" content="Song: {{ $song->name }} - Artist: {{ $artists_string }}">
    <meta property="og:url" content="{{ $currentUrl }}">
    <meta property="og:image" content="{{ $thumbnailUrl }}" alt="{{ $title . ' thumbnail' }}">
    <meta property="og:image:secure_url" content="{{ $thumbnailUrl }}" alt="{{ $title . ' thumbnail' }}">
    <meta property="og:image:width" content="460">
    <meta property="og:image:height" content="650">
    <meta property="og:image:alt" content="{{ $title . ' thumbnail' }}">
    <meta property="og:image:type" content="image/webp">
@endsection --}}

@section('meta')
    <meta name="song-id" content="{{ $song->id }}">
    @auth
        <meta name="score-format" content="{{ Auth::User()->score_format }}">
    @endauth
@endsection

@section('content')
    <div class="container">
        <!-- OPTIONS BUTTONS -->
        <div class="d-flex mb-2 gap-3 {{ $song->songVariants->count() > 1 ? '' : 'd-none' }}">
            @foreach ($song->songVariants as $variant)
                <button class="btn btn-sm btn-primary btnVersion" data-variant-id="{{ $variant->id }}">Version
                    {{ $variant->version_number }}</button>
            @endforeach
        </div>
        <!-- VIDEO CONTAINER -->
        <div class="mb-2 w-100" {{-- style="border: solid 1px red;" --}}>
            <div class="w-100" id="video_container">
                <video id="player"class="ratio ratio-16x9" controls autoplay>
                    <source id="video-source" src="" type="video/webm" />
                </video>
            </div>
            {{-- <div id="video_container" class="ratio ratio-16x9">
                <video id="player" controls>
                    <source id="video-source" src="" type="video/mp4">
                    Tu navegador no soporta video.
                </video>
            </div> --}}
        </div>

        <div class="">
            <div class="">
                <!--- TITLE -->
                <h1 class="fs-2">
                    <a href="{{ $post->url }}" class="">{{ $post->title }}
                        {{ $song->slug }}</a>
                </h1>
                {{-- <div class="row">
                    <div class="col-12 col-md-9 col-lg-10">
                        <h2>
                            <a href="{{ $post->url }}" class="text-decoration-none ">{{ $post->title }}
                                {{ $song->slug }}</a>
                        </h2>
                    </div>

                    <div class="col-12 col-md-3 col-lg-2">
                        <select class="form-select" aria-label="Default select example" id="select-variant">
                            @foreach ($song->songVariants as $variant)
                                <option value="{{ $variant->id }}">Version {{ $variant->version_number }}</option>
                            @endforeach
                        </select>
                    </div>
                </div> --}}

                <!-- SONG INFO -->
                <div class="mb-2">
                    <h5 class="">{{ $song->name }} -
                        @foreach ($song->artists as $index => $artist)
                            @php
                                if ($artist->name_jp != null) {
                                    $artistNameJp = '(' . $artist->name_jp . ')';
                                } else {
                                    $artistNameJp = '';
                                }
                            @endphp
                            <a class="text-decoration-none " href="{{ $artist->url }}">{{ $artist->name }}
                                {{ $artistNameJp }}</a>
                            @if ($index < count($song->artists) - 1)
                                ,
                            @endif
                        @endforeach
                    </h5>
                </div>
            </div>
            <!-- ACTIONS BUTTONS -->
            <div class="mb-3 d-flex justify-content-between">
                <div class="d-flex gap-3">
                    <div>
                        <button class="btn btn-primary rounded-pill">
                            <i class="fa-regular fa-eye"></i> <span> {{ $song->viewsString }}</span>
                        </button>
                    </div>
                    <!-- GUEST -->
                    @guest
                        <!-- LIKES -->
                        <div>
                            <button class="btn btn-primary rounded-pill"><i class="fa-regular fa-thumbs-up"></i>
                                <span id="like-counter">{{ $song->likesCount }}</span>
                            </button>
                        </div>
                        <!-- DISLIKES -->
                        <div>
                            <button class="btn btn-primary rounded-pill"><i class="fa-regular fa-thumbs-down"></i>
                                <span id="dislike-counter">{{ $song->dislikesCount }}</span>
                            </button>
                        </div>
                    @endguest
                    <!-- AUTH -->
                    @auth
                        <!-- LIKES -->
                        <div>
                            <button id="like-button" data-song-id="{{ $song->id }}" class="btn btn-primary rounded-pill"><i
                                    class="fa-regular fa-thumbs-up"></i>
                                <span id="like-counter">{{ $song->likes()->count() }}</span>
                            </button>
                        </div>
                        <!-- DISLIKES -->
                        <div>
                            <button id="dislike-button" data-song-id="{{ $song->id }}"
                                class="btn btn-primary rounded-pill">
                                <i class="fa-regular fa-thumbs-down"></i> <span
                                    id="dislike-counter">{{ $song->dislikes()->count() }}</span>
                            </button>
                        </div>
                    @endauth

                    <!-- SCORE -->
                    <div>
                        @php
                            if (Auth::check() && isset($song->rawUserScore)) {
                                $class = $song->rawUserScore != null ? 'btn-warning' : 'btn-primary';
                            } else {
                                $class = 'btn-primary';
                            }

                        @endphp
                        <button class="btn {{ $class }} rounded-pill d-flex flex-row gap-2 align-items-center"
                            data-bs-toggle="modal" data-bs-target="#modal-rating" id="rating-button">
                            <i class="fa fa-star" aria-hidden="true"></i>
                            <span id="score-span" class="">{{ $song->formattedScore }}</span>
                        </button>
                    </div>
                </div>
                <div class="d-flex gap-2">
                    {{-- En tu vista de posts --}}
                    @auth
                        <button class="btn btn-primary rounded-pill" id="add-to-playlist" data-song-id="{{ $song->id }}">
                            <i class="fa-solid fa-list"></i> <span>Playlist</span>
                        </button>
                    @endauth
                    <!-- FAVORITE BUTTON -->
                    <button type="submit"
                        class="btn {{ $song->isFavorited() ? 'btn-danger' : 'btn-primary' }} rounded-pill d-flex flex-row gap-2 align-items-center"
                        id="favorite-button" data-songid="{{ $song->id }}"><i
                            class="fa-{{ $song->isFavorited() ? 'solid' : 'regular' }} fa-heart" id="i-favorite"></i>
                        <span class="d-sm-block d-none ">
                            Favorite</span></button>

                    <!-- REPORT BUTTON -->
                    <button type="button" class="btn btn-primary rounded-pill d-flex flex-row gap-2 align-items-center"
                        data-bs-toggle="modal" data-bs-target="#modal-report" id="btn-modal-report">
                        <i class="fa-solid fa-triangle-exclamation"></i> <span class="d-sm-block d-none ">Report</span>
                    </button>
                </div>
            </div>
        </div>

        {{-- Add to Playlist Modal --}}
        @auth
            @include('components.add-to-playlist')
            @include('components.create-playlist-modal')
        @endauth

        {{-- Report Modal --}}
        @include('partials.songs.show.report-modal')

        {{-- Rate Modal --}}
        @include('partials.songs.show.rating.modal')

        {{-- Make Comment Section --}}

        <div class="mb-3">
            <h3 class="">Comments</h3>

            @include('partials.songs.show.comment-form')
        </div>

        {{-- All Coments Section --}}

        <div class="">
            <h4 class="">Recents comments</h4>
            <div class="">
                <div id="comments-container" data-song-id="{{ $song->id }}">
                    {{-- @include('partials.songs.show.comments.comments') --}}
                </div>

                <div class="d-flex justify-content-center" id="loader-comments">
                    <div class="spinner-border m-5" role="status">
                        <span class="visually-hidden">Loading...</span>
                    </div>
                </div>

                <div class="d-flex justify-content-center">
                    <button type="button" class="btn btn-primary" id="load-more-comments">
                        Load More Comments
                    </button>
                </div>
            </div>

        </div>
    </div>
@endsection

@section('script')
    <link rel="stylesheet" href="https://cdn.plyr.io/3.7.8/plyr.css" />
    <script src="https://cdn.plyr.io/3.7.8/plyr.js"></script>

    @vite(['resources/js/api_get_video.js', 'resources/js/modules/songs/get_comments.js'])

    @auth
        @vite(['resources/js/playlists.js', 'resources/js/modules/songs/delete_comment.js', 'resources/js/modules/songs/make_comment.js', 'resources/js/modules/songs/like.js', 'resources/js/modules/songs/dislike.js', 'resources/js/modules/songs/toggle_favorite.js', 'resources/js/modules/songs/rate.js', 'resources/js/modules/songs/report.js', 'resources/js/modules/comments/like.js', 'resources/js/modules/comments/dislike.js', 'resources/js/modules/comments/reply.js'])
    @endauth
@endsection
