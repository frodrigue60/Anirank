@extends('layouts.app')

@section('meta')
    <title>My Favorites | {{ config('app.name') }}</title>
    <meta title="My Favorites">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta name="description" content="View your favorite anime openings and endings.">
    <meta name="robots" content="noindex, nofollow">
@endsection

@section('content')
    @include('partials.user.banner')

    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10 md:py-8 flex flex-col gap-12">
        <div class="flex flex-col gap-8">
            {{-- Header --}}
            <div class="flex flex-col md:flex-row justify-between items-end gap-4">
                <div>
                    <h1 class="text-3xl font-black tracking-tight text-white mb-2 flex items-center gap-4">
                        <span class="material-symbols-outlined text-primary text-4xl">favorite</span>
                        Favorites
                    </h1>
                    <div class="h-1 w-20 bg-primary rounded-full"></div>
                </div>
            </div>

            {{-- Livewire Favorites Table --}}
            @livewire('favorites-table')
        </div>
    </div>
@endsection

@section('script')
    {{-- Legacy filter scripts removed --}}
@endsection