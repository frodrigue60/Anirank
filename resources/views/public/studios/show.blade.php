@extends('layouts.app')

@section('meta')
    <title>Studio {{ $studio->name }} | {{ config('app.name') }}</title>
    <meta title="Studio {{ $studio->name }}">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta name="description" content="Explore animes produced by {{ $studio->name }}.">
    <meta name="robots" content="index, follow, max-image-preview:standard">
@endsection

@section('content')
    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10 md:py-8 flex flex-col gap-12">
        {{-- Livewire Table component --}}
        @livewire('studio-animes-table', ['studioId' => $studio->id])
    </div>
@endsection

@section('script')
    {{-- Legacy filter scripts removed --}}
@endsection
