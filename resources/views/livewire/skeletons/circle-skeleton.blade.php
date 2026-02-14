<div class="animate-pulse">
    <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-8">
        @for ($i = 0; $i < 12; $i++)
            <div class="flex flex-col items-center gap-4">
                <div class="w-full aspect-square rounded-full bg-white/5 border border-white/5"></div>
                <div class="h-4 w-24 bg-white/5 rounded mx-auto"></div>
                <div class="h-6 w-20 bg-white/5 rounded-full mx-auto"></div>
            </div>
        @endfor
    </div>
</div>
