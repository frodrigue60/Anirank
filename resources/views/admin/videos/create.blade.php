@extends('layouts.app')

@section('content')
    <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header --}}
        <div class="mb-8">
            <a href="{{ route('admin.variants.videos', $song_variant->id) }}"
                class="text-blue-500 hover:text-blue-400 text-sm font-bold flex items-center mb-2 transition-colors">
                <i class="fa-solid fa-arrow-left mr-2"></i> BACK TO VIDEO MANAGEMENT
            </a>
            <h1 class="text-3xl font-bold text-white tracking-tight">Add Video Source</h1>
            <p class="text-zinc-400 mt-1">Configuring video for <span
                    class="text-blue-400 font-semibold">{{ $song_variant->song->song_romaji }}
                    ({{ $song_variant->slug }})</span></p>
        </div>

        {{-- Form Card --}}
        <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl shadow-xl overflow-hidden p-8">
            <form method="post" action="{{ route('admin.videos.store') }}" enctype="multipart/form-data" class="space-y-8">
                @csrf
                <input type="hidden" name="song_variant_id" value="{{ $song_variant->id }}">

                {{-- Source Options --}}
                <div class="space-y-6">
                    {{-- File Upload --}}
                    <div class="space-y-4 bg-zinc-950/30 p-6 rounded-2xl border border-zinc-800/50">
                        <label for="formFileBanner"
                            class="block text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center">
                            <i class="fa-solid fa-file-video mr-2 text-blue-500"></i> UPLOAD VIDEO FILE
                        </label>
                        <div
                            class="mt-2 flex justify-center px-6 pt-5 pb-6 border-2 border-zinc-800 border-dashed rounded-2xl hover:border-blue-500/50 transition-colors group">
                            <div class="space-y-1 text-center">
                                <i
                                    class="fa-solid fa-cloud-arrow-up text-3xl text-zinc-600 mb-3 group-hover:text-blue-400 transition-colors"></i>
                                <div class="flex text-sm text-zinc-400">
                                    <label for="formFileBanner"
                                        class="relative cursor-pointer bg-transparent rounded-md font-bold text-blue-500 hover:text-blue-400 transition-colors">
                                        <span>Select a file</span>
                                        <input id="formFileBanner" name="video" type="file" class="sr-only"
                                            accept="video/mp4,video/webm">
                                    </label>
                                    <p class="pl-1">or drag and drop</p>
                                </div>
                                <p class="text-[10px] text-zinc-500">MP4 or WEBM up to 100MB</p>
                            </div>
                        </div>
                    </div>

                    <div class="relative py-4 flex items-center">
                        <div class="flex-grow border-t border-zinc-800"></div>
                        <span
                            class="flex-shrink mx-4 text-zinc-600 text-[10px] font-bold uppercase tracking-widest">OR</span>
                        <div class="flex-grow border-t border-zinc-800"></div>
                    </div>

                    {{-- Embed Code --}}
                    <div class="space-y-4 bg-zinc-950/30 p-6 rounded-2xl border border-zinc-800/50">
                        <label for="embed"
                            class="block text-sm font-bold text-zinc-400 uppercase tracking-widest flex items-center">
                            <i class="fa-solid fa-code mr-2 text-emerald-500"></i> EMBED CODE / LINK
                        </label>
                        <input type="text" name="embed" id="embed" value="{{ old('embed') }}"
                            class="block w-full bg-zinc-900 border border-zinc-800 text-white rounded-xl px-4 py-3 focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all text-sm"
                            placeholder="Paste embed code or video URL source...">
                        <p class="text-[10px] text-zinc-500 italic">Use this if you are hosting the video on an external
                            service.</p>
                    </div>
                </div>

                {{-- Trigger --}}
                <div class="pt-4">
                    <button
                        class="w-full bg-blue-600 hover:bg-blue-500 text-white font-bold py-4 px-6 rounded-2xl transition-all shadow-lg shadow-blue-900/20 active:scale-[0.98] flex items-center justify-center gap-2 text-sm uppercase tracking-widest">
                        <i class="fa-solid fa-save"></i>
                        SAVE VIDEO SOURCE
                    </button>
                </div>
            </form>
        </div>
    </div>
@endsection
