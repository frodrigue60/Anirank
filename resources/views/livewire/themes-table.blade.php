<div class="max-w-[1440px] mx-auto px-4 md:px-8 py-8">

    {{-- Header & Filters --}}
    <div class="flex flex-col gap-6 mb-8">
        <div class="flex flex-col md:flex-row justify-between items-end gap-4">
            <div>
                <h1 class="text-3xl font-black tracking-tight text-white mb-2">Search Themes</h1>
                <div class="h-1 w-20 bg-primary rounded-full"></div>
            </div>
        </div>

        {{-- Filters Bar --}}
        <div
            class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 bg-surface-dark/30 p-4 rounded-xl border border-white/5">
            {{-- Search --}}
            <div class="relative group lg:col-span-1">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-focus-within:text-primary transition-colors">Search</label>
                <div class="relative">
                    <span
                        class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-white/30 group-focus-within:text-primary transition-colors">search</span>
                    <input wire:model.debounce.300ms="name" type="text" placeholder="Search themes..."
                        class="w-full !bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-10 pr-4 text-sm !text-white focus:outline-none focus:border-primary/50 transition-all placeholder:text-white/20">
                </div>
            </div>

            {{-- Type --}}
            <div class="relative group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Type</label>
                <div class="relative">
                    <select wire:model="type"
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-4 pr-10 text-sm text-white focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer hover:bg-surface-darker/80">
                        <option value="">All Types</option>
                        @foreach ($types as $typeOption)
                            <option value="{{ $typeOption['value'] }}">{{ $typeOption['name'] }}</option>
                        @endforeach
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-lg group-focus-within:text-primary transition-colors">expand_more</span>
                </div>
            </div>

            {{-- Year --}}
            <div class="relative group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Year</label>
                <div class="relative">
                    <select wire:model="year_id"
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

            {{-- Season --}}
            <div class="relative group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Season</label>
                <div class="relative">
                    <select wire:model="season_id"
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

            {{-- Sort --}}
            <div class="relative group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Sort</label>
                <div class="relative">
                    <select wire:model="sort"
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-4 pr-10 text-sm text-white focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer hover:bg-surface-darker/80">
                        @foreach ($sortMethods as $sortOption)
                            <option value="{{ $sortOption['value'] }}">{{ $sortOption['name'] }}</option>
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
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            @foreach ($songs as $song)
                <a href="{{ $song->url }}"
                    class="group relative bg-surface-darker p-4 rounded-xl hover:bg-surface-dark transition-colors cursor-pointer border border-white/5 flex gap-4 items-center">
                    <div class="relative w-20 h-20 shrink-0 rounded-lg overflow-hidden">
                        <img src="{{ $song->thumbnailUrl }}" alt="{{ $song->title }}"
                            class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500">
                        <div
                            class="absolute top-1 left-1 {{ $loop->iteration <= 3 ? 'bg-primary' : 'bg-surface-dark' }} text-white text-xs font-bold px-1.5 py-0.5 rounded shadow border border-white/10">
                            #{{ $loop->iteration }}</div>
                    </div>
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center justify-between">
                            <h3 class="font-bold text-white truncate text-lg" title="{{ $song->title }}">
                                {{ $song->name }}
                            </h3>
                            <div
                                class="flex items-center gap-1 bg-surface-dark px-2 py-0.5 rounded text-yellow-400 text-xs font-bold">
                                <span class="material-symbols-outlined filled text-[14px]">star</span>
                                {{ number_format($song->ratings_avg_rating ?? ($song->averageRating ?? 0), 1) }}
                            </div>
                        </div>
                        <p class="text-sm text-primary font-medium truncate">{{ $song->post->title }}</p>
                        <p class="text-xs text-white/50 truncate">
                            @foreach ($song->artists as $artist)
                                {{ $artist->name }}{{ !$loop->last ? ', ' : '' }}
                            @endforeach
                        </p>
                    </div>
                </a>
            @endforeach
        </div>

        {{-- Loader/Infinite Scroll --}}
        @if ($hasMorePages)
            <div x-data x-intersect="$wire.loadMore()" class="py-12 flex flex-col items-center gap-4">
                <div class="w-8 h-8 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
            </div>
        @endif

        {{-- Empty State --}}
        @if ($songs->isEmpty())
            <div class="py-20 flex flex-col items-center justify-center text-center">
                <span class="material-symbols-outlined text-6xl text-white/10 mb-4">music_off</span>
                <p class="text-white/40 text-lg font-medium">No themes found matching your criteria.</p>
                <button wire:click="$set('name', '')"
                    class="mt-4 text-primary hover:text-primary-light text-sm font-bold">Clear Filters</button>
            </div>
        @endif
    </div>
</div>