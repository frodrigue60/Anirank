@extends('layouts.app')

@section('content')
    <div class="max-w-[1440px] mx-auto px-6 py-8 space-y-8">
        {{-- Header --}}
        <div class="flex items-center justify-between">
            <div>
                <h1 class="text-3xl font-black text-white flex items-center gap-3">
                    <span class="material-symbols-outlined text-primary text-4xl">dashboard</span>
                    Admin Dashboard
                </h1>
                <p class="text-white/40 text-sm mt-1">Manage your platform content and settings</p>
            </div>
            <div class="flex items-center gap-2 bg-surface-dark px-4 py-2 rounded-xl border border-white/5">
                <span class="w-2 h-2 bg-green-500 rounded-full"></span>
                <span class="text-sm text-white/60">Logged in as <span
                        class="text-primary font-bold">{{ Auth::user()->name }}</span></span>
            </div>
        </div>

        @if (Auth::check() && Auth::user()->isStaff())
            {{-- Quick Stats (Optional - can be expanded later) --}}
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div class="bg-surface-dark rounded-2xl p-6 border border-white/5 hover:border-primary/30 transition-all">
                    <div class="flex items-center justify-between mb-2">
                        <span class="material-symbols-outlined text-primary text-3xl">movie</span>
                    </div>
                    <p class="text-2xl font-black text-white">{{ \App\Models\Post::count() }}</p>
                    <p class="text-xs text-white/40 uppercase tracking-widest mt-1">Total Posts</p>
                </div>
                <div class="bg-surface-dark rounded-2xl p-6 border border-white/5 hover:border-primary/30 transition-all">
                    <div class="flex items-center justify-between mb-2">
                        <span class="material-symbols-outlined text-primary text-3xl">music_note</span>
                    </div>
                    <p class="text-2xl font-black text-white">{{ \App\Models\Song::count() }}</p>
                    <p class="text-xs text-white/40 uppercase tracking-widest mt-1">Total Songs</p>
                </div>
                <div class="bg-surface-dark rounded-2xl p-6 border border-white/5 hover:border-primary/30 transition-all">
                    <div class="flex items-center justify-between mb-2">
                        <span class="material-symbols-outlined text-primary text-3xl">person</span>
                    </div>
                    <p class="text-2xl font-black text-white">{{ \App\Models\Artist::count() }}</p>
                    <p class="text-xs text-white/40 uppercase tracking-widest mt-1">Total Artists</p>
                </div>
                <div class="bg-surface-dark rounded-2xl p-6 border border-white/5 hover:border-primary/30 transition-all">
                    <div class="flex items-center justify-between mb-2">
                        <span class="material-symbols-outlined text-primary text-3xl">group</span>
                    </div>
                    <p class="text-2xl font-black text-white">{{ \App\Models\User::count() }}</p>
                    <p class="text-xs text-white/40 uppercase tracking-widest mt-1">Total Users</p>
                </div>
            </div>

            {{-- Management Sections --}}
            <div>
                <h2 class="text-xl font-bold text-white mb-4 flex items-center gap-2">
                    <span class="material-symbols-outlined text-primary">settings</span>
                    Content Management
                </h2>
                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {{-- Posts --}}
                    <a href="{{ route('admin.posts.index') }}"
                        class="group bg-surface-dark rounded-2xl p-6 border border-white/5 hover:border-primary/50 hover:bg-surface-darker transition-all">
                        <div class="flex items-start justify-between mb-4">
                            <div
                                class="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors">
                                <span class="material-symbols-outlined text-primary text-2xl">movie</span>
                            </div>
                            <span
                                class="material-symbols-outlined text-white/20 group-hover:text-primary transition-colors">arrow_forward</span>
                        </div>
                        <h3 class="text-lg font-bold text-white mb-1">Posts</h3>
                        <p class="text-sm text-white/40">Manage anime posts and metadata</p>
                    </a>

                    {{-- Artists --}}
                    <a href="{{ route('admin.artists.index') }}"
                        class="group bg-surface-dark rounded-2xl p-6 border border-white/5 hover:border-primary/50 hover:bg-surface-darker transition-all">
                        <div class="flex items-start justify-between mb-4">
                            <div
                                class="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors">
                                <span class="material-symbols-outlined text-primary text-2xl">person</span>
                            </div>
                            <span
                                class="material-symbols-outlined text-white/20 group-hover:text-primary transition-colors">arrow_forward</span>
                        </div>
                        <h3 class="text-lg font-bold text-white mb-1">Artists</h3>
                        <p class="text-sm text-white/40">Manage music artists and performers</p>
                    </a>

                    {{-- Years --}}
                    <a href="{{ route('admin.years.index') }}"
                        class="group bg-surface-dark rounded-2xl p-6 border border-white/5 hover:border-primary/50 hover:bg-surface-darker transition-all">
                        <div class="flex items-start justify-between mb-4">
                            <div
                                class="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors">
                                <span class="material-symbols-outlined text-primary text-2xl">calendar_today</span>
                            </div>
                            <span
                                class="material-symbols-outlined text-white/20 group-hover:text-primary transition-colors">arrow_forward</span>
                        </div>
                        <h3 class="text-lg font-bold text-white mb-1">Years</h3>
                        <p class="text-sm text-white/40">Manage broadcast years</p>
                    </a>

                    {{-- Seasons --}}
                    <a href="{{ route('admin.seasons.index') }}"
                        class="group bg-surface-dark rounded-2xl p-6 border border-white/5 hover:border-primary/50 hover:bg-surface-darker transition-all">
                        <div class="flex items-start justify-between mb-4">
                            <div
                                class="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors">
                                <span class="material-symbols-outlined text-primary text-2xl">wb_twilight</span>
                            </div>
                            <span
                                class="material-symbols-outlined text-white/20 group-hover:text-primary transition-colors">arrow_forward</span>
                        </div>
                        <h3 class="text-lg font-bold text-white mb-1">Seasons</h3>
                        <p class="text-sm text-white/40">Manage seasonal periods</p>
                    </a>

                    {{-- Users --}}
                    <a href="{{ route('admin.users.index') }}"
                        class="group bg-surface-dark rounded-2xl p-6 border border-white/5 hover:border-primary/50 hover:bg-surface-darker transition-all">
                        <div class="flex items-start justify-between mb-4">
                            <div
                                class="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors">
                                <span class="material-symbols-outlined text-primary text-2xl">group</span>
                            </div>
                            <span
                                class="material-symbols-outlined text-white/20 group-hover:text-primary transition-colors">arrow_forward</span>
                        </div>
                        <h3 class="text-lg font-bold text-white mb-1">Users</h3>
                        <p class="text-sm text-white/40">Manage user accounts and permissions</p>
                    </a>

                    {{-- Reports --}}
                    <a href="{{ route('admin.reports.index') }}"
                        class="group bg-surface-dark rounded-2xl p-6 border border-white/5 hover:border-primary/50 hover:bg-surface-darker transition-all">
                        <div class="flex items-start justify-between mb-4">
                            <div
                                class="w-12 h-12 rounded-xl bg-red-500/10 flex items-center justify-center group-hover:bg-red-500/20 transition-colors">
                                <span class="material-symbols-outlined text-red-500 text-2xl">flag</span>
                            </div>
                            <span
                                class="material-symbols-outlined text-white/20 group-hover:text-red-500 transition-colors">arrow_forward</span>
                        </div>
                        <h3 class="text-lg font-bold text-white mb-1">Reports</h3>
                        <p class="text-sm text-white/40">Review user-submitted reports</p>
                    </a>

                    {{-- Requests --}}
                    <a href="{{ route('admin.requests.index') }}"
                        class="group bg-surface-dark rounded-2xl p-6 border border-white/5 hover:border-primary/50 hover:bg-surface-darker transition-all">
                        <div class="flex items-start justify-between mb-4">
                            <div
                                class="w-12 h-12 rounded-xl bg-blue-500/10 flex items-center justify-center group-hover:bg-blue-500/20 transition-colors">
                                <span class="material-symbols-outlined text-blue-500 text-2xl">campaign</span>
                            </div>
                            <span
                                class="material-symbols-outlined text-white/20 group-hover:text-blue-500 transition-colors">arrow_forward</span>
                        </div>
                        <h3 class="text-lg font-bold text-white mb-1">Requests</h3>
                        <p class="text-sm text-white/40">Manage user content requests</p>
                    </a>
                </div>
            </div>
        @else
            <div class="bg-red-500/10 border border-red-500/20 rounded-2xl p-8 text-center">
                <span class="material-symbols-outlined text-red-500 text-5xl mb-4">block</span>
                <h3 class="text-xl font-bold text-red-400 mb-2">Access Denied</h3>
                <p class="text-white/60">You don't have permission to access the admin dashboard.</p>
            </div>
        @endif
    </div>
@endsection
