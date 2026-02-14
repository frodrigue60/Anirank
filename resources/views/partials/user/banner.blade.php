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

<div class="relative w-full overflow-hidden">
    {{-- Banner Image with Gradient Overlay --}}
    <div class="h-[250px] md:h-[350px] w-full bg-surface-dark relative bg-cover bg-center transition-all duration-700 group/banner"
        style="background-image: url('{{ $bannerUrl }}')">
        <div class="absolute inset-0 bg-gradient-to-t from-background via-background/40 to-transparent"></div>
        <div class="absolute inset-0 bg-black/20 group-hover/banner:bg-black/10 transition-colors duration-500"></div>
    </div>

    {{-- Content Area --}}
    <div class="max-w-[1440px] mx-auto px-4 md:px-8 relative -mt-20 md:-mt-24 pb-4">
        <div class="flex flex-col md:flex-row items-center md:items-end gap-6 md:gap-8">
            {{-- Avatar Wrapper --}}
            <div class="relative group/avatar">
                <div
                    class="absolute -inset-1 bg-gradient-to-tr from-primary to-primary-light rounded-full blur opacity-25 group-hover/avatar:opacity-50 transition duration-500">
                </div>
                <div
                    class="relative w-32 h-32 md:w-44 md:h-44 rounded-full border-4 border-background overflow-hidden bg-surface-dark shadow-2xl">
                    <img src="{{ $profileUrl }}" alt="{{ $user->name }}"
                        class="w-full h-full object-cover transition-transform duration-700 group-hover/avatar:scale-110">
                </div>
            </div>

            {{-- User Info --}}
            <div class="flex-1 text-center md:text-left pb-4 md:pb-6">
                <div class="flex flex-col gap-1 md:gap-2">
                    <h1 class="text-3xl md:text-5xl font-black tracking-tight text-white drop-shadow-xl">
                        {{ $user->name }}
                    </h1>
                    <div class="flex items-center justify-center md:justify-start gap-3">
                        <div class="h-1 w-12 bg-primary rounded-full"></div>
                        <span class="text-white/40 text-xs md:text-sm font-black uppercase tracking-[0.2em]">Collector
                            Archive</span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
