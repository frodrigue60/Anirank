<!DOCTYPE html>
<html lang="en" data-bs-theme="dark">

<head>
    <meta http-equiv="Content-Security-Policy" content="block-all-mixed-content">
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta property="og:site_name" content="{{ env('APP_NAME') }}">
    <meta http-equiv="x-ua-compatible" content="ie=edge">
    <meta name="Author" lang="en" content="Luis Rodz">
    <meta name="base-url" content="{{ env('APP_URL') }}">
    @yield('meta')

    <!-- CSRF Token -->
    <meta name="csrf-token" content="{{ csrf_token() }}">
    <meta property="fb:app_id" content="1363850827699525" />

    <link rel="icon" type="image/png" sizes="32x32" href="{{ asset('resources/images/favicon-32x32.png') }}">
    <link rel="icon" type="image/png" sizes="16x16" href="{{ asset('resources/images/favicon-16x16.png') }}">
    <link rel="shortcut icon" sizes="512x512" href="{{ asset('resources/images/logo3.svg') }}">

    <meta name="mobile-web-app-capable" content="yes">
    <meta name="theme-color" content="#0E3D5F">

    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="black">
    <meta name="apple-mobile-web-app-title" content="{{ env('APP_NAME') }}">
    <link rel="apple-touch-icon" sizes="180x180" href="{{ asset('resources/images/apple-touch-icon.png') }}">
    <link rel="mask-icon" href="{{ asset('resources/images/safari-pinned-tab.svg') }}" color="#0E3D5F">
    <meta name="msapplication-TileImage" content="{{ asset('resources/images/msapplication-icon-144x144.png') }}">
    <meta name="msapplication-TileColor" content="#0E3D5F">


    <link rel="stylesheet" href="{{ asset('resources/owlcarousel/assets/owl.carousel.min.css') }}">
    <link rel="stylesheet" href="{{ asset('resources/owlcarousel/assets/owl.theme.default.min.css') }}">

    <script>
        (function() {
            const savedTheme = localStorage.getItem('theme') ||
                (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
            document.documentElement.setAttribute('data-bs-theme', savedTheme);
        })();
    </script>

    @vite(['resources/js/app.js', 'resources/js/ajaxSearch.js', 'resources/js/theme_switch.js', 'resources/css/app.css', 'resources/css/userProfile.css', 'resources/css/post.css', 'resources/css/ranking.css', 'resources/css/fivestars.css', 'resources/sass/app.scss'])

    @auth
        @vite(['resources/js/make_request.js'])
    @endauth

    @if (config('app.env') === 'local')
        <link rel="stylesheet" href="{{ asset('resources/font-awesome-6.4.2/css/all.min.css') }}">
    @else
        <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.2/css/all.min.css"
            integrity="sha512-z3gLpd7yknf1YoNbCzqRKc4qyor8gaKU1qmn+CShxbuBusANI9QpRohGBreCFkKxLhei6S9CQXFEbbKuqLg0DA=="
            crossorigin="anonymous" referrerpolicy="no-referrer" />
    @endif
</head>

<body class="">
    <div id="app">
        {{-- <div class="loader-container">
            <div class="spinner-border text-primary" role="status">
                <span class="visually-hidden">Loading...</span>
              </div>
        </div> --}}
        @include('layouts.navbar')

        <main class="my-3">
            @isset($breadcrumb)
                @include('layouts.breadcrumb')
            @endisset

            @include('layouts.alerts')

            @yield('content')
            <!-- Modal Search -->
            @include('layouts.modal-search')

            @auth
                @include('partials.user.modal-request')
            @endauth
        </main>

        @if (Request::routeIs('/'))
            <script src="{{ asset('resources/owlcarousel/owl.carousel.min.js') }}" defer></script>
            <script src="{{ asset('resources/js/owCarouselConfig.js') }}" defer></script>
        @endif

        @if (config('app.env') === 'local')
            @if (Request::routeIs('/'))
                <script src="{{ asset('resources/js/jquery-3.6.3.slim.min.js') }}"></script>
            @endif

            <script src="{{ asset('resources/js/popper.min.js') }}"></script>
            <script src="{{ asset('resources/font-awesome-6.4.2/js/all.min.js') }}"></script>
        @else
            @if (Request::routeIs('/'))
                <script src="https://code.jquery.com/jquery-3.6.3.slim.min.js"
                    integrity="sha256-ZwqZIVdD3iXNyGHbSYdsmWP//UBokj2FHAxKuSBKDSo=" crossorigin="anonymous"></script>
            @endif

            <script src="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.2/js/all.min.js"
                integrity="sha512-uKQ39gEGiyUJl4AI6L+ekBdGKpGw4xJ55+xyJG7YFlJokPNYegn9KwQ3P8A7aFQAUtUsAQHep+d/lrGqrbPIDQ=="
                crossorigin="anonymous" referrerpolicy="no-referrer"></script>
        @endif

        {{-- <script>
            const loaderContainer = document.querySelector('.loader-container');
            document.addEventListener("DOMContentLoaded", function() {
                loaderContainer.style.display = 'none';
            });
        </script> --}}

        @yield('script')

        @include('layouts.footer.footer-v1')
        {{-- @include('layouts.footer.mobile') --}}
    </div>

</body>

</html>
