<div class="w-full flex flex-col items-center" x-data="{ score: @entangle('ratingValue').defer }">
    <style>
        input[type=range].rating-range-10::-webkit-slider-runnable-track {
            background: linear-gradient(to right, #7f13ec, #a855f7);
            height: 8px;
            border-radius: 4px;
        }
        input[type=range].rating-range-10::-webkit-slider-thumb {
            -webkit-appearance: none;
            height: 24px;
            width: 24px;
            border-radius: 50%;
            background: #ffffff;
            cursor: pointer;
            margin-top: -8px;
            box-shadow: 0 0 15px rgba(127, 19, 236, 0.5);
            border: 2px solid #7f13ec;
            transition: transform 0.1s ease;
        }
        input[type=range].rating-range-10:active::-webkit-slider-thumb {
            transform: scale(1.1);
        }
        input[type=range].rating-range-10 {
            -webkit-appearance: none;
            background: transparent;
        }
    </style>

    {{-- <div class="w-full flex justify-between items-start mb-8">
        <div class="text-left">
            <p class="text-[10px] font-bold uppercase tracking-[0.2em] text-primary mb-1">Rate this theme</p>
            <h3 class="text-2xl font-bold leading-tight">{{ $song->name }}</h3>
            <p class="text-sm text-white/50">{{ $song->post->title }} • {{ $song->type }} {{ $song->theme_num }}</p>
        </div>
    </div> --}}

    <div class="my-6 relative">
        <div class="text-[84px] font-bold leading-none tracking-tighter text-white drop-shadow-[0_0_20px_rgba(127,19,236,0.3)] text-center" x-text="(score / 10).toFixed(1)">
            {{ number_format($ratingValue / 10, 1) }}
        </div>
        <p class="text-sm text-white/40 font-medium uppercase tracking-widest mt-2">Personal Score</p>
    </div>

    <div class="w-full space-y-8 my-4">
        <div class="px-2">
            <input type="range" 
                class="rating-range-10 w-full h-2 bg-white/10 rounded-full appearance-none cursor-pointer focus:outline-none" 
                max="100" min="0" step="1" 
                x-model="score">
            
            <div class="flex justify-between mt-4 px-1">
                <span class="text-[10px] font-bold text-white/20">0.0</span>
                <span class="text-[10px] font-bold text-white/20">2.5</span>
                <span class="text-[10px] font-bold text-white/20">5.0</span>
                <span class="text-[10px] font-bold text-white/20">7.5</span>
                <span class="text-[10px] font-bold text-white/20">10.0</span>
            </div>
        </div>

        <div class="grid grid-cols-5 gap-2 w-full max-w-sm">
            @for ($i = 1; $i <= 10; $i++)
                <button @click="score = {{ $i * 10 }}" 
                    class="aspect-square flex items-center justify-center rounded-full border text-sm font-bold transition-all"
                    :class="Math.round(score/10) == {{ $i }} ? 'bg-primary/20 border-primary/40 text-primary' : 'bg-white/5 border-white/5 text-white/60 hover:bg-white/15 hover:border-white/20 hover:text-white'">
                    {{ $i }}
                </button>
            @endfor
        </div>
    </div>

</div>
