@extends('layouts.app')

@section('content')
    <div class="min-h-[calc(100vh-4rem)] flex items-center justify-center px-4 py-12">
        <div class="w-full max-w-md">
            {{-- Logo/Brand --}}
            <div class="text-center mb-8">
                <div
                    class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-primary shadow-lg shadow-primary/40 mb-4">
                    <span class="material-symbols-outlined text-white text-3xl">music_note</span>
                </div>
                <h1 class="text-3xl font-black text-white mb-2">Join Anirank</h1>
                <p class="text-white/40 text-sm">Create your account and start ranking anime music</p>
            </div>

            {{-- Register Card --}}
            <div class="bg-surface-dark rounded-3xl border border-white/5 shadow-2xl overflow-hidden">
                <div class="p-8">
                    <form method="POST" action="{{ route('register') }}">
                        @csrf

                        {{-- Name Field --}}
                        <div class="mb-6">
                            <label for="name" class="block text-sm font-bold text-white mb-2">
                                {{ __('Name') }}
                            </label>
                            <input id="name" type="text"
                                class="w-full bg-surface-darker border border-white/10 rounded-xl px-4 py-3 text-white placeholder:text-white/20 focus:border-primary focus:ring-1 focus:ring-primary transition-all @error('name') border-red-500 @enderror"
                                name="name" value="{{ old('name') }}" required autocomplete="name" autofocus
                                placeholder="Your name">

                            @error('name')
                                <p class="mt-2 text-sm text-red-400 flex items-center gap-1">
                                    <span class="material-symbols-outlined text-[16px]">error</span>
                                    {{ $message }}
                                </p>
                            @enderror
                        </div>

                        {{-- Email Field --}}
                        <div class="mb-6">
                            <label for="email" class="block text-sm font-bold text-white mb-2">
                                {{ __('Email Address') }}
                            </label>
                            <input id="email" type="email"
                                class="w-full bg-surface-darker border border-white/10 rounded-xl px-4 py-3 text-white placeholder:text-white/20 focus:border-primary focus:ring-1 focus:ring-primary transition-all @error('email') border-red-500 @enderror"
                                name="email" value="{{ old('email') }}" required autocomplete="email"
                                placeholder="your@email.com">

                            @error('email')
                                <p class="mt-2 text-sm text-red-400 flex items-center gap-1">
                                    <span class="material-symbols-outlined text-[16px]">error</span>
                                    {{ $message }}
                                </p>
                            @enderror
                        </div>

                        {{-- Password Field --}}
                        <div class="mb-6">
                            <label for="password" class="block text-sm font-bold text-white mb-2">
                                {{ __('Password') }}
                            </label>
                            <input id="password" type="password"
                                class="w-full bg-surface-darker border border-white/10 rounded-xl px-4 py-3 text-white placeholder:text-white/20 focus:border-primary focus:ring-1 focus:ring-primary transition-all @error('password') border-red-500 @enderror"
                                name="password" required autocomplete="new-password" placeholder="••••••••">

                            @error('password')
                                <p class="mt-2 text-sm text-red-400 flex items-center gap-1">
                                    <span class="material-symbols-outlined text-[16px]">error</span>
                                    {{ $message }}
                                </p>
                            @enderror
                        </div>

                        {{-- Confirm Password Field --}}
                        <div class="mb-6">
                            <label for="password-confirm" class="block text-sm font-bold text-white mb-2">
                                {{ __('Confirm Password') }}
                            </label>
                            <input id="password-confirm" type="password"
                                class="w-full bg-surface-darker border border-white/10 rounded-xl px-4 py-3 text-white placeholder:text-white/20 focus:border-primary focus:ring-1 focus:ring-primary transition-all"
                                name="password_confirmation" required autocomplete="new-password" placeholder="••••••••">
                        </div>

                        {{-- Submit Button --}}
                        <button type="submit"
                            class="w-full bg-primary hover:bg-primary/90 text-white font-bold py-3 rounded-xl transition-all shadow-lg shadow-primary/20 hover:shadow-primary/30 flex items-center justify-center gap-2 group">
                            <span>{{ __('Create Account') }}</span>
                            <span
                                class="material-symbols-outlined text-[20px] group-hover:translate-x-1 transition-transform">arrow_forward</span>
                        </button>
                    </form>
                </div>

                {{-- Login Link --}}
                <div class="bg-surface-darker/50 border-t border-white/5 px-8 py-4 text-center">
                    <p class="text-sm text-white/40">
                        Already have an account?
                        <a href="{{ route('login') }}" class="text-primary hover:text-white font-bold transition-colors">
                            Sign in
                        </a>
                    </p>
                </div>
            </div>
        </div>
    </div>
@endsection
