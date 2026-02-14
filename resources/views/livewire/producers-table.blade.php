<div wire:init="loadData" x-data="{}" class="flex flex-col gap-8">
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
                    <input wire:model.live.debounce.500ms="search" type="text" id="search"
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:outline-none focus:border-primary/50 transition-all placeholder:text-white/20"
                        placeholder="Search producer...">
                </div>
            </div>

            {{-- Sort --}}
            <div class="relative group">
                <label for="sort"
                    class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-focus-within:text-primary transition-colors">
                    Sort By
                </label>
                <div class="relative">
                    <select wire:model.live="sort" id="sort"
                        class="w-full bg-surface-darker border border-white/10 rounded-lg py-2.5 pl-4 pr-10 text-sm text-white focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer hover:bg-surface-darker/80">
                        <option value="name_asc">Name (A-Z)</option>
                        <option value="name_desc">Name (Z-A)</option>
                        <option value="most_series">Most Series</option>
                        <option value="least_series">Least Series</option>
                    </select>
                    <span
                        class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-lg group-focus-within:text-primary transition-colors">expand_more</span>
                </div>
            </div>
        </div>
    </div>

    {{-- Grid Section --}}
    <div class="min-h-[400px]">
        @if ($readyToLoad)
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                @foreach ($producers as $producer)
                    @php
                        $featuredPost = $producer->posts->first();
                    @endphp
                    <a wire:key="prod-{{ $producer->id }}" href="{{ route('producers.show', $producer->id) }}"
                        class="group relative overflow-hidden rounded-xl bg-slate-800 aspect-[16/10] border border-transparent hover:border-primary/50 transition-all cursor-pointer shadow-lg shadow-black/20">
                        <div class="absolute inset-0 bg-cover bg-center transition-transform duration-500 group-hover:scale-105"
                            style="background-image: url('{{ Storage::url($featuredPost->banner) }}');filter: brightness(0.5);">
                        </div>
                        <div
                            class="absolute inset-0 bg-gradient-to-t from-[#191022]/95 via-[#191022]/40 to-transparent">
                        </div>
                        <div class="absolute bottom-0 left-0 right-0 p-6 flex flex-col gap-1">
                            <div class="flex justify-between items-end">
                                <div>
                                    <h3
                                        class="text-2xl font-bold text-white group-hover:text-primary transition-colors">
                                        {{ $producer->name }}</h3>
                                </div>
                            </div>
                            <div class="mt-4 flex items-center justify-between border-t border-white/10 pt-4">
                                <div class="flex flex-col">
                                    <span
                                        class="text-[10px] text-slate-400 uppercase font-bold tracking-wider">Produced</span>
                                    <span class="text-white text-sm font-semibold">{{ $producer->posts_count }}
                                        Series</span>
                                </div>
                                <div class="flex flex-col items-end">
                                    <span
                                        class="text-[10px] text-slate-400 uppercase font-bold tracking-wider">Type</span>
                                    <span class="text-white text-sm font-semibold">Producer</span>
                                </div>
                            </div>
                        </div>
                    </a>
                @endforeach
            </div>

            {{-- Empty State --}}
            @if ($producers->isEmpty())
                <div class="flex flex-col items-center justify-center py-20 opacity-40">
                    <span class="material-symbols-outlined text-6xl mb-4">search_off</span>
                    <p class="text-xl font-bold">No producers found</p>
                </div>
            @endif
        @else
            @include('livewire.skeletons.rectangle-skeleton')
        @endif
    </div>

    {{-- Infinite Scroll Trigger --}}
    @if ($readyToLoad && $hasMorePages)
        <div wire:intersect="loadMore" wire:key="intersect-producers" class="flex justify-center py-12">
            <div class="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
        </div>
    @endif
</div>
