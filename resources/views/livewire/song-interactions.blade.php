<div>
    @if ($mode === 'bar')
        <div class="flex items-center gap-8">
            <button wire:click="toggleLike"
                class="flex flex-col items-center gap-1.5 group transition-all {{ $song->liked() ? 'text-primary' : 'text-white/40' }}">
                <span
                    class="material-symbols-outlined {{ $song->liked() ? 'filled' : '' }} group-hover:text-primary transition-colors">thumb_up</span>
                <span class="text-[10px] font-bold uppercase tracking-tighter">Like</span>
            </button>
            <button wire:click="toggleDislike"
                class="flex flex-col items-center gap-1.5 group transition-all {{ $song->disliked() ? 'text-red-500' : 'text-white/40' }}">
                <span
                    class="material-symbols-outlined {{ $song->disliked() ? 'filled' : '' }} group-hover:text-red-500 transition-colors">thumb_down</span>
                <span class="text-[10px] font-bold uppercase tracking-tighter">Dislike</span>
            </button>
            <button wire:click="toggleFavorite"
                class="flex flex-col items-center gap-1.5 group transition-all {{ $song->isFavorited() ? 'text-primary' : 'text-white/40' }}">
                <span
                    class="material-symbols-outlined {{ $song->isFavorited() ? 'filled' : '' }} hover:text-primary transition-colors">favorite</span>
                <span class="text-[10px] font-bold uppercase tracking-tighter">Favorite</span>
            </button>
        </div>
    @elseif($mode === 'score')
        <button wire:click="openRatingModal" class="flex items-center gap-1.5 group">
            <div class="flex flex-col items-end">
                <div class="flex items-center gap-1.5 text-yellow-400 font-black text-2xl">
                    <span
                        class="material-symbols-outlined filled group-hover:scale-110 transition-transform">star</span>
                    <span id="current-rating">{{ $song->formattedScore ?? 'N/A' }}</span>
                </div>
            </div>
        </button>
    @endif

    {{-- Rating Modal (Rendered only once per component instance) --}}
    <div x-data="{ open: @entangle('showRatingModal') }" x-show="open" style="display: none;"
        class="fixed inset-0 z-[200] flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
        @keydown.escape.window="open = false" x-transition.opacity>
        <div class="bg-[#1e1e2e] w-full max-w-sm rounded-3xl border border-white/10 shadow-2xl overflow-hidden"
            @click.away="open = false">
            <div class="px-6 py-5 border-b border-white/5 text-center bg-[#181825]">
                <h3 class="text-xl font-bold text-white">Rate this Song</h3>
                <p class="text-white/40 text-xs mt-1">{{ $song->name }}</p>
            </div>
            <div class="p-8">
                @auth
                    @switch(Auth::user()->score_format)
                        @case('POINT_100')
                            @include('partials.songs.show.rating.formats.point_100')
                        @break

                        @case('POINT_10_DECIMAL')
                            @include('partials.songs.show.rating.formats.point_10_decimal')
                        @break

                        @case('POINT_10')
                            @include('partials.songs.show.rating.formats.point_10')
                        @break

                        @case('POINT_5')
                            @include('partials.songs.show.rating.formats.point_5')
                        @break

                        @default
                            @include('partials.songs.show.rating.formats.point_10_decimal')
                    @endswitch
                @endauth
                @guest
                    <div class="text-center py-4">
                        <p class="text-white/60 mb-4">Please log in to rate this song.</p>
                        <a href="{{ route('login') }}"
                            class="inline-block bg-primary hover:bg-primary/90 text-white px-8 py-3 rounded-xl font-bold transition-all">
                            Log In
                        </a>
                    </div>
                @endguest
            </div>
            @auth
                <div class="bg-surface-darker p-6 border-t border-white/5">
                    <div class="grid grid-cols-2 gap-4">
                        <button @click="open = false"
                            class="bg-white/5 hover:bg-white/10 text-white/70 hover:text-white py-3.5 rounded-2xl font-bold text-sm transition-all border border-white/5">
                            Cancel
                        </button>
                        <button wire:click="rate"
                            class="bg-primary hover:bg-primary/80 text-white py-3.5 rounded-2xl font-bold text-sm transition-all shadow-xl shadow-primary/30 flex items-center justify-center gap-2 active:scale-[0.98]">
                            Submit Rating
                            <span class="material-symbols-outlined text-[18px]">check</span>
                        </button>
                    </div>
                </div>
            @else
                <div class="bg-surface-darker p-6 text-center border-t border-white/5">
                    <button @click="open = false"
                        class="text-white/40 hover:text-white text-sm font-bold transition-colors">Cancel</button>
                </div>
            @endauth
        </div>
    </div>
</div>
