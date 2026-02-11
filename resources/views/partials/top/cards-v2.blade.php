@foreach ($items as $index => $song)
    @isset($song->post)
        @php
            $img_url = null;
            if ($song->post->thumbnail != null) {
                if (Storage::disk('public')->exists($song->post->thumbnail)) {
                    $img_url = Storage::url($song->post->thumbnail);
                }
            } else {
                $img_url =
                    'https://static.vecteezy.com/system/resources/thumbnails/005/170/408/small/banner-abstract-geometric-white-and-gray-color-background-illustration-free-vector.jpg';
            }

            $rankNumber = $items->firstItem() + $index;
            $formattedRank = str_pad($rankNumber, 2, '0', STR_PAD_LEFT);
        @endphp

        <div
            class="ranking-row grid grid-cols-[60px_1fr_120px_140px] gap-4 px-8 py-5 items-center transition-colors border-b border-white/5 hover:bg-white/5 group">
            {{-- Rank Column --}}
            <div class="flex flex-col items-center gap-1">
                <span
                    class="text-2xl font-black {{ $rankNumber <= 3 ? 'text-primary' : 'text-white/90' }}">{{ $formattedRank }}</span>
                <div class="flex items-center text-white/20">
                    <span class="material-symbols-outlined text-sm">horizontal_rule</span>
                </div>
            </div>

            {{-- Theme Info Column --}}
            <div class="flex items-center gap-6">
                <div class="w-16 h-16 rounded-lg overflow-hidden shrink-0 shadow-lg shadow-black/40 border border-white/10">
                    <img alt="Cover"
                        class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                        src="{{ $img_url }}" />
                </div>
                <div class="min-w-0">
                    <h3 class="text-lg font-bold text-white truncate leading-tight mb-1">{{ $song->name }}</h3>
                    <div class="flex items-center gap-2 text-sm">
                        <span class="text-primary font-bold truncate">{{ $song->post->title }}</span>
                        <span class="w-1 h-1 rounded-full bg-white/20 shrink-0"></span>
                        <span class="text-white/60 truncate">
                            @if (isset($song->artists) && count($song->artists) != 0)
                                {{ implode(', ', $song->artists->pluck('name')->toArray()) }}
                            @else
                                N/A
                            @endif
                        </span>
                    </div>
                </div>
            </div>

            {{-- Score Column --}}
            <div class="text-center">
                <div class="text-2xl font-black text-white tracking-tight">{{ round($song->averageRating, 2) }}</div>
                <div class="text-[10px] font-bold text-white/30 uppercase tracking-widest">Avg Rating</div>
            </div>

            {{-- Actions Column --}}
            <div class="flex items-center justify-end gap-2">
                <button
                    class="w-10 h-10 rounded-full flex items-center justify-center bg-white/5 hover:bg-primary text-white transition-all shadow-lg hover:shadow-primary/20">
                    <span class="material-symbols-outlined text-[20px] filled">play_arrow</span>
                </button>
                <button
                    class="w-10 h-10 rounded-full flex items-center justify-center bg-white/5 hover:bg-white/10 text-white/40 hover:text-red-400 transition-all">
                    <span
                        class="material-symbols-outlined text-[20px] {{ $song->userScore ? 'filled text-red-400' : '' }}">favorite</span>
                </button>
                <button
                    class="w-10 h-10 rounded-full flex items-center justify-center bg-white/5 hover:bg-white/10 text-white/40 transition-all">
                    <span class="material-symbols-outlined text-[20px]">more_vert</span>
                </button>
            </div>
        </div>
    @endisset
@endforeach
