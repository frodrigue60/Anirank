@extends('layouts.app')
@section('meta')
    <meta name="robots" content="index, follow">
    <link rel="canonical" href="{{ url()->current() }}">
@endsection
@section('content')
    <div class="max-w-[1440px] mx-auto px-6 md:px-14 py-8 md:py-12">
        {{-- Hero Section --}}
        <div class="flex flex-col md:flex-row md:items-end justify-between gap-8 mb-12">
            <div>
                <div class="flex items-center gap-2 mb-3">
                    <span
                        class="bg-primary/20 text-primary text-[10px] font-black px-2.5 py-1 rounded-full border border-primary/30 uppercase tracking-[0.2em]">Live
                        Leaderboard</span>
                </div>
                <h1 id="section-header" class="text-4xl md:text-5xl lg:text-6xl font-black tracking-tighter text-white">
                    Global Ranking
                </h1>
                <p class="text-white/50 mt-4 text-lg font-medium max-w-2xl">
                    The definitive ranking of anime music, voted by the community and updated in real-time.
                </p>
            </div>

            {{-- Global/Seasonal Switch --}}
            <div
                class="relative flex items-center p-1.5 bg-surface-darker/50 rounded-2xl border border-white/5 w-fit select-none">
                <div id="ranking-type-bg"
                    class="absolute h-[calc(100%-12px)] w-[calc(50%-6px)] bg-primary rounded-xl shadow-lg shadow-primary/20 transition-all duration-300 ease-out left-1.5">
                </div>

                <button id="toggle-global"
                    class="relative z-10 flex items-center gap-2 px-8 py-3 rounded-xl text-white font-bold transition-all duration-300">
                    <span>Global</span>
                </button>

                <button id="toggle-seasonal"
                    class="relative z-10 flex items-center gap-2 px-8 py-3 rounded-xl text-white/40 font-bold transition-all duration-300">
                    <span>Seasonal</span>
                </button>

                {{-- Legacy trigger for compatibility --}}
                <button type="button" class="hidden" id="toggle-type-btn" disabled></button>
                <span id="toggle-type-btn-span" class="hidden">Seasonal</span>
            </div>
        </div>

        {{-- Leaderboard Section --}}
        <div class="bg-surface-dark/30 border border-white/5 rounded-3xl overflow-hidden mb-12 group/table">
            {{-- Table Controls --}}
            <div
                class="flex flex-col sm:flex-row justify-between items-center gap-4 px-8 py-6 border-b border-white/5 bg-surface-darker/30">
                <div class="flex items-center gap-3">
                    <span class="material-symbols-outlined text-primary">analytics</span>
                    <h2 class="text-xl font-bold text-white tracking-tight">Theme Leaderboard</h2>
                </div>

                <div class="flex items-center p-1 bg-background-dark/50 border border-white/5 rounded-xl w-fit">
                    <button id="show-ops"
                        class="px-6 py-2 rounded-lg bg-primary text-white font-bold text-sm shadow-lg transition-all">Openings</button>
                    <button id="show-eds"
                        class="px-6 py-2 rounded-lg text-white/40 hover:text-white font-bold text-sm transition-all">Endings</button>
                </div>
            </div>

            {{-- Table Header --}}
            <div
                class="grid grid-cols-[60px_1fr_120px_140px] gap-4 px-8 py-4 border-b border-white/5 text-[10px] font-black uppercase tracking-widest text-white/30 bg-surface-darker/50">
                <div class="text-center">Rank</div>
                <div>Theme Info</div>
                <div class="text-center">Score</div>
                <div class="text-right">Actions</div>
            </div>

            {{-- Table Body (Openings) --}}
            <div id="container-ops" class="flex flex-col">
                <!-- Data injected by JS -->
            </div>

            {{-- Table Body (Endings) --}}
            <div id="container-eds" class="flex flex-col hidden">
                <!-- Data injected by JS -->
            </div>

            {{-- Load More Section --}}
            <div class="p-8 border-t border-white/5 bg-surface-darker/30 flex flex-col items-center gap-4">
                {{-- Loaders (hidden by default) --}}
                <div id="loader-ops" class="hidden">
                    <div class="w-8 h-8 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                </div>
                <div id="loader-eds" class="hidden">
                    <div class="w-8 h-8 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                </div>

                <button id="load-more-op"
                    class="load-more-op flex items-center gap-2 text-sm font-bold text-primary hover:text-white transition-all group/btn">
                    <span>Show more openings</span>
                    <span
                        class="material-symbols-outlined transition-transform group-hover/btn:translate-y-0.5">expand_more</span>
                </button>

                <button id="load-more-ed"
                    class="load-more-ed hidden flex items-center gap-2 text-sm font-bold text-primary hover:text-white transition-all group/btn">
                    <span>Show more endings</span>
                    <span
                        class="material-symbols-outlined transition-transform group-hover/btn:translate-y-0.5">expand_more</span>
                </button>
            </div>
        </div>
    </div>
@endsection
@section('script')
    @vite(['resources/js/ranking_songs.js'])
@endsection
