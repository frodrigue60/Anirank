<div class="flex flex-col gap-8">
    {{-- Search Section --}}
    <div class="bg-surface-dark/30 p-4 rounded-xl border border-white/5 backdrop-blur-sm">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {{-- Search --}}
            <div class="relative group">
                <label for="search"
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-focus-within:text-primary transition-colors">
                    Search
                </label>
                <div class="relative">
                    <span
                        class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-white/30 group-focus-within:text-primary transition-colors">
                        search
                    </span>
                    <input wire:model.debounce.500ms="search" type="text" id="search"
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:outline-none focus:border-primary/50 transition-all placeholder:text-white/20"
                        placeholder="Search studio...">
                </div>
            </div>

            {{-- Sort --}}
            <div class="relative group">
                <label for="sort"
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-focus-within:text-primary transition-colors">
                    Sort By
                </label>
                <div class="relative">
                    <select wire:model="sort" id="sort"
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-4 pr-10 text-sm text-white focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer hover:bg-surface-darker/80">
                        <option value="name_asc">Name (A-Z)</option>
                        <option value="name_desc">Name (Z-A)</option>
                        <option value="series_count">Most Series</option>
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-lg group-focus-within:text-primary transition-colors">expand_more</span>
                </div>
            </div>
        </div>
    </div>

    {{-- Grid Section --}}
    <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 xl:grid-cols-6 gap-6 md:gap-8">
        @forelse($studios as $studio)
            @php
                $thumbnailUrl = '';
                if ($studio->thumbnail != null && Storage::disk('public')->exists($studio->thumbnail)) {
                    $thumbnailUrl = Storage::url($studio->thumbnail);
                } elseif ($studio->thumbnail_src != null) {
                    $thumbnailUrl = $studio->thumbnail_src;
                } else {
                    $thumbnailUrl = asset('resources/images/default-thumbnail.jpg');
                }
            @endphp
            <div class="group relative">
                <a href="{{ route('studios.show', $studio->slug) }}" class="block">
                    <div class="relative aspect-[2/3] overflow-hidden rounded-2xl border border-white/10 shadow-2xl">
                        {{-- Image --}}
                        <img src="{{ $thumbnailUrl }}" alt="{{ $studio->name }}"
                            class="w-full h-full object-cover transition-transform duration-700 group-hover:scale-110">

                        {{-- Overlay --}}
                        <div
                            class="absolute inset-0 bg-gradient-to-t from-background via-transparent to-transparent opacity-60">
                        </div>

                        {{-- Badges --}}
                        <div class="absolute top-3 right-3 flex flex-col gap-2">
                            <span
                                class="bg-primary/90 backdrop-blur-md text-white text-[9px] font-black px-2 py-1 rounded-md uppercase tracking-wider shadow-lg">
                                {{ $studio->post_count }} Series
                            </span>
                        </div>

                        {{-- Gloss effect --}}
                        <div
                            class="absolute inset-x-0 top-0 h-1/2 bg-gradient-to-b from-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500">
                        </div>
                    </div>

                    {{-- Info --}}
                    <div class="mt-4 px-2">
                        <h3
                            class="text-sm font-bold text-white group-hover:text-primary transition-colors line-clamp-2 leading-snug">
                            {{ $studio->name }}
                        </h3>
                    </div>
                </a>
            </div>
        @empty
            <div class="col-span-full flex flex-col items-center justify-center py-20 opacity-40">
                <span class="material-symbols-outlined text-6xl mb-4">search_off</span>
                <p class="text-xl font-bold">No studios found</p>
            </div>
        @endforelse
    </div>

    {{-- Infinite Scroll Trigger --}}
    @if($studios->hasMorePages())
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