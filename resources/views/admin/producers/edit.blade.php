@extends('layouts.app')

@section('content')
    <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="mb-8">
            <a href="{{ route('admin.producers.index') }}"
                class="text-blue-500 hover:text-blue-400 text-sm font-bold flex items-center mb-2 transition-colors">
                <i class="fa-solid fa-arrow-left mr-2"></i> BACK TO PRODUCERS
            </a>
            <h1 class="text-3xl font-bold text-white tracking-tight">Edit Producer</h1>
            <p class="text-zinc-400 mt-1">Updating details for <span
                    class="text-blue-400 font-semibold">{{ $producer->name }}</span></p>
        </div>

        {{-- Form Card --}}
        <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl shadow-xl overflow-hidden p-8">
            <form method="post" action="{{ route('admin.producers.update', $producer->id) }}" class="space-y-8">
                @csrf
                @method('PUT')

                <div class="space-y-6">
                    <h3 class="text-xs font-bold text-zinc-500 uppercase tracking-widest flex items-center">
                        <i class="fa-solid fa-building mr-2 text-blue-500"></i> PRODUCER DETAILS
                    </h3>

                    <div class="space-y-2">
                        <label for="nameProducer"
                            class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Producer Name</label>
                        <input type="text" name="name" id="nameProducer" required
                            value="{{ old('name', $producer->name) }}"
                            class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12 @error('name') border-red-500 @enderror"
                            placeholder="e.g. Aniplex">
                        @error('name')
                            <p class="text-red-500 text-xs mt-1">{{ $message }}</p>
                        @enderror
                    </div>
                </div>

                {{-- Action --}}
                <div class="pt-4 flex flex-col sm:flex-row gap-4">
                    <button
                        class="flex-1 bg-blue-600 hover:bg-blue-500 text-white font-bold py-4 px-6 rounded-2xl transition-all shadow-lg shadow-blue-900/20 active:scale-[0.98] flex items-center justify-center gap-2 text-sm uppercase tracking-widest">
                        <i class="fa-solid fa-check-double"></i>
                        SAVE CHANGES
                    </button>
                    <a href="{{ route('admin.producers.index') }}"
                        class="flex-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 font-bold py-4 px-6 rounded-2xl transition-all flex items-center justify-center gap-2 text-sm uppercase tracking-widest border border-zinc-700">
                        CANCEL
                    </a>
                </div>
            </form>
        </div>
    </div>
@endsection
