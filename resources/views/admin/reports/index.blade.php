@extends('layouts.app')

@section('content')
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="flex flex-col md:flex-row md:items-center md:justify-between mb-8 gap-4">
            <div>
                <h1 class="text-3xl font-bold text-white tracking-tight">System Reports</h1>
                <p class="text-zinc-400 mt-1">Review community flags, content issues, and broken links.</p>
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
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Context / Source
                            </th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest">Reporter ID</th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-center">
                                Status</th>
                            <th class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-right">
                                Actions</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-zinc-800/50">
                        @foreach ($reports as $report)
                            <tr class="hover:bg-zinc-800/30 transition-colors group">
                                <td class="px-6 py-4 text-sm font-mono text-zinc-500 text-center">#{{ $report->id }}</td>
                                <td class="px-6 py-4">
                                    <div class="flex flex-col">
                                        <a href="{{ $report->source }}"
                                            class="text-sm font-bold text-white hover:text-blue-400 transition-colors line-clamp-1">
                                            {{ $report->song->name }}
                                        </a>
                                        <span class="text-[10px] text-zinc-500 mt-1 font-medium truncate max-w-xs">
                                            {{ $report->song->name }} ({{ $report->song->slug }})
                                        </span>
                                    </div>
                                </td>
                                <td class="px-6 py-4 text-sm font-mono text-zinc-400">
                                    {{ $report->user_id }}
                                </td>
                                <td class="px-6 py-4 text-center">
                                    @if ($report->status == 'pending')
                                        <a href="{{ route('admin.reports.toggle', $report->id) }}"
                                            class="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-widest bg-amber-500/10 text-amber-500 border border-amber-500/20 hover:bg-amber-500 hover:text-white transition-all">
                                            <i class="fa-solid fa-clock mr-1.5"></i> Pending
                                        </a>
                                    @else
                                        <a href="{{ route('admin.reports.toggle', $report->id) }}"
                                            class="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-widest bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 hover:bg-emerald-500 hover:text-white transition-all">
                                            <i class="fa-solid fa-check mr-1.5"></i> Resolved
                                        </a>
                                    @endif
                                </td>
                                <td class="px-6 py-4">
                                    <div class="flex items-center justify-end gap-2">
                                        @if (Auth::user()->isAdmin() || Auth::user()->isEditor())
                                            <a href="{{ route('admin.reports.show', $report->id) }}"
                                                class="p-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700"
                                                title="View Report Details">
                                                <i class="fa-solid fa-eye"></i>
                                            </a>
                                            <form action="{{ route('admin.reports.destroy', $report->id) }}" method="post"
                                                class="inline">
                                                @csrf
                                                @method('DELETE')
                                                <button type="submit"
                                                    onclick="return confirm('Dismiss and delete this report?')"
                                                    class="p-2 bg-zinc-800 hover:bg-red-600 text-zinc-400 hover:text-white rounded-lg transition-all border border-zinc-700 hover:border-red-500"
                                                    title="Delete Report">
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
                    {{ $reports->links('vendor.pagination.tailwind') }}
                </div>
            </div>
        </div>
    </div>
@endsection
