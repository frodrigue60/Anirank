<div x-data="{}" class="max-w-[1440px] mx-auto px-6 md:px-14 py-8 md:py-12">
    {{-- Hero Section --}}
    <div class="flex flex-col md:flex-row md:items-end justify-between gap-8 mb-12">
        <div>
            <div class="flex items-center gap-2 mb-3">
                <span
                    class="bg-primary/20 text-primary text-[10px] font-black px-2.5 py-1 rounded-full border border-primary/30 uppercase tracking-[0.2em]">Live
                    Leaderboard</span>
            </div>
            <h1 id="section-header" class="text-xl md:text-2xl lg:text-3xl font-black tracking-tighter text-white">
                @if ($rankingType === '0')
                    Global Ranking
                @else
                    Ranking {{ $currentSeason->name }} {{ $currentYear->name }}
                @endif
            </h1>
            <p class="text-white/50 mt-4 text-lg font-medium max-w-2xl">
                The definitive ranking of anime music, voted by the community and updated in real-time.
            </p>
        </div>

        {{-- Global/Seasonal Switch --}}
        <div
            class="relative flex items-center p-1.5 bg-surface-darker/50 rounded-2xl border border-white/5 w-full md:w-fit select-none">
            <div id="ranking-type-bg"
                class="absolute h-[calc(100%-12px)] w-[calc(50%-6px)] bg-primary rounded-xl shadow-lg shadow-primary/20 transition-all duration-300 ease-out {{ $rankingType === '0' ? 'left-1.5' : 'left-[calc(50%+1.5px)]' }}">
            </div>

            <button wire:click="switchRankingType('0')"
                class="relative z-10 flex-1 md:flex-none flex items-center justify-center gap-2 px-8 py-3 rounded-xl {{ $rankingType === '0' ? 'text-white' : 'text-white/40' }} font-bold transition-all duration-300">
                <span>Global</span>
            </button>

            <button wire:click="switchRankingType('1')"
                class="relative z-10 flex-1 md:flex-none flex items-center justify-center gap-2 px-8 py-3 rounded-xl {{ $rankingType === '1' ? 'text-white' : 'text-white/40' }} font-bold transition-all duration-300">
                <span>Seasonal</span>
            </button>
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

            <div class="relative">
                <select wire:model="currentSection"
                    class="appearance-none bg-background-dark/50 border border-white/5 rounded-xl px-4 py-2.5 hover:border-primary/50 transition-all min-w-[160px] text-sm font-bold text-white tracking-tight cursor-pointer focus:outline-none focus:border-primary/50">
                    <option value="ALL">All Themes</option>
                    <option value="OP">Openings</option>
                    <option value="ED">Endings</option>
                </select>
                <div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none">
                    <span class="material-symbols-outlined text-white/20 text-lg">expand_more</span>
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

                                    <span class="text-white/60 truncate">
                                        @if ($song->artists->isNotEmpty())
                                            {{ implode(', ', $song->artists->pluck('name')->toArray()) }}
                                        @else
                                            N/A
                                        @endif
                                    </span>
                                </div>
                            </div>
                        </div>

                        {{-- Score Column --}}
                        <div class="text-center">
                            <div class="text-2xl font-black text-white tracking-tight">
                                {{ number_format($song->averageRating, 1) }}
                            </div>
                            <div class="text-[10px] font-bold text-white/30 uppercase tracking-widest">Avg Rating</div>
                        </div>

                        {{-- Actions Column --}}
                        <div class="flex items-center justify-end gap-2">
                            <button
                                @click.stop="console.log('RankingTable: Alpine triggered for ID', {{ $song->id }}); window.playSongGlobal({{ $song->id }})"
                                onclick="console.log('RankingTable: Native HTML click for ID {{ $song->id }}')"
                                class="w-10 h-10 rounded-full flex items-center justify-center bg-white/5 hover:bg-primary text-white transition-all shadow-lg hover:shadow-primary/20 cursor-pointer z-[99]">
                                <span
                                    class="material-symbols-outlined text-[20px] filled pointer-events-none">play_arrow</span>
                            </button>
                            <button wire:click="toggleFavorite({{ $song->id }})" wire:loading.attr="disabled"
                                class="w-10 h-10 rounded-full flex items-center justify-center bg-white/5 hover:bg-white/10 text-white/40 hover:text-red-400 transition-all">
                                <span
                                    class="material-symbols-outlined text-[20px] {{ $song->isFavorited() ? 'filled text-red-400' : '' }}">favorite</span>
                            </button>
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
                <span class="text-xs font-bold text-white/20 uppercase tracking-widest">Loading more
                    themes...</span>
            </div>
        @endif
    </div>
</div>
