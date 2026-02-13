<div class="flex flex-col gap-10">
    {{-- Filters Bar --}}
    <div class="bg-surface-dark/30 p-4 rounded-xl border border-white/5 backdrop-blur-sm">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
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

    {{-- Grid Section --}}
    <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-6 md:gap-8">
        @forelse($posts as $post)
            @php
                $thumbnailUrl = $post->thumbnail_src;
                if ($post->thumbnail && Storage::disk('public')->exists($post->thumbnail)) {
                    $thumbnailUrl = Storage::url($post->thumbnail);
                }
            @endphp
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
        @empty
            <div class="col-span-full flex flex-col items-center justify-center py-20 opacity-40">
                <span class="material-symbols-outlined text-6xl mb-4">movie_off</span>
                <p class="text-xl font-bold">No series found</p>
            </div>
        @endforelse
    </div>

    {{-- Infinite Scroll Trigger --}}
    @if ($posts->hasMorePages())
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
        }" x-init="observe()" class="flex justify-center py-12">
            <div class="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
        </div>
    @endif
</div>
