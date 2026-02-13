<div class="w-full flex flex-col items-center" x-data="{ score: @entangle('ratingValue').defer }">
    <style>
        input[type=range].rating-range::-webkit-slider-runnable-track {
            background: linear-gradient(to right, #7f13ec, #a855f7);
            height: 8px;
            border-radius: 4px;
        }
        input[type=range].rating-range::-webkit-slider-thumb {
            -webkit-appearance: none;
            height: 24px;
            width: 24px;
            border-radius: 50%;
            background: #ffffff;
            cursor: pointer;
            margin-top: -8px;
            box-shadow: 0 0 15px rgba(127, 19, 236, 0.5);
            border: 2px solid #7f13ec;
        }
        input[type=range].rating-range {
            -webkit-appearance: none;
            background: transparent;
        }
    </style>

    {{-- <div class="w-full flex justify-between items-start mb-6">
        <div class="text-left">
            <p class="text-[10px] font-bold uppercase tracking-[0.2em] text-primary mb-1">Now Rating</p>
            <h3 class="text-xl font-bold leading-tight">{{ $song->name }}</h3>
            <p class="text-sm text-white/60">
                {{ $song->post->title }} • {{ $song->type }} {{ $song->theme_num }}
            </p>
        </div>
    </div> --}}

    <div class="my-6">
        <div class="text-[84px] font-bold leading-none tracking-tighter text-white drop-shadow-[0_0_20px_rgba(127,19,236,0.3)] text-center" x-text="score">
            {{ $ratingValue }}
        </div>
        <p class="text-sm text-white/40 font-medium uppercase tracking-widest mt-2">Personal Score</p>
    </div>

    <div class="w-full space-y-8 my-4">
        <div class="px-2">
            <input type="range" 
                class="rating-range w-full h-2 bg-white/10 rounded-full appearance-none cursor-pointer focus:outline-none" 
                max="100" min="0" step="1"
                x-model="score">
            <div class="flex justify-between mt-4 px-1">
                <span class="text-[10px] font-bold text-white/20">0</span>
                <span class="text-[10px] font-bold text-white/20">25</span>
                <span class="text-[10px] font-bold text-white/20">50</span>
                <span class="text-[10px] font-bold text-white/20">75</span>
                <span class="text-[10px] font-bold text-white/20">100</span>
            </div>
        </div>
        <div class="grid grid-cols-4 gap-2">
            @foreach([50, 75, 90, 100] as $val)
                <button @click="score = {{ $val }}" 
                    class="py-2.5 rounded-xl border transition-all text-sm font-bold"
                    :class="score == {{ $val }} ? 'bg-primary/20 border-primary/30 text-primary' : 'bg-white/5 border-white/5 text-white/60 hover:bg-white/10 hover:border-white/10'">
                    {{ $val }}
                </button>
            @endforeach
        </div>
    </div>

</div>
