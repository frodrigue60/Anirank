<script lang="ts">
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

  let maxVal = $derived(Math.max(...data.map(d => d.count), 5));
  let minVal = 0;

  // Scale functions
  const x = (i: number) => padding.left + (i * (width - padding.left - padding.right)) / (data.length - 1);
  const y = (val: number) => height - padding.bottom - ((val - minVal) * (height - padding.top - padding.bottom)) / (maxVal - minVal);

  // Generate path
  let path = $derived.by(() => {
    if (data.length < 2) return "";
    return data.map((d, i) => `${i === 0 ? 'M' : 'L'} ${x(i)} ${y(d.count)}`).join(" ");
  });

  // Generate area path (for gradient)
  let areaPath = $derived.by(() => {
    if (data.length < 2) return "";
    const p = data.map((d, i) => `L ${x(i)} ${y(d.count)}`).join(" ");
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

<div class="relative group" onmousemove={handleMouseMove} onmouseleave={() => hoveredPoint = null} role="presentation">
  <svg 
    viewBox="0 0 {width} {height}" 
    class="w-full h-auto overflow-visible"
    preserveAspectRatio="none"
  >
    <defs>
      <linearGradient id="gradient-{label}" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color={color} stop-opacity="0.2" />
        <stop offset="100%" stop-color={color} stop-opacity="0" />
      </linearGradient>
    </defs>

    <!-- Grid lines (simplified) -->
    <line x1={padding.left} y1={height - padding.bottom} x2={width - padding.right} y2={height - padding.bottom} stroke="currentColor" stroke-opacity="0.05" />
    
    <!-- Area background -->
    <path d={areaPath} fill="url(#gradient-{label})" class="transition-all duration-700" />
    
    <!-- Main Line -->
    <path 
      d={path} 
      fill="none" 
      stroke={color} 
      stroke-width="3" 
      stroke-linecap="round" 
      stroke-linejoin="round"
      class="transition-all duration-700"
    />

    <!-- Active Point & Tooltip Indicator -->
    {#if hoveredPoint !== null && data[hoveredPoint]}
      <line 
        x1={x(hoveredPoint)} 
        y1={padding.top} 
        x2={x(hoveredPoint)} 
        y2={height - padding.bottom} 
        stroke={color} 
        stroke-width="1" 
        stroke-dasharray="4 4"
        class="opacity-50"
      />
      <circle 
        cx={x(hoveredPoint)} 
        cy={y(data[hoveredPoint].count)} 
        r="6" 
        fill={color} 
        stroke="white" 
        stroke-width="2" 
      />
    {/if}
  </svg>

  <!-- Tooltip overlay (HTML for better styling) -->
  {#if hoveredPoint !== null && data[hoveredPoint]}
    <div 
      class="absolute top-0 pointer-events-none bg-surface-dark border border-white/10 px-3 py-2 rounded-xl shadow-2xl z-20 -translate-x-1/2 -translate-y-full mb-4 transition-all duration-200"
      style="left: {(x(hoveredPoint) / width) * 100}%; top: {y(data[hoveredPoint].count)}px"
    >
      <div class="text-[10px] text-white/40 font-black uppercase tracking-widest leading-none mb-1">
        {data[hoveredPoint].date}
      </div>
      <div class="text-sm font-bold text-white flex items-center gap-2">
        <span class="w-2 h-2 rounded-full" style="background-color: {color}"></span>
        {data[hoveredPoint].count} {label}
      </div>
    </div>
  {/if}
</div>

<style>
  svg {
    filter: drop-shadow(0 0 10px rgba(0,0,0,0.1));
  }
</style>
