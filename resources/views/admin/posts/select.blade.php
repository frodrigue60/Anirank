@extends('layouts.app')

@section('content')
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header --}}
        <div class="mb-8">
            <a href="{{ route('admin.posts.index') }}"
                class="text-blue-500 hover:text-blue-400 text-sm font-bold flex items-center mb-2 transition-colors">
                <i class="fa-solid fa-arrow-left mr-2"></i> BACK TO POSTS
            </a>
            <h1 class="text-3xl font-bold text-white tracking-tight">Select Anime</h1>
            <p class="text-zinc-400 mt-1">Found <span class="text-blue-400 font-semibold">{{ count($posts) }}</span> matching
                results from Anilist. Choose one to import.</p>
        </div>

        {{-- Grid --}}
        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-6">
            @foreach ($posts as $post)
                @php
                    $title = $post->title->romaji ?? ($post->title->english ?? 'Untitled');
                    $anirank_id = ['id' => $post->id];
                @endphp
                <div
                    class="group relative flex flex-col bg-zinc-900 border border-zinc-800 rounded-2xl overflow-hidden hover:border-blue-500/50 hover:shadow-2xl hover:shadow-blue-900/10 transition-all hover:scale-[1.03]">
                    {{-- Cover Image --}}
                    <div class="aspect-[3/4] relative overflow-hidden">
                        <img class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                            src="{{ $post->coverImage->extraLarge }}" alt="{{ $title }}">

                        {{-- Overlay --}}
                        <div
                            class="absolute inset-0 bg-gradient-to-t from-zinc-950 via-zinc-950/20 to-transparent opacity-60">
                        </div>

                        {{-- Quick Info --}}
                        <div class="absolute top-2 left-2 flex flex-col gap-1">
                            @if ($post->format)
                                <span
                                    class="px-2 py-0.5 bg-black/60 backdrop-blur-md text-[9px] font-black text-white rounded uppercase tracking-widest border border-white/10">
                                    {{ $post->format }}
                                </span>
                            @endif
                        </div>

                        {{-- Action Overlay --}}
                        <div
                            class="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity bg-zinc-950/40 backdrop-blur-[2px]">
                            <a href="{{ route('admin.posts.get.by.id', $anirank_id) }}"
                                class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold rounded-xl shadow-lg transition-transform active:scale-90">
                                IMPORT THIS
                            </a>
                        </div>
                    </div>

                    {{-- Title Section --}}
                    <div class="p-3">
                        <h3
                            class="text-sm font-bold text-white line-clamp-2 leading-tight group-hover:text-blue-400 transition-colors">
                            {{ $title }}
                        </h3>
                        <div class="mt-2 flex items-center gap-2">
                            <span class="text-[10px] text-zinc-500 font-bold uppercase tracking-tighter">
                                {{ $post->season }} {{ $post->seasonYear }}
                            </span>
                        </div>
                    </div>

                    {{-- Invisible Link for Accessibility/SEO --}}
                    <a href="{{ route('admin.posts.get.by.id', $anirank_id) }}" class="absolute inset-0 z-10"
                        aria-label="Select {{ $title }}"></a>
                </div>
            @endforeach
        </div>
    </div>
@endsection
