@extends('layouts.app')

@section('meta')
    <title>{{ $playlist->name }} | Playlists</title>
    <meta name="description" content="Now playing: {{ $playlist->name }}. Enjoy your custom anime music collection.">
@endsection

@section('content')
    <div class="min-h-screen bg-background">
        {{-- Video Section (Cinema Mode) --}}
        <div class="relative w-full bg-black aspect-video lg:aspect-auto lg:h-[70vh] overflow-hidden shadow-2xl">
            <div id="player-container" class="w-full h-full flex items-center justify-center bg-surface-darker/20">
                <div class="flex flex-col items-center gap-4 opacity-20">
                    <div class="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                    <p class="text-sm font-black uppercase tracking-widest text-white">Initializing Player...</p>
                </div>
            </div>

            {{-- Subtle overlay for top-left info if needed, but keeping it clean --}}
            <div class="absolute top-6 left-6 pointer-events-none">
                <div class="flex items-center gap-3">
                    <span
                        class="bg-primary text-white text-[10px] font-black px-2 py-0.5 rounded uppercase tracking-wider shadow-lg">Live</span>
                    <span class="text-white/40 text-[10px] font-black uppercase tracking-widest drop-shadow-md">Cinema
                        Mode</span>
                </div>
            </div>
        </div>

        {{-- Main Control Bar (Outside Video) --}}
        <div class="sticky top-0 z-50 bg-surface-dark/80 backdrop-blur-2xl border-b border-white/5 shadow-2xl">
            {{-- Seek Bar --}}
            <div class="absolute top-0 left-0 right-0 h-1 bg-white/5 group hover:h-2 transition-all cursor-pointer">
                <input type="range" id="seek-slider" min="0" max="100" value="0"
                    class="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10">
                <div id="seek-progress"
                    class="absolute top-0 left-0 h-full bg-primary transition-all duration-100 ease-linear"
                    style="width: 0%"></div>
            </div>

            <div
                class="max-w-[1440px] mx-auto px-4 md:px-8 py-4 flex flex-col md:flex-row items-center justify-between gap-6">
                {{-- Metadata --}}
                <div class="flex items-center gap-4 flex-1 min-w-0">
                    <div id="current-thumbnail"
                        class="w-12 h-12 rounded-lg bg-surface-darker border border-white/10 overflow-hidden shrink-0 hidden md:block">
                        <img src="" class="w-full h-full object-cover opacity-0 transition-opacity duration-300">
                    </div>
                    <div class="flex flex-col min-w-0">
                        <h1 id="current-song-title" class="text-lg font-black text-white truncate drop-shadow-sm">Loading
                            theme...</h1>
                        <p id="playlist-title"
                            class="text-[10px] uppercase font-black tracking-widest text-primary truncate">
                            {{ $playlist->name }}
                        </p>
                    </div>
                </div>

                {{-- Playback Controls --}}
                <div class="flex items-center gap-6">
                    <button id="prev-btn"
                        class="w-10 h-10 rounded-xl hover:bg-white/5 text-white/60 hover:text-white transition-all flex items-center justify-center">
                        <span class="material-symbols-outlined text-[28px]">skip_previous</span>
                    </button>

                    <button id="play-pause-btn"
                        class="w-12 h-12 rounded-2xl bg-white text-black hover:bg-primary hover:text-white transition-all flex items-center justify-center shadow-xl hover:shadow-primary/20 group">
                        <span
                            class="material-symbols-outlined text-[32px] filled group-active:scale-90 transition-transform">play_arrow</span>
                    </button>

                    <button id="next-btn"
                        class="w-10 h-10 rounded-xl hover:bg-white/5 text-white/60 hover:text-white transition-all flex items-center justify-center">
                        <span class="material-symbols-outlined text-[28px]">skip_next</span>
                    </button>
                </div>

                {{-- Volume & Time --}}
                <div class="flex items-center gap-6 flex-1 justify-end">
                    {{-- Volume --}}
                    <div class="hidden lg:flex items-center gap-3 group">
                        <button id="mute-btn" class="text-white/40 hover:text-white transition-colors">
                            <span class="material-symbols-outlined text-[20px]">volume_up</span>
                        </button>
                        <div
                            class="w-24 h-1 bg-white/10 rounded-full relative overflow-hidden group-hover:h-1.5 transition-all">
                            <input type="range" id="volume-slider" min="0" max="100" value="100"
                                class="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10">
                            <div id="volume-progress"
                                class="absolute top-0 left-0 h-full bg-white/60 group-hover:bg-primary transition-all"
                                style="width: 100%"></div>
                        </div>
                    </div>

                    <div
                        class="flex items-center gap-3 px-4 py-2 bg-white/5 rounded-xl border border-white/5 font-black text-[11px] tracking-widest text-white/80 tabular-nums min-w-[110px] justify-center">
                        <span id="current-time">0:00</span>
                        <span class="text-white/20">/</span>
                        <span id="duration">0:00</span>
                    </div>

                    <button id="fullscreen-btn" class="text-white/40 hover:text-white transition-colors ml-2"
                        title="Toggle Fullscreen">
                        <span class="material-symbols-outlined text-[24px]">fullscreen</span>
                    </button>
                </div>
            </div>
        </div>

        {{-- Content Grid --}}
        <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10 md:py-12 grid grid-cols-1 lg:grid-cols-12 gap-10">
            {{-- Queue Section --}}
            <div class="lg:col-span-8 flex flex-col gap-6">
                <div class="flex items-center justify-between border-b border-white/5 pb-4">
                    <div class="flex items-center gap-3">
                        <span class="material-symbols-outlined text-primary">data_usage</span>
                        <h3 class="text-xl font-black text-white uppercase tracking-tight">Song Queue</h3>
                    </div>
                </div>

                <div id="queue-list" class="flex flex-col gap-3">
                    {{-- Populated by JS --}}
                </div>
            </div>

            {{-- Sidebar / Playlist Info --}}
            <div class="lg:col-span-4 flex flex-col gap-8">
                <div
                    class="bg-surface-dark/30 p-6 rounded-2xl border border-white/5 backdrop-blur-sm flex flex-col gap-6 sticky top-32">
                    <div class="flex flex-col gap-2">
                        <h4 class="text-[10px] uppercase font-black text-white/40 tracking-widest">About this playlist</h4>
                        <p class="text-sm text-white/80 leading-relaxed">
                            {{ $playlist->description ?? 'No description provided for this collection.' }}
                        </p>
                    </div>

                    <div class="flex flex-col gap-4">
                        <div class="flex items-center justify-between py-3 border-y border-white/5">
                            <span class="text-xs font-bold text-white/40 uppercase tracking-widest">Total Length</span>
                            <span class="text-xs font-black text-white tabular-nums">{{ count($queue) }} Themes</span>
                        </div>

                        <div class="grid grid-cols-2 gap-3">
                            <a href="{{ route('playlists.edit', $playlist->id) }}"
                                class="bg-white/5 hover:bg-white/10 text-white text-[10px] font-black uppercase py-3 rounded-xl transition-all flex items-center justify-center gap-2">
                                <span class="material-symbols-outlined text-[16px]">edit</span>
                                Edit Details
                            </a>
                            <a href="{{ route('playlists.index') }}"
                                class="bg-white/5 hover:bg-white/10 text-white text-[10px] font-black uppercase py-3 rounded-xl transition-all flex items-center justify-center gap-2">
                                <span class="material-symbols-outlined text-[16px]">arrow_back</span>
                                Back to List
                            </a>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Data from Laravel
        window.playlistQueue = @json($queue);
        window.currentIndex = 0;
    </script>
    <script>
        class PlaylistPlayer {
            constructor() {
                this.queue = window.playlistQueue || [];
                this.currentIndex = window.currentIndex || 0;

                // Elements
                this.playerContainer = document.getElementById('player-container');
                this.titleEl = document.getElementById('current-song-title');
                this.thumbContainer = document.getElementById('current-thumbnail');
                this.timeEl = document.getElementById('current-time');
                this.durationEl = document.getElementById('duration');

                this.playPauseBtn = document.getElementById('play-pause-btn');
                this.prevBtn = document.getElementById('prev-btn');
                this.nextBtn = document.getElementById('next-btn');
                this.muteBtn = document.getElementById('mute-btn');
                this.fullscreenBtn = document.getElementById('fullscreen-btn');

                this.seekSlider = document.getElementById('seek-slider');
                this.seekProgress = document.getElementById('seek-progress');
                this.volumeSlider = document.getElementById('volume-slider');
                this.volumeProgress = document.getElementById('volume-progress');

                this.queueList = document.getElementById('queue-list');

                // State
                this.currentPlayer = null;
                this.isPlaying = false;
                this.ytPlayer = null;
                this.youtubeTimeInterval = null;
                this.lastVolume = 100;

                this.init();
            }

            init() {
                if (this.queue.length === 0) {
                    this.showEmpty();
                    return;
                }

                this.renderQueue();
                this.loadCurrent();
                this.bindEvents();

                // Keyboard shortcuts
                document.addEventListener('keydown', (e) => {
                    if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;
                    if (e.key === ' ') { e.preventDefault(); this.togglePlay(); }
                    else if (e.key === 'ArrowRight') this.next();
                    else if (e.key === 'ArrowLeft') this.prev();
                });
            }

            loadCurrent() {
                const item = this.queue[this.currentIndex];
                if (!item) return;

                // UI Update
                this.titleEl.textContent = item.song_title;
                if (this.thumbContainer) {
                    const img = this.thumbContainer.querySelector('img');
                    img.src = item.thumbnail || '/images/default.jpg';
                    img.style.opacity = '1';
                }

                this.highlightCurrent();

                // Clear and Reset
                this.playerContainer.innerHTML = '';
                this.currentPlayer = null;
                this.ytPlayer = null;
                this.timeEl.textContent = '0:00';
                this.durationEl.textContent = '--:--';
                this.seekProgress.style.width = '0%';
                this.stopYouTubeTimeTracking();

                if (item.video_type === 'embed') {
                    const embedUrl = this.parseEmbedInput(item.video_url);
                    if (embedUrl) {
                        this.loadYouTubeEmbed(embedUrl);
                    } else {
                        this.showError('Invalid video source');
                    }
                } else {
                    this.loadLocalVideo(item.video_url);
                }
            }

            highlightCurrent() {
                document.querySelectorAll('.queue-item').forEach((el, i) => {
                    const isActive = i === this.currentIndex;
                    el.classList.toggle('bg-primary/10', isActive);
                    el.classList.toggle('border-primary/30', isActive);
                    el.classList.toggle('bg-surface-darker/50', !isActive);
                    el.classList.toggle('border-white/5', !isActive);

                    const activeBadge = el.querySelector('.active-badge');
                    if (activeBadge) activeBadge.classList.toggle('hidden', !isActive);
                });
            }

            parseEmbedInput(input) {
                if (!input) return null;
                const str = input.trim();
                const embedMatch = str.match(/<embed[^>]+src=["']([^"']+)["']/i) || str.match(/<iframe[^>]+src=["']([^"']+)["']/i);
                if (embedMatch) return this.buildEmbedUrl(embedMatch[1]);
                if (/^https?:\/\//i.test(str)) return this.buildEmbedUrl(str);
                return null;
            }

            buildEmbedUrl(url) {
                if (/youtube\.com|youtu\.be/i.test(url)) {
                    const id = this.extractYouTubeId(url);
                    return id ? `https://www.youtube.com/embed/${id}?autoplay=1&rel=0&modestbranding=1&enablejsapi=1` : null;
                }
                return url;
            }

            extractYouTubeId(url) {
                const patterns = [/v=([^"&?/ ]{11})/, /youtu\.be\/([^"&?/ ]{11})/, /embed\/([^"&?/ ]{11})/];
                for (const pattern of patterns) {
                    const match = url.match(pattern);
                    if (match) return match[1];
                }
                return null;
            }

            loadYouTubeEmbed(url) {
                const id = this.extractYouTubeId(url);
                if (!id) return;

                if (!window.YT) {
                    const tag = document.createElement('script');
                    tag.src = "https://www.youtube.com/iframe_api";
                    const firstScriptTag = document.getElementsByTagName('script')[0];
                    firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);
                }

                const container = document.createElement('div');
                container.id = 'yt-player-instance';
                container.className = "w-full h-full";
                this.playerContainer.appendChild(container);

                const createPlayer = () => {
                    this.ytPlayer = new YT.Player('yt-player-instance', {
                        videoId: id,
                        playerVars: { autoplay: 1, controls: 0, modestbranding: 1, rel: 0, iv_load_policy: 3 },
                        events: {
                            onReady: (e) => {
                                this.isPlaying = true;
                                this.updatePlayBtnIcon();
                                this.durationEl.textContent = this.formatTime(e.target.getDuration());
                                this.ytPlayer.setVolume(this.volumeSlider.value);
                                this.startYouTubeTimeTracking();
                            },
                            onStateChange: (e) => {
                                if (e.data === YT.PlayerState.ENDED) this.next();
                                else if (e.data === YT.PlayerState.PLAYING) {
                                    this.isPlaying = true;
                                    this.updatePlayBtnIcon();
                                    this.startYouTubeTimeTracking();
                                } else if (e.data === YT.PlayerState.PAUSED) {
                                    this.isPlaying = false;
                                    this.updatePlayBtnIcon();
                                    this.stopYouTubeTimeTracking();
                                }
                            }
                        }
                    });
                };

                if (window.YT && YT.Player) createPlayer();
                else window.onYouTubeIframeAPIReady = createPlayer;
            }

            loadLocalVideo(url) {
                const video = document.createElement('video');
                video.src = url;
                video.controls = false;
                video.autoplay = true;
                video.volume = this.volumeSlider.value / 100;
                video.className = 'w-full h-full object-contain';
                this.playerContainer.appendChild(video);
                this.currentPlayer = video;

                video.addEventListener('loadedmetadata', () => this.durationEl.textContent = this.formatTime(video.duration));
                video.addEventListener('timeupdate', () => {
                    this.timeEl.textContent = this.formatTime(video.currentTime);
                    const percent = (video.currentTime / video.duration) * 100;
                    this.seekProgress.style.width = percent + '%';
                    this.seekSlider.value = percent;
                });
                video.addEventListener('ended', () => this.next());
                video.addEventListener('play', () => { this.isPlaying = true; this.updatePlayBtnIcon(); });
                video.addEventListener('pause', () => { this.isPlaying = false; this.updatePlayBtnIcon(); });
                video.play().catch(() => { });
            }

            togglePlay() {
                if (this.ytPlayer && this.ytPlayer.getPlayerState) {
                    const state = this.ytPlayer.getPlayerState();
                    state === YT.PlayerState.PLAYING ? this.ytPlayer.pauseVideo() : this.ytPlayer.playVideo();
                } else if (this.currentPlayer) {
                    this.currentPlayer.paused ? this.currentPlayer.play() : this.currentPlayer.pause();
                }
            }

            updatePlayBtnIcon() {
                const icon = this.playPauseBtn.querySelector('.material-symbols-outlined');
                icon.textContent = this.isPlaying ? 'pause' : 'play_arrow';
            }

            next() { if (this.currentIndex < this.queue.length - 1) { this.currentIndex++; this.loadCurrent(); } }
            prev() { if (this.currentIndex > 0) { this.currentIndex--; this.loadCurrent(); } }
            playIndex(index) { this.currentIndex = index; this.loadCurrent(); }

            bindEvents() {
                this.playPauseBtn.addEventListener('click', () => this.togglePlay());
                this.nextBtn.addEventListener('click', () => this.next());
                this.prevBtn.addEventListener('click', () => this.prev());

                // Seek
                this.seekSlider.addEventListener('input', (e) => {
                    const percent = e.target.value;
                    const duration = this.getDuration();
                    if (duration > 0) {
                        const time = (percent / 100) * duration;
                        this.seekTo(time);
                        this.seekProgress.style.width = percent + '%';
                    }
                });

                // Volume
                this.volumeSlider.addEventListener('input', (e) => {
                    const val = e.target.value;
                    this.setVolume(val);
                    this.volumeProgress.style.width = val + '%';
                    this.updateVolumeIcon(val);
                });

                this.muteBtn.addEventListener('click', () => {
                    if (this.getVolume() > 0) {
                        this.lastVolume = this.getVolume();
                        this.setVolume(0);
                        this.volumeSlider.value = 0;
                        this.volumeProgress.style.width = '0%';
                        this.updateVolumeIcon(0);
                    } else {
                        const val = this.lastVolume || 100;
                        this.setVolume(val);
                        this.volumeSlider.value = val;
                        this.volumeProgress.style.width = val + '%';
                        this.updateVolumeIcon(val);
                    }
                });

                this.fullscreenBtn.addEventListener('click', () => this.toggleFullscreen());
            }

            toggleFullscreen() {
                if (!document.fullscreenElement) {
                    this.playerContainer.requestFullscreen().catch(err => {
                        console.error(`Error attempting to enable full-screen mode: ${err.message}`);
                    });
                } else {
                    document.exitFullscreen();
                }
            }

            getDuration() {
                if (this.ytPlayer && this.ytPlayer.getDuration) return this.ytPlayer.getDuration();
                if (this.currentPlayer) return this.currentPlayer.duration;
                return 0;
            }

            seekTo(time) {
                if (this.ytPlayer && this.ytPlayer.seekTo) this.ytPlayer.seekTo(time, true);
                else if (this.currentPlayer) this.currentPlayer.currentTime = time;
            }

            setVolume(val) {
                if (this.ytPlayer && this.ytPlayer.setVolume) this.ytPlayer.setVolume(val);
                else if (this.currentPlayer) this.currentPlayer.volume = val / 100;
            }

            getVolume() {
                if (this.ytPlayer && this.ytPlayer.getVolume) return this.ytPlayer.getVolume();
                if (this.currentPlayer) return this.currentPlayer.volume * 100;
                return 100;
            }

            updateVolumeIcon(val) {
                const icon = this.muteBtn.querySelector('.material-symbols-outlined');
                if (val == 0) icon.textContent = 'volume_off';
                else if (val < 50) icon.textContent = 'volume_down';
                else icon.textContent = 'volume_up';
            }

            renderQueue() {
                this.queueList.innerHTML = '';
                this.queue.forEach((item, i) => {
                    const div = document.createElement('div');
                    div.className = `queue-item group relative flex items-center gap-4 p-3 rounded-xl border transition-all cursor-pointer bg-surface-darker/50 border-white/5 hover:border-primary/30`;

                    div.innerHTML = `
                                        <div class="relative w-16 h-16 shrink-0 rounded-lg overflow-hidden border border-white/10">
                                            <img src="${item.thumbnail || '/images/default.jpg'}" class="w-full h-full object-cover">
                                            <div class="active-badge hidden absolute inset-0 bg-primary/20 backdrop-blur-sm flex items-center justify-center">
                                                <span class="material-symbols-outlined text-white text-[20px] animate-pulse">equalizer</span>
                                            </div>
                                        </div>
                                        <div class="flex-1 min-w-0">
                                            <h4 class="text-sm font-bold text-white truncate group-hover:text-primary transition-colors">${item.song_title || 'Unknown Song'}</h4>
                                            <p class="text-[10px] text-white/40 uppercase font-black tracking-widest mt-1">${item.variant_quality || 'Standard'}</p>
                                        </div>
                                        <div class="text-right shrink-0">
                                            <span class="text-[11px] font-black text-white/20 tabular-nums">${this.formatTime(item.duration || 0)}</span>
                                        </div>
                                    `;

                    div.addEventListener('click', () => this.playIndex(i));
                    this.queueList.appendChild(div);
                });
            }

            formatTime(s) {
                if (!s) return '0:00';
                const m = Math.floor(s / 60);
                const rs = Math.floor(s % 60);
                return `${m}:${rs.toString().padStart(2, '0')}`;
            }

            startYouTubeTimeTracking() {
                this.stopYouTubeTimeTracking();
                this.youtubeTimeInterval = setInterval(() => {
                    if (this.ytPlayer && this.ytPlayer.getCurrentTime) {
                        const current = this.ytPlayer.getCurrentTime();
                        const duration = this.ytPlayer.getDuration();
                        this.timeEl.textContent = this.formatTime(current);

                        if (duration > 0) {
                            const percent = (current / duration) * 100;
                            this.seekProgress.style.width = percent + '%';
                            this.seekSlider.value = percent;
                        }
                    }
                }, 1000);
            }

            stopYouTubeTimeTracking() { if (this.youtubeTimeInterval) { clearInterval(this.youtubeTimeInterval); this.youtubeTimeInterval = null; } }
            showEmpty() { this.playerContainer.innerHTML = '<div class="opacity-40 flex flex-col items-center gap-4"><span class="material-symbols-outlined text-6xl">videocam_off</span><p class="font-bold">No themes in this playlist</p></div>'; }
            showError(m) { this.playerContainer.innerHTML = `<div class="text-red-500 font-bold p-10 text-center">${m}</div>`; }
        }

        document.addEventListener('DOMContentLoaded', () => new PlaylistPlayer());
    </script>
@endsection