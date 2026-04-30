<script lang="ts">
  import Activity from "lucide-svelte/icons/activity";
  interface StatPoint {
    date: string;
    count: number;
  }

  let { 
    data = [], 
    color = "#7f13ec", 
    height = 200,
    label = ""
  } = $props<{
    data: StatPoint[];
    color?: string;
    height?: number;
    label?: string;
  }>();

  // Calculate chart dimensions
  const padding = { top: 20, right: 10, bottom: 20, left: 10 };
  const width = 1000; // Fixed internal coordinate system

  let maxVal = $derived(Math.max(...data.map((d: StatPoint) => d.count), 5));
  let minVal = 0;

  // Scale functions
  const x = (i: number) => padding.left + (i * (width - padding.left - padding.right)) / (data.length - 1);
  const y = (val: number) => height - padding.bottom - ((val - minVal) * (height - padding.top - padding.bottom)) / (maxVal - minVal);

  // Generate path
  let path = $derived.by(() => {
    if (data.length < 2) return "";
    return data.map((d: StatPoint, i: number) => `${i === 0 ? 'M' : 'L'} ${x(i)} ${y(d.count)}`).join(" ");
  });

  // Generate area path (for gradient)
  let areaPath = $derived.by(() => {
    if (data.length < 2) return "";
    const p = data.map((d: StatPoint, i: number) => `L ${x(i)} ${y(d.count)}`).join(" ");
    return `M ${x(0)} ${height - padding.bottom} ${p} L ${x(data.length - 1)} ${height - padding.bottom} Z`;
  });

  // Calculate active point on hover
  let hoveredPoint = $state<number | null>(null);

  function handleMouseMove(e: MouseEvent) {
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    const ratio = (e.clientX - rect.left) / rect.width;
    hoveredPoint = Math.round(ratio * (data.length - 1));
  }
</script>

{#if data.length > 0}
  <div class="relative group" onmousemove={handleMouseMove} onmouseleave={() => hoveredPoint = null} role="presentation">
    <svg 
      viewBox="0 0 {width} {height}" 
      class="w-full h-auto overflow-visible"
      preserveAspectRatio="none"
    >
      <defs>
        <linearGradient id="gradient-{label}" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="var(--color-primary)" stop-opacity="0.15" />
          <stop offset="100%" stop-color="var(--color-primary)" stop-opacity="0" />
        </linearGradient>
      </defs>

      <!-- Grid lines -->
      <line 
        x1={padding.left} 
        y1={height - padding.bottom} 
        x2={width - padding.right} 
        y2={height - padding.bottom} 
        stroke="var(--color-outline-variant)" 
        stroke-opacity="0.1" 
      />
      
      <!-- Area background -->
      <path d={areaPath} fill="url(#gradient-{label})" class="transition-all duration-700" />
      
      <!-- Main Line -->
      <path 
        d={path} 
        fill="none" 
        stroke="var(--color-primary)" 
        stroke-width="3" 
        stroke-linecap="round" 
        stroke-linejoin="round"
        class="transition-all duration-700 chart-line"
      />

      <!-- Active Point & Tooltip Indicator -->
      {#if hoveredPoint !== null && data[hoveredPoint]}
        <line 
          x1={x(hoveredPoint)} 
          y1={padding.top} 
          x2={x(hoveredPoint)} 
          y2={height - padding.bottom} 
          stroke="var(--color-primary)" 
          stroke-width="1" 
          stroke-dasharray="4 4"
          class="opacity-30"
        />
        <circle 
          cx={x(hoveredPoint)} 
          cy={y(data[hoveredPoint].count)} 
          r="6" 
          fill="var(--color-primary)" 
          stroke="var(--color-surface-container)" 
          stroke-width="3" 
          class="shadow-xl"
        />
      {/if}
    </svg>

    <!-- Tooltip overlay -->
    {#if hoveredPoint !== null && data[hoveredPoint]}
      <div 
        class="absolute pointer-events-none bg-surface-container border border-outline-variant/10 px-4 py-3 rounded-2xl shadow-2xl z-20 -translate-x-1/2 -translate-y-[calc(100%+12px)] transition-all duration-200 backdrop-blur-md"
        style="left: {(x(hoveredPoint) / width) * 100}%; top: {y(data[hoveredPoint].count)}px"
      >
        <div class="text-[9px] text-on-surface-variant/60 font-black uppercase tracking-[0.2em] leading-none mb-1.5 whitespace-nowrap">
          {data[hoveredPoint].date}
        </div>
        <div class="text-sm font-black text-on-surface flex items-center gap-2.5 whitespace-nowrap">
          <div class="w-2 h-2 rounded-full bg-primary shadow-[0_0_8px_rgba(var(--color-primary-rgb),0.4)]"></div>
          {data[hoveredPoint].count} <span class="text-on-surface-variant font-bold">{label}</span>
        </div>
      </div>
    {/if}
  </div>
{:else}
  <div class="flex items-center justify-center py-12 text-on-surface-variant/20">
    <div class="text-center">
      <Activity size={36} class="mb-2 opacity-10" />
      <p class="text-[10px] font-black uppercase tracking-widest">No data available</p>
    </div>
  </div>
{/if}

<style>
  svg {
    filter: drop-shadow(0 4px 12px rgba(0,0,0,0.05));
  }
  
  .chart-line {
    filter: drop-shadow(0 0 8px rgba(var(--color-primary-rgb), 0.2));
  }
</style>
