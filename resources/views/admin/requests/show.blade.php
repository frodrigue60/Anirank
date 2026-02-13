@extends('layouts.app')

@section('content')
    <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="mb-8">
            <a href="{{ route('admin.requests.index') }}"
                class="text-blue-500 hover:text-blue-400 text-sm font-bold flex items-center mb-2 transition-colors">
                <i class="fa-solid fa-arrow-left mr-2"></i> BACK TO REQUESTS
            </a>
            <h1 class="text-3xl font-bold text-white tracking-tight">Request Details</h1>
            <p class="text-zinc-400 mt-1">Submitted by <span
                    class="text-blue-400 font-bold">{{ $userRequest->user->name }}</span></p>
        </div>

        <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl shadow-xl overflow-hidden shadow-xl">
            {{-- Content Section --}}
            <div class="p-8">
                <h3 class="text-xs font-bold text-zinc-500 uppercase tracking-widest mb-6 flex items-center">
                    <i class="fa-solid fa-envelope-open-text mr-2 text-blue-500"></i> REQUEST CONTENT
                </h3>

                <div
                    class="p-6 bg-zinc-950/30 border border-zinc-800 rounded-2xl text-zinc-300 text-sm leading-relaxed whitespace-pre-wrap">
                    {{ $userRequest->content }}
                </div>
            </div>

            {{-- Info & Action Footer --}}
            <div
                class="bg-zinc-950/50 px-8 py-6 border-t border-zinc-800 flex flex-col sm:flex-row items-center justify-between gap-4">
                <div class="flex items-center gap-3">
                    <div
                        class="w-10 h-10 bg-zinc-800 rounded-full flex items-center justify-center text-zinc-400 border border-zinc-700">
                        <i class="fa-solid fa-user text-xs"></i>
                    </div>
                    <div class="flex flex-col">
                        <span class="text-sm font-bold text-white">{{ $userRequest->user->name }}</span>
                        <span class="text-[10px] text-zinc-500 uppercase font-black tracking-tighter">Member since
                            {{ $userRequest->user->created_at->format('M Y') }}</span>
                    </div>
                </div>

                <div class="flex items-center gap-3 w-full sm:w-auto">
                    @if ($userRequest->status == 'pending')
                        <a href="{{ route('admin.requests.toggle', $userRequest->id) }}"
                            class="flex-1 sm:flex-none inline-flex items-center justify-center px-6 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-black uppercase tracking-widest rounded-xl transition-all shadow-lg shadow-blue-900/20 active:scale-95">
                            <i class="fa-solid fa-check-circle mr-2"></i> MARK ATTENDED
                        </a>
                    @endif
                    <form action="{{ route('admin.requests.destroy', $userRequest->id) }}" method="post"
                        class="flex-1 sm:flex-none">
                        @csrf
                        @method('DELETE')
                        <button type="submit" onclick="return confirm('Permanentely delete this request?')"
                            class="w-full inline-flex items-center justify-center px-6 py-2 bg-zinc-800 hover:bg-rose-600 text-zinc-400 hover:text-white text-xs font-black uppercase tracking-widest rounded-xl transition-all border border-zinc-700 hover:border-rose-500 active:scale-95">
                            <i class="fa-solid fa-trash-can mr-2"></i> DELETE
                        </button>
                    </form>
                </div>
            </div>
        </div>
    </div>
@endsection
