<div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
    {{-- Profile Media Settings --}}
    <div class="flex flex-col gap-8">
        {{-- PROFILE PICTURE CARD --}}
        <div class="bg-surface-dark/30 p-6 rounded-2xl border border-white/5 shadow-2xl backdrop-blur-md">
            <div class="flex items-center gap-3 mb-6">
                <span class="material-symbols-outlined text-primary">account_circle</span>
                <h2 class="text-xl font-bold text-white uppercase tracking-tight">Profile Avatar</h2>
            </div>

            <form wire:submit.prevent="saveAvatar" class="space-y-4">
                <div class="relative group">
                    <label for="profile-file"
                        class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Upload
                        new avatar</label>

                    {{-- Avatar Preview --}}
                    @if ($image)
                        <div class="mb-4 relative w-20 h-20">
                            <img src="{{ $image->temporaryUrl() }}"
                                class="w-full h-full object-cover rounded-full border-2 border-primary shadow-lg shadow-primary/20">
                            <div class="absolute -top-1 -right-1 bg-primary text-white rounded-full p-0.5">
                                <span class="material-symbols-outlined text-xs">sync</span>
                            </div>
                        </div>
                    @endif

                    <div class="flex flex-col sm:flex-row gap-3">
                        <div class="flex-1 relative">
                            <input type="file" id="profile-file" wire:model="image"
                                accept="image/jpg, image/jpeg, image/png, image/webp"
                                class="w-full bg-surface-darker border border-white/10 rounded-lg py-2 px-3 text-sm text-white/60 file:mr-4 file:py-1.5 file:px-4 file:rounded-md file:border-0 file:text-xs file:font-black file:uppercase file:bg-primary/20 file:text-primary hover:file:bg-primary/30 transition-all cursor-pointer">
                            <div wire:loading wire:target="image" class="absolute right-3 top-1/2 -translate-y-1/2">
                                <div
                                    class="w-4 h-4 border-2 border-primary/20 border-t-primary rounded-full animate-spin">
                                </div>
                            </div>
                        </div>
                        <button
                            class="bg-primary hover:bg-primary-light text-white text-xs font-black uppercase px-6 py-2.5 rounded-lg transition-all shadow-lg shadow-primary/20 shrink-0 disabled:opacity-50 disabled:cursor-not-allowed"
                            type="submit" wire:loading.attr="disabled" wire:target="image, saveAvatar">
                            <span wire:loading.remove wire:target="saveAvatar">Save</span>
                            <span wire:loading wire:target="saveAvatar" class="flex items-center gap-2">
                                <div class="w-3 h-3 border-2 border-white/20 border-t-white rounded-full animate-spin">
                                </div>
                                Saving...
                            </span>
                        </button>
                    </div>

                    @error('image') <span class="text-xs text-red-500 mt-1 block">{{ $message }}</span> @enderror

                    @if (session()->has('avatar_success'))
                        <div class="mt-2 text-xs text-green-400 font-bold flex items-center gap-1">
                            <span class="material-symbols-outlined text-[14px]">check_circle</span>
                            {{ session('avatar_success') }}
                        </div>
                    @endif
                </div>
            </form>
        </div>

        {{-- BANNER PICTURE CARD --}}
        <div class="bg-surface-dark/30 p-6 rounded-2xl border border-white/5 shadow-2xl backdrop-blur-md">
            <div class="flex items-center gap-3 mb-6">
                <span class="material-symbols-outlined text-primary">image</span>
                <h2 class="text-xl font-bold text-white uppercase tracking-tight">Profile Banner</h2>
            </div>

            <form wire:submit.prevent="saveBanner" class="space-y-4">
                <div class="relative group">
                    <label for="banner-file"
                        class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Upload
                        new banner</label>

                    {{-- Banner Preview --}}
                    @if ($banner)
                        <div
                            class="mb-4 relative w-full aspect-[4/1] rounded-lg overflow-hidden border-2 border-primary shadow-lg shadow-primary/20">
                            <img src="{{ $banner->temporaryUrl() }}" class="w-full h-full object-cover">
                            <div class="absolute top-2 right-2 bg-primary text-white rounded-full p-1 shadow-lg">
                                <span class="material-symbols-outlined text-sm">sync</span>
                            </div>
                        </div>
                    @endif

                    <div class="flex flex-col sm:flex-row gap-3">
                        <div class="flex-1 relative">
                            <input type="file" id="banner-file" wire:model="banner"
                                accept="image/jpg, image/jpeg, image/png, image/webp"
                                class="w-full bg-surface-darker border border-white/10 rounded-lg py-2 px-3 text-sm text-white/60 file:mr-4 file:py-1.5 file:px-4 file:rounded-md file:border-0 file:text-xs file:font-black file:uppercase file:bg-primary/20 file:text-primary hover:file:bg-primary/30 transition-all cursor-pointer">
                            <div wire:loading wire:target="banner" class="absolute right-3 top-1/2 -translate-y-1/2">
                                <div
                                    class="w-4 h-4 border-2 border-primary/20 border-t-primary rounded-full animate-spin">
                                </div>
                            </div>
                        </div>
                        <button
                            class="bg-primary hover:bg-primary-light text-white text-xs font-black uppercase px-6 py-2.5 rounded-lg transition-all shadow-lg shadow-primary/20 shrink-0 disabled:opacity-50 disabled:cursor-not-allowed"
                            type="submit" wire:loading.attr="disabled" wire:target="banner, saveBanner">
                            <span wire:loading.remove wire:target="saveBanner">Save</span>
                            <span wire:loading wire:target="saveBanner" class="flex items-center gap-2">
                                <div class="w-3 h-3 border-2 border-white/20 border-t-white rounded-full animate-spin">
                                </div>
                                Saving...
                            </span>
                        </button>
                    </div>

                    @error('banner') <span class="text-xs text-red-500 mt-1 block">{{ $message }}</span> @enderror

                    @if (session()->has('banner_success'))
                        <div class="mt-2 text-xs text-green-400 font-bold flex items-center gap-1">
                            <span class="material-symbols-outlined text-[14px]">check_circle</span>
                            {{ session('banner_success') }}
                        </div>
                    @endif
                </div>
            </form>
        </div>
    </div>

    {{-- Site Preferences --}}
    <div class="flex flex-col gap-8">
        {{-- SCORE FORMAT CARD --}}
        <div
            class="bg-surface-dark/30 p-6 rounded-2xl border border-white/5 shadow-2xl backdrop-blur-md h-full flex flex-col">
            <div class="flex items-center gap-3 mb-6">
                <span class="material-symbols-outlined text-primary">settings</span>
                <h2 class="text-xl font-bold text-white uppercase tracking-tight">Site Settings</h2>
            </div>

            <form wire:submit.prevent="saveScoreFormat" class="flex flex-col flex-1">
                <div class="relative group flex-1">
                    <label for="select-score-format"
                        class="block text-[10px] uppercase font-black text-white/40 mb-1.5 ml-1 tracking-widest group-hover:text-primary transition-colors">Scoring
                        System Preference</label>
                    <div class="relative">
                        <select
                            class="w-full bg-surface-darker border border-white/10 rounded-lg py-3 pl-4 pr-10 text-sm text-white focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer hover:bg-surface-darker/80"
                            id="select-score-format" wire:model="score_format" required>
                            <option value="">Select Scoring System</option>
                            @foreach ($score_formats as $item)
                                <option value="{{ $item['value'] }}">
                                    {{ $item['name'] }}
                                </option>
                            @endforeach
                        </select>
                        <span
                            class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-white/30 pointer-events-none text-lg group-focus-within:text-primary transition-colors">expand_more</span>
                    </div>
                    <p class="mt-4 text-xs text-white/40 leading-relaxed italic">
                        Choose how you want to rate and see scores across the platform. This change will be
                        applied instantly to all ratings.
                    </p>
                </div>

                <div class="pt-6 mt-6 border-t border-white/5">
                    <button
                        class="w-full bg-primary hover:bg-primary-light text-white text-xs font-black uppercase py-4 rounded-xl transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2 disabled:opacity-50"
                        type="submit" wire:loading.attr="disabled">
                        <span wire:loading.remove wire:target="saveScoreFormat" class="flex items-center gap-2">
                            <span class="material-symbols-outlined text-[18px]">save</span>
                            Apply Site Preferences
                        </span>
                        <span wire:loading wire:target="saveScoreFormat" class="flex items-center gap-2">
                            <div class="w-3 h-3 border-2 border-white/20 border-t-white rounded-full animate-spin">
                            </div>
                            Applying...
                        </span>
                    </button>

                    @if (session()->has('settings_success'))
                        <div
                            class="mt-4 p-3 rounded-lg bg-green-500/10 border border-green-500/20 text-xs text-green-400 font-bold flex items-center gap-2 animate-pulse">
                            <span class="material-symbols-outlined text-[16px]">check_circle</span>
                            {{ session('settings_success') }}
                        </div>
                    @endif
                </div>
            </form>
        </div>
    </div>
</div>