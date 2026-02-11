<footer class="mt-20 border-t border-white/5 bg-background-dark/50 backdrop-blur-xl relative overflow-hidden">
    {{-- Decorative glow effect --}}
    <div class="absolute -top-24 -left-24 w-48 h-48 bg-primary/10 rounded-full blur-3xl pointer-events-none"></div>
    <div class="absolute -bottom-24 -right-24 w-48 h-48 bg-primary/10 rounded-full blur-3xl pointer-events-none"></div>

    <div class="max-w-[1440px] mx-auto px-6 py-16 relative">
        <div class="grid grid-cols-1 md:grid-cols-4 lg:grid-cols-5 gap-12">
            {{-- Brand Section --}}
            <div class="col-span-1 lg:col-span-2">
                <a class="flex items-center gap-2 group mb-6" href="{{ url('/') }}">
                    <div
                        class="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-white shadow-lg shadow-primary/40 group-hover:scale-105 transition-transform">
                        <span class="material-symbols-outlined text-[24px]">music_note</span>
                    </div>
                    <span class="text-2xl font-black tracking-tighter text-white">Anirank</span>
                </a>
                <p class="text-white/40 text-sm leading-relaxed max-w-xs">
                    Discover, rate, and explore the best anime openings and endings from all your favorite series. Join
                    our community of anime music enthusiasts.
                </p>
                <div class="flex items-center gap-4 mt-8">
                    <a href="#"
                        class="w-10 h-10 rounded-full bg-white/5 flex items-center justify-center text-white/40 hover:bg-primary hover:text-white transition-all">
                        <i class="fa-brands fa-x-twitter"></i>
                    </a>
                    <a href="#"
                        class="w-10 h-10 rounded-full bg-white/5 flex items-center justify-center text-white/40 hover:bg-primary hover:text-white transition-all">
                        <i class="fa-brands fa-discord"></i>
                    </a>
                    <a href="#"
                        class="w-10 h-10 rounded-full bg-white/5 flex items-center justify-center text-white/40 hover:bg-primary hover:text-white transition-all">
                        <i class="fa-brands fa-github"></i>
                    </a>
                </div>
            </div>

            {{-- Links Sections --}}
            <div>
                <h4 class="text-sm font-bold text-white uppercase tracking-widest mb-6 flex items-center gap-2">
                    <span class="material-symbols-outlined text-primary text-[18px]">info</span>
                    About
                </h4>
                <ul class="space-y-4">
                    <li><a href="#"
                            class="text-white/40 hover:text-primary transition-colors text-sm font-medium">Terms &
                            Privacy</a></li>
                    <li><a href="#"
                            class="text-white/40 hover:text-primary transition-colors text-sm font-medium">About Us</a>
                    </li>
                    <li><a href="#"
                            class="text-white/40 hover:text-primary transition-colors text-sm font-medium">Support</a>
                    </li>
                </ul>
            </div>

            <div>
                <h4 class="text-sm font-bold text-white uppercase tracking-widest mb-6 flex items-center gap-2">
                    <span class="material-symbols-outlined text-primary text-[18px]">explore</span>
                    Discover
                </h4>
                <ul class="space-y-4">
                    <li><a href="{{ route('animes') }}"
                            class="text-white/40 hover:text-primary transition-colors text-sm font-medium">Series</a>
                    </li>
                    <li><a href="{{ route('artists.index') }}"
                            class="text-white/40 hover:text-primary transition-colors text-sm font-medium">Artists</a>
                    </li>
                    <li><a href="{{ route('themes') }}"
                            class="text-white/40 hover:text-primary transition-colors text-sm font-medium">Themes</a>
                    </li>
                </ul>
            </div>

            <div>
                <h4 class="text-sm font-bold text-white uppercase tracking-widest mb-6 flex items-center gap-2">
                    <span class="material-symbols-outlined text-primary text-[18px]">help</span>
                    Help
                </h4>
                <ul class="space-y-4">
                    <li><a href="#"
                            class="text-white/40 hover:text-primary transition-colors text-sm font-medium">Report
                            Error</a></li>
                    <li><a href="#"
                            class="text-white/40 hover:text-primary transition-colors text-sm font-medium">DMCA</a></li>
                    <li><a href="#"
                            class="text-white/40 hover:text-primary transition-colors text-sm font-medium">Contact</a>
                    </li>
                </ul>
            </div>
        </div>

        <div class="mt-16 pt-8 border-t border-white/5 flex flex-col md:flex-row justify-between items-center gap-4">
            <p class="text-white/20 text-xs font-medium">
                &copy; {{ date('Y') }} Anirank. All rights reserved.
            </p>
            <div class="flex items-center gap-6">
                <span class="text-white/20 text-xs font-medium flex items-center gap-1">
                    Made with <span class="material-symbols-outlined text-primary text-[14px] filled">favorite</span> by
                    fans
                </span>
            </div>
        </div>
    </div>
</footer>
