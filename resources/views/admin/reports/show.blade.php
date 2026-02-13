@extends('layouts.app')

@section('content')
    <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="mb-8">
            <a href="{{ route('admin.reports.index') }}"
                class="text-blue-500 hover:text-blue-400 text-sm font-bold flex items-center mb-2 transition-colors">
                <i class="fa-solid fa-arrow-left mr-2"></i> BACK TO REPORTS
            </a>
            <h1 class="text-3xl font-bold text-white tracking-tight">Report Details</h1>
            <p class="text-zinc-400 mt-1">Audit report <span class="text-blue-400 font-mono">#{{ $report->id }}</span></p>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {{-- Main Content --}}
            <div class="lg:col-span-2 space-y-6">
                <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl p-8 shadow-xl">
                    <h3 class="text-xs font-bold text-zinc-500 uppercase tracking-widest mb-6 flex items-center">
                        <i class="fa-solid fa-file-lines mr-2"></i> REPORT CONTENT
                    </h3>

                    <div class="space-y-6">
                        <div>
                            <span class="text-[10px] font-black uppercase text-zinc-500 tracking-tighter block mb-1">Subject
                                / Title</span>
                            <p class="text-lg font-bold text-white tracking-tight">{{ $report->title }}</p>
                        </div>

                        <div>
                            <span
                                class="text-[10px] font-black uppercase text-zinc-500 tracking-tighter block mb-1">Description</span>
                            <div
                                class="p-4 bg-zinc-950/30 border border-zinc-800 rounded-2xl text-zinc-300 text-sm leading-relaxed">
                                {{ $report->content }}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {{-- Sidebar Actions --}}
            <div class="space-y-6">
                <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl p-6 shadow-xl">
                    <h3 class="text-xs font-bold text-zinc-500 uppercase tracking-widest mb-4 flex items-center">
                        <i class="fa-solid fa-bolt mr-2"></i> MODERATION
                    </h3>

                    <div class="space-y-3">
                        @if ($report->status == 'pending')
                            <a href="{{ route('admin.reports.toggle', $report->id) }}"
                                class="w-full inline-flex items-center justify-center px-4 py-3 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-emerald-900/20">
                                <i class="fa-solid fa-check-double mr-2"></i> MARK AS READ
                            </a>
                        @else
                            <a href="{{ route('admin.reports.toggle', $report->id) }}"
                                class="w-full inline-flex items-center justify-center px-4 py-3 bg-zinc-800 hover:bg-zinc-700 text-white font-bold rounded-xl transition-all border border-zinc-700">
                                <i class="fa-solid fa-undo mr-2 text-zinc-500"></i> REOPEN REPORT
                            </a>
                        @endif

                        <form action="{{ route('admin.reports.destroy', $report->id) }}" method="post" class="w-full">
                            @csrf
                            @method('DELETE')
                            <button type="submit" onclick="return confirm('Dismiss and delete permanentely?')"
                                class="w-full inline-flex items-center justify-center px-4 py-3 bg-rose-600/10 hover:bg-rose-600 text-rose-500 hover:text-white border border-rose-600/20 font-bold rounded-xl transition-all">
                                <i class="fa-solid fa-trash-can mr-2"></i> DISMISS REPORT
                            </button>
                        </form>
                    </div>

                    <div class="mt-6 pt-6 border-t border-zinc-800">
                        <span class="text-[10px] font-black uppercase text-zinc-500 tracking-tighter block mb-2">Attached
                            Source</span>
                        <a href="{{ $report->source }}"
                            class="group block p-3 bg-zinc-950/30 rounded-2xl border border-zinc-800 hover:border-blue-500/50 transition-all">
                            <div class="flex items-center gap-3">
                                <div class="w-8 h-10 bg-zinc-800 rounded border border-zinc-700 flex-shrink-0">
                                    <img src="{{ $report->song->post->image }}" alt=""
                                        class="w-full h-full object-cover rounded">
                                </div>
                                <div class="flex flex-col overflow-hidden">
                                    <span
                                        class="text-xs font-bold text-white truncate group-hover:text-blue-400 transition-colors">{{ $report->song->post->title }}</span>
                                    <span class="text-[10px] text-zinc-500 truncate">{{ $report->song->name }}</span>
                                </div>
                            </div>
                        </a>
                    </div>
                </div>

                <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl p-6 shadow-xl">
                    <span class="text-[10px] font-black uppercase text-zinc-500 tracking-tighter block mb-2">Reporter
                        Information</span>
                    <div class="flex items-center gap-3">
                        <div
                            class="w-10 h-10 bg-zinc-800 rounded-full flex items-center justify-center text-zinc-500 border border-zinc-700">
                            <i class="fa-solid fa-user text-xs"></i>
                        </div>
                        <div class="flex flex-col">
                            <span class="text-sm font-bold text-white">#{{ $report->user_id }}</span>
                            <span class="text-[10px] text-zinc-500">System Verified ID</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
@endsection
