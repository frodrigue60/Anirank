<script lang="ts">
  import { onMount } from "svelte";
  import Users from "lucide-svelte/icons/users";
import Music from "lucide-svelte/icons/music";
import Tv from "lucide-svelte/icons/tv";
import Star from "lucide-svelte/icons/star";
import MessageSquare from "lucide-svelte/icons/message-square";
import TrendingUp from "lucide-svelte/icons/trending-up";
import BarChart3 from "lucide-svelte/icons/bar-chart-3";
import PieChart from "lucide-svelte/icons/pie-chart";;
  import StatsChart from "$lib/components/StatsChart.svelte";
  import api from "$lib/api";

  let stats = $state<any>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  onMount(async () => {
    try {
      const response = await api.get("/site-statistics");
      stats = response.data;
    } catch (err: any) {
      error = err.message || "Failed to load statistics";
    } finally {
      loading = false;
    }
  });

  function formatNumber(num: number) {
    return new Intl.NumberFormat().format(num);
  }
</script>

<div class="min-h-screen bg-surface pt-12 pb-24 text-on-surface">
  <div class="max-w-6xl mx-auto px-6">
    <!-- Header -->
    <div class="mb-12">
      <h1
        class="text-4xl font-black tracking-tighter mb-2 animate-in fade-in slide-in-from-left-4 text-on-surface"
      >
        Site Statistics
      </h1>
      <p class="text-on-surface-variant/60 text-lg font-medium">
        Platform growth and community engagement metrics over the last 30 days.
      </p>
    </div>

    {#if loading}
      <div
        class="flex flex-col items-center justify-center py-32 animate-pulse"
      >
        <div
          class="w-12 h-12 border-4 border-outline-variant/20 border-t-primary rounded-full animate-spin mb-4"
        ></div>
        <div class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant/20">
          Aggregating Data...
        </div>
      </div>
    {:else if error}
      <div
        class="bg-red-500/5 border border-red-500/10 rounded-md p-12 text-center"
      >
        <div class="text-red-500 font-bold mb-6">{error}</div>
        <button
          onclick={() => window.location.reload()}
          class="bg-red-500 hover:bg-red-600 text-white px-10 py-4 rounded-sm font-black text-sm uppercase tracking-widest transition-all active:scale-95"
          >Retry Connection</button
        >
      </div>
    {:else if stats}
      <!-- Overview Grid (Optional refinement) -->

      <!-- Growth Charts Section -->
      <div class="grid grid-cols-1 gap-8 mb-12">
        <section
          class="bg-surface-container border border-outline-variant/10 rounded-md overflow-hidden shadow-2xl p-10"
        >
          <div class="flex justify-between items-center mb-10">
            <h2 class="text-xl font-black tracking-tight flex items-center gap-3 text-on-surface">
              <div class="w-10 h-10 rounded-md bg-primary/10 flex items-center justify-center text-primary">
                <TrendingUp size={20} />
              </div>
              New Ratings
            </h2>
            <div
              class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant/40 px-4 py-2 bg-surface-highest/50 rounded-full border border-outline-variant/5"
            >
              Last 30 Days
            </div>
          </div>
          <StatsChart
            data={stats.rating_growth}
            color="var(--color-primary)"
            label="ratings"
          />
        </section>

        <section
          class="bg-surface-container border border-outline-variant/10 rounded-md overflow-hidden shadow-2xl p-10"
        >
          <div class="flex justify-between items-center mb-10">
            <h2 class="text-xl font-black tracking-tight flex items-center gap-3 text-on-surface">
              <div class="w-10 h-10 rounded-md bg-secondary-container/10 flex items-center justify-center text-secondary-container">
                <Users size={20} />
              </div>
              New Users
            </h2>
            <div
              class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant/40 px-4 py-2 bg-surface-highest/50 rounded-full border border-outline-variant/5"
            >
              Last 30 Days
            </div>
          </div>
          <StatsChart 
            data={stats.user_growth} 
            color="var(--color-secondary-container)" 
            label="users" 
          />
        </section>
      </div>

      <!-- Distribution Section -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
        <section
          class="bg-surface-container border border-outline-variant/10 rounded-md overflow-hidden shadow-2xl p-10"
        >
          <div class="flex justify-between items-center mb-10">
            <h2 class="text-xl font-black tracking-tight flex items-center gap-3 text-on-surface">
              <div class="w-10 h-10 rounded-md bg-rating-star/10 flex items-center justify-center text-rating-star">
                <BarChart3 size={20} />
              </div>
              Score Distribution
            </h2>
          </div>
          <div class="flex items-end gap-2.5 h-64 px-2">
            {#each stats.score_distribution as point}
              {@const maxVal = Math.max(
                ...stats.score_distribution.map((p: any) => p.value),
                1,
              )}
              <div
                class="flex-1 h-full flex flex-col justify-end items-center gap-4 group"
              >
                <div
                  class="w-full bg-rating-star/10 hover:bg-rating-star/20 border-t-2 border-rating-star/30 rounded-t-md transition-all duration-500 relative cursor-pointer"
                  style="height: {(point.value / maxVal) * 100}%"
                >
                  <div
                    class="absolute -top-12 left-1/2 -translate-x-1/2 bg-surface-highest border border-outline-variant/20 px-3 py-1.5 rounded-md text-[10px] font-black text-on-surface shadow-xl opacity-0 scale-90 group-hover:opacity-100 group-hover:scale-100 transition-all whitespace-nowrap z-10 pointer-events-none"
                  >
                    {point.value} ratings
                  </div>
                </div>
                <span class="text-[9px] font-black uppercase tracking-tighter text-on-surface-variant/30 group-hover:text-on-surface transition-colors"
                  >{point.label}</span
                >
              </div>
            {/each}
          </div>
        </section>

        <section
          class="bg-surface-container border border-outline-variant/10 rounded-md overflow-hidden shadow-2xl p-10"
        >
          <div class="flex justify-between items-center mb-10">
            <h2 class="text-xl font-black tracking-tight flex items-center gap-3 text-on-surface">
              <div class="w-10 h-10 rounded-md bg-green-500/10 flex items-center justify-center text-green-500">
                <PieChart size={20} />
              </div>
              Level Distribution
            </h2>
          </div>
          <div class="flex items-end gap-2.5 h-64 px-2">
            {#each stats.level_distribution as point}
              {@const maxVal = Math.max(
                ...stats.level_distribution.map((p: any) => p.value),
                1,
              )}
              <div
                class="flex-1 h-full flex flex-col justify-end items-center gap-4 group"
              >
                <div
                  class="w-full bg-green-500/10 hover:bg-green-500/20 border-t-2 border-green-500/30 rounded-t-md transition-all duration-500 relative cursor-pointer"
                  style="height: {(point.value / maxVal) * 100}%"
                >
                  <div
                    class="absolute -top-12 left-1/2 -translate-x-1/2 bg-surface-highest border border-outline-variant/20 px-3 py-1.5 rounded-md text-[10px] font-black text-on-surface shadow-xl opacity-0 scale-90 group-hover:opacity-100 group-hover:scale-100 transition-all whitespace-nowrap z-10 pointer-events-none"
                  >
                    {point.value} users
                  </div>
                </div>
                <span class="text-[9px] font-black uppercase tracking-tighter text-on-surface-variant/30 group-hover:text-on-surface transition-colors"
                  >{point.label}</span
                >
              </div>
            {/each}
          </div>
        </section>
      </div>
    {/if}
  </div>
</div>
