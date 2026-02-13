@extends('layouts.app')

@section('content')
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="flex flex-col md:flex-row md:items-center md:justify-between mb-8 gap-4">
            <div>
                <h1 class="text-3xl font-bold text-white tracking-tight">Post Management</h1>
                <p class="text-zinc-400 mt-1">Manage and curate your anime directory.</p>
            </div>
            <div class="flex flex-wrap gap-3">
                <a href="{{ route('admin.posts.create') }}"
                    class="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-xl transition-all shadow-lg shadow-blue-900/20 hover:scale-105 active:scale-95">
                    <i class="fa-solid fa-plus mr-2"></i> CREATE POST
                </a>
                <a href="{{ route('admin.posts.sync-all') }}"
                    onclick="return confirm('This will update all posts with AniList data. It may take some time. Continue?')"
                    class="inline-flex items-center px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-semibold rounded-xl transition-all shadow-lg shadow-emerald-900/20 hover:scale-105 active:scale-95">
                    <i class="fa-solid fa-sync mr-2"></i> BULK SYNC ANILIST
                </a>
                <a href="{{ route('admin.posts.wipe') }}"
                    class="inline-flex items-center px-4 py-2 bg-zinc-800 hover:bg-red-600 text-zinc-300 hover:text-white text-sm font-semibold rounded-xl transition-all shadow-lg hover:scale-105 active:scale-95 border border-zinc-700 hover:border-red-500">
                    <i class="fa-solid fa-trash-can mr-2"></i> WIPE ALL
                </a>
            </div>
        </div>

        <div class="space-y-6">
            {{-- Search & API Options Card --}}
            <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl shadow-xl overflow-hidden p-6">
                @if (Auth::user()->isAdmin())
                    <div x-data="{ open: false }" class="mb-6">
                        <button @click="open = !open"
                            class="flex items-center justify-between w-full px-4 py-3 bg-zinc-950/50 hover:bg-zinc-800/50 rounded-2xl transition-all border border-zinc-800 group">
                            <span class="text-sm font-bold text-zinc-300 group-hover:text-white transition-colors">
                                <i class="fa-solid fa-cloud-arrow-down mr-2 text-blue-500"></i> API OPTIONS
                            </span>
                            <i class="fa-solid fa-chevron-down text-zinc-500 transition-transform duration-300"
                                :class="open ? 'rotate-180' : ''"></i>
                        </button>

                        <div x-show="open" x-collapse x-cloak class="mt-4 space-y-4">
                            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                                {{-- Search API --}}
                                <form action="{{ route('admin.posts.search.animes') }}" method="POST"
                                    class="bg-zinc-950/30 p-4 rounded-2xl border border-zinc-800/50">
                                    @csrf
                                    <h3 class="text-xs font-bold text-zinc-500 uppercase tracking-widest mb-3">Search API
                                    </h3>
                                    <div class="flex flex-col sm:flex-row gap-2">
                                        <input
                                            class="flex-1 bg-zinc-900 border-zinc-700 text-white rounded-xl text-sm focus:ring-blue-500 focus:border-blue-500"
                                            type="text" name="q" placeholder="Anime title..." required />
                                        <select
                                            class="bg-zinc-900 border-zinc-700 text-white rounded-xl text-sm focus:ring-blue-500 focus:border-blue-500"
                                            name="type">
                                            <option value="TV">TV</option>
                                            <option value="TV_SHORT">TV SHORT</option>
                                            <option value="MOVIE">MOVIE</option>
                                            <option value="SPECIAL">SPECIAL</option>
                                            <option value="OVA">OVA</option>
                                            <option value="ONA">ONA</option>
                                        </select>
                                        <button
                                            class="px-4 py-2 bg-blue-600/20 hover:bg-blue-600 text-blue-400 hover:text-white border border-blue-600/30 rounded-xl text-sm font-bold transition-all"
                                            type="submit">Search</button>
                                    </div>
                                </form>

                                {{-- Seasonal API --}}
                                <form action="{{ route('admin.posts.get.seasonal.animes') }}" method="POST"
                                    class="bg-zinc-950/30 p-4 rounded-2xl border border-zinc-800/50">
                                    @csrf
                                    <h3 class="text-xs font-bold text-zinc-500 uppercase tracking-widest mb-3">Batch Import
                                    </h3>
                                    <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
                                        <input
                                            class="col-span-1 bg-zinc-900 border-zinc-700 text-white rounded-xl text-sm focus:ring-blue-500 focus:border-blue-500"
                                            type="number" name="year" placeholder="Year" required />
                                        <select
                                            class="col-span-1 bg-zinc-900 border-zinc-700 text-white rounded-xl text-sm focus:ring-blue-500 focus:border-blue-500"
                                            name="season">
                                            <option value="">Season</option>
                                            <option value="WINTER">WINTER</option>
                                            <option value="SPRING">SPRING</option>
                                            <option value="SUMMER">SUMMER</option>
                                            <option value="FALL">FALL</option>
                                        </select>
                                        <select
                                            class="col-span-1 bg-zinc-900 border-zinc-700 text-white rounded-xl text-sm focus:ring-blue-500 focus:border-blue-500"
                                            name="type">
                                            <option value="TV">TV</option>
                                            <option value="TV_SHORT">TV SHORT</option>
                                            <option value="MOVIE">MOVIE</option>
                                            <option value="SPECIAL">SPECIAL</option>
                                            <option value="OVA">OVA</option>
                                            <option value="ONA">ONA</option>
                                        </select>
                                        <button
                                            class="col-span-1 px-4 py-2 bg-emerald-600/20 hover:bg-emerald-600 text-emerald-400 hover:text-white border border-emerald-600/30 rounded-xl text-sm font-bold transition-all"
                                            type="submit">Fetch</button>
                                    </div>
                                </form>
                            </div>
                        </div>
                    </div>
                @endif

                {{-- Local Search --}}
                <form action="{{ route('admin.posts.search') }}" method="POST" class="relative group">
                    @csrf
                    <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                        <i
                            class="fa-solid fa-magnifying-glass text-zinc-500 group-focus-within:text-blue-500 transition-colors"></i>
                    </div>
                    <input type="text" name="q"
                        class="block w-full pl-11 pr-32 py-4 bg-zinc-950/50 border border-zinc-800 text-white placeholder-zinc-500 rounded-2xl focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm"
                        placeholder="Search posts in database..." required>
                    <div class="absolute inset-y-2 right-2 flex items-center">
                        <button type="submit"
                            class="px-6 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold rounded-xl transition-all shadow-lg active:scale-95">
                            SEARCH
                        </button>
                    </div>
                </form>
            </div>

            {{-- Table Card --}}
            <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl shadow-xl overflow-hidden">
                <div class="overflow-x-auto">
                    <table class="w-full text-left border-collapse">
                        <thead>
                            <tr class="bg-zinc-950/50 border-b border-zinc-800">
                                <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">ID</th>
                                <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Title</th>
                                <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Filters</th>
                                <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-center">
                                    Songs</th>
                                <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Status</th>
                                @if (Auth::user()->isAdmin() || Auth::user()->isEditor())
                                    <th
                                        class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-right">
                                        Actions</th>
                                @endif
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-zinc-800/50">
                            @foreach ($posts as $post)
                                <tr class="hover:bg-zinc-800/30 transition-colors group">
                                    <td class="px-6 py-4 text-sm font-mono text-zinc-500">#{{ $post->id }}</td>
                                    <td class="px-6 py-4">
                                        <div class="flex items-center gap-3">
                                            <div
                                                class="w-10 h-14 bg-zinc-800 rounded-md overflow-hidden flex-shrink-0 border border-zinc-700">
                                                <img src="{{ asset('storage/' . $post->thumbnail ?? ($post->thumbnail_src ?? 'images/no-image.png')) }}"
                                                    alt="" class="w-full h-full object-cover">
                                            </div>
                                            <a href="{{ $post->url }}"
                                                class="text-sm font-semibold text-white hover:text-blue-400 transition-colors line-clamp-2 leading-snug">
                                                {{ $post->title }}
                                            </a>
                                        </div>
                                    </td>
                                    <td class="px-6 py-4">
                                        <div class="flex flex-wrap gap-1.5">
                                            @isset($post->season->id)
                                                <a href="{{ route('seasons.show', $post->season->id) }}"
                                                    class="px-2.5 py-1 bg-zinc-800 text-zinc-300 text-[10px] font-bold uppercase rounded-md hover:bg-blue-600 hover:text-white transition-all">
                                                    {{ $post->season->name }}
                                                </a>
                                            @endisset
                                            @isset($post->year->id)
                                                <a href="{{ route('years.show', $post->year->id) }}"
                                                    class="px-2.5 py-1 bg-zinc-800 text-zinc-300 text-[10px] font-bold uppercase rounded-md hover:bg-blue-600 hover:text-white transition-all">
                                                    {{ $post->year->name }}
                                                </a>
                                            @endisset
                                        </div>
                                    </td>
                                    <td class="px-6 py-4 text-center">
                                        <span
                                            class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-zinc-800 text-zinc-300 border border-zinc-700">
                                            {{ $post->songs->count() }}
                                        </span>
                                    </td>
                                    <td class="px-6 py-4 text-center">
                                        @if (Auth::user()->isCreator())
                                            @if ($post->status === null)
                                                <span
                                                    class="px-2.5 py-1 bg-zinc-800 text-zinc-500 text-[10px] font-bold uppercase rounded-md border border-zinc-700">N/A</span>
                                            @elseif ($post->status == false)
                                                <span
                                                    class="px-2.5 py-1 bg-amber-500/10 text-amber-500 text-[10px] font-bold uppercase rounded-md border border-amber-500/20">
                                                    <i class="fa-solid fa-clock mr-1"></i> Pending
                                                </span>
                                            @else
                                                <span
                                                    class="px-2.5 py-1 bg-blue-500/10 text-blue-500 text-[10px] font-bold uppercase rounded-md border border-blue-500/20">
                                                    <i class="fa-solid fa-check mr-1"></i> Active
                                                </span>
                                            @endif
                                        @endif

                                        @if (Auth::user()->isAdmin() || Auth::user()->isEditor())
                                            <form action="{{ route('admin.posts.toggle.status', $post->id) }}"
                                                method="post">
                                                @csrf
                                                <button type="submit"
                                                    class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-[10px] font-bold uppercase border transition-all {{ $post->status == true ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/30 hover:bg-emerald-500 hover:text-white' : 'bg-amber-500/10 text-amber-500 border-amber-500/30 hover:bg-amber-500 hover:text-white' }}">
                                                    <i
                                                        class="fa-solid {{ $post->status == true ? 'fa-toggle-on' : 'fa-toggle-off' }} text-sm"></i>
                                                    {{ $post->status == true ? 'Public' : 'Private' }}
                                                </button>
                                            </form>
                                        @endif
                                    </td>
                                    @if (Auth::user()->isAdmin() || Auth::user()->isEditor())
                                        <td class="px-6 py-4">
                                            <div class="flex items-center justify-end gap-2 shrink-0">
                                                <a href="{{ route('admin.posts.edit', $post->id) }}"
                                                    class="p-2 bg-zinc-800 hover:bg-blue-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-blue-500">
                                                    <i class="fa-solid fa-pencil"></i>
                                                </a>
                                                <a href="{{ route('admin.posts.songs', $post->id) }}"
                                                    class="p-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700">
                                                    <i class="fa-solid fa-list"></i>
                                                </a>
                                                <a href="{{ route('admin.posts.show', $post->id) }}"
                                                    class="p-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700">
                                                    <i class="fa-solid fa-eye"></i>
                                                </a>
                                                <form action="{{ route('admin.posts.destroy', $post->id) }}"
                                                    method="post" class="inline">
                                                    @csrf
                                                    @method('DELETE')
                                                    <button type="submit"
                                                        onclick="return confirm('Are you sure you want to delete this post?')"
                                                        class="p-2 bg-zinc-800 hover:bg-red-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-red-500">
                                                        <i class="fa-solid fa-trash-can"></i>
                                                    </button>
                                                </form>
                                            </div>
                                        </td>
                                    @endif
                                </tr>
                            @endforeach
                        </tbody>
                    </table>
                </div>

                {{-- Pagination Footer --}}
                <div class="bg-zinc-950/50 px-6 py-6 border-t border-zinc-800">
                    <div class="flex flex-col items-center">
                        <div class="text-zinc-500 text-xs mb-4">
                            Showing <span class="text-white font-bold">{{ $posts->firstItem() }}</span> to <span
                                class="text-white font-bold">{{ $posts->lastItem() }}</span> of <span
                                class="text-white font-bold">{{ $posts->total() }}</span> results
                        </div>
                        <div class="pagination-custom">
                            {{ $posts->links() }}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
@endsection

@push('styles')
    <style>
        /* Custom Scrollbar for the table */
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

        /* Alpine JS cloak */
        [x-cloak] {
            display: none !important;
        }

        /* Fix Laravel/Tailwind pagination common alignment issues */
        .pagination-custom nav svg {
            display: inline-block;
            width: 1.25rem;
        }

        .pagination-custom nav div:first-child {
            display: none;
        }

        .pagination-custom nav div:last-child {
            background: transparent !important;
            border: none !important;
            box-shadow: none !important;
        }
    </style>
@endpush
