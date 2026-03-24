<script lang="ts">
  let { 
    xp = 0, 
    level = 1, 
    nextLevelXP = 0, 
    currentLevelMinXP = 0,
    accentColor = "#3db4f2"
  } = $props();

  let progress = $derived(() => {
    if (nextLevelXP <= currentLevelMinXP) return 0;
    const currentProgress = xp - currentLevelMinXP;
    const needed = nextLevelXP - currentLevelMinXP;
    return Math.min(100, Math.max(0, (currentProgress / needed) * 100));
  });
</script>

<div class="flex flex-col gap-2 w-full max-w-md">
  <div class="flex justify-between items-end">
    <div class="flex items-center gap-2">
      <span class="text-xs font-black uppercase tracking-widest" style="color: {accentColor}"
        >Level</span
      >
      <span class="text-2xl font-black text-white italic leading-none"
        >{level}</span
      >
    </div>
    <div class="text-right">
      <span class="text-[10px] font-bold text-slate-500 uppercase tracking-tighter"
        >Experience Points</span
      >
      <div class="flex items-baseline justify-end gap-1">
        <span class="text-sm font-black text-white">{xp.toLocaleString()}</span>
        <span class="text-[10px] font-bold text-slate-500"
          >/ {nextLevelXP.toLocaleString()}</span
        >
      </div>
    </div>
  </div>

  <div class="relative h-2 w-full bg-white/5 rounded-full overflow-hidden border border-white/5">
    <div
      class="absolute inset-y-0 left-0 transition-all duration-1000 ease-out shadow-lg"
      style="width: {progress()}%; background-color: {accentColor}; box-shadow: 0 0 15px {accentColor}80"
    >
      <div class="absolute inset-0 bg-linear-to-r from-transparent via-white/20 to-transparent animate-shimmer"></div>
    </div>
  </div>
</div>

<style>
  @keyframes shimmer {
    0% { transform: translateX(-100%); }
    100% { transform: translateX(100%); }
  }
  .animate-shimmer {
    animation: shimmer 2s infinite linear;
  }
</style>
