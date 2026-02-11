<div id="modal-search" class="fixed inset-0 z-[100] hidden overflow-y-auto" aria-labelledby="modal-title" role="dialog"
    aria-modal="true">
    <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        {{-- Backdrop --}}
        <div class="fixed inset-0 bg-black/80 backdrop-blur-sm transition-opacity" aria-hidden="true"
            onclick="document.getElementById('modal-search').classList.add('hidden')"></div>

        {{-- Center cheat --}}
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        {{-- Modal Content --}}
        <div
            class="inline-block w-full max-w-2xl overflow-hidden text-left align-bottom transition-all transform glass-panel bg-surface-darker/95 rounded-2xl shadow-2xl sm:my-8 sm:align-middle border border-white/10 relative">

            {{-- Header --}}
            <div class="px-6 py-4 border-b border-white/5 flex justify-between items-center">
                <div class="flex items-center gap-2">
                    <span class="material-symbols-outlined text-primary">search</span>
                    <h3 class="text-xl font-bold tracking-tight text-white" id="modal-title">Search Music</h3>
                </div>
                <button type="button" class="text-white/40 hover:text-white transition-colors"
                    onclick="document.getElementById('modal-search').classList.add('hidden')">
                    <span class="material-symbols-outlined">close</span>
                </button>
            </div>

            <div class="p-6">
                {{-- Search Input --}}
                <div class="relative group mb-8">
                    <div class="absolute inset-y-0 left-4 flex items-center pointer-events-none">
                        <span
                            class="material-symbols-outlined text-white/20 group-focus-within:text-primary transition-colors">search</span>
                    </div>
                    <form role="search" id="form-search" data-url-base="{{ config('app.url') }}">
                        <input id="searchInputModal"
                            class="w-full bg-surface-dark border border-white/10 rounded-full py-4 pl-12 pr-6 text-white placeholder:text-white/20 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all font-medium text-lg"
                            type="search" placeholder="Search anime, artists, users..." autocomplete="off">
                    </form>
                </div>

                {{-- Results Body --}}
                <div id="modalBody" class="max-h-[50vh] overflow-y-auto custom-scrollbar pr-2">
                    <div class="res hidden space-y-10">
                        {{-- Animes Section --}}
                        <section>
                            <div class="flex items-center gap-4 mb-4">
                                <span
                                    class="flex items-center gap-1.5 px-3 py-1 rounded-full bg-blue-500/10 text-blue-400 text-[10px] font-black uppercase tracking-widest border border-blue-500/20">
                                    <span class="material-symbols-outlined text-[14px] filled">movie</span>
                                    Animes
                                </span>
                                <div class="h-px flex-grow bg-white/5"></div>
                            </div>
                            <div id="posts" class="grid grid-cols-1 gap-3"></div>
                        </section>

                        {{-- Artists Section --}}
                        <section>
                            <div class="flex items-center gap-4 mb-4">
                                <span
                                    class="flex items-center gap-1.5 px-3 py-1 rounded-full bg-purple-500/10 text-purple-400 text-[10px] font-black uppercase tracking-widest border border-purple-500/20">
                                    <span class="material-symbols-outlined text-[14px] filled">person</span>
                                    Artists
                                </span>
                                <div class="h-px flex-grow bg-white/5"></div>
                            </div>
                            <div id="artists" class="grid grid-cols-1 gap-3"></div>
                        </section>

                        {{-- Users Section --}}
                        <section>
                            <div class="flex items-center gap-4 mb-4">
                                <span
                                    class="flex items-center gap-1.5 px-3 py-1 rounded-full bg-green-500/10 text-green-400 text-[10px] font-black uppercase tracking-widest border border-green-500/20">
                                    <span class="material-symbols-outlined text-[14px] filled">group</span>
                                    Users
                                </span>
                                <div class="h-px flex-grow bg-white/5"></div>
                            </div>
                            <div id="users" class="grid grid-cols-1 gap-3"></div>
                        </section>
                    </div>
                </div>
            </div>

            {{-- Footer Decorations --}}
            <div
                class="absolute -bottom-12 -right-12 w-32 h-32 bg-primary/10 rounded-full blur-3xl pointer-events-none">
            </div>
        </div>
    </div>
</div>
