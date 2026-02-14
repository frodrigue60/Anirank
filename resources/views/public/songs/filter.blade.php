@extends('layouts.app')
@section('meta')
    @if (Request::routeIs('themes'))
        <title>Search Openings & Endings</title>
        <meta title="Search Openings & Endings">
        <link rel="canonical" href="{{ url()->current() }}">
        <meta name="description" content="Search Openings & Endings by type, season, order as you want, and filter by letter">
        <meta name="robots" content="index, follow, max-image-preview:standard">
    @endif
    @if (Request::routeIs('user.list'))
        <title>{{ $user->name }} - list</title>
        <meta title="{{ $user->name }} - Themes list">
        <link rel="canonical" href="{{ url()->current() }}">
        <meta name="description" content="Openings & Endings from {{ $user->name }}">
        <meta name="robots" content="index, follow, max-image-preview:standard">
    @endif
@endsection

@section('content')
    @if (Request::routeIs('user.list') || Request::routeIs('favorites'))
        @include('partials.user.banner')
    @endif
    <div class="container d-flex flex-column ">
        @if (Request::routeIs('themes'))
            <div class="top-header color1 py-1">
                <h2 class=" m-0">Search Themes</h2>
            </div>
        @endif
        @if (Request::routeIs('favorites'))
            <div class="top-header color1 py-1">
                <h2 class=" m-0">My Favorites</h2>
            </div>
        @endif
        @if (Request::routeIs('user.list') && isset($user))
            <div class="top-header color1 py-1">
                <h2 class=" m-0"><strong>{{ $user->name }}</strong> favorites</h2>
            </div>
        @endif
        @if (Request::routeIs('artists.show'))
            <div class="top-header color1 py-1">
                <h2 class=" m-0">{{ $artist->name }}</h2>
            </div>
        @endif
        {{-- SEARCH PANEL --}}

        <section class="my-2">
            @if (Request::routeIs('favorites'))
                <form class="d-flex gap-3 flex-wrap" id="form-filter" action="{{ route('profile.favorites') }}" method="get">
            @endif
            @if (Request::routeIs('themes'))
                <form class="d-flex gap-3 flex-wrap" id="form-filter" action="{{ route('themes') }}" method="get">
            @endif
            @if (Request::routeIs('user.list'))
                <form class="d-flex gap-3 flex-wrap" id="form-filter" action="{{ route('user.list', $user->id) }}"
                    method="get">
            @endif
            @if (Request::routeIs('artists.show'))
                <form class="d-flex gap-3 flex-wrap" id="form-filter"
                    action="{{ route('artists.show', [$artist->id, $artist->slug]) }}" method="get">
                    <input type="hidden" name="artist_id" id="artist_id" value="{{ $artist->id }}">
            @endif
            {{-- NAME --}}
            <div class="">
                <label for="input-name" class="">Name</label>
                <input type="text" class="form-control" name="name" id="input-name">
            </div>
            {{-- TYPE --}}
            <div class="">
                <label for="select-type" class="">Type</label>
                <select class="form-select" aria-label="Default select example" id="select-type" name="type">
                    <option value="" selected>Any</option>
                    @foreach ($types as $item)
                        <option value="{{ $item['value'] }}">
                            {{ $item['name'] }}
                        </option>
                    @endforeach
                </select>
            </div>

            {{-- YEAR --}}
            <div class="">
                <label class="" for="select-year">Year</label>
                <select class="form-select" aria-label="Default select example" name="year_id" id="select-year">
                    <option selected value="">Any</option>
                    @foreach ($years as $year)
                        <option value="{{ $year->id }}">{{ $year->name }}
                        </option>
                    @endforeach
                </select>
            </div>
            {{-- SEASON --}}
            <div class="">
                <label class="" for="select-season">Season</label>
                <select class="form-select" aria-label="Default select example" name="season_id" id="select-season">
                    <option selected value="">Season</option>
                    @foreach ($seasons as $season)
                        <option value="{{ $season->id }}">{{ $season->name }}
                        </option>
                    @endforeach
                </select>
            </div>

            {{-- SORT --}}
            <div class="">
                <label for="select-sort" class="">Sort</label>
                <select class="form-select" aria-label="Default select example" id="select-sort" name="sort">
                    <option value="">Any</option>
                    @foreach ($sortMethods as $item)
                        <option value="{{ $item['value'] }}">
                            {{ $item['name'] }}
                        </option>
                    @endforeach
                </select>
            </div>

            {{-- <div class="d-flex mt-1">
                        <button class="btn btn-primary w-100" type="submit">Do it</button>
                    </div> --}}
            </form>
        </section>

        {{-- POSTS --}}
        <section class=" mb-3">
            <div class="contenedor-tarjetas-filtro" id="data">
                {{--  @include('layouts.variant.cards') --}}
            </div>
            <div class="d-flex m-5 justify-content-center" id="loader">
                <div class="spinner-border" role="status">
                    <span class="visually-hidden">Loading...</span>
                </div>
            </div>
        </section>

    </div>
@endsection
@section('script')
    @if (Request::routeIs('favorites'))
        @if (config('app.env') === 'local')
            @vite(['resources/js/favorites_infinite_scroll.js'])
        @else
            <script src="{{ asset('build/favorites_infinite_scroll.js') }}"></script>
        @endif
    @endif
    @if (Request::routeIs('user.list'))
        @if (config('app.env') === 'local')
            @vite(['resources/js/favorites_infinite_scroll.js'])
        @else
            <script src="{{ asset('build/favorites_infinite_scroll.js') }}"></script>
        @endif
    @endif

    @if (Request::routeIs('themes'))
        @if (config('app.env') === 'local')
            @vite(['resources/js/themes_infinite_scroll.js'])
        @else
            <script src="{{ asset('build/themes_infinite_scroll.js') }}"></script>
        @endif
    @endif

    @if (Request::routeIs('artists.show'))
        @if (config('app.env') === 'local')
            @vite(['resources/js/artists_infinite_scroll.js'])
        @else
            <script src="{{ asset('build/artists_infinite_scroll.js') }}"></script>
        @endif
    @endif
@endsection
