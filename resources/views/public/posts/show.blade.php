@extends('layouts.app')

@section('meta')
    @php
        if ($post->thumbnail != null && Storage::disk('public')->exists($post->thumbnail)) {
            $thumbnail_url = Storage::url($post->thumbnail);
        } else {
            $thumbnail_url = $post->thumbnail_src;
        }

        if ($post->banner != null && Storage::disk('public')->exists($post->banner)) {
            $banner_url = Storage::url($post->banner);
        } else {
            $banner_url = $post->banner_src;
        }
    @endphp
    <title>{{ $post->title }} - Themes List</title>
    <meta name="title" content="{{ $post->title }} Themes - Anirank">
    <meta name="description" content="Explore and listen to the openings and endings of {{ $post->title }}.">
    <meta name="robots" content="index, follow, max-image-preview:standard">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta property="og:type" content="article">
    <meta property="og:title" content="{{ $post->title }} Themes">
    <meta property="og:image" content="{{ $thumbnail_url }}">
    <meta property="og:url" content="{{ url()->current() }}">
@endsection

@section('content')
    {{-- Hero Section --}}
    <div class="relative w-full h-[320px] md:h-[500px] overflow-hidden group">
        {{-- Banner Background --}}
        <div class="absolute inset-0 transition-transform duration-1000 group-hover:scale-110">
            <img src="{{ $banner_url }}" alt="{{ $post->title }}"
                class="w-full h-full object-cover saturate-[0.8] brightness-[0.6]">

            {{-- Scrims & Gradients --}}
            <div class="absolute inset-0 bg-gradient-to-t from-background via-background/90 to-transparent opacity-90">
            </div>
            <div class="absolute inset-0 bg-gradient-to-b from-background/80 via-transparent to-transparent h-1/3"></div>
            {{-- Top scrim for Navbar --}}
            <div class="absolute inset-0 bg-gradient-to-r from-background via-transparent to-background/40"></div>
        </div>

        <div class="relative max-w-[1440px] mx-auto h-full flex flex-col justify-end px-4 md:px-8 pb-12">
            <div class="flex flex-col md:flex-row items-end gap-8 md:gap-12">
                {{-- Cover Image in Sidebar position --}}
                <div class="hidden md:block w-56 h-auto aspect-[2/3] -mb-24 z-20 shrink-0 relative group/poster">
                    <div
                        class="absolute -inset-1 bg-primary/20 rounded-[22px] blur-xl opacity-0 group-hover/poster:opacity-100 transition-opacity duration-500">
                    </div>
                    <img src="{{ $thumbnail_url }}" alt="{{ $post->title }}"
                        class="relative w-full h-full object-cover rounded-2xl shadow-[0_30px_60px_rgba(0,0,0,0.9)] border border-white/20 ring-1 ring-white/10">
                </div>

                {{-- Basic Info --}}
                <div class="flex-1 mb-0 md:mb-4">
                    <div class="flex flex-wrap items-center gap-3 mb-6">
                        <div
                            class="flex items-center gap-1 bg-white/5 backdrop-blur-md border border-white/10 px-2.5 py-1 rounded-lg">
                            <span class="text-primary text-[10px] font-black uppercase tracking-widest">
                                {{ $post->format->name ?? 'TV' }}
                            </span>
                        </div>
                        <div class="w-1 h-1 bg-white/20 rounded-full"></div>
                        <div class="flex items-center gap-1.5 text-white/50 text-sm font-bold tracking-wide">
                            <span class="material-symbols-outlined text-[18px]">calendar_today</span>
                            {{ $post->season->name ?? '' }} {{ $post->year->name ?? '' }}
                        </div>
                    </div>
                    <h1
                        class="text-4xl md:text-7xl font-black text-white leading-[1.1] md:leading-[1.05] tracking-tight mb-4 drop-shadow-[0_10px_10px_rgba(0,0,0,0.5)]">
                        <span class="bg-gradient-to-b from-white to-white/70 bg-clip-text text-transparent">
                            {{ $post->title }}
                        </span>
                    </h1>
                </div>
            </div>

            {{-- Staff Actions (Fixed for Tailwind) --}}
            @auth
                @if (Auth::User()->isStaff())
                    <div
                        class="flex items-center gap-2 mb-4 md:mb-6 backdrop-blur-xl bg-white/5 border border-white/10 p-2 rounded-2xl h-fit">
                        <a href="{{ route('admin.posts.edit', $post->id) }}" title="Edit"
                            class="w-10 h-10 flex items-center justify-center bg-white/5 hover:bg-emerald-500/20 text-emerald-400 rounded-xl transition-all border border-white/5">
                            <span class="material-symbols-outlined text-[20px]">edit</span>
                        </a>
                        <a href="{{ route('admin.posts.songs', $post->id) }}" title="Manage Songs"
                            class="w-10 h-10 flex items-center justify-center bg-white/5 hover:bg-primary/20 text-primary rounded-xl transition-all border border-white/5">
                            <span class="material-symbols-outlined text-[20px]">list</span>
                        </a>
                        <a href="{{ route('admin.posts.force.update', $post->id) }}" title="Force Update"
                            class="w-10 h-10 flex items-center justify-center bg-white/5 hover:bg-amber-500/20 text-amber-500 rounded-xl transition-all border border-white/5">
                            <span class="material-symbols-outlined text-[20px]">sync</span>
                        </a>
                        <form action="{{ route('admin.posts.destroy', $post->id) }}" method="post" class="m-0">
                            @csrf
                            @method('DELETE')
                            <button type="submit" title="Delete"
                                class="w-10 h-10 flex items-center justify-center bg-white/5 hover:bg-red-500/20 text-red-500 rounded-xl transition-all border border-white/5">
                                <span class="material-symbols-outlined text-[20px]">delete</span>
                            </button>
                        </form>
                    </div>
                @endif
            @endauth
        </div>
    </div>
    </div>

    {{-- Content Layout --}}
    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10 md:py-20 mt-4 md:mt-0">
        <div class="flex flex-col lg:flex-row gap-12">

            {{-- Sidebar --}}
            <aside class="w-full lg:w-[300px] shrink-0">
                <div class="flex flex-col gap-8 sticky top-32">
                    {{-- Mobile Thumbnail Only --}}
                    <div class="md:hidden w-full flex justify-center mb-6">
                        <img src="{{ $thumbnail_url }}" alt="{{ $post->title }}"
                            class="w-48 aspect-[2/3] object-cover rounded-2xl shadow-2xl border border-white/10">
                    </div>

                    {{-- Metadata Cards --}}
                    <div class="bg-surface-dark/30 rounded-3xl border border-white/5 p-6 backdrop-blur-sm">
                        <h4
                            class="text-[11px] uppercase font-black text-white/30 tracking-[0.2em] mb-6 flex items-center gap-2">
                            <div class="w-1 h-3 bg-primary rounded-full"></div>
                            Anime Details
                        </h4>

                        <div class="flex flex-col gap-6">
                            @isset($post->format)
                                <div class="flex flex-col gap-1.5">
                                    <span class="text-xs font-black text-white/20 uppercase tracking-widest">Format</span>
                                    <p class="text-sm text-white font-bold">{{ $post->format->name }}</p>
                                </div>
                            @endisset

                            <div class="flex flex-col gap-1.5">
                                <span class="text-xs font-black text-white/20 uppercase tracking-widest">Release</span>
                                <p class="text-sm text-white font-bold">
                                    {{ $post->season->name ?? 'N/A' }} {{ $post->year->name ?? '' }}
                                </p>
                            </div>

                            @if($post->producers->count() > 0)
                                <div class="flex flex-col gap-3">
                                    <span class="text-xs font-black text-white/20 uppercase tracking-widest">Producers</span>
                                    <div class="flex flex-col gap-2">
                                        @foreach ($post->producers as $item)
                                            <a href="{{ route('studios.show', $item->slug) }}"
                                                class="flex items-center gap-2 text-sm text-white hover:text-primary transition-colors font-bold group/studio">
                                                <span
                                                    class="material-symbols-outlined text-[16px] text-white/20 group-hover/studio:text-primary">business</span>
                                                {{ $item->name }}
                                            </a>
                                        @endforeach
                                    </div>
                                </div>
                            @endif

                            @if($post->studios->count() > 0)
                                <div class="flex flex-col gap-3">
                                    <span class="text-xs font-black text-white/20 uppercase tracking-widest">Studios</span>
                                    <div class="flex flex-col gap-2">
                                        @foreach ($post->studios as $item)
                                            <a href="{{ route('studios.show', $item->slug) }}"
                                                class="flex items-center gap-2 text-sm text-white hover:text-primary transition-colors font-bold group/studio">
                                                <span
                                                    class="material-symbols-outlined text-[16px] text-white/20 group-hover/studio:text-primary">business</span>
                                                {{ $item->name }}
                                            </a>
                                        @endforeach
                                    </div>
                                </div>
                            @endif

                            @if($post->externalLinks->count() > 0)
                                <div class="flex flex-col gap-3">
                                    <span class="text-xs font-black text-white/20 uppercase tracking-widest">External
                                        Links</span>
                                    <div class="flex flex-wrap gap-2">
                                        @foreach ($post->externalLinks as $item)
                                            <a href="{{ $item->url }}" target="_blank"
                                                class="h-9 px-4 flex items-center gap-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-xl text-xs text-white/80 hover:text-white font-bold transition-all">
                                                <span class="material-symbols-outlined text-[16px]">link</span>
                                                {{ $item->name }}
                                            </a>
                                        @endforeach
                                    </div>
                                </div>
                            @endif
                        </div>
                    </div>
                </div>
            </aside>

            {{-- Main Content --}}
            <main class="flex-1 min-w-0">
                <div class="flex flex-col gap-14">

                    {{-- Synopsis Section --}}
                    <section>
                        <h2 class="text-2xl font-black text-white mb-6 flex items-center gap-4">
                            Synopsis
                            <div class="flex-1 h-px bg-white/5"></div>
                        </h2>
                        <div class="relative group">
                            <div
                                class="absolute -inset-4 bg-primary/5 rounded-[2rem] opacity-0 group-hover:opacity-100 transition-opacity duration-500">
                            </div>
                            <div
                                class="relative prose prose-invert prose-p:text-white/60 prose-p:leading-relaxed prose-p:text-lg max-w-none">
                                {!! $post->description !!}
                            </div>
                        </div>
                    </section>

                    {{-- Openings Section --}}
                    <section>
                        <div class="flex justify-between items-center mb-8">
                            <h2 class="text-2xl font-black text-white flex items-center gap-4">
                                Openings
                                <div class="flex-1 h-px bg-white/5 hidden sm:block w-20"></div>
                            </h2>
                            <span
                                class="text-[10px] uppercase font-black px-3 py-1 bg-primary/10 text-primary border border-primary/20 rounded-full tracking-widest leading-none">
                                {{ $openings->count() ?? 0 }} Themes
                            </span>
                        </div>

                        <div class="grid grid-cols-1 gap-4">
                            @forelse ($openings->sortBy('theme_num') as $song)
                                @include('partials.posts.show.song-card-premium')
                            @empty
                                <div
                                    class="flex flex-col items-center justify-center py-12 bg-surface-dark/20 rounded-3xl border border-white/5">
                                    <span class="material-symbols-outlined text-4xl text-white/10 mb-2">music_off</span>
                                    <p class="text-white/20 font-bold uppercase tracking-widest text-xs">No openings recorded
                                    </p>
                                </div>
                            @endforelse
                        </div>
                    </section>

                    {{-- Endings Section --}}
                    <section>
                        <div class="flex justify-between items-center mb-8">
                            <h2 class="text-2xl font-black text-white flex items-center gap-4">
                                Endings
                                <div class="flex-1 h-px bg-white/5 hidden sm:block w-20"></div>
                            </h2>
                            <span
                                class="text-[10px] uppercase font-black px-3 py-1 bg-secondary/10 text-secondary border border-secondary/20 rounded-full tracking-widest leading-none">
                                {{ $endings->count() ?? 0 }} Themes
                            </span>
                        </div>

                        <div class="grid grid-cols-1 gap-4">
                            @forelse ($endings->sortBy('theme_num') as $song)
                                @include('partials.posts.show.song-card-premium')
                            @empty
                                <div
                                    class="flex flex-col items-center justify-center py-12 bg-surface-dark/20 rounded-3xl border border-white/5">
                                    <span class="material-symbols-outlined text-4xl text-white/10 mb-2">music_off</span>
                                    <p class="text-white/20 font-bold uppercase tracking-widest text-xs">No endings recorded</p>
                                </div>
                            @endforelse
                        </div>
                    </section>
                </div>
            </main>

        </div>
    </div>
@endsection