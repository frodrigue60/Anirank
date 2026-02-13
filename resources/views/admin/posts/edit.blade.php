@extends('layouts.app')

@section('content')
    <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="mb-8">
            <a href="{{ route('admin.posts.index') }}"
                class="text-blue-500 hover:text-blue-400 text-sm font-bold flex items-center mb-2 transition-colors">
                <i class="fa-solid fa-arrow-left mr-2"></i> BACK TO POSTS
            </a>
            <h1 class="text-3xl font-bold text-white tracking-tight">Edit Post</h1>
            <p class="text-zinc-400 mt-1">Updating entry for <span
                    class="text-blue-400 font-semibold">{{ $post->title }}</span></p>
        </div>

        {{-- Form Card --}}
        <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl shadow-xl overflow-hidden p-8">
            <form method="post" action="{{ route('admin.posts.update', $post->id) }}" enctype="multipart/form-data"
                class="space-y-8">
                @csrf
                @method('PUT')

                {{-- Basic Information --}}
                <div class="space-y-6">
                    <h3 class="text-xs font-bold text-zinc-500 uppercase tracking-widest flex items-center">
                        <i class="fa-solid fa-circle-info mr-2 text-blue-500"></i> BASIC INFORMATION
                    </h3>

                    <div class="space-y-4">
                        <div class="space-y-2">
                            <label for="titleAnime"
                                class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Post Title</label>
                            <input type="text" name="title" id="titleAnime" required
                                value="{{ old('title', $post->title) }}"
                                class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12"
                                placeholder="e.g. Shingeki no Kyojin">
                        </div>

                        <div class="space-y-2">
                            <label for="description"
                                class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Description</label>
                            <textarea name="description" id="description" rows="5"
                                class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm"
                                placeholder="Enter anime synopsis...">{{ old('description', $post->description) }}</textarea>
                        </div>
                    </div>
                </div>

                {{-- Status & Metadata --}}
                <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
                    @if (Auth::user()->isAdmin() || Auth::user()->isEditor())
                        <div class="space-y-2">
                            <label for="statusId"
                                class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Publication
                                Status</label>
                            <select name="postStatus" id="statusId"
                                class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12">
                                <option value="">Select a status</option>
                                @foreach ($postStatus as $item)
                                    <option value="{{ $item['value'] }}"
                                        {{ old('postStatus', $post->status) == $item['value'] ? 'selected' : '' }}>
                                        {{ $item['name'] }}
                                    </option>
                                @endforeach
                            </select>
                        </div>
                    @endif

                    <div class="space-y-2">
                        <label for="year"
                            class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Release Year</label>
                        <select name="year" id="year" required
                            class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12">
                            @foreach ($years as $year)
                                <option value="{{ $year->id }}"
                                    {{ old('year', $post->year_id) == $year->id ? 'selected' : '' }}>
                                    {{ $year->name }}
                                </option>
                            @endforeach
                        </select>
                    </div>

                    <div class="space-y-2">
                        <label for="season"
                            class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Season</label>
                        <select name="season" id="season" required
                            class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12">
                            @foreach ($seasons as $season)
                                <option value="{{ $season->id }}"
                                    {{ old('season', $post->season_id) == $season->id ? 'selected' : '' }}>
                                    {{ $season->name }}
                                </option>
                            @endforeach
                        </select>
                    </div>
                </div>

                {{-- Visuals --}}
                <div class="space-y-6 bg-zinc-950/30 p-6 rounded-3xl border border-zinc-800/50">
                    <h3 class="text-xs font-bold text-zinc-500 uppercase tracking-widest flex items-center">
                        <i class="fa-solid fa-image mr-2 text-blue-500"></i> MEDIA ASSETS
                    </h3>

                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {{-- Thumbnail --}}
                        <div class="space-y-4">
                            <div class="space-y-2">
                                <label for="thumbnail_src" class="block text-sm font-bold text-zinc-400">Thumbnail
                                    URL</label>
                                @if ($post->image)
                                    <div
                                        class="w-20 h-28 bg-zinc-800 rounded-xl overflow-hidden mb-2 border border-zinc-700">
                                        <img src="{{ $post->image }}" alt="" class="w-full h-full object-cover">
                                    </div>
                                @endif
                                <input type="text" name="thumbnail_src" id="thumbnail_src"
                                    value="{{ old('thumbnail_src', $post->thumbnail_src) }}"
                                    class="block w-full bg-zinc-900 border border-zinc-800 text-white rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm"
                                    placeholder="https://...">
                            </div>
                            <div class="space-y-2">
                                <label for="formFile" class="block text-sm font-bold text-zinc-400">OR Upload
                                    Thumbnail</label>
                                <input type="file" name="file" id="formFile"
                                    class="block w-full text-xs text-zinc-500 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-xs file:font-black file:bg-zinc-800 file:text-zinc-300 hover:file:bg-zinc-700 transition-all">
                            </div>
                        </div>

                        {{-- Banner --}}
                        <div class="space-y-4">
                            <div class="space-y-2">
                                <label for="banner_src" class="block text-sm font-bold text-zinc-400">Banner URL</label>
                                <input type="text" name="banner_src" id="banner_src"
                                    value="{{ old('banner_src', $post->banner_src) }}"
                                    class="block w-full bg-zinc-900 border border-zinc-800 text-white rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm"
                                    placeholder="https://...">
                            </div>
                            <div class="space-y-2">
                                <label for="formFileBanner" class="block text-sm font-bold text-zinc-400">OR Upload
                                    Banner</label>
                                <input type="file" name="banner" id="formFileBanner"
                                    class="block w-full text-xs text-zinc-500 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-xs file:font-black file:bg-zinc-800 file:text-zinc-300 hover:file:bg-zinc-700 transition-all">
                            </div>
                        </div>
                    </div>
                </div>

                {{-- Action --}}
                <div class="pt-4 flex flex-col sm:flex-row gap-4">
                    <button
                        class="flex-1 bg-blue-600 hover:bg-blue-500 text-white font-bold py-4 px-6 rounded-2xl transition-all shadow-lg shadow-blue-900/20 active:scale-[0.98] flex items-center justify-center gap-2 text-sm uppercase tracking-widest">
                        <i class="fa-solid fa-check-double"></i>
                        SAVE CHANGES
                    </button>
                    <a href="{{ route('admin.posts.index') }}"
                        class="flex-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 font-bold py-4 px-6 rounded-2xl transition-all flex items-center justify-center gap-2 text-sm uppercase tracking-widest border border-zinc-700">
                        CANCEL
                    </a>
                </div>
            </form>
        </div>
    </div>
@endsection
