@php
    // Robust URL resolution
    $profileUrl = asset('resources/images/default-avatar.jpg');
    $bannerUrl = asset('resources/images/default-banner.jpg');

    if ($user->image && Storage::disk('public')->exists($user->image)) {
        $profileUrl = Storage::disk('public')->url($user->image);
    }

    if ($user->banner && Storage::disk('public')->exists($user->banner)) {
        $bannerUrl = Storage::disk('public')->url($user->banner);
    }
@endphp

<div class="header">
    <div class="banner-user" style="background-image: url('{{ $bannerUrl }}')" id="banner-image">
        <div class="data-container">
            <div class="shadow-banner"></div>
            <div class="banner-content container">
                <div class="relative group/avatar">
                    <img class="avatar border-4 border-surface-dark shadow-2xl" src="{{ $profileUrl }}"
                        alt="{{ $user->name }}" id="avatar-image">
                </div>
                <div class="name-wrapper">
                    <h1 class="name font-black tracking-tight text-white drop-shadow-lg">{{ $user->name }}</h1>
                </div>
            </div>
        </div>
    </div>
</div>