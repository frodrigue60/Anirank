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
                <h1 class="text-3xl font-black text-white mb-2">Welcome Back</h1>
                <p class="text-white/40 text-sm">Sign in to your Anirank account</p>
            </div>

            {{-- Login Card --}}
            <div class="bg-surface-dark rounded-3xl border border-white/5 shadow-2xl overflow-hidden">
                <div class="p-8">
                    <form method="POST" action="{{ route('login') }}" id="loginForm">
                        @csrf

                        {{-- Email Field --}}
                        <div class="mb-6">
                            <label for="email" class="block text-sm font-bold text-white mb-2">
                                {{ __('Email Address') }}
                            </label>
                            <input id="email" type="email"
                                class="w-full bg-surface-darker border border-white/10 rounded-xl px-4 py-3 text-white placeholder:text-white/20 focus:border-primary focus:ring-1 focus:ring-primary transition-all @error('email') border-red-500 @enderror"
                                name="email" value="{{ old('email') }}" required autocomplete="email" autofocus
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
                                name="password" required autocomplete="current-password" placeholder="••••••••">

                            @error('password')
                                <p class="mt-2 text-sm text-red-400 flex items-center gap-1">
                                    <span class="material-symbols-outlined text-[16px]">error</span>
                                    {{ $message }}
                                </p>
                            @enderror
                        </div>

                        {{-- Remember Me --}}
                        <div class="flex items-center justify-between mb-6">
                            <label class="flex items-center gap-2 cursor-pointer group">
                                <input type="checkbox" name="remember" id="remember" {{ old('remember') ? 'checked' : '' }}
                                    class="w-4 h-4 rounded border-white/10 bg-surface-darker text-primary focus:ring-1 focus:ring-primary">
                                <span
                                    class="text-sm text-white/60 group-hover:text-white transition-colors">{{ __('Remember Me') }}</span>
                            </label>

                            @if (Route::has('password.request'))
                                <a href="{{ route('password.request') }}"
                                    class="text-sm text-primary hover:text-white transition-colors">
                                    {{ __('Forgot Password?') }}
                                </a>
                            @endif
                        </div>

                        {{-- Submit Button --}}
                        <button type="submit"
                            class="w-full bg-primary hover:bg-primary/90 text-white font-bold py-3 rounded-xl transition-all shadow-lg shadow-primary/20 hover:shadow-primary/30 flex items-center justify-center gap-2 group">
                            <span>{{ __('Login') }}</span>
                            <span
                                class="material-symbols-outlined text-[20px] group-hover:translate-x-1 transition-transform">arrow_forward</span>
                        </button>

                        {{-- Response Message --}}
                        <div id="responseMessage" class="mt-4"></div>
                    </form>
                </div>

                {{-- Register Link --}}
                @if (Route::has('register'))
                    <div class="bg-surface-darker/50 border-t border-white/5 px-8 py-4 text-center">
                        <p class="text-sm text-white/40">
                            Don't have an account?
                            <a href="{{ route('register') }}"
                                class="text-primary hover:text-white font-bold transition-colors">
                                Sign up
                            </a>
                        </p>
                    </div>
                @endif
            </div>
        </div>
    </div>
@endsection

@section('script')
    <script>
        document.getElementById('loginForm').addEventListener('submit', function(event) {
            event.preventDefault();
            loginWithAjax();
        });

        function loginWithAjax() {
            const form = document.getElementById('loginForm');
            const formData = new FormData(form);
            const responseDiv = document.getElementById('responseMessage');
            const submitBtn = form.querySelector('button[type="submit"]');

            // Disable button and show loading
            submitBtn.disabled = true;
            submitBtn.innerHTML =
                '<span class="material-symbols-outlined animate-spin">progress_activity</span> Logging in...';

            fetch(form.action, {
                    method: 'POST',
                    body: formData,
                    headers: {
                        'X-Requested-With': 'XMLHttpRequest',
                        'Accept': 'application/json',
                    },
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        if (data.token) {
                            localStorage.setItem('api_token', data.token);
                        }
                        responseDiv.innerHTML =
                            '<div class="bg-green-500/10 border border-green-500/20 rounded-xl p-3 text-sm text-green-400 flex items-center gap-2"><span class="material-symbols-outlined text-[18px]">check_circle</span>Success! Redirecting...</div>';
                        setTimeout(() => {
                            window.location.href = data.redirect || '/';
                        }, 500);
                    } else {
                        responseDiv.innerHTML =
                            '<div class="bg-red-500/10 border border-red-500/20 rounded-xl p-3 text-sm text-red-400 flex items-center gap-2"><span class="material-symbols-outlined text-[18px]">error</span>' +
                            (data.message || 'Login failed') + '</div>';
                        submitBtn.disabled = false;
                        submitBtn.innerHTML =
                            '<span>{{ __('Login') }}</span><span class="material-symbols-outlined text-[20px] group-hover:translate-x-1 transition-transform">arrow_forward</span>';
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    responseDiv.innerHTML =
                        '<div class="bg-red-500/10 border border-red-500/20 rounded-xl p-3 text-sm text-red-400 flex items-center gap-2"><span class="material-symbols-outlined text-[18px]">error</span>Network error. Please try again.</div>';
                    submitBtn.disabled = false;
                    submitBtn.innerHTML =
                        '<span>{{ __('Login') }}</span><span class="material-symbols-outlined text-[20px] group-hover:translate-x-1 transition-transform">arrow_forward</span>';
                });
        }
    </script>
@endsection
