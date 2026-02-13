<div class="max-w-[1440px] mx-auto px-4 md:px-8 py-8">

    {{-- Header & Filters --}}
    <div class="flex flex-col gap-6 mb-8">
        <div class="flex flex-col md:flex-row justify-between items-end gap-4">
            <div>
                <h1 class="text-3xl font-black tracking-tight text-white mb-2">Search Anime</h1>
                <div class="h-1 w-20 bg-primary rounded-full"></div>
            </div>

            {{-- View Mode Switcher --}}
            <div class="flex items-center gap-1 bg-surface-darker/50 p-1 rounded-lg border border-white/5">
                <button wire:click="setViewMode('grid_small')"
                    class="p-2 rounded-md transition-all {{ $viewMode === 'grid_small' ? 'bg-primary/20 text-primary' : 'text-white/40 hover:text-white hover:bg-white/5' }}"
                    title="Small Grid">
                    <span class="material-symbols-outlined text-[20px]">grid_view</span>
                </button>
                <button wire:click="setViewMode('grid_large')"
                    class="p-2 rounded-md transition-all {{ $viewMode === 'grid_large' ? 'bg-primary/20 text-primary' : 'text-white/40 hover:text-white hover:bg-white/5' }}"
                    title="Large Grid">
                    <span class="material-symbols-outlined text-[20px]">window</span>
                </button>
                <button wire:click="setViewMode('list')"
                    class="p-2 rounded-md transition-all {{ $viewMode === 'list' ? 'bg-primary/20 text-primary' : 'text-white/40 hover:text-white hover:bg-white/5' }}"
                    title="List View">
                    <span class="material-symbols-outlined text-[20px]">view_list</span>
                </button>
            </div>
        </div>

        {{-- Filters Bar --}}
        <div
            class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 bg-surface-dark/30 p-4 rounded-xl border border-white/5">
            {{-- Search --}}
            <div class="relative group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-focus-within:text-primary transition-colors">Search</label>
                <div class="relative">
                    <span
                        class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-white/30 group-focus-within:text-primary transition-colors">search</span>
                    <input wire:model.debounce.300ms="name" type="text" placeholder="Search anime..."
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:outline-none focus:border-primary/50 transition-all placeholder:text-white/20">
                </div>
            </div>

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

            {{-- Format --}}
            <div class="relative group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Format</label>
                <div class="relative">
                    <select wire:model="format_id"
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-4 pr-10 text-sm text-white focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer hover:bg-surface-darker/80">
                        <option value="">All Formats</option>
                        @foreach ($formats as $format)
                            <option value="{{ $format->id }}">{{ $format->name }}</option>
                        @endforeach
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-lg group-focus-within:text-primary transition-colors">expand_more</span>
                </div>
            </div>
        </div>
    </div>

    {{-- Content Grid/List --}}
    <div class="min-h-[400px]">
        @if ($viewMode === 'grid_small')
            <div class="grid grid-cols-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 md:gap-6">
                @foreach ($posts as $post)
                    <div class="group relative">
                        <div class="aspect-[2/3] rounded-lg overflow-hidden bg-surface-darker shadow-lg relative">
                            {{-- Cover Image --}}
                            @php
                                $thumbnailUrl = $post->thumbnail_src;
                                if ($post->thumbnail && Storage::disk('public')->exists($post->thumbnail)) {
                                    $thumbnailUrl = Storage::url($post->thumbnail);
                                }
                            @endphp
                            <img src="{{ $thumbnailUrl }}" alt="{{ $post->title }}" loading="lazy"
                                class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110">

                            <div class="absolute top-3 right-3 flex items-end gap-1.5 z-20">
                                <span
                                    class="px-2 py-1 rounded bg-black/60 backdrop-blur-md border border-white/10 text-[10px] font-black uppercase tracking-widest text-white shadow-xl">
                                    {{ $post->format->name }}
                                </span>
                                <span
                                    class="px-2 py-1 rounded bg-primary/80 backdrop-blur-md border border-primary/20 text-[10px] font-black uppercase tracking-widest text-white shadow-lg flex items-center gap-1">
                                    <span class="material-symbols-outlined text-[14px] leading-none">music_note</span>
                                    {{ $post->songs->count() }}
                                </span>
                            </div>
                            {{-- Hover Overlay --}}
                            <div
                                class="absolute inset-0 bg-gradient-to-t from-black/90 via-black/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex flex-col justify-end p-3">
                                <a href="{{ route('post.show', $post->slug) }}" class="absolute inset-0 z-10"></a>
                            </div>
                        </div>
                        <h3
                            class="mt-2 text-sm font-bold text-white leading-tight line-clamp-2 group-hover:text-primary transition-colors">
                            <a href="{{ route('post.show', $post->slug) }}">{{ $post->title }}</a>
                        </h3>
                    </div>
                @endforeach
            </div>
        @elseif($viewMode === 'grid_large')
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6">
                @foreach ($posts as $post)
                    <div
                        class="flex bg-surface-dark rounded-lg overflow-hidden shadow-sm hover:shadow-md transition-shadow group relative border border-white/5">

                        {{-- COVER (Left) --}}
                        <div class="w-[120px] md:w-[180px] shrink-0 relative aspect-[2/3]">
                            @php
                                $thumbnailUrl = $post->thumbnail_src;
                                if ($post->thumbnail && Storage::disk('public')->exists($post->thumbnail)) {
                                    $thumbnailUrl = Storage::url($post->thumbnail);
                                }
                            @endphp
                            <a href="{{ route('post.show', $post->slug) }}" class="block w-full h-full">
                                <img src="{{ $thumbnailUrl }}" alt="{{ $post->title }}" loading="lazy"
                                    class="w-full h-full object-cover">
                            </a>

                            {{-- Overlay: Title & Studio --}}
                            <div
                                class="absolute inset-0 bg-gradient-to-t from-black/90 via-transparent to-transparent flex flex-col justify-end p-2 pointer-events-none">
                                <h3
                                    class="text-sm font-bold text-white leading-tight line-clamp-2 mb-0.5 text-shadow-sm">
                                    {{ $post->title }}
                                </h3>
                                @if ($post->producers->isNotEmpty())
                                    <div class="text-[12px] font-bold text-primary truncate">
                                        {{ $post->producers->first()->name }}
                                    </div>
                                @endif
                            </div>
                        </div>

                        {{-- DATA (Right) --}}
                        <div class="flex-1 p-3 flex flex-col min-w-0 bg-surface-darker/30">
                            {{-- Header: Season/Year & Format --}}
                            <div
                                class="flex items-center justify-start text-[11px] font-bold text-white/50 mb-2 uppercase tracking-wide gap-2">
                                <div class="flex items-center gap-2">
                                    @if ($post->season && $post->year)
                                        <span>{{ $post->season->name }} {{ $post->year->name }}</span>
                                    @else
                                        <span>Start Date N/A</span>
                                    @endif

                                    @if ($post->format)
                                        <span>&bull;</span>
                                        <span> {{ $post->format->name }}</span>
                                    @endif

                                    @if ($post->songs->isNotEmpty())
                                        <span>&bull;</span>
                                        <span> {{ $post->songs->count() }} Songs</span>
                                    @endif
                                </div>
                            </div>
                            <div>

                            </div>


                            {{-- Body: Description --}}
                            <div class="text-xs text-white/60 overflow-hidden line-clamp-6 leading-relaxed mb-auto">
                                {!! strip_tags($post->description) !!}
                            </div>

                            {{-- Footer: Tags Placeholder --}}
                            <div
                                class="mt-3 pt-3 border-t border-white/5 flex flex-wrap gap-1.5 opacity-60 hover:opacity-100 transition-opacity">

                                @if ($post->tags)
                                    @foreach ($post->tags->take(3) as $tag)
                                        <a href="#"
                                            class="px-2 py-0.5 rounded-full bg-white/5 text-[10px] font-bold text-primary hover:bg-primary hover:text-white transition-colors">
                                            {{ $tag->name }}
                                        </a>
                                    @endforeach
                                @else
                                    <span
                                        class="px-2 py-0.5 rounded-full bg-white/5 text-[10px] font-bold text-white/40">No
                                        Tags</span>
                                @endif

                                {{-- Example Placeholder --}}
                                {{--
                                <span
                                    class="px-2 py-0.5 rounded-full bg-white/5 text-[10px] font-bold text-white/40">Action</span>
                                <span
                                    class="px-2 py-0.5 rounded-full bg-white/5 text-[10px] font-bold text-white/40">Fantasy</span>
                                --}}
                            </div>
                        </div>
                    </div>
                @endforeach
            </div>
        @elseif($viewMode === 'list')
            <div class="flex flex-col gap-2">
                @foreach ($posts as $index => $post)
                    <div
                        class="flex items-center gap-4 bg-surface-dark/30 border border-white/5 p-3 rounded-lg hover:bg-white/5 transition-colors group">
                        {{-- Cover --}}
                        @php
                            $thumbnailUrl = $post->thumbnail_src;
                            if ($post->thumbnail && Storage::disk('public')->exists($post->thumbnail)) {
                                $thumbnailUrl = Storage::url($post->thumbnail);
                            }
                        @endphp
                        <div class="w-12 h-12 rounded overflow-hidden shrink-0">
                            <img src="{{ $thumbnailUrl }}" alt="{{ $post->title }}" loading="lazy"
                                class="w-full h-full object-cover">
                        </div>

                        {{-- Title & Meta --}}
                        <div class="flex-1 min-w-0">
                            <h3
                                class="text-sm font-bold text-white leading-tight truncate px-1 group-hover:text-primary transition-colors">
                                <a href="{{ route('post.show', $post->slug) }}">{{ $post->title }}</a>
                            </h3>
                            <div class="flex items-center gap-3 text-xs text-white/40 mt-0.5 px-1">
                                @if ($post->format)
                                    <span>{{ $post->format->name }}</span>
                                @endif
                                @if ($post->season && $post->year)
                                    <span>&bull;</span>
                                    <span>{{ $post->season->name }} {{ $post->year->name }}</span>
                                @endif
                                @if ($post->songs->isNotEmpty())
                                    <span>&bull;</span>
                                    <span>{{ $post->songs->count() }} Songs</span>
                                @endif
                            </div>
                        </div>

                        {{-- Action --}}
                        <div class="ml-auto">
                            <a href="{{ route('post.show', $post->slug) }}"
                                class="w-8 h-8 flex items-center justify-center rounded-full bg-white/5 text-white/40 hover:bg-primary hover:text-white transition-all">
                                <span class="material-symbols-outlined text-lg">chevron_right</span>
                            </a>
                        </div>
                    </div>
                @endforeach
            </div>
        @endif

        {{-- Loader/Infinite Scroll --}}
        @if ($hasMorePages)
            <div x-data x-intersect="$wire.loadMore()" class="py-12 flex flex-col items-center gap-4">
                <div class="w-8 h-8 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
            </div>
        @endif

        {{-- Empty State --}}
        @if ($posts->isEmpty())
            <div class="py-20 flex flex-col items-center justify-center text-center">
                <span class="material-symbols-outlined text-6xl text-white/10 mb-4">search_off</span>
                <p class="text-white/40 text-lg font-medium">No animes found matching your criteria.</p>
                <button wire:click="$set('name', '')"
                    class="mt-4 text-primary hover:text-primary-light text-sm font-bold">Clear Filters</button>
            </div>
        @endif

    </div>
</div>
