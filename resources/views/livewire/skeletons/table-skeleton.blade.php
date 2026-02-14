<div class="animate-pulse flex flex-col gap-1">
    @for ($i = 0; $i < 10; $i++)
        <div class="grid grid-cols-[60px_1fr_120px_140px] gap-4 px-8 py-5 border-b border-white/5 bg-surface-dark/10">
            <div class="h-8 w-8 bg-white/5 rounded-full mx-auto"></div>
            <div class="flex items-center gap-4">
                <div class="w-16 h-16 rounded-lg bg-white/5"></div>
                <div class="flex-1 space-y-2">
                    <div class="h-4 bg-white/5 rounded w-3/4"></div>
                    <div class="h-3 bg-white/5 rounded w-1/2"></div>
                </div>
            </div>
            <div class="h-8 w-16 bg-white/5 rounded mx-auto mt-4"></div>
            <div class="flex justify-end gap-2 mt-4">
                <div class="h-10 w-10 bg-white/5 rounded-full"></div>
                <div class="h-10 w-10 bg-white/5 rounded-full"></div>
            </div>
        </div>
    @endfor
</div>
