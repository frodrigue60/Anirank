<!DOCTYPE html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}" class="dark">

<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="csrf-token" content="{{ csrf_token() }}">

    {{-- Fonts & Icons --}}
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Spline+Sans:wght@300;400;500;600;700&display=swap"
        rel="stylesheet">
    <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap"
        rel="stylesheet">

    <title>@yield('title', config('app.name'))</title>
    @yield('meta')

    {{-- Main Assets --}}
    @vite(['resources/sass/app.scss', 'resources/js/app.js'])

    @stack('styles')
    <style>
        [x-cloak] {
            display: none !important;
        }

        body {
            overflow: hidden;
        }
    </style>
</head>

<body class="bg-background-dark text-white font-display antialiased">
    @yield('content')

    @livewireScripts
    @stack('scripts')
</body>

</html>