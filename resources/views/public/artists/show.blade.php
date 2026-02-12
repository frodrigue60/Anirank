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

            {{-- Livewire Table Component --}}
            <livewire:artist-themes-table :artist="$artist" />

        </div>
    </div>
@endsection