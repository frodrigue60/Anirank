@extends('layouts.app')

@section('title', 'Ranking Anime Openings & Endings | ' . config('app.name'))

@section('meta')
    <meta name="title" content="Search, play, and rate anime openings and endings">
    <meta name="description"
        content="The site you were looking for to rate openings and endings of your favorite animes. Discover which are the most popular opening and endings.">
    <meta name="keywords"
        content="top anime openings, top anime endings, ranking openings anime, ranking endings anime, Best Anime Openings Of All Time, openings anime, endings anime">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta name="robots" content="index, follow, max-image-preview:standard">
    <meta property="og:type" content="website" />
    <meta property="og:image" content="{{ asset('resources/images/og-image-wide.png') }}">
    <meta property="og:image:secure_url" content="{{ asset('resources/images/og-image-wide.png') }}">
    <meta property="og:image:type" content="image/png">
    <meta property="og:image:width" content="828">
    <meta property="og:image:height" content="450">
    <meta property="og:url" content="{{ url()->current() }}" />
    <meta property="og:image:alt" content="Anirank banner" />
    <meta name="twitter:card" content="summary" />
    <meta name="twitter:site" content="@frodrigue60" />
    <meta name="twitter:creator" content="@frodrigue60" />
    <meta property="og:title" content="Search, play, and rate anime openings and endings" />
    <meta property="og:description"
        content="The site you were looking for to rate openings and endings of your favorite animes. Discover which are the most popular opening and endings." />
@endsection

@push('styles')
    <link rel="stylesheet" href="{{ asset('resources/owlcarousel/assets/owl.carousel.min.css') }}">
    <link rel="stylesheet" href="{{ asset('resources/owlcarousel/assets/owl.theme.default.min.css') }}">
@endpush

@section('content')
    {{-- @guest
        <section class="relative h-screen flex items-center justify-center overflow-hidden">
            <div class="absolute inset-0 z-0">
                <div class="absolute inset-0 bg-gradient-to-t from-background-dark via-background-dark/60 to-transparent z-10">
                </div>
                <div
                    class="absolute inset-0 bg-gradient-to-r from-background-dark/80 via-transparent to-background-dark/80 z-10">
                </div>
                <img class="w-full h-full object-cover scale-105" data-alt="Cinematic wide shot of anime cityscape at night"
                    src="https://lh3.googleusercontent.com/aida-public/AB6AXuC4NKMsiVFSxKoMy0QJq1bdQInEoybDwBon5RstY0kyN604vExTI9pu8NcCmoEH1v6KITdZ4av_LtDghv2w-_XaKXQElk4leiUH9GVq1u0IyulYjmeCIqspuNtkNr8PCg2bZhyubZmFlVu3i7A2-Ug2FVNBcK4uF7HEtkwYOOHS-swOGUyTo14YgLkUPJsdetP6SpSqK0fw-RR74Zhd-zOxh-7r1ruGpFQDSftrvuYB_oXOcfl7wzCHPUY-joO8Cn9Nr-6SuEVYS6A" />
            </div>
            <div class="relative z-20 container mx-auto px-6 text-center max-w-4xl">
                <h1
                    class="text-5xl md:text-7xl font-black mb-6 tracking-tight leading-none text-transparent bg-clip-text bg-gradient-to-b from-white to-white/60">
                    Your Anime Soundtrack Journey Starts Here
                </h1>
                <div class="flex flex-wrap justify-center gap-8 mb-10 text-lg">
                    <div class="flex items-center gap-2 text-white/90">
                        <span class="material-symbols-outlined text-primary">analytics</span>
                        <span>Track favorites</span>
                    </div>
                    <div class="flex items-center gap-2 text-white/90">
                        <span class="material-symbols-outlined text-primary">stars</span>
                        <span>Influence rankings</span>
                    </div>
                    <div class="flex items-center gap-2 text-white/90">
                        <span class="material-symbols-outlined text-primary">playlist_add_check</span>
                        <span>Curate playlists</span>
                    </div>
                </div>
                <a href="{{ route('songs.index') }}"
                    class="group relative inline-flex items-center gap-3 bg-primary px-10 py-5 rounded-xl font-bold text-xl transition-all hover:scale-105 active:scale-95 neon-glow">
                    <span>Start Exploring</span>
                    <span class="material-symbols-outlined group-hover:translate-x-1 transition-transform">arrow_forward</span>
                </a>
            </div>
            <div class="absolute bottom-10 left-1/2 -translate-x-1/2 z-20 animate-bounce opacity-50">
                <span class="material-symbols-outlined text-3xl">keyboard_double_arrow_down</span>
            </div>
        </section>
    @endguest --}}
    <main class="flex-1 w-full max-w-[1440px] mx-auto grid grid-cols-1 lg:grid-cols-12 gap-8 px-6 py-8">
        <div class="lg:col-span-9 flex flex-col gap-10">
            {{-- FEATURED THEME --}}
            @isset($featuredSong)
                <section class="relative w-full rounded-2xl overflow-hidden bg-surface-dark group">
                    <div class="absolute inset-0 bg-cover bg-center opacity-60 mix-blend-overlay transition-transform duration-700 group-hover:scale-105"
                        style="background-image: url('{{ Storage::url($featuredSong->post->banner) }}');">
                    </div>
                    <div class="absolute inset-0 bg-gradient-to-r from-background-dark via-background-dark/80 to-transparent">
                    </div>
                    <div class="relative z-10 p-8 md:p-12 flex flex-col md:flex-row gap-8 items-center md:items-end">
                        <div class="relative shrink-0 hero-glow">
                            <div
                                class="w-48 h-48 md:w-64 md:h-64 rounded-xl shadow-2xl overflow-hidden relative border-2 border-white/10">
                                <img src="{{ Storage::url($featuredSong->post->thumbnail) }}" alt="Anime cover art"
                                    class="w-full h-full object-cover">
                                <div
                                    class="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/80 to-transparent p-4 pt-10">
                                    <div class="flex items-center gap-1 text-yellow-400 font-bold text-lg">
                                        <span class="material-symbols-outlined filled text-[20px]">star</span>
                                        <span>{{ number_format($featuredSong->averageRating ?? 0, 1) }}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="flex flex-col gap-4 text-center md:text-left flex-1">
                            <div
                                class="inline-flex items-center gap-2 self-center md:self-start bg-white/5 border border-white/10 px-3 py-1 rounded-full backdrop-blur-md">
                                <span class="material-symbols-outlined text-primary text-sm filled">auto_awesome</span>
                                <span class="text-xs font-bold uppercase tracking-wider text-white/90">Featured Theme</span>
                            </div>
                            <div>
                                <h1
                                    class="text-xl md:text-2xl font-black leading-tight tracking-tight text-white drop-shadow-xl mb-2 truncate">
                                    {{ $featuredSong->name }}</h1>
                                <div
                                    class="flex flex-col md:flex-col items-center md:items-start gap-1 md:gap-3 text-md md:text-lg text-white/80">
                                    <span class="font-bold text-primary truncate">{{ $featuredSong->post->title }}</span>
                                    <span class="font-medium text-white truncate">
                                        @foreach ($featuredSong->artists as $artist)
                                            {{ $artist->name }}{{ !$loop->last ? ', ' : '' }}
                                        @endforeach
                                    </span>
                                </div>
                            </div>
                            <div class="flex items-center gap-4 mt-2 justify-center md:justify-start">
                                <a href="{{ route('songs.show', $featuredSong->id) }}"
                                    class="bg-primary hover:bg-primary/80 text-white h-12 px-8 rounded-full font-bold flex items-center gap-2 transition-all hover:scale-105 active:scale-95 shadow-lg shadow-primary/30">
                                    <span class="material-symbols-outlined filled">play_circle</span>
                                    <span>Play Now</span>
                                </a>
                            </div>
                        </div>
                    </div>
                </section>
            @endisset

            {{-- WEEKLY RANKINGS --}}
            <section>
                <div class="flex flex-col sm:flex-row sm:items-center justify-between mb-6 px-1 gap-4">
                    <h2 class="text-2xl font-bold tracking-tight flex items-center gap-2 text-white">
                        <span class="material-symbols-outlined text-primary">leaderboard</span>
                        Weekly Rankings
                    </h2>
                    {{-- <div class="bg-surface-darker p-1 rounded-lg flex items-center border border-white/5 self-start sm:self-auto">
                        <button class="px-4 py-1.5 rounded-md bg-primary text-white text-sm font-bold shadow-sm transition-all">Openings (OP)</button>
                        <button class="px-4 py-1.5 rounded-md text-white/60 hover:text-white text-sm font-medium transition-all">Endings (ED)</button>
                    </div> --}}
                </div>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    @foreach ($weaklyRanking->chunk(ceil($weaklyRanking->count() / 2)) as $chunk)
                        <div class="flex flex-col gap-4">
                            <div class="flex items-center justify-between mb-2">
                                <span class="text-white/50 text-xs font-bold uppercase tracking-wider">Top
                                    {{ $loop->first ? 'Openings' : 'Endings' }}</span>
                            </div>
                            @foreach ($chunk as $song)
                                <div
                                    class="group relative bg-surface-darker p-4 rounded-xl hover:bg-surface-dark transition-colors border border-white/5 flex gap-4 items-center">
                                    <div class="relative w-20 h-20 shrink-0 rounded-lg overflow-hidden">
                                        <img src="{{ Storage::url($song->post->thumbnail) }}" alt="{{ $song->name }}"
                                            class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500">
                                        <div
                                            class="absolute top-1 left-1 {{ $loop->parent->first && $loop->first ? 'bg-primary' : 'bg-surface-dark' }} text-white text-xs font-bold px-1.5 py-0.5 rounded shadow border border-white/10">
                                            #{{ $loop->iteration }}</div>
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <div class="flex items-center justify-between">
                                            <a href="{{ route('songs.show', $song->id) }}"
                                                class="font-bold text-white truncate text-lg hover:text-primary transition-colors">{{ $song->name }}</a>
                                            <div
                                                class="flex items-center gap-1 bg-surface-dark px-2 py-0.5 rounded text-yellow-400 text-xs font-bold">
                                                <span class="material-symbols-outlined filled text-[14px]">star</span>
                                                {{ number_format($song->averageRating ?? 0, 1) }}
                                            </div>
                                        </div>
                                        <p class="text-sm text-primary font-medium truncate">{{ $song->post->title }}</p>
                                        <p class="text-xs text-white/50 truncate">
                                            @foreach ($song->artists as $artist)
                                                {{ $artist->name }}{{ !$loop->last ? ', ' : '' }}
                                            @endforeach
                                        </p>
                                    </div>
                                </div>
                            @endforeach
                        </div>
                    @endforeach
                </div>
            </section>

            {{-- TABS SECTION --}}
            <section class="pb-12">
                <div class="border-b border-white/5 mb-6">
                    <div class="flex gap-8" id="tabs">
                        <button id="recently-tab"
                            class="pb-4 border-b-2 border-primary text-white font-bold text-sm tracking-wide transition-all active-tab-link">Recently
                            Added</button>
                        <button id="popular-tab"
                            class="pb-4 border-b-2 border-transparent text-white/40 hover:text-white font-medium text-sm tracking-wide transition-colors">Most
                            Popular</button>
                        <button id="viewed-tab"
                            class="pb-4 border-b-2 border-transparent text-white/40 hover:text-white font-medium text-sm tracking-wide transition-colors">Most
                            Viewed</button>
                    </div>
                </div>
                <div id="tab-content">
                    <div id="recently" class="tab-pane">
                        <div class="owl-carousel">
                            @include('partials.songs.cards-v2', ['songs' => $recently])
                        </div>
                    </div>
                    <div id="popular" class="tab-pane hidden">
                        <div class="owl-carousel">
                            @include('partials.songs.cards-v2', ['songs' => $popular])
                        </div>
                    </div>
                    <div id="viewed" class="tab-pane hidden">
                        <div class="owl-carousel">
                            @include('partials.songs.cards-v2', ['songs' => $viewed])
                        </div>
                    </div>
                </div>
            </section>
        </div>

        {{-- SIDEBAR --}}
        <aside class="lg:col-span-3 flex flex-col gap-8">
            <div class="bg-surface-darker rounded-2xl p-6 border border-white/5">
                <div class="flex items-center justify-between mb-6">
                    <h3 class="font-bold text-white text-lg">Featured Artists</h3>
                    <a href="{{ route('artists.index') }}"
                        class="text-primary text-xs font-bold uppercase tracking-wide hover:underline">View All</a>
                </div>
                <div class="flex flex-col gap-5">
                    @foreach ($artists->take(5) as $artist)
                        <div class="flex items-center justify-between group">
                            <a href="{{ route('artists.show', $artist->slug) }}" class="flex items-center gap-3 min-w-0">
                                <div
                                    class="w-12 h-12 rounded-full overflow-hidden border-2 border-transparent group-hover:border-primary transition-colors flex-shrink-0">
                                    @if ($artist->image)
                                        <img src="{{ Storage::url($artist->image) }}" alt="{{ $artist->name }}"
                                            class="w-full h-full object-cover">
                                    @else
                                        <div
                                            class="w-full h-full bg-primary/20 flex items-center justify-center text-primary">
                                            <span class="material-symbols-outlined">person</span>
                                        </div>
                                    @endif
                                </div>
                                <div class="min-w-0">
                                    <h4
                                        class="text-sm font-bold text-white group-hover:text-primary transition-colors truncate">
                                        {{ $artist->name }}</h4>
                                    <p class="text-xs text-white/40 truncate">{{ $artist->songs_count ?? '0' }} Themes</p>
                                </div>
                            </a>
                            <a href="{{ route('artists.show', $artist->slug) }}"
                                class="w-8 h-8 rounded-full bg-white/5 hover:bg-white/10 flex items-center justify-center text-primary transition-colors">
                                <span class="material-symbols-outlined text-[18px]">arrow_forward</span>
                            </a>
                        </div>
                    @endforeach
                </div>
            </div>

            {{-- EVENT/COMMUNITY BOX --}}
            <div class="relative rounded-2xl overflow-hidden aspect-[4/5] flex flex-col justify-end p-6 group">
                <div class="absolute inset-0 bg-cover bg-center transition-transform duration-700 group-hover:scale-110"
                    style="background-image: url('https://lh3.googleusercontent.com/aida-public/AB6AXuAH_og0-9n8HCwoymOtvxCZ2iSwhi1s0A9GouQSn-K7NzkZ13nzTIqMOJ6cq1utUw0s-FL7CewN6C8VFcJizGXn7mHZASt9HtY3-Lhm7ktTQLi7ouj5QP8adxWImQ_5RxcUGZRHyujm7HGEcipxIes_YOw3FD7C2XnhIYlfxnYJ8zeWxbouotxhLpbP4NwK2OAuftUr0xAL6mD6jN8392YwYPYusIbgoJXxmVZ38rnjbBuYc5uaDeXAuuWdDfBjH2c72v2y7Btneg8');">
                </div>
                <div class="absolute inset-0 bg-gradient-to-t from-black via-black/50 to-transparent"></div>
                <div class="relative z-10 text-white">
                    <span
                        class="bg-[#5865F2] text-white text-[10px] font-bold uppercase px-2 py-1 rounded mb-2 inline-block">Discord</span>
                    <h3 class="text-2xl font-bold mb-2 leading-tight">Join Our Community</h3>
                    <p class="text-white/70 text-sm mb-4">Discuss your favorite themes with other fans!</p>
                    <a href="#"
                        class="w-full bg-[#5865F2] text-white font-bold py-3 rounded-lg hover:bg-[#4752C4] transition-colors inline-block text-center">Join
                        Now</a>
                </div>
            </div>

            {{-- ADS placeholder --}}
            <div
                class="p-8 rounded-2xl text-center border border-dashed border-white/10 bg-surface-darker/50 flex flex-col items-center justify-center min-h-[250px] group">
                <span class="material-symbols-outlined text-[32px] text-white/10 mb-2">ad_units</span>
                <div class="text-white/20 text-[10px] font-black tracking-[0.2em] uppercase">Advertisement</div>
            </div>
        </aside>
    </main>
    </div>
@endsection

@push('scripts')
    @if (config('app.env') === 'local')
        <script src="{{ asset('resources/js/jquery-3.6.3.slim.min.js') }}"></script>
    @else
        <script src="https://code.jquery.com/jquery-3.6.3.slim.min.js"
            integrity="sha256-ZwqZIVdD3iXNyGHbSYdsmWP//UBokj2FHAxKuSBKDSo=" crossorigin="anonymous"></script>
    @endif

    <script src="{{ asset('resources/owlcarousel/owl.carousel.min.js') }}"></script>
    <script src="{{ asset('resources/js/owCarouselConfig.js') }}"></script>

    <script>
        $(function() {
            const $tabs = $('#tabs button');
            const $panes = $('#tab-content .tab-pane');

            function setActiveTab($clickedTab) {
                // Update button styles
                $tabs.removeClass('border-primary text-white font-bold active-tab-link')
                    .addClass('border-transparent text-white/40 font-medium');

                $clickedTab.addClass('border-primary text-white font-bold active-tab-link')
                    .removeClass('border-transparent text-white/40 font-medium');

                // Show/Hide panes
                const targetId = $clickedTab.attr('id').replace('-tab', '');
                $panes.addClass('hidden');
                $('#' + targetId).removeClass('hidden');

                // Refresh and Reset Owl Carousel in the active pane
                setTimeout(() => {
                    window.dispatchEvent(new Event('resize'));
                    const $activeCarousel = $('#' + targetId + ' .owl-carousel');
                    $activeCarousel.trigger('to.owl.carousel', [0, 0]);
                    $activeCarousel.trigger('refresh.owl.carousel');
                }, 50);
            }

            $tabs.on('click', function(e) {
                e.preventDefault();
                setActiveTab($(this));
            });

            // Set initial state
            if ($tabs.length > 0) {
                setActiveTab($tabs.first());
            }
        });
    </script>
@endpush
