@extends('layouts.app')

@section('meta')
    <title>Studios</title>
    <meta title="Search Studios">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta name="description" content="Explore anime production studios.">
    <meta name="robots" content="index, follow, max-image-preview:standard">
@endsection

@section('content')
    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10">
        <div class="flex flex-col gap-8">
            {{-- Header --}}
            <div class="flex items-center gap-6">
                <div>
                    <h1 class="text-4xl font-black text-white mb-2">Studios</h1>
                    <div class="h-1 w-20 bg-primary rounded-full"></div>
                </div>
            </div>

            {{-- Filter Panel --}}
            <section class="bg-surface-dark/30 p-6 rounded-2xl border border-white/5 shadow-2xl backdrop-blur-md">
                @include('components.filter.container', [
                    'apiEndpoint' => '',
                    'method' => 'GET',
                    'fields' => ['name'],
                ])
            </section>

            {{-- Data Container --}}
            <section class="min-h-[400px]">
                <div class="results grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-6" id="data">
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
    @vite(['resources/js/filter_studios.js'])
@endsection
