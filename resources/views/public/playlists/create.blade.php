@extends('layouts.app')

@section('meta')
    <title>Create Playlist | {{ config('app.name') }}</title>
@endsection

@section('content')
    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10 md:py-8 flex flex-col gap-12">
        <div class="max-w-2xl mx-auto w-full">
            {{-- Header --}}
            <div class="flex flex-col gap-4 mb-8 text-center">
                <div class="mx-auto w-16 h-16 bg-primary/10 rounded-2xl flex items-center justify-center mb-2">
                    <span class="material-symbols-outlined text-primary text-4xl">playlist_add</span>
                </div>
                <h1 class="text-3xl font-black tracking-tight text-white">Create New Playlist</h1>
                <p class="text-white/40 text-sm">Organize your favorite themes into custom collections.</p>
            </div>

            {{-- Form Card --}}
            <div class="bg-surface-dark/30 p-8 rounded-2xl border border-white/5 shadow-2xl backdrop-blur-md">
                <form action="{{ route('playlists.store') }}" method="post" class="flex flex-col gap-6">
                    @csrf
                    @method('POST')

                    {{-- Name --}}
                    <div class="relative group">
                        <label for="name"
                            class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-focus-within:text-primary transition-colors">
                            Playlist Name
                        </label>
                        <div class="relative">
                            <span
                                class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-white/30 group-focus-within:text-primary transition-colors text-xl">label</span>
                            <input type="text" id="name" name="name" required placeholder="e.g. My Favorite OPs 2024"
                                class="w-full bg-surface-darker border border-white/10 rounded-lg py-3 pl-10 pr-4 text-sm text-white focus:outline-none focus:border-primary/50 transition-all placeholder:text-white/20">
                        </div>
                        @error('name') <span class="text-xs text-red-500 mt-1 block">{{ $message }}</span> @enderror
                    </div>

                    {{-- Description --}}
                    <div class="relative group">
                        <label for="description"
                            class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-focus-within:text-primary transition-colors">
                            Description (Optional)
                        </label>
                        <div class="relative">
                            <textarea id="description" name="description" rows="4" maxlength="255"
                                placeholder="Tell us what this playlist is about..."
                                class="w-full bg-surface-darker border border-white/10 rounded-lg p-4 text-sm text-white focus:outline-none focus:border-primary/50 transition-all placeholder:text-white/20 resize-none"></textarea>
                        </div>
                        @error('description') <span class="text-xs text-red-500 mt-1 block">{{ $message }}</span> @enderror
                    </div>

                    {{-- Actions --}}
                    <div class="flex items-center gap-4 pt-4 border-t border-white/5 mt-2">
                        <a href="{{ route('playlists.index') }}"
                            class="flex-1 bg-white/5 hover:bg-white/10 text-white text-xs font-black uppercase py-4 rounded-xl transition-all text-center">
                            Cancel
                        </a>
                        <button type="submit"
                            class="flex-[2] bg-primary hover:bg-primary-light text-white text-xs font-black uppercase py-4 rounded-xl transition-all shadow-lg shadow-primary/20">
                            Create Playlist
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>
@endsection