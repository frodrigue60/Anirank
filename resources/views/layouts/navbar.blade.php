<header class="sticky top-0 z-50 glass-panel border-b border-white/5 w-full">
    <div class="max-w-[1440px] mx-auto px-6 h-16 flex items-center justify-between gap-4">
        <div class="flex items-center gap-10">
            <a class="flex items-center gap-2 group" href="{{ url('/') }}">
                <div
                    class="w-8 h-8 rounded-lg bg-primary flex items-center justify-center text-white shadow-lg shadow-primary/40 group-hover:scale-105 transition-transform">
                    <span class="material-symbols-outlined text-[20px]">music_note</span>
                </div>
                <span class="text-xl font-bold tracking-tight text-white">Anirank</span>
            </a>
            <nav class="hidden lg:flex items-center gap-6">
                <a class="text-sm font-medium {{ Request::is('seasonal') ? 'text-primary' : 'text-white/60 hover:text-white' }} transition-colors"
                    href="{{ route('songs.seasonal') }}">Season</a>
                <a class="text-sm font-medium {{ Request::is('ranking') ? 'text-primary' : 'text-white/60 hover:text-white' }} transition-colors"
                    href="{{ route('songs.ranking') }}">Ranking</a>

                {{-- Discover Dropdown --}}
                <div class="relative">
                    <button id="discover-menu-button"
                        class="flex items-center gap-1 text-sm font-medium text-white/60 hover:text-white transition-colors">
                        Discover <span class="material-symbols-outlined text-[18px]">expand_more</span>
                    </button>

                    <div id="discover-dropdown"
                        class="z-50 hidden absolute left-0 mt-3 w-48 glass-panel rounded-xl border border-white/10 shadow-2xl overflow-hidden py-1 bg-surface-darker">
                        <a href="{{ route('posts.animes') }}"
                            class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                            <span class="material-symbols-outlined text-[18px]">movie</span> Animes
                        </a>
                        <a href="{{ route('artists.index') }}"
                            class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                            <span class="material-symbols-outlined text-[18px]">person</span> Artists
                        </a>
                        <a href="{{ route('posts.themes') }}"
                            class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                            <span class="material-symbols-outlined text-[18px]">library_music</span> Themes
                        </a>
                        <a href="{{ route('studios.index') }}"
                            class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                            <span class="material-symbols-outlined text-[18px]">business</span> Studios
                        </a>
                        <a href="{{ route('producers.index') }}"
                            class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                            <span class="material-symbols-outlined text-[18px]">token</span> Producers
                        </a>
                    </div>
                </div>
            </nav>
        </div>

        <div class="flex items-center gap-4 flex-1 justify-end max-w-lg">
            {{-- Search Trigger --}}
            <button onclick="document.getElementById('modal-search').classList.remove('hidden')"
                class="relative w-full max-w-xs hidden sm:flex items-center group bg-surface-darker/50 border border-white/10 rounded-full h-10 px-4 text-white/40 hover:border-primary/50 transition-all">
                <span class="material-symbols-outlined text-[20px] mr-2">search</span>
                <span class="text-sm">Search anime, songs...</span>
            </button>

            <div class="h-8 w-[1px] bg-white/10 hidden sm:block"></div>

            {{-- User Toggle --}}
            <div class="relative">
                <button id="user-menu-button" class="relative group flex items-center">
                    @auth
                        @php
                            $userImage =
                                Auth::user()->image && Storage::disk('public')->exists(Auth::user()->image)
                                    ? Storage::url(Auth::user()->image)
                                    : 'https://www.gravatar.com/avatar/' .
                                        md5(strtolower(trim(Auth::user()->email))) .
                                        '?d=mp';
                        @endphp
                        <div class="w-9 h-9 rounded-full bg-cover bg-center ring-2 ring-transparent group-hover:ring-primary transition-all shadow-lg"
                            style="background-image: url('{{ $userImage }}');"></div>
                        <div
                            class="absolute bottom-0 right-0 w-2.5 h-2.5 bg-green-500 rounded-full border-2 border-background-dark">
                        </div>
                    @else
                        <div
                            class="w-9 h-9 rounded-full bg-surface-dark flex items-center justify-center text-white/60 group-hover:text-primary transition-colors border border-white/10 shadow-lg">
                            <span class="material-symbols-outlined">person</span>
                        </div>
                    @endauth
                </button>

                <!-- User dropdown -->
                <div id="user-dropdown"
                    class="z-50 hidden absolute right-0 mt-3 w-56 rounded-xl border border-white/10 shadow-2xl overflow-hidden py-1 bg-[#1e1e2e]">
                    @auth
                        <div class="px-4 py-3 border-b border-white/5">
                            <p class="text-sm font-bold text-white truncate">{{ Auth::user()->name }}</p>
                            <p class="text-xs text-white/40 truncate">{{ Auth::user()->email }}</p>
                        </div>
                        <div class="py-1">
                            @if (Auth::user()->isStaff())
                                <a href="{{ route('admin.dashboard') }}"
                                    class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                                    <span class="material-symbols-outlined text-[18px]">dashboard</span> Dashboard
                                </a>
                            @endif
                            <a href="{{ route('users.list', Auth::user()->slug) }}"
                                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                                <span class="material-symbols-outlined text-[18px]">person</span> Profile
                            </a>
                            <a href="{{ route('profile.favorites') }}"
                                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                                <span class="material-symbols-outlined text-[18px]">favorite</span> Favorites
                            </a>
                            <a href="{{ route('playlists.index') }}"
                                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                                <span class="material-symbols-outlined text-[18px]">playlist_play</span> Playlists
                            </a>
                            <button onclick="window.showModal('requestModal')"
                                class="w-full flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                                <span class="material-symbols-outlined text-[18px]">campaign</span> Request
                            </button>
                        </div>
                        <div class="py-1 border-t border-white/5">
                            <a href="{{ route('logout') }}"
                                onclick="event.preventDefault(); localStorage.removeItem('api_token'); document.getElementById('logout-form').submit();"
                                class="flex items-center gap-3 px-4 py-2 text-sm text-red-400 hover:bg-red-400/10 transition-colors">
                                <span class="material-symbols-outlined text-[18px]">logout</span> Logout
                            </a>
                            <form id="logout-form" action="{{ route('logout') }}" method="POST" class="hidden">
                                @csrf
                            </form>
                        </div>
                    @else
                        <div class="py-1">
                            <a href="{{ route('login') }}"
                                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                                <span class="material-symbols-outlined text-[18px]">login</span> Login
                            </a>
                            <a href="{{ route('register') }}"
                                class="flex items-center gap-3 px-4 py-2 text-sm text-white/80 hover:bg-white/5 hover:text-primary transition-colors">
                                <span class="material-symbols-outlined text-[18px]">person_add</span> Register
                            </a>
                        </div>
                    @endauth
                </div>
            </div>

            {{-- Mobile Toggle --}}
            <button id="mobile-menu-button"
                class="lg:hidden w-10 h-10 flex items-center justify-center text-white/60 hover:text-white transition-colors">
                <span class="material-symbols-outlined">menu</span>
            </button>
        </div>
    </div>

</header>

{{-- Mobile Menu Drawer --}}
<div id="mobile-menu-drawer" class="fixed inset-0 z-[100] hidden">
    {{-- Backdrop --}}
    <div id="mobile-menu-backdrop"
        class="absolute inset-0 bg-black/60 backdrop-blur-sm opacity-0 transition-opacity duration-300"></div>

    {{-- Drawer Content --}}
    <div id="mobile-menu-content"
        class="absolute top-0 left-0 w-[280px] h-full glass-panel bg-surface-darker/95 border-r border-white/5 shadow-2xl transform -translate-x-full transition-transform duration-300 ease-out flex flex-col">
        <div class="px-6 h-16 flex items-center justify-between border-b border-white/5">
            <a class="flex items-center gap-2" href="{{ url('/') }}">
                <div
                    class="w-8 h-8 rounded-lg bg-primary flex items-center justify-center text-white shadow-lg shadow-primary/40">
                    <span class="material-symbols-outlined text-[18px]">music_note</span>
                </div>
                <span class="text-lg font-bold tracking-tight text-white">Anirank</span>
            </a>
            <button id="mobile-menu-close" class="text-white/40 hover:text-white transition-colors">
                <span class="material-symbols-outlined">close</span>
            </button>
        </div>

        <nav class="flex-1 py-6 px-4 overflow-y-auto space-y-2">
            <div class=" border-t border-white/5">
                <button
                    onclick="document.getElementById('modal-search').classList.remove('hidden'); document.getElementById('mobile-menu-close').click();"
                    class="w-full h-11 flex items-center justify-center gap-2 bg-primary/10 border border-primary/20 rounded-xl text-primary font-bold transition-all hover:bg-primary/20 group">
                    <span class="material-symbols-outlined text-[20px]">search</span>
                    <span class="text-sm">Search</span>
                </button>
            </div>
            <a href="{{ route('songs.seasonal') }}"
                class="flex items-center gap-3 px-4 py-3 rounded-xl text-white/60 hover:text-white hover:bg-white/5 transition-all {{ Request::is('seasonal') ? 'bg-primary/10 text-primary' : '' }}">
                <span
                    class="material-symbols-outlined {{ Request::is('seasonal') ? 'filled' : '' }}">calendar_today</span>
                <span class="font-medium text-sm">Season Charts</span>
            </a>
            <a href="{{ route('songs.ranking') }}"
                class="flex items-center gap-3 px-4 py-3 rounded-xl text-white/60 hover:text-white hover:bg-white/5 transition-all {{ Request::is('ranking') ? 'bg-primary/10 text-primary' : '' }}">
                <span
                    class="material-symbols-outlined {{ Request::is('ranking') ? 'filled' : '' }}">military_tech</span>
                <span class="font-medium text-sm">Rankings</span>
            </a>

            <div class="h-[1px] bg-white/5 my-4"></div>
            <div class="px-4 mb-2 text-[10px] font-bold text-white/20 uppercase tracking-[0.2em]">Discover</div>

            <a href="{{ route('posts.animes') }}"
                class="flex items-center gap-3 px-4 py-3 rounded-xl text-white/60 hover:text-white hover:bg-white/5 transition-all">
                <span class="material-symbols-outlined">movie</span>
                <span class="font-medium text-sm">Series</span>
            </a>
            <a href="{{ route('artists.index') }}"
                class="flex items-center gap-3 px-4 py-3 rounded-xl text-white/60 hover:text-white hover:bg-white/5 transition-all">
                <span class="material-symbols-outlined">person</span>
                <span class="font-medium text-sm">Artists</span>
            </a>
            <a href="{{ route('posts.themes') }}"
                class="flex items-center gap-3 px-4 py-3 rounded-xl text-white/60 hover:text-white hover:bg-white/5 transition-all">
                <span class="material-symbols-outlined">library_music</span>
                <span class="font-medium text-sm">Themes</span>
            </a>
            <a href="{{ route('studios.index') }}"
                class="flex items-center gap-3 px-4 py-3 rounded-xl text-white/60 hover:text-white hover:bg-white/5 transition-all">
                <span class="material-symbols-outlined">business</span>
                <span class="font-medium text-sm">Studios</span>
            </a>
            <a href="{{ route('producers.index') }}"
                class="flex items-center gap-3 px-4 py-3 rounded-xl text-white/60 hover:text-white hover:bg-white/5 transition-all">
                <span class="material-symbols-outlined">token</span>
                <span class="font-medium text-sm">Producers</span>
            </a>
        </nav>


    </div>
</div>

<script>
    document.addEventListener('DOMContentLoaded', function() {
        // Dropdown Utility
        function setupDropdown(buttonId, menuId) {
            const btn = document.getElementById(buttonId);
            const menu = document.getElementById(menuId);
            if (!btn || !menu) return;

            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                menu.classList.toggle('hidden');
                menu.classList.toggle('animate-in');
                menu.classList.toggle('fade-in');
                menu.classList.toggle('slide-in-from-top-2');
            });

            document.addEventListener('click', (e) => {
                if (!btn.contains(e.target) && !menu.contains(e.target)) {
                    menu.classList.add('hidden');
                }
            });
        }

        setupDropdown('user-menu-button', 'user-dropdown');
        setupDropdown('discover-menu-button', 'discover-dropdown');

        // Mobile Menu Logic
        const mobileBtn = document.getElementById('mobile-menu-button');
        const mobileDrawer = document.getElementById('mobile-menu-drawer');
        const mobileContent = document.getElementById('mobile-menu-content');
        const mobileBackdrop = document.getElementById('mobile-menu-backdrop');
        const mobileClose = document.getElementById('mobile-menu-close');

        function openMobileMenu() {
            mobileDrawer.classList.remove('hidden');
            setTimeout(() => {
                mobileBackdrop.classList.replace('opacity-0', 'opacity-100');
                mobileContent.classList.replace('-translate-x-full', 'translate-x-0');
            }, 10);
            document.body.style.overflow = 'hidden';
        }

        function closeMobileMenu() {
            mobileBackdrop.classList.replace('opacity-100', 'opacity-0');
            mobileContent.classList.replace('translate-x-0', '-translate-x-full');
            setTimeout(() => {
                mobileDrawer.classList.add('hidden');
            }, 300);
            document.body.style.overflow = '';
        }

        if (mobileBtn) mobileBtn.addEventListener('click', openMobileMenu);
        if (mobileClose) mobileClose.addEventListener('click', closeMobileMenu);
        if (mobileBackdrop) mobileBackdrop.addEventListener('click', closeMobileMenu);
    });
</script>
