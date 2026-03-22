<script lang="ts">
  import { onMount } from "svelte";
  import {
    Users,
    Music,
    Tv,
    Star,
    MessageSquare,
    TrendingUp,
    BarChart3,
    PieChart,
  } from "lucide-svelte";
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

<div class="min-h-screen bg-background-dark pt-12 pb-24 text-white">
  <div class="max-w-6xl mx-auto px-6">
    <!-- Header -->
    <div class="mb-12">
      <h1
        class="text-4xl font-black tracking-tighter mb-2 animate-in fade-in slide-in-from-left-4"
      >
        Site Statistics
      </h1>
      <p class="text-white/40 text-lg font-medium">
        Platform growth and community engagement metrics over the last 30 days.
      </p>
    </div>

    {#if loading}
      <div
        class="flex flex-col items-center justify-center py-32 animate-pulse"
      >
        <div
          class="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin mb-4"
        ></div>
        <div class="text-sm font-black uppercase tracking-widest text-white/20">
          Aggregating Data...
        </div>
      </div>
    {:else if error}
      <div
        class="bg-red-500/10 border border-red-500/20 rounded-3xl p-12 text-center"
      >
        <div class="text-red-500 font-bold mb-4">{error}</div>
        <button
          onclick={() => window.location.reload()}
          class="bg-red-500 text-white px-8 py-3 rounded-xl font-bold text-sm uppercase tracking-widest"
          >Retry</button
        >
      </div>
    {:else if stats}
      <!-- Overview Grid -->
      <!-- <div class="grid grid-cols-2 md:grid-cols-5 gap-6 mb-12">
        <div class="bg-surface-dark border border-white/5 p-6 rounded-3xl shadow-xl hover:scale-105 transition-transform duration-300">
          <div class="text-primary mb-3"><Users size={24} /></div>
          <div class="text-2xl font-black tracking-tighter">{formatNumber(stats.overviews.total_users)}</div>
          <div class="text-[10px] font-black uppercase tracking-[0.2em] text-white/20">Total Users</div>
        </div>
        <div class="bg-surface-dark border border-white/5 p-6 rounded-3xl shadow-xl hover:scale-105 transition-transform duration-300">
          <div class="text-blue-400 mb-3"><Tv size={24} /></div>
          <div class="text-2xl font-black tracking-tighter">{formatNumber(stats.overviews.total_animes)}</div>
          <div class="text-[10px] font-black uppercase tracking-[0.2em] text-white/20">Total Animes</div>
        </div>
        <div class="bg-surface-dark border border-white/5 p-6 rounded-3xl shadow-xl hover:scale-105 transition-transform duration-300">
          <div class="text-purple-400 mb-3"><Music size={24} /></div>
          <div class="text-2xl font-black tracking-tighter">{formatNumber(stats.overviews.total_songs)}</div>
          <div class="text-[10px] font-black uppercase tracking-[0.2em] text-white/20">Total Songs</div>
        </div>
        <div class="bg-surface-dark border border-white/5 p-6 rounded-3xl shadow-xl hover:scale-105 transition-transform duration-300">
          <div class="text-yellow-400 mb-3"><Star size={24} /></div>
          <div class="text-2xl font-black tracking-tighter">{formatNumber(stats.overviews.total_ratings)}</div>
          <div class="text-[10px] font-black uppercase tracking-[0.2em] text-white/20">Total Ratings</div>
        </div>
        <div class="bg-surface-dark border border-white/5 p-6 rounded-3xl shadow-xl hover:scale-105 transition-transform duration-300">
          <div class="text-green-400 mb-3"><MessageSquare size={24} /></div>
          <div class="text-2xl font-black tracking-tighter">{formatNumber(stats.overviews.total_comments)}</div>
          <div class="text-[10px] font-black uppercase tracking-[0.2em] text-white/20">Total Comments</div>
        </div>
      </div> -->

      <!-- Growth Charts Section -->
      <div class="grid grid-cols-1 gap-8 mb-12">
        <section
          class="bg-surface-dark border border-white/5 rounded-3xl overflow-hidden shadow-2xl p-8"
        >
          <div class="flex justify-between items-center mb-8">
            <h2 class="text-lg font-bold flex items-center gap-2">
              <TrendingUp size={18} class="text-primary" /> New Ratings
            </h2>
            <div
              class="text-[10px] font-black uppercase tracking-widest text-white/20 px-3 py-1 bg-white/5 rounded-full"
            >
              Last 30 Days
            </div>
          </div>
          <StatsChart
            data={stats.rating_growth}
            color="#7f13ec"
            label="ratings"
          />
        </section>

        <section
          class="bg-surface-dark border border-white/5 rounded-3xl overflow-hidden shadow-2xl p-8"
        >
          <div class="flex justify-between items-center mb-8">
            <h2 class="text-lg font-bold flex items-center gap-2">
              <Users size={18} class="text-blue-400" /> New Users
            </h2>
            <div
              class="text-[10px] font-black uppercase tracking-widest text-white/20 px-3 py-1 bg-white/5 rounded-full"
            >
              Last 30 Days
            </div>
          </div>
          <StatsChart data={stats.user_growth} color="#3db4f2" label="users" />
        </section>
      </div>

      <!-- Distribution Section -->
      <div class="grid grid-cols-1 gap-8">
        <section
          class="bg-surface-dark border border-white/5 rounded-3xl overflow-hidden shadow-2xl p-8"
        >
          <div class="flex justify-between items-center mb-8">
            <h2 class="text-lg font-bold flex items-center gap-2">
              <BarChart3 size={18} class="text-yellow-400" /> Score Distribution
            </h2>
          </div>
          <div class="flex items-end gap-2 h-48 px-2">
            {#each stats.score_distribution as point}
              {@const maxVal = Math.max(
                ...stats.score_distribution.map((p: any) => p.value),
                1,
              )}
              <div
                class="flex-1 h-full flex flex-col justify-end items-center gap-2 group"
              >
                <div
                  class="w-full bg-yellow-400/20 hover:bg-yellow-400/40 border-t border-yellow-400/30 rounded-t-lg transition-all duration-500 relative"
                  style="height: {(point.value / maxVal) * 100}%"
                >
                  <div
                    class="absolute -top-8 left-1/2 -translate-x-1/2 bg-surface-dark border border-white/10 px-2 py-1 rounded text-[10px] font-bold opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap z-10"
                  >
                    {point.value} ratings
                  </div>
                </div>
                <span class="text-[10px] font-black uppercase text-white/20"
                  >{point.label}</span
                >
              </div>
            {/each}
          </div>
        </section>

        <section
          class="bg-surface-dark border border-white/5 rounded-3xl overflow-hidden shadow-2xl p-8"
        >
          <div class="flex justify-between items-center mb-8">
            <h2 class="text-lg font-bold flex items-center gap-2">
              <PieChart size={18} class="text-green-400" /> Level Distribution
            </h2>
          </div>
          <div class="flex items-end gap-2 h-48 px-2">
            {#each stats.level_distribution as point}
              {@const maxVal = Math.max(
                ...stats.level_distribution.map((p: any) => p.value),
                1,
              )}
              <div
                class="flex-1 h-full flex flex-col justify-end items-center gap-2 group"
              >
                <div
                  class="w-full bg-green-400/20 hover:bg-green-400/40 border-t border-green-400/30 rounded-t-lg transition-all duration-500 relative"
                  style="height: {(point.value / maxVal) * 100}%"
                >
                  <div
                    class="absolute -top-8 left-1/2 -translate-x-1/2 bg-surface-dark border border-white/10 px-2 py-1 rounded text-[10px] font-bold opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap z-10"
                  >
                    {point.value} users
                  </div>
                </div>
                <span class="text-[10px] font-black uppercase text-white/20"
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
