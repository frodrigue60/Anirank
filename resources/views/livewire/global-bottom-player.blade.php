<div id="global-bottom-player" x-data="{
    visible: @entangle('isVisible'),
    playing: @entangle('isPlaying'),
    progress: 0,
    currentTime: '0:00',
    duration: '0:00',
    songTitle: '',
    animeTitle: '',
    artistNames: '',
    thumbnailImg: '',
    hasVideoLoaded: @entangle('hasVideoLoaded')
}" x-show="visible"
    x-transition:enter="transition ease-out duration-300" x-transition:enter-start="translate-y-full"
    x-transition:enter-end="translate-y-0"
    @song-loaded.window="
        console.log('GlobalBottomPlayer: Song loaded event received', $event.detail);
        songTitle = $event.detail.title;
        animeTitle = $event.detail.anime;
        artistNames = $event.detail.artists || 'Various Artists';
        thumbnailImg = $event.detail.thumbnail || 'https://lh3.googleusercontent.com/aida-public/65b3671a742c8d20fd4f8e568ba7a6f23fcf6ca0e872e41132644917a86f9f52';
    "
    class="fixed bottom-0 left-0 w-full z-[100] bg-background-dark/80 backdrop-blur-2xl border-t border-primary/30 h-24 flex flex-col shadow-[0_-10px_40px_rgba(0,0,0,0.5)]">

    {{-- Neon Glow Top Border --}}
    <div class="absolute top-0 left-0 w-full h-[2px] bg-primary shadow-[0_0_15px_#7f13ec]"></div>

    <style>
        #mini-player-container video,
        #mini-player-container iframe {
            width: 100% !important;
            height: 100% !important;
            object-fit: contain !important;
            background: #000;
            border: none !important;
        }

        .volume-slider {
            -webkit-appearance: none;
            width: 100%;
            height: 4px;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 10px;
            outline: none;
            overflow: hidden;
        }

        .volume-slider::-webkit-slider-thumb {
            -webkit-appearance: none;
            width: 12px;
            height: 12px;
            background: #fff;
            border-radius: 50%;
            cursor: pointer;
            box-shadow: -407px 0 0 400px #7f13ec;
            transition: all 0.2s ease-in-out;
            border: 2px solid #fff;
        }

        .volume-slider:hover::-webkit-slider-thumb {
            background: #fff;
            transform: scale(1.2);
            box-shadow: -407px 0 0 400px #9733ff;
        }
    </style>

    <div class="flex-1 flex items-center px-6 md:px-12">
        {{-- Left: Song Info --}}
        <div class="w-1/3 md:w-1/4 flex items-center gap-4">
            <div class="w-32 aspect-video rounded-xl bg-surface-darker/50 border border-white/10 overflow-hidden shrink-0 shadow-2xl relative group"
                wire:ignore>
                {{-- Mini Player Container --}}
                <div id="mini-player-container" class="absolute inset-0 w-full h-full z-1"></div>

                {{-- Placeholder when no video is loaded --}}
                <div id="mini-player-placeholder"
                    class="absolute inset-0 flex flex-col items-center justify-center bg-surface-darker/80 backdrop-blur-sm z-2">
                    <span class="material-symbols-outlined text-white/20 text-4xl">movie</span>
                </div>
            </div>
            <div class="min-w-0">
                <div class="text-white font-black text-base truncate leading-tight mb-1"
                    x-text="songTitle || 'Loading...'"></div>
                <div class="flex flex-col">
                    <div class="text-[10px] font-black uppercase tracking-[0.2em] text-primary truncate"
                        x-text="animeTitle"></div>
                    <div class="text-[10px] font-bold text-white/40 truncate" x-text="artistNames || 'Various Artists'">
                    </div>
                </div>
            </div>
        </div>

        {{-- Center: Controls & Progress --}}
        <div class="flex-1 flex flex-col items-center gap-2 px-8">
            <div class="flex items-center gap-6">
                <button class="text-white/40 hover:text-white transition-colors">
                    <span class="material-symbols-outlined text-2xl">skip_previous</span>
                </button>
                <button @click="$wire.togglePlay()"
                    class="w-10 h-10 rounded-full bg-white text-black flex items-center justify-center hover:bg-primary hover:text-white transition-all transform active:scale-90 shadow-xl shadow-white/5">
                    <span class="material-symbols-outlined text-3xl filled"
                        x-text="playing ? 'pause' : 'play_arrow'"></span>
                </button>
                <button class="text-white/40 hover:text-white transition-colors">
                    <span class="material-symbols-outlined text-2xl">skip_next</span>
                </button>
            </div>

            <div class="w-full flex items-center gap-4 group">
                <div class="text-[10px] font-mono text-white/40 w-10 text-right tabular-nums" x-text="currentTime">0:00
                </div>
                <div @click="const rect = $el.getBoundingClientRect(); const percent = (($event.clientX - rect.left) / rect.width) * 100; seekTo(percent);"
                    class="flex-1 h-1.5 bg-white/5 rounded-full relative overflow-hidden cursor-pointer group-hover:h-2 transition-all shadow-inner">
                    <div class="absolute left-0 top-0 h-full bg-primary shadow-[0_0_15px_rgba(127,19,236,0.8)] transition-all ease-linear"
                        :style="'width: ' + progress + '%'"></div>
                </div>
                <div class="text-[10px] font-mono text-white/40 w-10 tabular-nums" x-text="duration">0:00</div>
            </div>
        </div>

        {{-- Right: Rating & Actions --}}
        <div class="w-1/3 md:w-1/4 flex items-center justify-end gap-6">

            <div class="flex items-center gap-4">
                <div class="hidden lg:flex items-center gap-2 group/volume relative">
                    <button class="text-white/40 hover:text-white transition-colors">
                        <span class="material-symbols-outlined text-xl">volume_up</span>
                    </button>
                    <input type="range" min="0" max="100" value="50"
                        @input="setVolume($event.target.value)" class="volume-slider w-20">
                </div>
                <button @click="window.toggleFullscreen()" class="text-white/40 hover:text-primary transition-colors">
                    <span class="material-symbols-outlined text-2xl">fullscreen</span>
                </button>

            </div>
        </div>
    </div>

    {{-- Hidden Player Engine (Unused, moved to thumbnail) --}}
    <div id="hidden-player-container" class="hidden"></div>

    <script>
        document.addEventListener('livewire:load', function() {
            console.log('GlobalBottomPlayer: Component initialized');

            // Global helper to trigger playback from anywhere
            window.playSongGlobal = function(songId) {
                console.log('GlobalBottomPlayer: playSongGlobal called for ID', songId);
                window.Livewire.emit('playSong', songId);
            };

            Livewire.on('playSong', (songId) => {
                console.log('GlobalBottomPlayer: Received playSong via Livewire JS', songId);
            });

            let player = null;
            let ytPlayer = null;
            let ytInterval = null;
            let currentType = null;

            // Load YouTube API once
            if (!window.YT) {
                const tag = document.createElement('script');
                tag.src = "https://www.youtube.com/iframe_api";
                const firstScriptTag = document.getElementsByTagName('script')[0];
                firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);
            }

            window.addEventListener('song-loaded', event => {
                const data = event.detail;
                console.log('GlobalBottomPlayer: Song loaded event received', data);
                const container = document.getElementById('mini-player-container');
                container.innerHTML = '';

                // Hide placeholder
                const placeholder = document.getElementById('mini-player-placeholder');
                if (placeholder) placeholder.style.display = 'none';

                // Reset state
                if (ytInterval) clearInterval(ytInterval);
                if (ytPlayer) {
                    try {
                        ytPlayer.destroy();
                    } catch (e) {}
                    ytPlayer = null;
                }
                player = null;
                currentType = data.type;

                const playerEl = document.getElementById('global-bottom-player');
                if (!playerEl) return;
                const alpine = window.Alpine ? Alpine.$data(playerEl) : null;
                if (!alpine) return;

                if (data.type === 'embed') {
                    const ytId = extractYTId(data.url);
                    const ytDiv = document.createElement('div');
                    ytDiv.id = 'yt-player-instance';
                    container.appendChild(ytDiv);

                    const createYT = () => {
                        ytPlayer = new YT.Player('yt-player-instance', {
                            videoId: ytId,
                            playerVars: {
                                autoplay: 1,
                                controls: 0,
                                modestbranding: 1,
                                rel: 0
                            },
                            events: {
                                onReady: (e) => {
                                    alpine.duration = formatTime(e.target.getDuration());
                                    e.target.setVolume(50);
                                    startYTTracking();
                                },
                                onStateChange: (e) => {
                                    if (e.data === YT.PlayerState.ENDED) {
                                        alpine.playing = false;
                                        clearInterval(ytInterval);
                                    }
                                    if (e.data === YT.PlayerState.PLAYING) alpine.playing =
                                        true;
                                    if (e.data === YT.PlayerState.PAUSED) alpine.playing =
                                        false;
                                }
                            }
                        });
                    };

                    if (window.YT && YT.Player) createYT();
                    else window.onYouTubeIframeAPIReady = createYT;

                } else {
                    const video = document.createElement('video');
                    video.src = data.url;
                    video.autoplay = true;
                    video.controls = false;
                    video.volume = 0.5;
                    container.appendChild(video);
                    player = video;

                    video.addEventListener('loadedmetadata', () => {
                        alpine.duration = formatTime(video.duration);
                    });

                    video.addEventListener('timeupdate', () => {
                        alpine.progress = (video.currentTime / video.duration) * 100;
                        alpine.currentTime = formatTime(video.currentTime);
                    });

                    video.addEventListener('ended', () => {
                        alpine.playing = false;
                    });

                    video.onplay = () => {
                        console.log('GlobalBottomPlayer: Video started playing');
                        alpine.playing = true;
                    };

                    video.onpause = () => {
                        console.log('GlobalBottomPlayer: Video paused');
                        alpine.playing = false;
                    };
                }
            });

            window.addEventListener('toggle-playback', event => {
                const isPlaying = event.detail.playing;
                if (currentType === 'embed' && ytPlayer) {
                    isPlaying ? ytPlayer.playVideo() : ytPlayer.pauseVideo();
                } else if (player) {
                    isPlaying ? player.play() : player.pause();
                }
            });

            // Seek logic
            window.seekTo = function(percent) {
                const duration = getDuration();
                const time = (percent / 100) * duration;
                if (currentType === 'embed' && ytPlayer) ytPlayer.seekTo(time, true);
                else if (player) player.currentTime = time;
            };

            // Fullscreen logic
            window.toggleFullscreen = function() {
                const container = document.getElementById('mini-player-container');
                if (!container) return;

                // Find the actual element (video or iframe) to match native behavior
                const target = container.querySelector('video, iframe') || container;

                if (!document.fullscreenElement) {
                    if (target.requestFullscreen) {
                        target.requestFullscreen();
                    } else if (target.webkitRequestFullscreen) {
                        target.webkitRequestFullscreen();
                    } else if (target.msRequestFullscreen) {
                        target.msRequestFullscreen();
                    }
                } else {
                    if (document.exitFullscreen) {
                        document.exitFullscreen();
                    }
                }
            };

            // Listener to toggle controls based on fullscreen state
            const handleFullscreenChange = () => {
                const container = document.getElementById('mini-player-container');
                if (!container) return;
                const media = container.querySelector('video, iframe');
                if (!media) return;

                const isFullscreen = !!document.fullscreenElement;

                if (media.tagName === 'VIDEO') {
                    media.controls = isFullscreen;
                }
                // YouTube iframes (API) don't easily allow toggling controls post-init.
                // We'll leave them as they were configured (controls: 0).
            };

            document.addEventListener('fullscreenchange', handleFullscreenChange);
            document.addEventListener('webkitfullscreenchange', handleFullscreenChange);
            document.addEventListener('mozfullscreenchange', handleFullscreenChange);
            document.addEventListener('MSFullscreenChange', handleFullscreenChange);

            // Volume logic
            window.setVolume = function(val) {
                if (currentType === 'embed' && ytPlayer) ytPlayer.setVolume(val);
                else if (player) player.volume = val / 100;
            };

            function getDuration() {
                if (currentType === 'embed' && ytPlayer) return ytPlayer.getDuration();
                if (player) return player.duration;
                return 0;
            }

            function startYTTracking() {
                ytInterval = setInterval(() => {
                    if (ytPlayer && ytPlayer.getCurrentTime) {
                        const playerEl = document.getElementById('global-bottom-player');
                        if (!playerEl) return;
                        const alpine = window.Alpine ? Alpine.$data(playerEl) : null;
                        if (!alpine) return;
                        const current = ytPlayer.getCurrentTime();
                        const dur = ytPlayer.getDuration();
                        alpine.currentTime = formatTime(current);
                        if (dur > 0) alpine.progress = (current / dur) * 100;
                    }
                }, 1000);
            }

            function extractYTId(url) {
                const match = url.match(/(?:v=|youtu\.be\/|embed\/)([^"&?/ ]{11})/);
                return match ? match[1] : null;
            }

            function formatTime(s) {
                if (!s || isNaN(s)) return '0:00';
                const m = Math.floor(s / 60);
                const rs = Math.floor(s % 60);
                return `${m}:${rs.toString().padStart(2, '0')}`;
            }
        });
    </script>
</div>
