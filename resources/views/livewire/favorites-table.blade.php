<div class="flex flex-col gap-10">
    {{-- Filters Bar --}}
    <div class="bg-surface-dark/30 p-4 rounded-xl border border-white/5 backdrop-blur-sm">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
            {{-- Search --}}
            <div class="relative group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-focus-within:text-primary transition-colors">Search</label>
                <div class="relative">
                    <span
                        class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-white/30 group-focus-within:text-primary transition-colors">search</span>
                    <input wire:model.debounce.300ms="name" type="text" placeholder="Anime title..."
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:outline-none focus:border-primary/50 transition-all placeholder:text-white/20">
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
                    <div class="relative w-24 h-24 shrink-0 rounded-lg overflow-hidden border border-white/10 shadow-lg">
                        <img src="{{ $song->thumbnailUrl }}" alt="{{ $song->title }}"
                            class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500">
                        <div
                            class="absolute top-1 left-1 bg-primary/90 backdrop-blur-md text-white text-[10px] font-black px-1.5 py-0.5 rounded shadow-lg border border-white/10">
                            {{ $song->type }}
                        </div>
                    </div>
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center justify-between gap-2 mb-1">
                            <h3 class="font-bold text-white truncate text-base lg:text-lg group-hover:text-primary transition-colors"
                                title="{{ $song->name }}">
                                {{ $song->name }}
                            </h3>
                            <div
                                class="flex items-center gap-1 bg-surface-dark px-2 py-0.5 rounded-md text-yellow-400 text-[11px] font-black border border-white/5 shadow-sm shrink-0">
                                <span class="material-symbols-outlined filled text-[14px]">star</span>
                                {{ $song->scoreString }}
                            </div>
                        </div>
                        <p class="text-xs font-bold text-primary truncate mb-1">
                            {{ $song->post->title }}
                        </p>
                        <div class="flex flex-col gap-0.5">
                            <p class="text-[11px] text-white/40 truncate font-medium">
                                @foreach ($song->artists as $artist)
                                    {{ $artist->name }}{{ !$loop->last ? ', ' : '' }}
                                @endforeach
                            </p>
                            <span class="text-[10px] font-black text-white/20 uppercase tracking-widest mt-1">
                                {{ $song->post->season->name ?? '' }} {{ $song->post->year->name ?? '' }}
                            </span>
                        </div>
                    </div>
                </a>
            @endforeach
        </div>

        {{-- Loader/Infinite Scroll --}}
        @if ($songs->hasMorePages())
            <div x-data="{
                    observe() {
                        let observer = new IntersectionObserver((entries) => {
                            entries.forEach(entry => {
                                if (entry.isIntersecting) {
                                    @this.call('loadMore')
                                }
                            })
                        }, {
                            rootMargin: '200px',
                        })
                        observer.observe(this.$el)
                    }
                }" x-init="observe()" class="py-12 flex flex-col items-center gap-4">
                <div class="w-8 h-8 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
            </div>
        @endif

        {{-- Empty State --}}
        @if ($songs->isEmpty())
            <div class="py-32 flex flex-col items-center justify-center text-center opacity-40">
                <span class="material-symbols-outlined text-7xl mb-4">favorite_border</span>
                <p class="text-xl font-bold">No favorites found</p>
                <p class="text-sm font-medium mt-2">Try adjusting your filters or add some themes to your favorites!</p>
            </div>
        @endif
    </div>
</div>