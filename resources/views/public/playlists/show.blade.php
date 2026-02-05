@extends ('layouts.app')

@section('title', 'Playlists {{ $playlist->name }}')

@section('content')
    {{-- <div class="container">
        <h1>{{ $playlist->name }}</h1>
        <p>{{ $playlist->description != null ? $playlist->description : 'description' }}</p>

        @foreach ($playlist->songs as $item)
            <p>
                <a href="{{ $item->url }}">{{ $item->name }}</a>
            </p>
        @endforeach
    </div> --}}

    {{-- resources/views/playlist/show.blade.php --}}
    <div class="container">
        <h1 id="playlist-title">{{ $playlist->name }}</h1>
        <small class="text-muted">{{ $playlist->songs->count() != null ? $playlist->songs->count() : '0' }} canciones</small>

        <div>
            <div id="player-container" class="ratio ratio-16x9 bg-dark">
                <!-- Aquí se inyecta iframe o video -->
            </div>
        </div>
        <!-- Reproductor -->
        <div class="card my-4 w-100">
            <div class="card-body p-0 m-0">
                <div class="p-3">
                    <h5 id="current-song-title">Cargando...</h5>
                    <div class="d-flex justify-content-between align-items-center">
                        <span id="current-time">0:00</span>
                        <div>
                            <button id="prev-btn" class="btn btn-sm btn-outline-secondary me-2">Previous</button>
                            <button id="play-pause-btn" class="btn btn-sm btn-primary me-2">Play</button>
                            <button id="next-btn" class="btn btn-sm btn-outline-secondary">Next</button>
                        </div>
                        <span id="duration">0:00</span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Lista de canciones -->
        <div id="queue-list" class="list-group mt-4">
            <!-- Se llena con JS -->
        </div>
    </div>

    <script>
        // Datos de la playlist (desde Laravel)
        window.playlistQueue = @json($queue);
        window.currentIndex = 0;
    </script>
    <script>
        class PlaylistPlayer {
            constructor() {
                this.queue = window.playlistQueue || [];
                this.currentIndex = window.currentIndex || 0;
                this.playerContainer = document.getElementById('player-container');
                this.titleEl = document.getElementById('current-song-title');
                this.timeEl = document.getElementById('current-time');
                this.durationEl = document.getElementById('duration');
                this.playPauseBtn = document.getElementById('play-pause-btn');
                this.prevBtn = document.getElementById('prev-btn');
                this.nextBtn = document.getElementById('next-btn');
                this.queueList = document.getElementById('queue-list');

                this.currentPlayer = null;
                this.isPlaying = false;

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
                setTimeout(() => this.play(), 500);
            }

            loadCurrent() {
                const item = this.queue[this.currentIndex];
                if (!item) return;

                this.titleEl.textContent = item.song_title;
                this.highlightCurrent();
                this.playerContainer.innerHTML = '';
                this.currentPlayer = null;

                // Resetear tiempos
                this.timeEl.textContent = '0:00';
                this.durationEl.textContent = '--:--';

                this.stopYouTubeTimeTracking();

                if (item.video_type === 'embed') {
                    const embedUrl = this.parseEmbedInput(item.video_url);
                    if (embedUrl) {
                        this.loadEmbed(embedUrl);
                    }
                } else {
                    this.loadLocalVideo(item.video_url);
                }
            }

            // === NUEVA FUNCIÓN: PARSEA <embed>, <iframe>, URL ===
            parseEmbedInput(input) {
                if (!input) return null;

                const str = input.trim();

                // 1. <embed src="...">
                const embedMatch = str.match(/<embed[^>]+src=["']([^"']+)["']/i);
                if (embedMatch) {
                    return this.buildEmbedUrl(embedMatch[1]);
                }

                // 2. <iframe src="...">
                const iframeMatch = str.match(/<iframe[^>]+src=["']([^"']+)["']/i);
                if (iframeMatch) {
                    return this.buildEmbedUrl(iframeMatch[1]);
                }

                // 3. URL limpia
                if (/^https?:\/\//i.test(str)) {
                    return this.buildEmbedUrl(str);
                }

                return null;
            }

            buildEmbedUrl(url) {
                // YouTube
                if (/youtube\.com|youtu\.be/i.test(url)) {
                    const id = this.extractYouTubeId(url);
                    return id ? `https://www.youtube.com/embed/${id}?autoplay=1&rel=0&modestbranding=1` : null;
                }

                // Vimeo
                if (/vimeo\.com/i.test(url)) {
                    const id = url.split('/').pop().split('?')[0];
                    return id ? `https://player.vimeo.com/video/${id}?autoplay=1` : null;
                }

                return null;
            }

            extractYouTubeId(url) {
                const patterns = [
                    /youtube\.com.*v=([^"&?/ ]{11})/,
                    /youtu\.be\/([^"&?/ ]{11})/,
                    /youtube\.com\/embed\/([^"&?/ ]{11})/
                ];
                for (const pattern of patterns) {
                    const match = url.match(pattern);
                    if (match) return match[1];
                }
                return null;
            }

            // === EMBED ===
            loadEmbed(url) {
                const isYouTube = /youtube\.com|youtu\.be/i.test(url);
                const iframe = document.createElement('iframe');

                if (isYouTube) {
                    this.loadYouTubeEmbed(url);
                } else {
                    this.loadGenericEmbed(url);
                }

                // Duración simulada
                this.durationEl.textContent = this.formatTime(this.queue[this.currentIndex].duration || 180);
                this.timeEl.textContent = '0:00';
            }

            // === LOCAL VIDEO ===
            loadLocalVideo(url) {
                const video = document.createElement('video');
                video.src = url;
                video.controls = true;
                video.muted = false;
                video.autoplay = true;
                video.className = 'w-100 h-100';

                this.playerContainer.appendChild(video);
                this.currentPlayer = video;

                video.addEventListener('canplay', () => {
                    video.play().catch(e => {
                        console.warn('Autoplay bloqueado:', e);
                        this.playPauseBtn.textContent = 'Play';
                    });
                });

                video.addEventListener('loadedmetadata', () => {
                    this.durationEl.textContent = this.formatTime(video.duration);
                });

                video.addEventListener('timeupdate', () => {
                    this.timeEl.textContent = this.formatTime(video.currentTime);
                });

                video.addEventListener('ended', () => this.next());

                video.addEventListener('play', () => {
                    this.isPlaying = true;
                    this.playPauseBtn.textContent = 'Pause';
                });

                video.addEventListener('pause', () => {
                    this.isPlaying = false;
                    this.playPauseBtn.textContent = 'Play';
                });
            }

            // === CONTROLES ===
            play() {
                if (!this.currentPlayer) return;
                if (this.currentPlayer.tagName === 'VIDEO') {
                    this.currentPlayer.play();
                }
            }

            pause() {
                if (!this.currentPlayer) return;
                if (this.currentPlayer.tagName === 'VIDEO') {
                    this.currentPlayer.pause();
                }
            }

            next() {
                this.stopYouTubeTimeTracking();
                if (this.ytPlayer && this.ytPlayer.destroy) {
                    this.ytPlayer.destroy();
                }
                if (this.currentIndex < this.queue.length - 1) {
                    this.currentIndex++;
                    this.loadCurrent();
                }
            }

            prev() {
                this.stopYouTubeTimeTracking();
                if (this.ytPlayer && this.ytPlayer.destroy) {
                    this.ytPlayer.destroy();
                }
                if (this.currentIndex > 0) {
                    this.currentIndex--;
                    this.loadCurrent();
                }
            }

            playIndex(index) {
                this.currentIndex = index;
                this.loadCurrent();
            }

            bindEvents() {
                this.playPauseBtn.addEventListener('click', () => {
                    this.isPlaying ? this.pause() : this.play();
                });

                this.nextBtn.addEventListener('click', () => this.next());
                this.prevBtn.addEventListener('click', () => this.prev());

                document.addEventListener('keydown', (e) => {
                    if (e.key === ' ') {
                        e.preventDefault();
                        this.playPauseBtn.click();
                    } else if (e.key === 'ArrowRight') this.next();
                    else if (e.key === 'ArrowLeft') this.prev();
                });
            }

            renderQueue() {
                this.queueList.innerHTML = '';
                this.queue.forEach((item, i) => {
                    const div = document.createElement('div');
                    div.className =
                        `list-group-item list-group-item-action d-flex justify-content-between align-items-center`;
                    if (i === this.currentIndex) div.classList.add('active');

                    div.innerHTML = `
                    <div class="d-flex align-items-center">
                        <img src="${item.thumbnail || '/images/default.jpg'}" width="40" class="me-3 rounded">
                        <div>
                            <strong>${item.song_title}</strong>
                            <small class="text-muted d-block">${item.variant_quality || ''}</small>
                        </div>
                    </div>
                    <span class="badge bg-secondary">${this.formatTime(item.duration || 0)}</span>
                `;

                    div.addEventListener('click', () => this.playIndex(i));
                    this.queueList.appendChild(div);
                });
            }

            highlightCurrent() {
                document.querySelectorAll('#queue-list .list-group-item').forEach((el, i) => {
                    el.classList.toggle('active', i === this.currentIndex);
                });
            }

            formatTime(seconds) {
                if (!seconds) return '0:00';
                const m = Math.floor(seconds / 60);
                const s = Math.floor(seconds % 60);
                return `${m}:${s.toString().padStart(2, '0')}`;
            }

            showEmpty() {
                this.playerContainer.innerHTML = '<div class="text-center text-muted p-5">No hay videos</div>';
            }

            showError(msg) {
                this.playerContainer.innerHTML = `<div class="text-center text-danger p-5">${msg}</div>`;
            }

            loadYouTubeAPI() {
                if (document.getElementById('youtube-api-script')) return;

                const script = document.createElement('script');
                script.id = 'youtube-api-script';
                script.src = 'https://www.youtube.com/iframe_api';
                document.body.appendChild(script);
            }

            loadYouTubeEmbed(url) {
                const id = this.extractYouTubeId(url);
                if (!id) return this.showError('ID de YouTube no válido');

                this.loadYouTubeAPI();

                const container = document.createElement('div');
                container.id = `youtube-player-${Date.now()}`;
                this.playerContainer.appendChild(container);

                const createPlayer = () => {
                    this.ytPlayer = new YT.Player(container.id, {
                        videoId: id,
                        playerVars: {
                            autoplay: 1,
                            rel: 0,
                            modestbranding: 1,
                            playsinline: 1,
                            controls: 0, // ← OCULTA LOS CONTROLES
                            showinfo: 0, // ← (obsoleto, pero por si acaso)
                            fs: 0, // ← opcional: oculta botón fullscreen
                            iv_load_policy: 3 // ← oculta anotaciones
                        },
                        events: {
                            onReady: (e) => {
                                this.isPlaying = true;
                                this.playPauseBtn.textContent = 'Pause';

                                const duration = e.target.getDuration();
                                this.durationEl.textContent = this.formatTime(duration);

                                this.startYouTubeTimeTracking();
                            },
                            onStateChange: (e) => {
                                if (e.data === YT.PlayerState.ENDED) {
                                    this.next();
                                } else if (e.data === YT.PlayerState.PLAYING) {
                                    this.isPlaying = true;
                                    this.playPauseBtn.textContent = 'Pause';
                                    this.startYouTubeTimeTracking();
                                } else if (e.data === YT.PlayerState.PAUSED) {
                                    this.isPlaying = false;
                                    this.playPauseBtn.textContent = 'Play';
                                    this.stopYouTubeTimeTracking();
                                }
                            }
                        }
                    });
                };

                if (window.YT && window.YT.Player) {
                    createPlayer();
                } else {
                    window.onYouTubeIframeAPIReady = createPlayer;
                }
            }

            loadGenericEmbed(url) {
                const iframe = document.createElement('iframe');
                iframe.src = url;
                iframe.allow = 'autoplay; encrypted-media; fullscreen';
                iframe.allowFullscreen = true;
                iframe.className = 'w-100 h-100 border-0';

                this.playerContainer.appendChild(iframe);
                this.currentPlayer = iframe;
            }

            // === TIEMPO EN YOUTUBE ===
            startYouTubeTimeTracking() {
                this.stopYouTubeTimeTracking(); // Evitar duplicados

                this.youtubeTimeInterval = setInterval(() => {
                    if (!this.ytPlayer || !this.ytPlayer.getCurrentTime) return;

                    const current = this.ytPlayer.getCurrentTime();
                    this.timeEl.textContent = this.formatTime(current);
                }, 500); // cada 0.5s
            }

            stopYouTubeTimeTracking() {
                if (this.youtubeTimeInterval) {
                    clearInterval(this.youtubeTimeInterval);
                    this.youtubeTimeInterval = null;
                }
            }
        }

        document.addEventListener('DOMContentLoaded', () => {
            new PlaylistPlayer();
        });
    </script>
@endsection
