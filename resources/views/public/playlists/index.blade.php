@extends('layouts.app')

@section('meta')
    <title>My Playlists | {{ config('app.name') }}</title>
    <meta title="My Playlists">
    <meta name="description" content="Manage your custom anime music playlists.">
@endsection

@section('content')
    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10 md:py-8 flex flex-col gap-12">
        <div class="flex flex-col gap-8">
            {{-- Header --}}
            <div class="flex flex-col md:flex-row justify-between items-end gap-4">
                <div>
                    <h1 class="text-3xl font-black tracking-tight text-white mb-2 flex items-center gap-4">
                        <span class="material-symbols-outlined text-primary text-4xl">featured_play_list</span>
                        My Playlists
                    </h1>
                    <div class="h-1 w-20 bg-primary rounded-full"></div>
                </div>

                <a href="{{ route('playlists.create') }}"
                    class="bg-primary hover:bg-primary-light text-white text-xs font-black uppercase px-6 py-3 rounded-xl transition-all shadow-lg shadow-primary/20 flex items-center gap-2 group">
                    <span
                        class="material-symbols-outlined text-[20px] group-hover:rotate-90 transition-transform">add</span>
                    Create New Playlist
                </a>
            </div>

            {{-- Grid Section --}}
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                @forelse ($playlists as $playlist)
                    @php
                        $cover = asset('images/default.jpg');
                        $firstSong = $playlist->songs->first();
                        if ($firstSong && $firstSong->post && $firstSong->post->thumbnail && Storage::disk('public')->exists($firstSong->post->thumbnail)) {
                            $cover = Storage::url($firstSong->post->thumbnail);
                        }
                    @endphp
                    <div
                        class="group relative bg-surface-dark/30 rounded-2xl border border-white/5 p-5 backdrop-blur-sm hover:bg-surface-dark/50 transition-all flex flex-col gap-4 overflow-hidden">

                        {{-- Cover Image Background --}}
                        <div class="absolute inset-0 opacity-10 group-hover:opacity-20 transition-opacity">
                            <img src="{{ $cover }}" class="w-full h-full object-cover blur-sm scale-110">
                            <div class="absolute inset-0 bg-gradient-to-b from-surface-darker/80 to-surface-darker"></div>
                        </div>

                        <div class="relative flex items-start justify-between gap-4">
                            <div class="flex items-center gap-4 flex-1 min-w-0">
                                <div
                                    class="w-14 h-14 rounded-xl overflow-hidden border border-white/10 shrink-0 shadow-lg group-hover:border-primary/50 transition-colors">
                                    <img src="{{ $cover }}" class="w-full h-full object-cover">
                                </div>
                                <div class="flex-1 min-w-0">
                                    <h3
                                        class="text-xl font-bold text-white group-hover:text-primary transition-colors truncate mb-1">
                                        {{ $playlist->name }}
                                    </h3>
                                    <div class="flex items-center gap-2">
                                        <span
                                            class="bg-primary/10 text-primary text-[10px] font-black px-2 py-0.5 rounded uppercase tracking-wider">
                                            {{ $playlist->songs_count ?? 0 }} Songs
                                        </span>
                                    </div>
                                </div>
                            </div>
                            <div class="flex gap-2">
                                <a href="{{ route('playlists.edit', $playlist->id) }}"
                                    class="w-8 h-8 rounded-lg bg-white/5 hover:bg-green-500/20 hover:text-green-400 flex items-center justify-center transition-all text-white/40"
                                    title="Edit Playlist">
                                    <span class="material-symbols-outlined text-[18px]">edit</span>
                                </a>
                                <form action="{{ route('playlists.destroy', $playlist->id) }}" method="POST" class="inline">
                                    @csrf
                                    @method('DELETE')
                                    <button type="submit"
                                        class="w-8 h-8 rounded-lg bg-white/5 hover:bg-red-500/20 hover:text-red-400 flex items-center justify-center transition-all text-white/40"
                                        onclick="return confirm('Are you sure you want to delete this playlist?')"
                                        title="Delete Playlist">
                                        <span class="material-symbols-outlined text-[18px]">delete</span>
                                    </button>
                                </form>
                            </div>
                        </div>

                        <p class="text-sm text-white/40 line-clamp-2 leading-relaxed h-10">
                            {{ $playlist->description ?? 'No description provided.' }}
                        </p>

                        <div class="mt-2">
                            <a href="{{ route('playlists.show', $playlist->id) }}"
                                class="w-full bg-white/5 hover:bg-white/10 text-white text-[11px] font-black uppercase py-3 rounded-xl transition-all flex items-center justify-center gap-2 group-hover:border-primary/30 border border-transparent">
                                <span class="material-symbols-outlined text-[18px]">play_circle</span>
                                View Playlist
                            </a>
                        </div>

                        {{-- Hover decoration --}}
                        <div
                            class="absolute -inset-px rounded-2xl border border-primary/20 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none">
                        </div>
                    </div>
                @empty
                    <div class="col-span-full py-24 flex flex-col items-center justify-center text-center opacity-40">
                        <span class="material-symbols-outlined text-7xl mb-4">playlist_add</span>
                        <p class="text-xl font-bold">No playlists yet</p>
                        <p class="text-sm font-medium mt-2">Create your first playlist and start collecting your favorite
                            themes!</p>
                        <a href="{{ route('playlists.create') }}" class="mt-6 text-primary font-bold hover:underline">Create
                            Playlist</a>
                    </div>
                @endforelse
            </div>
        </div>
    </div>
@endsection