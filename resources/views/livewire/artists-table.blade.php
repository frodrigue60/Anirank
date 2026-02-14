<div wire:init="loadData" x-data="{}" class="max-w-[1440px] mx-auto px-6 py-12">
    <div class="mb-4">
        <h1 class="text-3xl font-black tracking-tight text-white mb-2">Search Artists</h1>
        <div class="h-1 w-20 bg-primary rounded-full"></div>
    </div>
    <div class="flex flex-col gap-10">
        {{-- Search & Filters Section --}}
        <section class="flex flex-col gap-6">
            <div class="flex flex-wrap items-center justify-center gap-4">
                {{-- Search --}}
                <div class="relative flex-1 min-w-[300px] group">
                    <label
                        class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-focus-within:text-primary transition-colors">Search</label>
                    <div class="relative">
                        <span
                            class="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-primary text-[22px] group-focus-within:scale-110 transition-transform">search</span>
                        <input wire:model.live.debounce.300ms="name"
                            class="w-full h-11 bg-surface-darker/50 border border-white/10 rounded-xl pl-12 pr-4 text-sm text-white focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-white/20 transition-all"
                            placeholder="Search for an artist..." type="text" />
                    </div>
                </div>

                <div class="relative min-w-[200px] group">
                    <label
                        class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Sort
                        Alphabetical</label>
                    <div class="relative">
                        <select wire:model.live="sortBy"
                            class="w-full bg-surface-dark border-white/10 rounded-xl text-sm text-white/80 focus:ring-primary focus:border-primary py-2.5 pl-4 pr-10 appearance-none cursor-pointer transition-all hover:bg-surface-dark/80 focus:bg-surface-darker">
                            <option value="A-Z">A - Z</option>
                            <option value="Z-A">Z - A</option>
                            <option value="most_themes">Most Themes</option>
                            <option value="least_themes">Least Themes</option>
                        </select>
                        <span
                            class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-xl group-focus-within:text-primary transition-colors">expand_more</span>
                    </div>
                </div>
            </div>
        </section>

        {{-- Artists Grid Section --}}
        <section class="mt-8">
            @if ($readyToLoad)
                <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-8">
                    @foreach ($artists as $artist)
                        <a wire:key="artist-{{ $artist->id }}" href="{{ route('artists.show', $artist->slug) }}"
                            class="group flex flex-col items-center gap-4 cursor-pointer">
                            <div
                                class="relative w-full aspect-square rounded-full overflow-hidden card-shadow ring-4 ring-transparent group-hover:ring-primary/50 transition-all duration-300">
                                @if ($artist->thumbnail)
                                    <img alt="{{ $artist->name }}"
                                        class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                                        src="{{ Storage::url($artist->thumbnail) }}" loading="lazy" />
                                @else
                                    <div
                                        class="w-full h-full bg-surface-darker flex flex-col items-center justify-center text-white/10">
                                        <span class="material-symbols-outlined text-5xl">person</span>
                                    </div>
                                @endif
                                <div
                                    class="absolute inset-0 bg-gradient-to-t from-primary/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity">
                                </div>
                            </div>
                            <div class="text-center">
                                <h3
                                    class="font-bold text-white group-hover:text-primary transition-colors text-lg truncate max-w-[150px]">
                                    {{ $artist->name }}
                                </h3>
                                <div
                                    class="mt-1.5 inline-flex items-center px-3 py-1 rounded-full bg-primary/20 text-primary text-[11px] font-bold border border-primary/20">
                                    {{ $artist->songs_count }} Themes
                                </div>
                            </div>
                        </a>
                    @endforeach
                </div>

                {{-- Empty State --}}
                @if ($artists->isEmpty())
                    <div class="py-20 flex flex-col items-center justify-center text-center">
                        <div class="w-20 h-20 rounded-full bg-surface-dark flex items-center justify-center mb-6">
                            <span class="material-symbols-outlined text-5xl text-white/10">person_off</span>
                        </div>
                        <h3 class="text-xl font-bold text-white mb-2">No artists found</h3>
                        <p class="text-white/40 max-w-xs mx-auto mb-6">
                            We couldn't find any artists matching your search or filters.
                        </p>
                        <button wire:click="clearFilters"
                            class="bg-primary hover:bg-primary/90 text-white px-8 py-3 rounded-xl font-bold transition-all shadow-lg shadow-primary/20">
                            Clear Search
                        </button>
                    </div>
                @endif
            @else
                @include('livewire.skeletons.circle-skeleton')
            @endif

            @if ($hasMorePages && $readyToLoad)
                <div wire:intersect="loadMore" wire:key="intersect-artists"
                    class="py-12 flex flex-col items-center gap-4">
                    <div class="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                </div>
            @elseif($readyToLoad)
                <div class="py-12 text-center text-white/20 text-sm font-medium">
                    Showing all artists
                </div>
            @endif
        </section>
    </div>
</div>
