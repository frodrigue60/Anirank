<div class="w-full flex flex-col items-center" x-data="{ score: @entangle('ratingValue').defer }">
    <style>
        .star-gradient-fill {
            background: linear-gradient(135deg, #7f13ec, #a855f7);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            filter: drop-shadow(0 0 8px rgba(127, 19, 236, 0.6));
        }
    </style>

    <div class="my-6 relative">
        <div class="text-[84px] font-bold leading-none tracking-tighter text-white drop-shadow-[0_0_20px_rgba(127,19,236,0.3)] text-center" x-text="(score / 20).toFixed(1)">
            {{ number_format($ratingValue / 20, 1) }}
        </div>
        <p class="text-sm text-white/40 font-medium uppercase tracking-widest mt-2">Personal Score</p>
    </div>

    <div class="flex items-center justify-center gap-1 my-8">
        @for ($i = 1; $i <= 5; $i++)
            <button @click="score = {{ $i * 20 }}" class="group p-1 relative">
                {{-- Empty Star --}}
                <span class="material-symbols-outlined text-[48px] text-white/10"
                    :class="Math.round(score/20) >= {{ $i }} ? 'star-gradient-fill filled' : ''">
                    star
                </span>
            </button>
        @endfor
    </div>

</div>
