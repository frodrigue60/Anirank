@extends('layouts.app')

@section('content')
    <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {{-- Header Section --}}
        <div class="mb-8">
            <a href="{{ route('admin.users.index') }}"
                class="text-blue-500 hover:text-blue-400 text-sm font-bold flex items-center mb-2 transition-colors">
                <i class="fa-solid fa-arrow-left mr-2"></i> BACK TO USER DIRECTORY
            </a>
            <h1 class="text-3xl font-bold text-white tracking-tight">Create User Account</h1>
            <p class="text-zinc-400 mt-1">Onboard a new platform member or staff account.</p>
        </div>

        {{-- Form Card --}}
        <div class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl shadow-xl overflow-hidden p-8">
            <form method="POST" action="{{ route('admin.users.store') }}" class="space-y-8">
                @csrf

                {{-- User Credentials --}}
                <div class="space-y-6">
                    <h3 class="text-xs font-bold text-zinc-500 uppercase tracking-widest flex items-center">
                        <i class="fa-solid fa-user-lock mr-2 text-blue-500"></i> ACCOUNT DETAILS
                    </h3>

                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div class="space-y-2">
                            <label for="nameUser"
                                class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Full Name</label>
                            <input type="text" name="name" id="nameUser" required value="{{ old('name') }}"
                                class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12"
                                placeholder="John Doe">
                        </div>

                        <div class="space-y-2">
                            <label for="emailUser"
                                class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Email
                                Address</label>
                            <input type="email" name="email" id="emailUser" required value="{{ old('email') }}"
                                class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12"
                                placeholder="john@example.com">
                        </div>
                    </div>

                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div class="space-y-2">
                            <label for="userPass"
                                class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Secure
                                Password</label>
                            <input type="password" name="password" id="userPass" required
                                class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12"
                                placeholder="••••••••">
                        </div>

                        <div class="space-y-2">
                            <label for="userType"
                                class="block text-sm font-bold text-zinc-400 uppercase tracking-widest">Platform
                                Role</label>
                            <select name="userType" id="userType" required
                                class="block w-full bg-zinc-950/50 border border-zinc-800 text-white rounded-2xl px-4 py-3 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all text-sm h-12">
                                <option value="">Select a role</option>
                                @foreach ($type as $item)
                                    <option value="{{ $item['value'] }}"
                                        {{ old('userType') == $item['value'] ? 'selected' : '' }}>
                                        {{ $item['name'] }}
                                    </option>
                                @endforeach
                            </select>
                        </div>
                    </div>
                </div>

                {{-- Action --}}
                <div class="pt-4">
                    <button
                        class="w-full bg-blue-600 hover:bg-blue-500 text-white font-bold py-4 px-6 rounded-2xl transition-all shadow-lg shadow-blue-900/20 active:scale-[0.98] flex items-center justify-center gap-2 text-sm uppercase tracking-widest">
                        <i class="fa-solid fa-user-plus"></i>
                        CREATE ACCOUNT
                    </button>
                </div>
            </form>
        </div>
    </div>
@endsection
