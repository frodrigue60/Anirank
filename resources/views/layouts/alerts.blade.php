@if (session('success'))
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-4">
        <div class="flex items-center p-4 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 rounded-2xl backdrop-blur-xl shadow-xl shadow-emerald-900/10 animate-in fade-in slide-in-from-top-4 duration-500"
            role="alert">
            <div class="flex-shrink-0 w-8 h-8 flex items-center justify-center bg-emerald-500/20 rounded-full mr-3">
                <i class="fa-solid fa-check text-sm"></i>
            </div>
            <div class="text-sm font-bold tracking-tight">
                {{ session('success') }}
            </div>
            <button type="button"
                class="ml-auto w-8 h-8 flex items-center justify-center text-emerald-400/50 hover:text-emerald-400 hover:bg-emerald-500/20 rounded-xl transition-all"
                onclick="this.closest('[role=alert]').remove()" aria-label="Close">
                <i class="fa-solid fa-xmark"></i>
            </button>
        </div>
    </div>
@endif

@if (session('warning'))
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-4">
        <div class="flex items-center p-4 bg-amber-500/10 border border-amber-500/20 text-amber-400 rounded-2xl backdrop-blur-xl shadow-xl shadow-amber-900/10 animate-in fade-in slide-in-from-top-4 duration-500"
            role="alert">
            <div class="flex-shrink-0 w-8 h-8 flex items-center justify-center bg-amber-500/20 rounded-full mr-3">
                <i class="fa-solid fa-triangle-exclamation text-sm"></i>
            </div>
            <div class="text-sm font-bold tracking-tight">
                {{ session('warning') }}
            </div>
            <button type="button"
                class="ml-auto w-8 h-8 flex items-center justify-center text-amber-400/50 hover:text-amber-400 hover:bg-amber-500/20 rounded-xl transition-all"
                onclick="this.closest('[role=alert]').remove()" aria-label="Close">
                <i class="fa-solid fa-xmark"></i>
            </button>
        </div>
    </div>
@endif

@if (session('error') || session('danger'))
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-4">
        <div class="flex items-center p-4 bg-rose-500/10 border border-rose-500/20 text-rose-400 rounded-2xl backdrop-blur-xl shadow-xl shadow-rose-900/10 animate-in fade-in slide-in-from-top-4 duration-500"
            role="alert">
            <div class="flex-shrink-0 w-8 h-8 flex items-center justify-center bg-rose-500/20 rounded-full mr-3">
                <i class="fa-solid fa-circle-xmark text-sm"></i>
            </div>
            <div class="text-sm font-bold tracking-tight">
                {{ session('error') ?? session('danger') }}
            </div>
            <button type="button"
                class="ml-auto w-8 h-8 flex items-center justify-center text-rose-400/50 hover:text-rose-400 hover:bg-rose-500/20 rounded-xl transition-all"
                onclick="this.closest('[role=alert]').remove()" aria-label="Close">
                <i class="fa-solid fa-xmark"></i>
            </button>
        </div>
    </div>
@endif

@if ($errors->any())
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-4">
        <div class="p-5 bg-rose-500/10 border border-rose-500/20 text-rose-400 rounded-2xl backdrop-blur-xl shadow-xl shadow-rose-900/10 animate-in fade-in slide-in-from-top-4 duration-500"
            role="alert">
            <div class="flex items-center mb-3">
                <div class="flex-shrink-0 w-8 h-8 flex items-center justify-center bg-rose-500/20 rounded-full mr-3">
                    <i class="fa-solid fa-bug text-sm"></i>
                </div>
                <span class="text-sm font-black uppercase tracking-widest text-white">Validation Errors</span>
                <button type="button"
                    class="ml-auto w-8 h-8 flex items-center justify-center text-rose-400/50 hover:text-rose-400 hover:bg-rose-500/20 rounded-xl transition-all"
                    onclick="this.closest('[role=alert]').remove()" aria-label="Close">
                    <i class="fa-solid fa-xmark"></i>
                </button>
            </div>
            <ul class="space-y-1 ml-11">
                @foreach ($errors->all() as $error)
                    <li class="text-xs font-medium text-rose-400/80">• {{ $error }}</li>
                @endforeach
            </ul>
        </div>
    </div>
@endif
