@extends('layouts.app')

@section('meta')
    <title>{{ $user->name }} - List | {{ config('app.name') }}</title>
    <meta title="{{ $user->name }} - Favorites Archive">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta name="description"
        content="Explore the personal collection of openings & endings from {{ $user->name }} on Anirank.">
    <meta name="robots" content="index, follow, max-image-preview:standard">
@endsection

@section('content')
    {{-- Modern User Banner --}}
    @include('partials.user.banner')

    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-12 md:py-16">
        {{-- Livewire Favorites Table --}}
        @livewire('user-favorites-table', ['userId' => $user->id])
    </div>
@endsection

@section('script')
    {{-- Legacy AJAX filter scripts removed --}}
@endsection
