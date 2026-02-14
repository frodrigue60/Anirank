<div wire:init="loadData">
    <div class="flex flex-col gap-6 mb-8">
        {{-- Filters Bar --}}
        <div
            class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 bg-surface-dark/30 p-4 rounded-xl border border-white/5 shadow-2xl backdrop-blur-sm">

            {{-- Search Input --}}
            <div class="relative group lg:col-span-1">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-focus-within:text-primary transition-colors">Search</label>
                <div class="relative">
                    <span
                        class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-white/30 group-focus-within:text-primary transition-colors text-base">search</span>
                    <input wire:model.live.debounce.300ms="name" type="text" placeholder="Search anime..."
                        class="w-full !bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-10 pr-4 text-sm !text-white focus:outline-none focus:border-primary/50 transition-all placeholder:text-white/20">
                </div>
            </div>

            {{-- Type Filter --}}
            <div class="relative group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Type</label>
                <div class="relative">
                    <select wire:model.live="type"
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-4 pr-10 text-sm text-white focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer hover:bg-surface-darker/80">
                        <option value="">All Types</option>
                        @foreach ($types as $t)
                            <option value="{{ $t['value'] }}">{{ $t['name'] }}</option>
                        @endforeach
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-lg group-focus-within:text-primary transition-colors">expand_more</span>
                </div>
            </div>

            {{-- Year Filter --}}
            <div class="relative group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Year</label>
                <div class="relative">
                    <select wire:model.live="year_id"
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-4 pr-10 text-sm text-white focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer hover:bg-surface-darker/80">
                        <option value="">All Years</option>
                        @foreach ($years as $year)
                            <option value="{{ $year->id }}">{{ $year->name }}</option>
                        @endforeach
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-lg group-focus-within:text-primary transition-colors">expand_more</span>
                </div>
            </div>

            {{-- Season Filter --}}
            <div class="relative group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Season</label>
                <div class="relative">
                    <select wire:model.live="season_id"
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-4 pr-10 text-sm text-white focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer hover:bg-surface-darker/80">
                        <option value="">All Seasons</option>
                        @foreach ($seasons as $season)
                            <option value="{{ $season->id }}">{{ $season->name }}</option>
                        @endforeach
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-lg group-focus-within:text-primary transition-colors">expand_more</span>
                </div>
            </div>

            {{-- Sort Filter --}}
            <div class="relative group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Sort
                    By</label>
                <div class="relative">
                    <select wire:model.live="sort"
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-4 pr-10 text-sm text-white focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer hover:bg-surface-darker/80">
                        @foreach ($sortMethods as $m)
                            <option value="{{ $m['value'] }}">{{ $m['name'] }}</option>
                        @endforeach
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-lg group-focus-within:text-primary transition-colors">sort</span>
                </div>
            </div>
        </div>
    </div>

    {{-- Content Grid --}}
    <div class="min-h-[400px]">
        @if ($readyToLoad)
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-4">
                @foreach ($songs as $song)
                    <div wire:key="user-fav-{{ $song->id }}"
                        class="group relative overflow-hidden rounded-xl h-48 card-hover transition-all duration-300 border border-primary/10 bg-background-dark">
                        <div class="absolute inset-0 bg-cover bg-center transition-transform duration-500 group-hover:scale-105"
                            style="background-image: url('{{ Storage::url($song->post->banner) }}'); filter: brightness(0.5);">
                        </div>
                        <div
                            class="absolute inset-0 bg-gradient-to-r from-background-dark via-background-dark/80 to-transparent">
                        </div>
                        <div class="relative h-full p-6 flex items-center justify-between">
                            <div class="space-y-1">
                                <div
                                    class="inline-flex items-center px-2 py-0.5 rounded bg-primary text-[10px] font-bold text-white mb-2 uppercase tracking-wider">
                                    {{ $song->slug }}</div>
                                <h3
                                    class="text-2xl font-bold text-white group-hover:text-primary transition-colors text-glow">
                                    {{ $song->name }}</h3>
                                <div class="text-slate-300 text-sm font-medium">
                                    @foreach ($song->artists as $artist)
                                        <span
                                            class="hover:text-primary transition-colors cursor-pointer">{{ $artist->name }}</span>
                                        @if (!$loop->last)
                                            ,
                                        @endif
                                    @endforeach
                                </div>
                                <div class="text-slate-500 text-xs italic mt-2">
                                    {{ $song->post->title }}</div>
                            </div>
                            <div class="flex flex-col items-end gap-2">
                                <div
                                    class="glass px-3 py-2 rounded-lg border-primary/30 flex items-center gap-1.5 shadow-lg">
                                    <span class="material-symbols-outlined text-primary text-sm fill-1">star</span>
                                    <span
                                        class="text-white font-bold text-lg">{{ number_format($song->ratings_avg_rating ?? 0, 1) }}</span>
                                </div>
                                <a href="{{ route('songs.show', [$song->post->slug, $song->slug]) }}"
                                    class="mt-4 flex items-center justify-center h-10 w-10 rounded-full bg-white/10 hover:bg-primary transition-all text-white backdrop-blur-sm border border-white/10 group-hover:border-primary/50">
                                    <span class="material-symbols-outlined">play_arrow</span>
                                </a>
                            </div>
                        </div>
                    </div>
                @endforeach
            </div>

            {{-- Empty State --}}
            @if ($songs->isEmpty())
                <div class="py-24 flex flex-col items-center justify-center text-center opacity-40">
                    <span class="material-symbols-outlined text-7xl mb-4">favorite_border</span>
                    <p class="text-xl font-bold">No favorite themes found matching your filters.</p>
                    <button wire:click="$set('name', '')"
                        class="mt-4 text-primary hover:text-primary-light text-sm font-bold uppercase tracking-widest">Clear
                        Searches</button>
                </div>
            @endif
        @else
            @include('livewire.skeletons.card-skeleton')
        @endif

        {{-- Loader/Infinite Scroll --}}
        @if ($hasMorePages && $readyToLoad)
            <div wire:intersect="loadMore" wire:key="intersect-user-favorites"
                class="py-12 flex flex-col items-center gap-4">
                <div class="w-8 h-8 border-4 border-primary/20 border-t-primary rounded-full animate-spin">
                </div>
            </div>
        @endif
    </div>
</div>
