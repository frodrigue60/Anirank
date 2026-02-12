@extends('layouts.app')

@section('meta')
    <title>Search Animes</title>
    <meta title="Search Animes">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta name="description" content="Search your favorite animes by year, season, and format.">
    <meta name="robots" content="index, follow, max-image-preview:standard">
@endsection

@section('content')
    <livewire:animes-table />
@endsection
