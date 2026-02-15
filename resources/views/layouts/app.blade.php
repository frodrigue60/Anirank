<!DOCTYPE html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}" class="dark">

<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta http-equiv="x-ua-compatible" content="ie=edge">
    <meta name="csrf-token" content="{{ csrf_token() }}">
    <meta http-equiv="Content-Security-Policy" content="block-all-mixed-content">

    {{-- New Theme Fonts & Icons --}}
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link
        href="https://fonts.googleapis.com/css2?family=Nunito:wght@400;500;600;700&family=Spline+Sans:wght@300;400;500;600;700&family=Noto+Sans:wght@400;500;700&display=swap"
        rel="stylesheet">
    <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap"
        rel="stylesheet">

    <title>@yield('title', config('app.name'))</title>

    {{-- Meta Tags & SEO --}}
    <meta name="author" content="Luis Rodz">
    <meta name="base-url" content="{{ config('app.url') }}">
    <meta property="og:site_name" content="{{ config('app.name') }}">
    <meta property="fb:app_id" content="1363850827699525" />
    @yield('meta')

    {{-- Favicons --}}
    <link rel="icon" type="image/png" sizes="32x32" href="{{ asset('resources/images/favicon-32x32.png') }}">
    <link rel="icon" type="image/png" sizes="16x16" href="{{ asset('resources/images/favicon-16x16.png') }}">
    <link rel="shortcut icon" sizes="512x512" href="{{ asset('resources/images/logo3.svg') }}">

    {{-- PWA & Mobile --}}
    <meta name="mobile-web-app-capable" content="yes">
    <meta name="theme-color" content="#0E3D5F">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="black">
    <meta name="apple-mobile-web-app-title" content="{{ config('app.name') }}">
    <link rel="apple-touch-icon" sizes="180x180" href="{{ asset('resources/images/apple-touch-icon.png') }}">
    <link rel="mask-icon" href="{{ asset('resources/images/safari-pinned-tab.svg') }}" color="#0E3D5F">
    <meta name="msapplication-TileImage" content="{{ asset('resources/images/msapplication-icon-144x144.png') }}">
    <meta name="msapplication-TileColor" content="#0E3D5F">

    {{-- Theme Detection (Blocking to prevent flickers) --}}
    <script>
        (function() {
            const theme = localStorage.getItem('theme') || 'dark';
            if (theme === 'dark') {
                document.documentElement.classList.add('dark');
            } else {
                document.documentElement.classList.remove('dark');
            }
        })();
    </script>

    {{-- Main Assets (Vite) --}}
    @vite(['resources/css/app.css', 'resources/css/userProfile.css', 'resources/css/post.css', 'resources/css/ranking.css', 'resources/css/fivestars.css', 'resources/js/app.js', 'resources/js/ajaxSearch.js', 'resources/js/theme_switch.js'])

    @auth
        @vite(['resources/js/make_request.js'])
    @endauth

    {{-- Third Party & Conditional Styles --}}
    @if (config('app.env') === 'local')
        <link rel="stylesheet" href="{{ asset('resources/font-awesome-6.4.2/css/all.min.css') }}">
    @else
        <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.2/css/all.min.css"
            integrity="sha512-z3gLpd7yknf1YoNbCzqRKc4qyor8gaKU1qmn+CShxbuBusANI9QpRohGBreCFkKxLhei6S9CQXFEbbKuqLg0DA=="
            crossorigin="anonymous" referrerpolicy="no-referrer" />
    @endif

    @livewireStyles
    <style>
        [x-cloak] {
            display: none !important;
        }
    </style>
    @stack('styles')
</head>

<body
    class="bg-background-dark text-white font-display min-h-screen flex flex-col overflow-x-hidden antialiased selection:bg-primary selection:text-white transition-colors duration-300">
    <div id="app">
        @include('layouts.navbar')

        <main class="">
            @isset($breadcrumb)
                @include('layouts.breadcrumb')
            @endisset

            @include('layouts.alerts')

            @yield('content')

            @include('layouts.modal-search')

            @auth
                @include('partials.user.modal-request')
            @endauth
        </main>

        @include('layouts.footer.footer-v1')

        <x-toast />
    </div>

    {{-- Scripts --}}
    @if (config('app.env') === 'local')
        <script src="{{ asset('resources/font-awesome-6.4.2/js/all.min.js') }}"></script>
    @else
        <script src="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.2/js/all.min.js"
            integrity="sha512-uKQ39gEGiyUJl4AI6L+ekBdGKpGw4xJ55+xyJG7YFlJokPNYegn9KwQ3P8A7aFQAUtUsAQHep+d/lrGqrbPIDQ=="
            crossorigin="anonymous" referrerpolicy="no-referrer"></script>
    @endif

    @livewireScripts
    @stack('scripts')
    @yield('script')

    @if (Route::is('seasonal') || Route::is('ranking'))
        <livewire:global-bottom-player />
    @endif
</body>

</html>
