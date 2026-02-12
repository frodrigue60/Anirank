@extends('layouts.app')
@section('meta')
    @if (Request::is('home'))
        <title>Profile {{ Auth::user()->name }} | {{ config('app.name') }}</title>
        <meta title="Profile">
    @endif
@endsection

@section('content')
    @if (Request::routeIs('profile'))
        @include('partials.user.banner')
    @endif

    <div class="max-w-[1440px] mx-auto px-4 md:px-8 py-10 md:py-8 flex flex-col gap-12">
        <div class="flex flex-col gap-8">
            {{-- Header --}}
            <div class="flex flex-col md:flex-row justify-between items-end gap-4">
                <div>
                    <h1 class="text-3xl font-black tracking-tight text-white mb-2 flex items-center gap-4">
                        <span class="material-symbols-outlined text-primary text-4xl">dashboard_customize</span>
                        User Dashboard
                    </h1>
                    <div class="h-1 w-20 bg-primary rounded-full"></div>
                </div>
            </div>

            {{-- Livewire User Settings Component --}}
            @livewire('user-settings')
        </div>
    </div>
@endsection

@section('script')
    <script>
        document.addEventListener('livewire:load', function () {
            // Listen for avatar updates
            Livewire.on('avatarUpdated', avatarUrl => {
                const avatarImg = document.getElementById('avatar-image');
                if (avatarImg) {
                    avatarImg.src = avatarUrl;
                }
            });

            // Listen for banner updates
            Livewire.on('bannerUpdated', bannerUrl => {
                const bannerDiv = document.getElementById('banner-image');
                if (bannerDiv) {
                    bannerDiv.style.backgroundImage = `url(${bannerUrl})`;
                }
            });
        });
    </script>
@endsection