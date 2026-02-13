@extends('layouts.app')

@section('content')
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="flex flex-col md:flex-row md:items-center md:justify-between mb-8 gap-4">
            <div>
                <a href="{{ route('admin.posts.songs', $song->post_id) }}"
                    class="text-blue-500 hover:text-blue-400 text-sm font-bold flex items-center mb-2 transition-colors">
                    <i class="fa-solid fa-arrow-left mr-2"></i> BACK TO SONGS
                </a>
                <h1 class="text-3xl font-bold text-white tracking-tight">Song Variants</h1>
                <p class="text-zinc-400 mt-1">Manage versions for <span
                        class="text-blue-400 font-semibold">{{ $song->song_romaji ?? 'this song' }}</span></p>
            </div>
            <div class="flex flex-wrap gap-3">
                <a href="{{ route('admin.songs.variants.add', $song->id) }}"
                    class="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-xl transition-all shadow-lg shadow-blue-900/20 hover:scale-105 active:scale-95">
                    <i class="fa-solid fa-plus mr-2"></i> CREATE VARIANT
                </a>
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
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Variant Details
                            </th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-center">
                                Video Status</th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-right">
                                Actions</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-zinc-800/50">
                        @isset($song_variants)
                            @foreach ($song_variants as $variant)
                                <tr class="hover:bg-zinc-800/30 transition-colors group">
                                    <td class="px-6 py-4 text-sm font-mono text-zinc-500 text-center">#{{ $variant->id }}</td>
                                    <td class="px-6 py-4">
                                        <div class="flex flex-col">
                                            <span class="text-sm font-bold text-white">{{ $variant->slug ?? 'Standard' }}</span>
                                            <span class="text-[10px] text-zinc-500 font-mono mt-1">{{ $variant->id }}</span>
                                        </div>
                                    </td>
                                    <td class="px-6 py-4 text-center">
                                        @if ($variant->video)
                                            <a href="{{ route('admin.variants.videos', $variant->id) }}"
                                                class="inline-flex items-center px-3 py-1 bg-emerald-500/10 text-emerald-500 text-xs font-bold rounded-lg border border-emerald-500/20 hover:bg-emerald-500 hover:text-white transition-all">
                                                <i class="fa-solid fa-play mr-2"></i> CONFIGURE VIDEO
                                            </a>
                                        @else
                                            <span
                                                class="inline-flex items-center px-3 py-1 bg-zinc-800 text-zinc-500 text-xs font-bold rounded-lg border border-zinc-700">
                                                <i class="fa-solid fa-circle-exclamation mr-2"></i> NO VIDEO
                                            </span>
                                        @endif
                                    </td>
                                    <td class="px-6 py-4">
                                        <div class="flex items-center justify-end gap-2">
                                            <a href="{{ route('admin.variants.edit', $variant->id) }}"
                                                class="p-2 bg-zinc-800 hover:bg-blue-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-blue-500"
                                                title="Edit Variant">
                                                <i class="fa-solid fa-pencil"></i>
                                            </a>
                                            <a href="{{ route('admin.variants.videos.add', $variant->id) }}"
                                                class="p-2 bg-zinc-800 hover:bg-emerald-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-emerald-500"
                                                title="Add Video">
                                                <i class="fa-solid fa-plus"></i>
                                            </a>
                                            <form action="{{ route('admin.variants.destroy', $variant->id) }}" method="post"
                                                class="inline">
                                                @csrf
                                                @method('DELETE')
                                                <button type="submit" onclick="return confirm('Delete this variant?')"
                                                    class="p-2 bg-zinc-800 hover:bg-red-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-red-500"
                                                    title="Delete Variant">
                                                    <i class="fa-solid fa-trash-can"></i>
                                                </button>
                                            </form>
                                        </div>
                                    </td>
                                </tr>
                            @endforeach
                        @endisset
                    </tbody>
                </table>
            </div>

            <div class="bg-zinc-950/50 px-6 py-4 border-t border-zinc-800">
                <div class="flex justify-center italic text-zinc-500 text-xs">
                    Managing variants for song ID: {{ $song->id }}
                </div>
            </div>
        </div>
    </div>
@endsection
