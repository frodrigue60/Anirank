@extends('layouts.app')

@section('content')
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="flex flex-col md:flex-row md:items-center md:justify-between mb-8 gap-4">
            <div>
                <a href="{{ route('admin.posts.index') }}"
                    class="text-blue-500 hover:text-blue-400 text-sm font-bold flex items-center mb-2 transition-colors">
                    <i class="fa-solid fa-arrow-left mr-2"></i> BACK TO POSTS
                </a>
                <h1 class="text-3xl font-bold text-white tracking-tight">Manage Songs</h1>
                @isset($post)
                    <p class="text-zinc-400 mt-1">Manage themes for <span
                            class="text-blue-400 font-semibold">{{ $post->title }}</span></p>
                @endisset
            </div>
            <div class="flex flex-wrap gap-3">
                @isset($post)
                    <a href="{{ route('admin.posts.songs.add', $post->id) }}"
                        class="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-xl transition-all shadow-lg shadow-blue-900/20 hover:scale-105 active:scale-95">
                        <i class="fa-solid fa-plus mr-2"></i> ADD SONG
                    </a>
                @endisset
            </div>
        </div>

        {{-- Table Card --}}
        <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl shadow-xl overflow-hidden">
            <div class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                    <thead>
                        <tr class="bg-zinc-950/50 border-b border-zinc-800">
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-center">ID
                            </th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Song Details
                            </th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-center">
                                Artists</th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Theme Info</th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-center">
                                Variants</th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-right">
                                Actions</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-zinc-800/50">
                        @isset($post)
                            @foreach ($post->songs as $song)
                                @php
                                    $song_name =
                                        $song->song_romaji ?? ($song->song_en ?? ($song->song_jp ?? 'Untitled Song'));
                                @endphp
                                <tr class="hover:bg-zinc-800/30 transition-colors group">
                                    <td class="px-6 py-4 text-sm font-mono text-zinc-500 text-center">#{{ $song->id }}</td>
                                    <td class="px-6 py-4">
                                        <div class="flex flex-col">
                                            <a href="{{ $song->post->url }}"
                                                class="text-sm font-bold text-white hover:text-blue-400 transition-colors">
                                                {{ $song_name }}
                                            </a>
                                            <div class="flex gap-2 mt-1">
                                                <span
                                                    class="text-[10px] bg-zinc-800 text-zinc-400 px-1.5 py-0.5 rounded font-bold uppercase tracking-wider border border-zinc-700">
                                                    {{ $song->type }}
                                                </span>
                                            </div>
                                        </div>
                                    </td>
                                    <td class="px-6 py-4 text-center">
                                        <span
                                            class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20">
                                            <i class="fa-solid fa-microphone mr-1.5"></i>
                                            {{ isset($song->artists) ? count($song->artists) : 0 }}
                                        </span>
                                    </td>
                                    <td class="px-6 py-4">
                                        <div class="flex flex-col space-y-1">
                                            <div class="flex items-center text-xs text-zinc-300">
                                                <i class="fa-solid fa-calendar-days mr-1.5 text-zinc-500 w-4"></i>
                                                {{ $song->season->name }} {{ $song->year->name }}
                                            </div>
                                            <div class="flex items-center text-xs text-zinc-300">
                                                <i class="fa-solid fa-tag mr-1.5 text-zinc-500 w-4"></i>
                                                {{ $song->slug ?? $song->type }}
                                            </div>
                                        </div>
                                    </td>
                                    <td class="px-6 py-4 text-center">
                                        <div class="flex items-center justify-center">
                                            <div
                                                class="relative flex items-center justify-center w-8 h-8 rounded-lg bg-zinc-800 text-zinc-300 font-bold text-xs border border-zinc-700">
                                                {{ count($song->songVariants) }}
                                            </div>
                                        </div>
                                    </td>
                                    <td class="px-6 py-4">
                                        <div class="flex items-center justify-end gap-2">
                                            @if (Auth::user()->isEditor() || Auth::user()->isAdmin())
                                                <a href="{{ route('admin.songs.edit', $song->id) }}"
                                                    class="p-2 bg-zinc-800 hover:bg-blue-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-blue-500"
                                                    title="Edit Song">
                                                    <i class="fa-solid fa-pencil"></i>
                                                </a>
                                                <a href="{{ route('admin.songs.variants', $song->id) }}"
                                                    class="p-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700"
                                                    title="Manage Variants">
                                                    <i class="fa-solid fa-list"></i>
                                                </a>
                                                <form action="{{ route('admin.songs.destroy', $song->id) }}" method="post"
                                                    class="inline">
                                                    @csrf
                                                    @method('DELETE')
                                                    <button type="submit"
                                                        onclick="return confirm('Are you sure you want to delete this song?')"
                                                        class="p-2 bg-zinc-800 hover:bg-red-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-red-500"
                                                        title="Delete Song">
                                                        <i class="fa-solid fa-trash-can"></i>
                                                    </button>
                                                </form>
                                            @endif
                                        </div>
                                    </td>
                                </tr>
                            @endforeach
                        @endisset
                    </tbody>
                </table>
            </div>

            <div class="bg-zinc-950/50 px-6 py-6 border-t border-zinc-800">
                <div class="flex justify-center italic text-zinc-500 text-xs">
                    Total: {{ isset($post) ? $post->songs->count() : 0 }} songs managed for this post
                </div>
            </div>
        </div>
    </div>
@endsection

@push('styles')
    <style>
        .overflow-x-auto::-webkit-scrollbar {
            height: 6px;
        }

        .overflow-x-auto::-webkit-scrollbar-track {
            background: transparent;
        }

        .overflow-x-auto::-webkit-scrollbar-thumb {
            background: #27272a;
            border-radius: 10px;
        }

        .overflow-x-auto::-webkit-scrollbar-thumb:hover {
            background: #3f3f46;
        }
    </style>
@endpush
