<div class="max-w-[1440px] mx-auto px-4 md:px-8 py-8" x-data="{ playlistModalOpen: @entangle('showPlaylistModal'), ratingModalOpen: @entangle('showRatingModal') }">
    <div class="space-y-8">

        {{-- Video Player Section --}}
        <div
            class="relative w-full aspect-video rounded-3xl overflow-hidden bg-black shadow-2xl shadow-primary/20 group border border-white/5">
            @if ($currentVariant)
                <video id="player" class="w-full h-full object-contain" controls autoplay crossorigin playsinline>
                    <source src="{{ $this->getVideoUrl() }}" type="video/mp4">
                </video>
            @else
                <div class="absolute inset-0 flex items-center justify-center bg-surface-darker">
                    <div class="text-center">
                        <span class="material-symbols-outlined text-6xl text-white/20 mb-4">videocam_off</span>
                        <p class="text-white/40">No video available</p>
                    </div>
                </div>
            @endif
        </div>

        {{-- Title & Actions Section --}}
        <div class="flex flex-col md:flex-row items-start justify-between gap-6">
            <div>
                <div class="flex items-center gap-3 mb-2">
                    <span
                        class="bg-primary text-white text-[10px] font-bold px-2 py-0.5 rounded uppercase tracking-wider">
                        {{ $song->type }} {{ $song->theme_num ?? '' }}
                    </span>
                    @if ($song->songVariants->count() > 1)
                        <div class="flex bg-white/5 rounded-lg p-0.5 backdrop-blur-md">
                            @foreach ($song->songVariants->sortBy('version_number') as $variant)
                                <button wire:click="switchVariant({{ $variant->id }})"
                                    class="px-3 py-1 text-xs font-bold rounded-md transition-all {{ $currentVariant && $currentVariant->id === $variant->id ? 'bg-primary text-white shadow-lg' : 'text-white/60 hover:text-white hover:bg-white/10' }}">
                                    v{{ $variant->version_number }}
                                </button>
                            @endforeach
                        </div>
                    @else
                        <span class="text-xs font-bold text-white/40 uppercase tracking-widest">Version 1</span>
                    @endif
                </div>
                <h1 class="text-2xl md:text-4xl font-black text-white leading-tight mb-2">
                    {{ $song->name }}
                </h1>
                <div class="flex items-center gap-2 text-white/60 text-sm font-medium">
                    <span>by</span>
                    @foreach ($song->artists as $artist)
                        <a href="{{ route('artists.show', $artist->slug) }}"
                            class="text-primary hover:text-white transition-colors">
                            {{ $artist->name }}
                        </a>
                        @if (!$loop->last)
                            ,
                        @endif
                    @endforeach
                </div>
            </div>
        </div>

        {{-- Info & Stats Bar --}}
        <div
            class="bg-surface-dark rounded-3xl p-6 border border-white/5 shadow-xl flex flex-col md:flex-row items-center justify-between gap-6">
            <div class="flex items-center gap-6">
                {{-- Views --}}
                <div class="flex flex-col">
                    <span class="text-white/40 text-[10px] font-bold uppercase tracking-widest mb-1">Views</span>
                    <span class="text-xl font-black text-white">{{ $song->views }}</span>
                </div>
                {{-- Rating --}}
                <div class="flex flex-col gap-2 min-w-[160px]">
                    @php
                        $likes = $song->likes_count;
                        $dislikes = $song->dislikes_count;
                        $total = $likes + $dislikes;
                        $percent = $total > 0 ? ($likes / $total) * 100 : 50;
                    @endphp
                    <div class="flex items-center justify-between text-white font-black group/stats">
                        <div class="flex items-center gap-2">
                            <span class="material-symbols-outlined text-[20px] text-white">thumb_up</span>
                            <span class="text-sm tracking-tighter">{{ number_format($likes) }}</span>
                        </div>
                        <div
                            class="flex items-center gap-2 text-white/40 group-hover/stats:text-white/60 transition-colors">
                            <span class="material-symbols-outlined text-[20px]">thumb_down</span>
                            <span class="text-sm tracking-tighter">{{ number_format($dislikes) }}</span>
                        </div>
                    </div>
                    <div class="h-1 w-full bg-white/10 rounded-full overflow-hidden">
                        <div class="h-full bg-white transition-all duration-700 ease-out"
                            style="width: {{ $percent }}%"></div>
                    </div>
                </div>
                <div class="h-8 w-[1px] bg-white/10"></div>
                {{-- Rating Trigger --}}
                <button wire:click="openRatingModal" class="flex flex-col text-left group">
                    <span
                        class="text-white/40 text-[10px] font-bold uppercase tracking-widest mb-1 group-hover:text-primary transition-colors">Detailed
                        Score</span>
                    <div class="flex items-center gap-2">
                        <span class="material-symbols-outlined text-yellow-400 text-[20px] filled">star</span>
                        <span
                            class="text-yellow-400 font-black text-xl">{{ number_format($song->averageRating, 1) ?? 'N/A' }}</span>
                    </div>
                </button>
            </div>
            {{-- Quick Actions --}}
            <div class="flex items-center gap-3">
                <button wire:click="toggleLike"
                    class="group flex items-center gap-2 px-5 py-2.5 rounded-2xl border border-white/10 transition-all {{ $song->liked() ? 'bg-primary/20 border-primary/50 text-primary shadow-lg shadow-primary/10' : 'bg-surface-dark hover:bg-surface-darker text-white/60 hover:text-white' }}">
                    <span
                        class="material-symbols-outlined {{ $song->liked() ? 'filled' : '' }} text-[20px]">thumb_up</span>
                </button>

                <button wire:click="toggleDislike"
                    class="group flex items-center gap-2 px-5 py-2.5 rounded-2xl border border-white/10 transition-all {{ $song->disliked() ? 'bg-primary/20 border-primary/50 text-primary shadow-lg shadow-primary/10' : 'bg-surface-dark hover:bg-surface-darker text-white/60 hover:text-white' }}">
                    <span
                        class="material-symbols-outlined {{ $song->disliked() ? 'filled' : '' }} text-[20px]">thumb_down</span>
                </button>

                <button wire:click="toggleFavorite"
                    class="group px-4 py-2.5 flex items-center justify-center rounded-2xl border border-white/10 transition-all {{ $song->isFavorited() ? 'bg-red-500/20 border-red-500/50 text-red-500 shadow-lg shadow-red-500/10' : 'bg-surface-dark hover:bg-surface-darker text-white/60 hover:text-red-400' }}">
                    <span
                        class="material-symbols-outlined {{ $song->isFavorited() ? 'filled' : '' }} text-[22px]">favorite</span>
                </button>
                <button wire:click="openPlaylistModal"
                    class="flex items-center gap-2 px-5 py-2.5 bg-primary/10 hover:bg-primary/20 text-primary border border-primary/20 rounded-xl transition-all text-sm font-bold shadow-lg shadow-primary/5">
                    <span class="material-symbols-outlined text-[18px]">playlist_add</span> Add to Playlist
                </button>
            </div>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
            {{-- Left Column: Comments --}}
            <div class="lg:col-span-7 space-y-8">
                {{-- Join Discussion --}}
                <div class="space-y-4">
                    <h2 class="text-2xl font-bold flex items-center gap-3 text-white">
                        <span class="material-symbols-outlined text-primary">forum</span>
                        Join the Discussion
                    </h2>

                    @auth
                        <div class="bg-surface-darker rounded-3xl p-1 border border-white/5 shadow-inner">
                            <textarea wire:model.defer="commentBody"
                                class="w-full bg-transparent border-none rounded-2xl p-4 text-sm text-white placeholder:text-white/20 min-h-[100px] resize-none focus:ring-0"
                                placeholder="What do you think about this song?"></textarea>
                            <div class="flex justify-between items-center px-2 pb-2">
                                <div class="flex gap-1"></div>
                                <button wire:click="postComment"
                                    class="bg-primary hover:bg-primary/80 text-white px-6 py-2 rounded-xl font-bold text-sm transition-all shadow-lg shadow-primary/20">
                                    Post Comment
                                </button>
                            </div>
                        </div>
                        @error('commentBody')
                            <span class="text-red-400 text-xs ml-2">{{ $message }}</span>
                        @enderror
                    @else
                        <div class="bg-surface-darker rounded-2xl p-8 border border-white/5 text-center">
                            <p class="text-white/60 mb-4">Please log in to share your thoughts.</p>
                            <a href="{{ route('login') }}"
                                class="inline-block bg-primary hover:bg-primary/90 text-white px-8 py-3 rounded-xl font-bold transition-all">
                                Log In
                            </a>
                        </div>
                    @endauth
                </div>

                {{-- Comments List --}}
                <div class="space-y-6">
                    <div class="flex items-center justify-between">
                        <span class="text-sm font-bold text-white/60">{{ $comments->count() }} Main Comments</span>
                    </div>

                    @forelse($comments as $comment)
                        @include('partials.comments.item', ['comment' => $comment, 'isReply' => false])
                    @empty
                        <div class="text-center py-12 bg-surface-dark/20 rounded-3xl border border-white/5">
                            <span
                                class="material-symbols-outlined text-4xl text-white/10 mb-2">chat_bubble_outline</span>
                            <p class="text-white/40 text-sm">No comments yet. Be the first!</p>
                        </div>
                    @endforelse
                </div>
            </div>

            {{-- Right Column: Related --}}
            <div class="lg:col-span-5 space-y-6">
                <h3 class="text-xl font-bold text-white mb-4">More from {{ $post->title }}</h3>
                <div class="grid grid-cols-1 gap-4">
                    @foreach ($relatedSongs as $related)
                        <a href="{{ $related->url }}"
                            class="flex gap-4 p-3 rounded-2xl bg-surface-dark hover:bg-surface-darker border border-white/5 hover:border-primary/30 transition-all group">
                            <div class="w-24 aspect-video rounded-lg overflow-hidden relative shrink-0">
                                @php
                                    $thumb = $related->post->thumbnail;
                                    if (Storage::disk('public')->exists($thumb)) {
                                        $thumbUrl = Storage::url($thumb);
                                    } else {
                                        $thumbUrl = $related->post->thumbnail_src;
                                    }
                                @endphp
                                <img src="{{ $thumbUrl }}"
                                    class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                                    alt="{{ $related->name }}">
                                <div class="absolute inset-0 bg-black/20 group-hover:bg-transparent transition-colors">
                                </div>
                            </div>
                            <div class="flex-1 flex flex-col justify-center">
                                <div class="flex items-center gap-2 mb-1">
                                    <span
                                        class="px-1.5 py-0.5 rounded text-[9px] font-bold uppercase {{ $related->type == 'OP' ? 'bg-primary/20 text-primary' : 'bg-blue-500/20 text-blue-400' }}">
                                        {{ $related->type }} {{ $related->theme_num }}
                                    </span>
                                </div>
                                <h4
                                    class="font-bold text-sm text-white group-hover:text-primary transition-colors line-clamp-1__">
                                    {{ $related->name }}
                                </h4>
                                <p class="text-xs text-white/40 line-clamp-1">
                                    {{ $related->artists->pluck('name')->join(', ') }}
                                </p>
                            </div>
                        </a>
                    @endforeach
                </div>
            </div>
        </div>

    </div>

    {{-- Playlist Modal --}}
    <div x-show="playlistModalOpen" style="display: none;"
        class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
        x-transition.opacity>
        <div class="bg-[#1e1e2e] w-full max-w-md rounded-3xl border border-white/10 shadow-2xl overflow-hidden"
            @click.away="playlistModalOpen = false">
            <div class="px-6 py-5 border-b border-white/5 flex justify-between items-center bg-[#181825]">
                <div>
                    <h3 class="text-xl font-bold text-white leading-tight">Add to Playlist</h3>
                    <p class="text-white/40 text-xs mt-1">{{ $song->name }}</p>
                </div>
                <button @click="playlistModalOpen = false" class="text-white/40 hover:text-white transition-colors">
                    <span class="material-symbols-outlined">close</span>
                </button>
            </div>
            <div class="p-6 space-y-4 max-h-[60vh] overflow-y-auto custom-scrollbar">
                @if (count($userPlaylists) > 0)
                    <div class="space-y-2">
                        @foreach ($userPlaylists as $playlist)
                            <button wire:click="togglePlaylist({{ $playlist->id }})"
                                class="w-full flex items-center justify-between p-3 rounded-xl hover:bg-white/5 transition-colors group text-left">
                                <span
                                    class="font-medium text-white {{ $playlist->songs_count > 0 ? 'text-primary' : '' }}">{{ $playlist->name }}</span>
                                @if ($playlist->songs_count > 0)
                                    <span
                                        class="material-symbols-outlined text-primary text-[20px]">check_circle</span>
                                @else
                                    <span
                                        class="material-symbols-outlined text-white/10 group-hover:text-white/40 text-[20px]">radio_button_unchecked</span>
                                @endif
                            </button>
                        @endforeach
                    </div>
                @else
                    <p class="text-white/40 text-center text-sm py-4">No playlists found.</p>
                @endif

                <div class="pt-4 border-t border-white/5">
                    <label class="text-xs font-bold text-white/40 uppercase tracking-widest mb-2 block">Create
                        New</label>
                    <div class="flex gap-2">
                        <input wire:model.defer="newPlaylistName" type="text"
                            class="flex-1 bg-surface-darker border border-white/10 rounded-xl px-4 py-2 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary placeholder:text-white/20"
                            placeholder="My Awesome Playlist">
                        <button wire:click="createPlaylist"
                            class="bg-primary hover:bg-primary/90 text-white p-2 rounded-xl transition-colors">
                            <span class="material-symbols-outlined">add</span>
                        </button>
                    </div>
                </div>
            </div>
            <div class="bg-surface-darker p-4 text-center border-t border-white/5">
                <button @click="playlistModalOpen = false"
                    class="text-white/40 hover:text-white text-sm font-bold transition-colors">Close</button>
            </div>
        </div>
    </div>

    {{-- Rating Modal --}}
    <div x-show="ratingModalOpen" style="display: none;"
        class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
        x-transition.opacity>
        <div class="bg-[#1e1e2e] w-full max-w-sm rounded-3xl border border-white/10 shadow-2xl overflow-hidden"
            @click.away="ratingModalOpen = false">
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
                        <button @click="ratingModalOpen = false"
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
                    <button @click="ratingModalOpen = false"
                        class="text-white/40 hover:text-white text-sm font-bold transition-colors">Cancel</button>
                </div>
            @endauth
        </div>
    </div>

    <script>
        document.addEventListener('livewire:load', function() {
            const player = new Plyr('#player', {
                controls: ['play-large', 'play', 'progress', 'current-time', 'mute', 'volume',
                    'fullscreen'
                ],
            });

            window.addEventListener('video-changed', event => {
                const video = document.getElementById('player');
                const source = video.querySelector('source');
                source.setAttribute('src', event.detail.src);
                video.load();
                // Plyr update handled by re-init or source setter if consistent
                player.source = {
                    type: 'video',
                    sources: [{
                        src: event.detail.src,
                        type: 'video/mp4',
                    }, ],
                };
                player.play();
            });
        });
    </script>
</div>
