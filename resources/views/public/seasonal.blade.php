@extends('layouts.app')
@section('meta')
    <meta name="robots" content="index, follow" />
    <link rel="canonical" href="{{ url()->current() }}" />
    @if (isset($currentSeason->name) && isset($currentYear->name))
        <meta name="keywords"
            content="ranking, top, anime openings {{ $currentSeason->name }}, openings anime {{ $currentSeason->name }}, anime endings {{ $currentSeason->name }}, endings anime {{ $currentSeason->name }}, of {{ $currentSeason->name }}">
        @if (Request::is('openings'))
            <title>Openings {{ $currentSeason->name }}</title>
            <meta title="Openings {{ $currentSeason->name }}">
            <meta name="description" content="Openings of {{ $currentSeason->name }} anime season">

            <meta property="og:type" content="website" />
            {{-- <meta property="og:image" content="{{ asset('resources/images/og-image-wide.png') }}">
        <meta property="og:image:secure_url" content="{{ asset('resources/images/og-image-wide.png') }}">
        <meta property="og:image:type" content="image/png">
        <meta property="og:image:width" content="828">
        <meta property="og:image:height" content="450">
        <meta property="og:image:alt" content="Anirank banner" /> --}}
            <meta property="og:url" content="{{ url()->current() }}" />

            <meta name="twitter:card" content="summary" />
            <meta name="twitter:site" content="@frodrigue60" />
            <meta name="twitter:creator" content="@frodrigue60" />
            <meta property="og:title" content="Openings {{ $currentSeason->name }}" />
            <meta property="og:description" content="Openings of {{ $currentSeason->name }} anime season" />
        @endif
        @if (Request::is('endings'))
            <title>Endings {{ $currentSeason->name }}</title>
            <meta title="Endings {{ $currentSeason->name }}">
            <meta name="description" content="Endings of {{ $currentSeason->name }} anime season">

            <meta property="og:type" content="website" />
            {{-- <meta property="og:image" content="{{ asset('resources/images/og-image-wide.png') }}">
        <meta property="og:image:secure_url" content="{{ asset('resources/images/og-image-wide.png') }}">
        <meta property="og:image:type" content="image/png">
        <meta property="og:image:width" content="828">
        <meta property="og:image:height" content="450">
        <meta property="og:image:alt" content="Anirank banner" /> --}}
            <meta property="og:url" content="{{ url()->current() }}" />

            <meta name="twitter:card" content="summary" />
            <meta name="twitter:site" content="@frodrigue60" />
            <meta name="twitter:creator" content="@frodrigue60" />
            <meta property="og:title" content="Endings {{ $currentSeason->name }}" />
            <meta property="og:description" content="Endings of {{ $currentSeason->name }} anime season" />
        @endif
    @endif
@endsection
@section('content')
    @if ($currentSeason && $currentYear)
        <livewire:seasonal-table :season="$currentSeason->id" :year="$currentYear->id" />
    @endif
@endsection
