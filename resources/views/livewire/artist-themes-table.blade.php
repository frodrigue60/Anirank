<div class="flex flex-col gap-8">
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
                    <input wire:model.debounce.300ms="name"
                        class="w-full h-11 bg-surface-darker/50 border border-white/10 rounded-xl pl-12 pr-4 text-sm text-white focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-white/20 transition-all"
                        placeholder="Search themes..." type="text" />
                </div>
            </div>

            {{-- Type Filter --}}
            <div class="relative min-w-[140px] group">
                <label
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Type</label>
                <div class="relative">
                    <select wire:model="type"
                        class="w-full h-11 bg-surface-dark border-white/10 rounded-xl text-sm text-white/80 focus:ring-primary focus:border-primary py-2 pl-4 pr-10 appearance-none cursor-pointer transition-all hover:bg-surface-dark/80 focus:bg-surface-darker">
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
                    <select wire:model="year_id"
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
                    <select wire:model="season_id"
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
                    <select wire:model="sort"
                        class="w-full h-11 bg-surface-dark border-white/10 rounded-xl text-sm text-white/80 focus:ring-primary focus:border-primary py-2 pl-4 pr-10 appearance-none cursor-pointer transition-all hover:bg-surface-dark/80 focus:bg-surface-darker">
                        @foreach ($sortMethods as $m)
                            <option value="{{ $m['value'] }}">{{ $m['name'] }}</option>
                        @endforeach
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-xl group-focus-within:text-primary transition-colors">sort</span>
                </div>
            </div>

            {{-- Clear Button --}}
            <button wire:click="clearFilters"
                class="self-end h-11 bg-primary/10 hover:bg-primary/20 text-primary px-6 rounded-xl font-bold text-sm transition-all border border-primary/20 flex items-center gap-2 group/btn">
                <span
                    class="material-symbols-outlined text-lg group-hover/btn:rotate-180 transition-transform">close</span>
                Clear
            </button>
        </div>
    </section>

    {{-- Data Container --}}
    <div class="min-h-[400px]">
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            @foreach ($songs as $song)
                <a href="{{ $song->url }}"
                    class="group relative bg-surface-darker p-4 rounded-xl hover:bg-surface-dark transition-colors cursor-pointer border border-white/5 flex gap-4 items-center">
                    <div class="relative w-24 h-24 shrink-0 rounded-lg overflow-hidden shadow-lg border border-white/5">
                        <img src="{{ $song->thumbnailUrl }}" alt="{{ $song->title }}"
                            class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500">
                        <div
                            class="absolute top-1 left-1 {{ $loop->iteration <= 3 ? 'bg-primary' : 'bg-surface-dark' }} text-white text-[10px] font-black px-1.5 py-0.5 rounded shadow border border-white/10 uppercase tracking-tighter">
                            #{{ $loop->iteration }}</div>
                    </div>
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center justify-between gap-2 mb-1">
                            <h3 class="font-bold text-white truncate text-lg group-hover:text-primary transition-colors"
                                title="{{ $song->title }}">
                                {{ $song->title }}
                            </h3>
                        </div>
                        <p class="text-sm text-primary font-black truncate mb-1.5 uppercase tracking-wide">
                            {{ $song->post->title }}</p>

                        <div class="flex flex-wrap items-center gap-x-4 gap-y-2 mt-2">
                            <div
                                class="flex items-center gap-1.5 bg-black/40 px-2 py-0.5 rounded-lg border border-white/5 text-yellow-400 text-xs font-black">
                                <span class="material-symbols-outlined filled text-[14px]">star</span>
                                {{ number_format($song->averageRating / 10, 1) }}/10
                            </div>

                            <div class="flex items-center gap-1.5 text-white/30">
                                <span class="material-symbols-outlined text-[16px]">visibility</span>
                                <span
                                    class="text-[11px] font-black tracking-wider uppercase">{{ number_format($song->view_count) }}</span>
                            </div>

                            <div class="flex items-center gap-1.5 text-white/30">
                                <span class="material-symbols-outlined text-[16px]">calendar_today</span>
                                <span
                                    class="text-[11px] font-black tracking-wider uppercase">{{ $song->post->year->name ?? 'N/A' }}</span>
                            </div>
                        </div>

                        <div class="mt-3 flex items-center gap-2">
                            <span
                                class="bg-white/5 text-white/40 text-[9px] font-black px-2 py-0.5 rounded uppercase tracking-widest border border-white/5">
                                {{ $song->type }}{{ $song->number }}
                            </span>
                        </div>
                    </div>
                </a>
            @endforeach
        </div>

        {{-- Infinite Scroll Trigger --}}
        @if ($hasMorePages)
            <div x-data="{
                        observe() {
                            let observer = new IntersectionObserver((entries) => {
                                entries.forEach(entry => {
                                    if (entry.isIntersecting) {
                                        @this.call('loadMore')
                                    }
                                })
                            }, {
                                rootMargin: '200px'
                            })
                            observer.observe(this.$el)
                        }
                    }" x-init="observe()" class="flex justify-center py-12">
                <div class="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
            </div>
        @endif

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
    </div>
</div>