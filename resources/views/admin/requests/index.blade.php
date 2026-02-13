@extends('layouts.app')

@section('content')
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="flex flex-col md:flex-row md:items-center md:justify-between mb-8 gap-4">
            <div>
                <h1 class="text-3xl font-bold text-white tracking-tight">Content Requests</h1>
                <p class="text-zinc-400 mt-1">Manage user-submitted requests for additions or updates.</p>
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
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Request Content
                            </th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-right">
                                Actions</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-zinc-800/50">
                        @foreach ($requests as $request)
                            <tr class="hover:bg-zinc-800/30 transition-colors group">
                                <td class="px-6 py-4 text-sm font-mono text-zinc-500 text-center">#{{ $request->id }}</td>
                                <td class="px-6 py-4">
                                    <div class="text-sm text-zinc-300 leading-relaxed line-clamp-2 max-w-2xl">
                                        {{ $request->content }}
                                    </div>
                                </td>
                                <td class="px-6 py-4">
                                    <div class="flex items-center justify-end gap-2">
                                        @if (Auth::user()->isAdmin() || Auth::user()->isEditor())
                                            @if ($request->status == 'pending')
                                                <a href="{{ route('admin.requests.show', $request->id) }}"
                                                    class="inline-flex items-center px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 hover:text-white text-xs font-bold rounded-xl transition-all border border-zinc-700">
                                                    VIEW & ATTEND
                                                </a>
                                            @else
                                                <span
                                                    class="inline-flex items-center px-4 py-2 bg-emerald-500/10 text-emerald-500 text-[10px] font-black uppercase tracking-widest rounded-xl border border-emerald-500/20">
                                                    <i class="fa-solid fa-check mr-1.5"></i> Attended
                                                </span>
                                            @endif

                                            <form action="{{ route('admin.requests.destroy', $request->id) }}"
                                                method="post" class="inline">
                                                @csrf
                                                @method('DELETE')
                                                {{-- The existing code uses a link with destroy route, but this should be a form for security --}}
                                                <button type="submit"
                                                    onclick="return confirm('Reject and delete this request?')"
                                                    class="p-2 bg-zinc-800 hover:bg-red-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-red-500"
                                                    title="Force Delete">
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
                    {{ $requests->links('vendor.pagination.tailwind') }}
                </div>
            </div>
        </div>
    </div>
@endsection
