@extends('layouts.app')

@section('meta')
    <title>Producer {{ $producer->name }} | {{ config('app.name') }}</title>
    <meta title="Producer {{ $producer->name }}">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta name="description" content="Explore animes produced by {{ $producer->name }}.">
    <meta name="robots" content="index, follow, max-image-preview:standard">
@endsection

@section('content')
    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10 md:py-8 flex flex-col gap-12">
        {{-- Livewire Table component --}}
        @livewire('producer-animes-table', ['producerId' => $producer->id])
    </div>
@endsection
