<div wire:init="loadData" x-data="{}" class="flex flex-col gap-8">
    {{-- Filter Panel --}}
    <section class="bg-surface-dark/30 p-6 rounded-2xl border border-white/5 shadow-2xl backdrop-blur-md">
        <div class="flex flex-wrap items-center gap-4">
            {{-- Search --}}
            <div class="relative flex-1 min-w-[300px] group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-focus-within:text-primary transition-colors">Search
                    Theme</label>
                <div class="relative">
                    <span
                        class="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-primary text-[22px] group-focus-within:scale-110 transition-transform">search</span>
                    <input wire:model.live.debounce.300ms="name"
                        class="w-full h-11 bg-surface-darker/50 border border-white/10 rounded-xl pl-12 pr-4 text-sm text-white focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-white/20 transition-all"
                        placeholder="Search themes..." type="text" />
                </div>
            </div>

            {{-- Type Filter --}}
            <div class="relative min-w-[140px] group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Type</label>
                <div class="relative">
                    <select wire:model.live="type"
                        class="w-full h-11 bg-surface-dark border-white/10 rounded-xl text-sm text-white/80 focus:ring-primary focus:border-primary py-2 pl-4 pr-10 appearance-none cursor-pointer transition-all hover:bg-surface-dark/80 focus:bg-surface-darker">
                        <option value="">All Types</option>
                        @foreach ($types as $t)
                            <option value="{{ $t['value'] }}">{{ $t['name'] }}</option>
                        @endforeach
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-xl group-focus-within:text-primary transition-colors">expand_more</span>
                </div>
            </div>

            {{-- Year Filter --}}
            <div class="relative min-w-[120px] group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Year</label>
                <div class="relative">
                    <select wire:model.live="year_id"
                        class="w-full h-11 bg-surface-dark border-white/10 rounded-xl text-sm text-white/80 focus:ring-primary focus:border-primary py-2 pl-4 pr-10 appearance-none cursor-pointer transition-all hover:bg-surface-dark/80 focus:bg-surface-darker">
                        <option value="">All Years</option>
                        @foreach ($years as $y)
                            <option value="{{ $y->id }}">{{ $y->name }}</option>
                        @endforeach
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-xl group-focus-within:text-primary transition-colors">calendar_today</span>
                </div>
            </div>

            {{-- Season Filter --}}
            <div class="relative min-w-[140px] group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Season</label>
                <div class="relative">
                    <select wire:model.live="season_id"
                        class="w-full h-11 bg-surface-dark border-white/10 rounded-xl text-sm text-white/80 focus:ring-primary focus:border-primary py-2 pl-4 pr-10 appearance-none cursor-pointer transition-all hover:bg-surface-dark/80 focus:bg-surface-darker">
                        <option value="">All Seasons</option>
                        @foreach ($seasons as $s)
                            <option value="{{ $s->id }}">{{ $s->name }}</option>
                        @endforeach
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-xl group-focus-within:text-primary transition-colors">routine</span>
                </div>
            </div>

            {{-- Sort Filter --}}
            <div class="relative min-w-[160px] group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Sort
                    By</label>
                <div class="relative">
                    <select wire:model.live="sort"
                        class="w-full h-11 bg-surface-dark border-white/10 rounded-xl text-sm text-white/80 focus:ring-primary focus:border-primary py-2 pl-4 pr-10 appearance-none cursor-pointer transition-all hover:bg-surface-dark/80 focus:bg-surface-darker">
                        @foreach ($sortMethods as $m)
                            <option value="{{ $m['value'] }}">{{ $m['name'] }}</option>
                        @endforeach
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-xl group-focus-within:text-primary transition-colors">sort</span>
                </div>
            </div>
        </div>
    </section>

    {{-- Data Container --}}
    <div class="min-h-[400px]">
        @if ($readyToLoad)
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-4">
                @foreach ($songs as $song)
                    <div wire:key="artist-theme-{{ $song->id }}"
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
                                    {{ $song->type }}{{ $song->number }}</div>
                                <h3
                                    class="text-2xl font-bold text-white group-hover:text-primary transition-colors text-glow">
                                    {{ $song->name }}</h3>
                                <div class="text-slate-300 text-sm font-medium">
                                    {{ $song->artists->pluck('name')->implode(', ') }}
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
                <div
                    class="py-20 flex flex-col items-center justify-center text-center bg-surface-dark/10 rounded-3xl border-2 border-dashed border-white/5">
                    <span class="material-symbols-outlined text-6xl text-white/10 mb-4">music_off</span>
                    <p class="text-white/40 text-lg font-black uppercase tracking-widest">No themes found</p>
                    <button wire:click="clearFilters"
                        class="mt-4 text-primary hover:text-primary-light text-sm font-black uppercase tracking-widest">Clear
                        Filters</button>
                </div>
            @endif

            @if ($hasMorePages && $readyToLoad)
                <div wire:intersect="loadMore" wire:key="intersect-artist-themes" class="flex justify-center py-12">
                    <div class="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                </div>
            @endif
        @else
            @include('livewire.skeletons.card-skeleton')
        @endif
    </div>
</div>
