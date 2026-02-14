<div class="animate-pulse">
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        @for ($i = 0; $i < 6; $i++)
            <div class="relative overflow-hidden rounded-xl bg-white/5 aspect-[16/10] border border-white/5">
                <div class="absolute bottom-0 left-0 right-0 p-6 flex flex-col gap-3">
                    <div class="h-8 w-3/4 bg-white/10 rounded"></div>
                    <div class="flex justify-between border-t border-white/5 pt-4">
                        <div class="space-y-2">
                            <div class="h-3 w-12 bg-white/5 rounded"></div>
                            <div class="h-4 w-20 bg-white/10 rounded"></div>
                        </div>
                        <div class="space-y-2 flex flex-col items-end">
                            <div class="h-3 w-12 bg-white/5 rounded"></div>
                            <div class="h-4 w-16 bg-white/10 rounded"></div>
                        </div>
                    </div>
                </div>
            </div>
        @endfor
    </div>
</div>
