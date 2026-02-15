@extends('layouts.app')

@section('meta')
    <title>Search Openings & Endings</title>
    <meta title="Search Openings & Endings">
    <link rel="canonical" href="{{ url()->current() }}">
    <meta name="description" content="Search Openings & Endings by type, season, order as you want, and filter by letter">
    <meta name="robots" content="index, follow, max-image-preview:standard">
@endsection

@section('content')
    <livewire:songs-table />
@endsection
