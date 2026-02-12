@extends('layouts.app')

@section('meta')
    <title>{{ $artist->name }} - Themes</title>
    <meta title="{{ $artist->name }} - Themes">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta name="description" content="Explore themes by {{ $artist->name }}.">
    <meta name="robots" content="index, follow, max-image-preview:standard">
@endsection

@section('content')
    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10">
        <div class="flex flex-col gap-8">
            {{-- Header --}}
            <div class="flex items-center gap-6">
                @php
                    $thumbnailUrl = null;
                    if ($artist->thumbnail != null && Storage::disk('public')->exists($artist->thumbnail)) {
                        $thumbnailUrl = Storage::url($artist->thumbnail);
                    } elseif ($artist->thumbnail_src != null) {
                        $thumbnailUrl = $artist->thumbnail_src;
                    }
                @endphp

                @if ($thumbnailUrl)
                    <div class="w-24 h-24 rounded-full overflow-hidden ring-4 ring-primary/20 shadow-xl shrink-0">
                        <img src="{{ $thumbnailUrl }}" alt="{{ $artist->name }}" class="w-full h-full object-cover">
                    </div>
                @else
                    <div
                        class="w-24 h-24 rounded-full bg-surface-darker flex items-center justify-center text-white/10 shrink-0">
                        <span class="material-symbols-outlined text-4xl">person</span>
                    </div>
                @endif

                <div>
                    <h1 class="text-3xl md:text-4xl font-black text-white mb-2">{{ $artist->name }}</h1>
                    <div class="h-1 w-20 bg-primary rounded-full"></div>
                </div>
            </div>

            {{-- Filter Panel --}}
            <section class="bg-surface-dark/30 p-6 rounded-2xl border border-white/5 shadow-2xl backdrop-blur-md">
                @include('components.filter.container', [
                    'apiEndpoint' => '',
                    'method' => 'GET',
                    'fields' => ['name', 'type', 'year', 'season', 'sort', 'artist-id'],
                ])
            </section>

            {{-- Data Container --}}
            <section class="min-h-[400px]">
                <div class="results grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6" id="data">
                    {{-- AJAX Results --}}
                </div>

                {{-- Loader --}}
                <div class="flex justify-center py-20" id="loader">
                    <div class="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                </div>
            </section>
        </div>
    </div>
@endsection

@section('script')
    @vite(['resources/js/filter_artist_themes.js'])
@endsection
