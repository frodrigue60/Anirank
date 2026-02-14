<div class="animate-pulse">
    <div class="grid grid-cols-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 md:gap-6">
        @for ($i = 0; $i < 12; $i++)
            <div class="flex flex-col gap-2">
                <div class="aspect-[2/3] rounded-lg bg-white/5"></div>
                <div class="h-4 bg-white/5 rounded w-3/4 mt-1"></div>
            </div>
        @endfor
    </div>
</div>
