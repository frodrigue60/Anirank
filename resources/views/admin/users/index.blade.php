@extends('layouts.app')

@section('content')
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="flex flex-col md:flex-row md:items-center md:justify-between mb-8 gap-4">
            <div>
                <h1 class="text-3xl font-bold text-white tracking-tight">User Directory</h1>
                <p class="text-zinc-400 mt-1">Audit platform accounts, manage permissions, and staff roles.</p>
            </div>
            <div class="flex flex-wrap gap-3">
                <a href="{{ route('admin.users.create') }}"
                    class="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-xl transition-all shadow-lg shadow-blue-900/20 hover:scale-105 active:scale-95">
                    <i class="fa-solid fa-user-plus mr-2"></i> CREATE USER
                </a>
            </div>
        </div>

        <div class="space-y-6">
            {{-- Search Bar --}}
            <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl shadow-xl overflow-hidden p-6">
                <form action="{{ route('admin.users.search') }}" method="GET" class="relative group">
                    <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                        <i
                            class="fa-solid fa-magnifying-glass text-zinc-500 group-focus-within:text-blue-500 transition-colors"></i>
                    </div>
                    <input type="text" name="q"
                        class="block w-full pl-11 pr-32 py-4 bg-zinc-950/50 border border-zinc-800 text-white placeholder-zinc-500 rounded-2xl focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm"
                        placeholder="Search users by name, email, or ID..." required>
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
                                <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">User Profile
                                </th>
                                <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Email
                                    Address</th>
                                <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-center">
                                    Type</th>
                                <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-right">
                                    Actions</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-zinc-800/50">
                            @foreach ($users as $user)
                                <tr class="hover:bg-zinc-800/30 transition-colors group">
                                    <td class="px-6 py-4 text-sm font-mono text-zinc-500">#{{ $user->id }}</td>
                                    <td class="px-6 py-4">
                                        <div class="flex items-center gap-3">
                                            <div
                                                class="w-10 h-10 bg-zinc-800 rounded-full flex items-center justify-center text-zinc-500 border border-zinc-700 overflow-hidden">
                                                @if ($user->image)
                                                    <img src="{{ $user->image }}" alt=""
                                                        class="w-full h-full object-cover">
                                                @else
                                                    <i class="fa-solid fa-user text-xs"></i>
                                                @endif
                                            </div>
                                            <div class="flex flex-col">
                                                <a href="{{ route('user.list', $user->slug) }}"
                                                    class="text-sm font-bold text-white hover:text-blue-400 transition-colors">
                                                    {{ $user->name }}
                                                </a>
                                                <span class="text-[10px] text-zinc-500 font-mono">@
                                                    {{ $user->slug }}</span>
                                            </div>
                                        </div>
                                    </td>
                                    <td class="px-6 py-4 text-sm text-zinc-400 font-medium">
                                        {{ $user->email }}
                                    </td>
                                    <td class="px-6 py-4 text-center">
                                        @php
                                            $typeColor = match ($user->type) {
                                                'admin' => 'bg-rose-500/10 text-rose-500 border-rose-500/20',
                                                'editor' => 'bg-blue-500/10 text-blue-500 border-blue-500/20',
                                                'creator' => 'bg-amber-500/10 text-amber-500 border-amber-500/20',
                                                default => 'bg-zinc-800 text-zinc-400 border-zinc-700',
                                            };
                                        @endphp
                                        <span
                                            class="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-widest border {{ $typeColor }}">
                                            {{ $user->type }}
                                        </span>
                                    </td>
                                    <td class="px-6 py-4">
                                        <div class="flex items-center justify-end gap-2">
                                            @if (Auth::user()->isAdmin())
                                                <a href="{{ route('admin.users.edit', $user->id) }}"
                                                    class="p-2 bg-zinc-800 hover:bg-blue-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-blue-500">
                                                    <i class="fa-solid fa-pencil"></i>
                                                </a>
                                                <form action="{{ route('admin.users.destroy', $user->id) }}" method="post"
                                                    class="inline">
                                                    @csrf
                                                    @method('DELETE')
                                                    <button type="submit"
                                                        onclick="return confirm('Delete this user account?')"
                                                        class="p-2 bg-zinc-800 hover:bg-red-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-red-500">
                                                        <i class="fa-solid fa-trash-can"></i>
                                                    </button>
                                                </form>
                                            @endif
                                        </div>
                                    </td>
                                </tr>
                            @endforeach
                        </tbody>
                    </table>
                </div>

                {{-- Pagination --}}
                <div class="bg-zinc-950/50 px-6 py-8 border-t border-zinc-800">
                    <div class="flex justify-center">
                        {{ $users->links('vendor.pagination.tailwind') }}
                    </div>
                </div>
            </div>
        </div>
    </div>
@endsection
