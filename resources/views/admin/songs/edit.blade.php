@extends('layouts.app')

@section('content')
    <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header --}}
        <div class="mb-8">
            <a href="{{ route('admin.posts.songs', $song->post_id) }}"
                class="text-blue-500 hover:text-blue-400 text-sm font-bold flex items-center mb-2 transition-colors">
                <i class="fa-solid fa-arrow-left mr-2"></i> BACK TO MANAGE SONGS
            </a>
            <h1 class="text-3xl font-bold text-white tracking-tight">Edit Song</h1>
            <p class="text-zinc-400 mt-1">Updating entry for <span
                    class="text-blue-400 font-semibold">{{ $song->song_romaji ?? ($song->song_en ?? 'Theme') }}</span></p>
        </div>

        {{-- Form Card --}}
        <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl shadow-xl overflow-hidden p-8">
            <form method="post" action="{{ route('admin.songs.update', $song->id) }}" enctype="multipart/form-data"
                class="space-y-6">
                @method('put')
                @csrf

                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {{-- OP/ED Number --}}
                    <div class="space-y-2">
                        <label for="theme_num" class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">OP/ED
                            Number</label>
                        <input type="number" name="theme_num" id="theme_num" value="{{ $song->theme_num }}"
                            class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12"
                            placeholder="e.g. 1">
                    </div>

                    {{-- Type --}}
                    <div class="space-y-2">
                        <label for="type" class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Theme
                            Type</label>
                        <select name="type" id="type"
                            class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12">
                            @foreach ($types as $item)
                                <option value="{{ $item['value'] }}" {{ $song->type == $item['value'] ? 'selected' : '' }}>
                                    {{ $item['name'] }}
                                </option>
                            @endforeach
                        </select>
                    </div>
                </div>

                {{-- Song Names --}}
                <div class="space-y-6 bg-zinc-950/30 p-6 rounded-2xl border border-zinc-800/50">
                    <h3 class="text-xs font-bold text-zinc-500 uppercase tracking-widest flex items-center">
                        <i class="fa-solid fa-music mr-2"></i> SONG METADATA
                    </h3>

                    <div class="space-y-4">
                        <div class="space-y-2">
                            <label for="songRomaji" class="block text-sm font-bold text-zinc-400">Song Name (Romaji)</label>
                            <input type="text" name="song_romaji" id="songRomaji" value="{{ $song->song_romaji }}"
                                class="block w-full bg-zinc-900 border border-zinc-800 text-white rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm"
                                placeholder="Enter romaji title...">
                        </div>

                        <div class="space-y-2">
                            <label for="songJp" class="block text-sm font-bold text-zinc-400">Song Name (Japanese)</label>
                            <input type="text" name="song_jp" id="songJp" value="{{ $song->song_jp }}"
                                class="block w-full bg-zinc-900 border border-zinc-800 text-white rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm"
                                placeholder="Enter original Japanese title...">
                        </div>

                        <div class="space-y-2">
                            <label for="songEn" class="block text-sm font-bold text-zinc-400">Song Name (English)</label>
                            <input type="text" name="song_en" id="songEn" value="{{ $song->song_en }}"
                                class="block w-full bg-zinc-900 border border-zinc-800 text-white rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm"
                                placeholder="Enter localized English title...">
                        </div>
                    </div>
                </div>

                {{-- Artists --}}
                <div class="space-y-2">
                    @php
                        $artists_string = '';
                        if (isset($song->artists) && $song->artists->count() != 0) {
                            $artists_string = implode(',', $song->artists->pluck('name')->toArray());
                        }
                    @endphp
                    <label for="artists-input"
                        class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Artists</label>
                    <input type="text" name="artists" id="artists-input" value="{{ old('artists', $artists_string) }}"
                        required
                        class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm"
                        placeholder="Artist 1, Artist 2, ...">
                    <p class="text-[10px] text-zinc-500 mt-1 italic">Please separate multiple artists with commas.</p>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {{-- Season --}}
                    <div class="space-y-2">
                        <label for="season_id"
                            class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Season</label>
                        <select name="season_id" id="season_id"
                            class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12">
                            @foreach ($seasons as $season)
                                <option value="{{ $season->id }}"
                                    {{ old('season_id', $song->season_id) == $season->id ? 'selected' : '' }}>
                                    {{ $season->name }}
                                </option>
                            @endforeach
                        </select>
                    </div>

                    {{-- Year --}}
                    <div class="space-y-2">
                        <label for="year_id"
                            class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Year</label>
                        <select name="year_id" id="year_id"
                            class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12">
                            @foreach ($years as $year)
                                <option value="{{ $year->id }}"
                                    {{ old('year_id', $song->year_id) == $year->id ? 'selected' : '' }}>
                                    {{ $year->name }}
                                </option>
                            @endforeach
                        </select>
                    </div>
                </div>

                {{-- Trigger --}}
                <div class="pt-4">
                    <button
                        class="w-full bg-blue-600 hover:bg-blue-500 text-white font-bold py-4 px-6 rounded-2xl transition-all shadow-lg shadow-blue-900/20 active:scale-[0.98] flex items-center justify-center gap-2 text-sm uppercase tracking-widest">
                        <i class="fa-solid fa-cloud-arrow-up"></i>
                        UPDATE SONG ENTRY
                    </button>
                </div>
            </form>
        </div>
    </div>
@endsection
