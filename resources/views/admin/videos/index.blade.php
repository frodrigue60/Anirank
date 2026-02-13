@extends('layouts.app')

@section('content')
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="flex flex-col md:flex-row md:items-center md:justify-between mb-8 gap-4">
            <div>
                <a href="{{ route('admin.songs.variants', $song_variant->song_id) }}"
                    class="text-blue-500 hover:text-blue-400 text-sm font-bold flex items-center mb-2 transition-colors">
                    <i class="fa-solid fa-arrow-left mr-2"></i> BACK TO VARIANTS
                </a>
                <h1 class="text-3xl font-bold text-white tracking-tight">Manage Video</h1>
                <p class="text-zinc-400 mt-1">Video configuration for <span
                        class="text-blue-400 font-semibold">{{ $song_variant->song->song_romaji }}
                        ({{ $song_variant->slug }})</span></p>
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
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Type</th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Source/Content
                            </th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-right">
                                Actions</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-zinc-800/50">
                        @php
                            $video = $song_variant->video;
                        @endphp
                        @if ($video)
                            <tr class="hover:bg-zinc-800/30 transition-colors group">
                                <td class="px-6 py-4 text-sm font-mono text-zinc-500 text-center">#{{ $video->id }}</td>
                                <td class="px-6 py-4">
                                    <span
                                        class="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-bold uppercase bg-zinc-800 text-zinc-300 border border-zinc-700">
                                        {{ $video->type }}
                                    </span>
                                </td>
                                <td class="px-6 py-4">
                                    <div class="text-sm text-zinc-300 font-mono truncate max-w-md">
                                        {{ $video->video_src ?? $video->embed_code }}
                                    </div>
                                </td>
                                <td class="px-6 py-4 text-right">
                                    <div class="flex items-center justify-end gap-2">
                                        <a href="{{ route('admin.videos.edit', $video->id) }}"
                                            class="p-2 bg-zinc-800 hover:bg-blue-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-blue-500">
                                            <i class="fa-solid fa-pencil"></i>
                                        </a>
                                        <a href="{{ route('admin.videos.show', $video->id) }}"
                                            class="p-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700">
                                            <i class="fa-solid fa-eye"></i>
                                        </a>
                                        <form action="{{ route('admin.videos.destroy', $video->id) }}" method="post"
                                            class="inline">
                                            @csrf
                                            @method('DELETE')
                                            <button type="submit" onclick="return confirm('Delete video?')"
                                                class="p-2 bg-zinc-800 hover:bg-red-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-red-500">
                                                <i class="fa-solid fa-trash-can"></i>
                                            </button>
                                        </form>
                                    </div>
                                </td>
                            </tr>
                        @else
                            <tr>
                                <td colspan="4" class="px-6 py-12 text-center">
                                    <p class="text-zinc-500 italic">No video configured for this variant.</p>
                                    <a href="{{ route('admin.variants.videos.add', $song_variant->id) }}"
                                        class="mt-4 inline-flex items-center text-blue-500 hover:text-blue-400 font-bold">
                                        <i class="fa-solid fa-plus mr-2"></i> ADD VIDEO NOW
                                    </a>
                                </td>
                            </tr>
                        @endif
                    </tbody>
                </table>
            </div>
        </div>
    </div>
@endsection
