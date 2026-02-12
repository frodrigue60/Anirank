@extends('layouts.app')

@section('meta')
    <title>Studios | {{ config('app.name') }}</title>
    <meta title="Search Studios">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta name="description" content="Explore anime production studios.">
    <meta name="robots" content="index, follow, max-image-preview:standard">
@endsection

@section('content')
    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10 md:py-8 flex flex-col gap-12">
        {{-- Header Section --}}
        <div class="flex flex-col md:flex-row justify-between items-end gap-4">
            <div>
                <h1 class="text-3xl font-black tracking-tight text-white mb-2">Search Studios</h1>
                <div class="h-1 w-20 bg-primary rounded-full"></div>
            </div>
        </div>
        {{-- Livewire Table component --}}
        @livewire('studios-table')
    </div>
@endsection

@section('script')
    {{-- Legacy filter scripts removed --}}
@endsection