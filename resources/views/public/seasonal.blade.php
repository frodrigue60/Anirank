@extends('layouts.app')
@section('meta')
    <meta name="robots" content="index, follow" />
    <link rel="canonical" href="{{ url()->current() }}" />
    @if (isset($currentSeason->name) && isset($currentYear->name))
        <meta name="keywords"
            content="ranking, top, anime openings {{ $currentSeason->name }}, openings anime {{ $currentSeason->name }}, anime endings {{ $currentSeason->name }}, endings anime {{ $currentSeason->name }}, of {{ $currentSeason->name }}">
        @if (Request::is('openings'))
            <title>Openings {{ $currentSeason->name }}</title>
            <meta title="Openings {{ $currentSeason->name }}">
            <meta name="description" content="Openings of {{ $currentSeason->name }} anime season">

            <meta property="og:type" content="website" />
            {{-- <meta property="og:image" content="{{ asset('resources/images/og-image-wide.png') }}">
        <meta property="og:image:secure_url" content="{{ asset('resources/images/og-image-wide.png') }}">
        <meta property="og:image:type" content="image/png">
        <meta property="og:image:width" content="828">
        <meta property="og:image:height" content="450">
        <meta property="og:image:alt" content="Anirank banner" /> --}}
            <meta property="og:url" content="{{ url()->current() }}" />

            <meta name="twitter:card" content="summary" />
            <meta name="twitter:site" content="@frodrigue60" />
            <meta name="twitter:creator" content="@frodrigue60" />
            <meta property="og:title" content="Openings {{ $currentSeason->name }}" />
            <meta property="og:description" content="Openings of {{ $currentSeason->name }} anime season" />
        @endif
        @if (Request::is('endings'))
            <title>Endings {{ $currentSeason->name }}</title>
            <meta title="Endings {{ $currentSeason->name }}">
            <meta name="description" content="Endings of {{ $currentSeason->name }} anime season">

            <meta property="og:type" content="website" />
            {{-- <meta property="og:image" content="{{ asset('resources/images/og-image-wide.png') }}">
        <meta property="og:image:secure_url" content="{{ asset('resources/images/og-image-wide.png') }}">
        <meta property="og:image:type" content="image/png">
        <meta property="og:image:width" content="828">
        <meta property="og:image:height" content="450">
        <meta property="og:image:alt" content="Anirank banner" /> --}}
            <meta property="og:url" content="{{ url()->current() }}" />

            <meta name="twitter:card" content="summary" />
            <meta name="twitter:site" content="@frodrigue60" />
            <meta name="twitter:creator" content="@frodrigue60" />
            <meta property="og:title" content="Endings {{ $currentSeason->name }}" />
            <meta property="og:description" content="Endings of {{ $currentSeason->name }} anime season" />
        @endif
    @endif
@endsection
@section('content')
    <div class="max-w-[1440px] mx-auto px-6 md:px-14 py-8 md:py-12">
        <div class="mb-12">
            <div class="flex flex-col md:flex-row md:items-end justify-between gap-8">
                <div>
                    <div class="flex items-center gap-2 mb-3">
                        <span
                            class="bg-primary/20 text-primary text-[10px] font-black px-2.5 py-1 rounded-full border border-primary/30 uppercase tracking-[0.2em]">Active
                            Season</span>
                    </div>
                    @if (isset($currentSeason) && isset($currentYear))
                        <h1 class="text-4xl md:text-5xl lg:text-6xl font-black tracking-tighter text-white">
                            {{ $currentSeason->name }} <span class="text-primary">{{ $currentYear->name }}</span>
                        </h1>
                    @else
                        <h1 class="text-4xl md:text-5xl lg:text-6xl font-black tracking-tighter text-white/40">Season: N/A
                        </h1>
                    @endif
                    <p class="text-white/50 mt-4 text-lg font-medium max-w-2xl">Exploring the latest themes and musical
                        trends from this season's anime releases.</p>
                </div>

                {{-- Toggle Switcher (Two-sided style) --}}
                <div
                    class="relative flex items-center p-1 bg-surface-darker/50 rounded-2xl border border-white/5 w-fit select-none">
                    {{-- Sliding Background --}}
                    <div id="toggle-bg"
                        class="absolute h-[calc(100%-8px)] w-[calc(50%-4px)] bg-primary rounded-xl shadow-lg shadow-primary/20 transition-all duration-300 ease-out left-1">
                    </div>

                    <button data-type="OP" id="toggle-op"
                        class="relative z-10 flex items-center gap-2 px-8 py-3 rounded-xl text-white font-bold transition-all duration-300">
                        <span class="material-symbols-outlined text-[20px] filled">play_circle</span>
                        <span>Openings</span>
                    </button>

                    <button data-type="ED" id="toggle-ed"
                        class="relative z-10 flex items-center gap-2 px-8 py-3 rounded-xl text-white/40 font-bold transition-all duration-300">
                        <span class="material-symbols-outlined text-[20px]">skip_next</span>
                        <span>Endings</span>
                    </button>

                    {{-- Hidden input/trigger for compatibility with existing JS --}}
                    <button id="toggle-type-btn" class="hidden"></button>
                    <span id="section-header" class="hidden">OPENINGS</span>
                    <span id="btn-toggle-text" class="hidden">Endings</span>
                </div>
            </div>
        </div>

        <div class="pb-20">
            <div class="results grid grid-cols-1 gap-6" id="data">
                <!-- DATA INJECTED BY JS --->
            </div>
        </div>
    </div>
@endsection

@section('script')
    @vite(['resources/js/seasonal.js'])
@endsection
