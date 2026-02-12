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
        </div>
    </div>

    {{-- Leaderboard Section --}}
    <div class="bg-surface-dark/30 border border-white/5 rounded-3xl overflow-hidden mb-12 group/table">
        {{-- Table Controls --}}
        <div
            class="flex flex-row justify-between items-center gap-4 px-8 py-6 border-b border-white/5 bg-surface-darker/30">
            <div class="flex items-center gap-3">
                <span class="material-symbols-outlined text-primary">analytics</span>
                <h2 class="text-xl font-bold text-white tracking-tight">Leaderboard</h2>
            </div>

            <div class="relative" x-data="{ open: false, selected: @entangle('currentSection') }">
                <button @click="open = !open" @click.away="open = false"
                    class="flex items-center bg-background-dark/50 border border-white/5 rounded-xl px-4 py-2.5 hover:border-primary/50 transition-all min-w-[160px] justify-between group/select">
                    <div class="flex items-center gap-2">
                        <span class="material-symbols-outlined text-sm text-primary">list_alt</span>
                        <span class="text-sm font-bold text-white tracking-tight"
                            x-text="selected === 'OP' ? 'Openings' : 'Endings'"></span>
                    </div>
                    <span class="material-symbols-outlined text-white/20 text-lg transition-transform duration-300"
                        :class="open ? 'rotate-180' : ''">expand_more</span>
                </button>

                {{-- Custom Options Portal/Dropdown --}}
                <div x-show="open" x-cloak x-transition:enter="transition ease-out duration-200"
                    x-transition:enter-start="opacity-0 translate-y-2"
                    x-transition:enter-end="opacity-100 translate-y-0"
                    class="absolute right-0 mt-2 w-full min-w-[180px] bg-surface-darker/95 backdrop-blur-xl border border-white/10 rounded-2xl overflow-hidden shadow-2xl z-50">

                    <button @click="selected = 'OP'; open = false;"
                        class="w-full flex items-center gap-3 px-4 py-3 hover:bg-white/5 transition-colors group/opt"
                        :class="selected === 'OP' ? 'bg-primary/10' : ''">
                        <span class="material-symbols-outlined text-sm"
                            :class="selected === 'OP' ? 'text-primary' : 'text-white/40'">music_note</span>
                        <span class="text-sm font-bold"
                            :class="selected === 'OP' ? 'text-white' : 'text-white/60'">Openings</span>
                        <span x-show="selected === 'OP'"
                            class="material-symbols-outlined text-primary text-sm ms-auto">check</span>
                    </button>

                    <button @click="selected = 'ED'; open = false;"
                        class="w-full flex items-center gap-3 px-4 py-3 hover:bg-white/5 transition-colors group/opt"
                        :class="selected === 'ED' ? 'bg-primary/10' : ''">
                        <span class="material-symbols-outlined text-sm"
                            :class="selected === 'ED' ? 'text-primary' : 'text-white/40'">album</span>
                        <span class="text-sm font-bold"
                            :class="selected === 'ED' ? 'text-white' : 'text-white/60'">Endings</span>
                        <span x-show="selected === 'ED'"
                            class="material-symbols-outlined text-primary text-sm ms-auto">check</span>
                    </button>
                </div>
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

        {{-- Table Body --}}
        <div class="flex flex-col">
            @foreach ($songs as $index => $song)
                @isset($song->post)
                    @php
                        $img_url = null;
                        if ($song->post->thumbnail != null) {
                            if (Storage::disk('public')->exists($song->post->thumbnail)) {
                                $img_url = Storage::url($song->post->thumbnail);
                            }
                        } else {
                            $img_url =
                                'https://static.vecteezy.com/system/resources/thumbnails/005/170/408/small/banner-abstract-geometric-white-and-gray-color-background-illustration-free-vector.jpg';
                        }

                        $rankNumber = $index + 1;
                        $formattedRank = str_pad($rankNumber, 2, '0', STR_PAD_LEFT);
                    @endphp

                    <div
                        class="ranking-row grid grid-cols-[60px_1fr_120px_140px] gap-4 px-8 py-5 items-center transition-colors border-b border-white/5 hover:bg-white/5 group">
                        {{-- Rank Column --}}
                        <div class="flex flex-col items-center gap-1">
                            <span
                                class="text-2xl font-black {{ $rankNumber <= 3 ? 'text-primary' : 'text-white/90' }}">{{ $formattedRank }}</span>
                            <div class="flex items-center text-white/20">
                                <span class="material-symbols-outlined text-sm">horizontal_rule</span>
                            </div>
                        </div>

                        {{-- Theme Info Column --}}
                        <div class="flex items-center gap-6">
                            <div
                                class="w-16 h-16 rounded-lg overflow-hidden shrink-0 shadow-lg shadow-black/40 border border-white/10">
                                <img alt="Cover"
                                    class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                                    src="{{ $img_url }}" />
                            </div>
                            <div class="min-w-0">
                                <a href="{{ route('songs.show', [$song->post->slug, $song->slug]) }}"
                                    class="text-lg font-bold text-white truncate leading-tight mb-1">{{ $song->name }}
                                </a>
                                <div class="flex flex-col items-start gap-1 text-sm">
                                    <a href="{{ route('post.show', $song->post->slug) }}"
                                        class="text-primary font-bold truncate">{{ $song->post->title }}</a>

                                    @if ($song->artists->isNotEmpty())
                                        @foreach ($song->artists as $artist)
                                            <a href="{{ route('artists.show', $artist->slug) }}"
                                                class="text-white/60 truncate">
                                                {{ $artist->name }}
                                            </a>
                                        @endforeach
                                    @else
                                        <span class="text-white/60 truncate">N/A</span>
                                    @endif

                                </div>
                            </div>
                        </div>

                        {{-- Score Column --}}
                        <div class="text-center">
                            <div class="text-2xl font-black text-white tracking-tight">{{ round($song->averageRating, 2) }}
                            </div>
                            <div class="text-[10px] font-bold text-white/30 uppercase tracking-widest">Avg Rating</div>
                        </div>

                        {{-- Actions Column --}}
                        <div class="flex items-center justify-end gap-2">
                            <button
                                class="w-10 h-10 rounded-full flex items-center justify-center bg-white/5 hover:bg-primary text-white transition-all shadow-lg hover:shadow-primary/20">
                                <span class="material-symbols-outlined text-[20px] filled">play_arrow</span>
                            </button>
                            <button
                                class="w-10 h-10 rounded-full flex items-center justify-center bg-white/5 hover:bg-white/10 text-white/40 hover:text-red-400 transition-all">
                                <span
                                    class="material-symbols-outlined text-[20px] {{ $song->userScore ? 'filled text-red-400' : '' }}">favorite</span>
                            </button>
                            <div class="relative" x-data="{ open: false }">
                                <button @click="open = !open"
                                    class="w-10 h-10 rounded-full flex items-center justify-center bg-white/5 hover:bg-white/10 text-white/40 transition-all">
                                    <span class="material-symbols-outlined text-[20px]">more_vert</span>
                                </button>
                            </div>
                        </div>
                    </div>
                @endisset
            @endforeach
        </div>

        {{-- Infinite Scroll Trigger --}}
        @if ($hasMorePages)
            <div x-data x-intersect="$wire.loadMore()"
                class="p-8 border-t border-white/5 bg-surface-darker/30 flex flex-col items-center gap-4">
                <div class="w-8 h-8 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                <span class="text-xs font-bold text-white/20 uppercase tracking-widest">Loading more themes...</span>
            </div>
        @endif
    </div>
</div>
