@extends('layouts.app')

@section('meta')
    <link rel="stylesheet"
        href="https://cdn.jsdelivr.net/gh/habibmhamadi/multi-select-tag@2.0.1/dist/css/multi-select-tag.css">

    <style>
        .mult-select-tag ul li {
            color: black;
        }
    </style>
@endsection

@section('content')
    <div class="container">
        <div class="row justify-content-center">
            <div class="card  ">
                <div class="card-header">
                    <h5 class="card-title">Add song</h5>
                </div>
                <div class="card-body">
                    <form method="post" action="{{ route('admin.songs.store') }}" enctype="multipart/form-data">
                        @csrf
                        <input type="hidden" name="post_id" value="{{ $post->id }}">
                        <div class="mb-3">
                            <label for="theme_num" class="form-label">OP/ED Number</label>
                            <input type="number" class="form-control" placeholder="Opening Number" id="theme_num"
                                name="theme_num" value="{{ old('theme_num') }}">
                        </div>
                        <div class="mb-3">
                            <label for="songRomaji" class="form-label">Song name (romaji)</label>
                            <input type="text" class="form-control" placeholder="Song Name Romaji + (Kanji)" id="songRomaji"
                                name="song_romaji" value="{{ old('song_romaji') }}">
                        </div>
                        <div class="mb-3">
                            <label for="songJp" class="form-label">Song name (JP)</label>
                            <input type="text" class="form-control" placeholder="Song Name JP" id="songJp"
                                name="song_jp" value="{{ old('song_jp') }}">
                        </div>
                        <div class="mb-3">
                            <label for="songEn" class="form-label">Song name (EN)</label>
                            <input type="text" class="form-control" placeholder="Song Name EN" id="songEn"
                                name="song_en" value="{{ old('song_en') }}">
                        </div>
                        <div class="mb-3">
                            <label for="artists-input" class="form-label">Artists</label>
                            <input type="text" class="form-control" placeholder="Artist 1 + (Kanji), Artist 2  + (Kanji), Artist 3  + (Kanji)..." id="artists-input"
                                name="artists" value="{{ old('artists') }}" required>
                        </div>

                        {{-- <div class="row">
                            <div class="col-md mb-3">
                                <label for="artists-select" class="form-label">Artist</label>
                                <select class="form-select" multiple name="artists[]" id="artists-select">
                                    <option value="">Select a artist</option>
                                    @foreach ($artists as $artist)
                                        <option {{ old('artists') == $artist->id ? 'selected' : '' }}
                                            value="{{ $artist->id }}">{{ $artist->name }}</option>
                                    @endforeach
                                </select>
                            </div>
                            <div class="col-md mb-3">
                                <label for="type" class="form-label">Type</label>
                                <select class="form-select" name="type" id="type">
                                    @foreach ($types as $item)
                                        <option {{ old('type') == $item['value'] ? 'selected' : '' }}
                                            value="{{ $item['value'] }}">{{ $item['name'] }}</option>
                                    @endforeach
                                </select>
                            </div>
                        </div> --}}

                        <div class="mb-3">
                            <label for="type" class="form-label">Type</label>
                            <select class="form-select" name="type" id="type">
                                @foreach ($types as $item)
                                    <option {{ old('type') == $item['value'] ? 'selected' : '' }}
                                        value="{{ $item['value'] }}">{{ $item['name'] }}</option>
                                @endforeach
                            </select>
                        </div>
                        <div class="mb-3">
                            <div class="">
                                <label for="" class="form-label">Season</label>
                                <select class="form-select" name="season_id" id="">
                                    @foreach ($seasons as $season)
                                        <option value="{{ $season->id }}"
                                            {{ old('season_id', $post->season_id) == $season->id ? 'selected' : '' }}>
                                            {{ $season->name }}</option>
                                    @endforeach
                                </select>
                            </div>
                        </div>
                        <div class="mb-3">
                            <div class="">
                                <label for="" class="form-label">Year</label>
                                <select class="form-select" name="year_id" id="">
                                    @foreach ($years as $year)
                                        <option value="{{ $year->id }}"
                                            {{ old('year_id', $post->year_id) == $year->id ? 'selected' : '' }}>
                                            {{ $year->name }}</option>
                                    @endforeach
                                </select>
                            </div>
                        </div>
                        <div class="d-flex">
                            <button class="btn btn-primary w-100" type="submit">Submit</button>
                        </div>
                    </form>
                </div>
            </div>
        </div>

    </div>
@endsection

@section('script')
    {{-- <script src="https://cdn.jsdelivr.net/gh/habibmhamadi/multi-select-tag@2.0.1/dist/js/multi-select-tag.js"></script> --}}

    {{-- <script>
        new MultiSelectTag('artists-select');
    </script> --}}
@endsection
