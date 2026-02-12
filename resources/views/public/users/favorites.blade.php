@extends('layouts.app')

@section('meta')
    <title>My Favorites</title>
    <meta title="My Favorites">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta name="description" content="View your favorite anime openings and endings.">
    <meta name="robots" content="noindex, nofollow">
@endsection

@section('content')
    @include('partials.user.banner')

    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10">
        <div class="flex flex-col gap-8">
            {{-- Header --}}
            <div class="flex items-center justify-between">
                <h2 class="text-3xl font-black text-white flex items-center gap-4">
                    <span class="material-symbols-outlined text-primary text-4xl">favorite</span>
                    My Favorites
                </h2>
            </div>

            {{-- Filter Panel --}}
            <section class="bg-surface-dark/30 p-6 rounded-2xl border border-white/5 shadow-2xl backdrop-blur-md">
                @include('components.filter.container', [
                    'apiEndpoint' => '',
                    'method' => 'post',
                    'fields' => ['name', 'type', 'year', 'season', 'sort'],
                ])
            </section>

            {{-- Data Container --}}
            <section class="min-h-[400px]">
                <div class="results grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6" id="data">
                    {{-- AJAX Results --}}
                </div>

                {{-- Loader --}}
                <div class="flex justify-center py-20" id="loader">
                    <div class="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                </div>
            </section>
        </div>
    </div>
@endsection

@section('script')
    @vite(['resources/js/filter_favorites.js'])
@endsection
