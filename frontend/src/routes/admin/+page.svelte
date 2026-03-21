<script lang="ts">
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";

  let { data } = $props();

  // Mapping backend stats to minimalist display format focus on Moderation & Activity
  let displayStats = $derived([
    {
      name: "Pending Animes",
      value: data.stats?.pending_animes.toLocaleString() || "0",
      icon: "movie",
      color: "bg-blue-500/10 text-blue-400",
      link: "/admin/animes?status=false",
    },
    {
      name: "Pending Songs",
      value: data.stats?.pending_songs.toLocaleString() || "0",
      icon: "music_note",
      color: "bg-emerald-500/10 text-emerald-400",
      link: "/admin/songs?status=false",
    },
    {
      name: "Pending Variants",
      value: data.stats?.pending_variants.toLocaleString() || "0",
      icon: "layers",
      color: "bg-indigo-500/10 text-indigo-400",
      link: "/admin/variants?status=false",
    },
    {
      name: "Pending Videos",
      icon: "videocam",
      value: data.stats?.pending_videos.toLocaleString() || "0",
      color: "bg-amber-500/10 text-amber-400",
      link: "/admin/variants?status=false", // Videos are part of variants
    },
    {
      name: "Pending Artists",
      value: data.stats?.pending_artists.toLocaleString() || "0",
      icon: "person",
      color: "bg-purple-500/10 text-purple-400",
      link: "/admin/artists?status=false",
    },
    {
      name: "Song Reports",
      value: data.stats?.song_reports.toLocaleString() || "0",
      icon: "report",
      color: "bg-rose-500/10 text-rose-400",
      link: "/admin/reports/songs",
    },
    {
      name: "Comment Reports",
      value: data.stats?.comment_reports.toLocaleString() || "0",
      icon: "forum",
      color: "bg-pink-500/10 text-pink-400",
      link: "/admin/reports/comments",
    },
    {
      name: "Pending Requests",
      value: data.stats?.pending_requests.toLocaleString() || "0",
      icon: "help",
      color: "bg-cyan-500/10 text-cyan-400",
      link: "/admin/requests",
    },
    {
      name: "Active Tournaments",
      value: data.stats?.active_tournaments.toLocaleString() || "0",
      icon: "emoji_events",
      color: "bg-orange-500/10 text-orange-400",
      link: "/admin/tournaments",
    },
    {
      name: "Active Users (24h)",
      value: data.stats?.active_users_day.toLocaleString() || "0",
      icon: "bolt",
      color: "bg-sky-500/10 text-sky-400",
    },
  ]);

  let isUpdating = $state(false);

  async function handleSnapshot() {
    isUpdating = true;
    try {
      await api.post("/admin/ranking/snapshot");
      toastState.addToast("Ranking snapshot completed successfully", "success");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(
        `Failed to update rankings: ${err.message || err}`,
        "error",
      );
    } finally {
      isUpdating = false;
    }
  }

  // --- Chart Logic ---
  let chartData = $derived(data.metrics || []);
  let maxViews = $derived(
    Math.max(...chartData.map((m: any) => m.views_count), 1),
  );

  // Calculate SVG path for the line chart
  let chartPath = $derived(() => {
    if (chartData.length < 2) return "";
    const width = 800;
    const height = 150;
    const padding = 20;
    const innerWidth = width - padding * 2;
    const innerHeight = height - padding * 2;

    const points = chartData.map((m: any, i: number) => {
      const x = padding + i * (innerWidth / (chartData.length - 1));
      const y = height - padding - (m.views_count / maxViews) * innerHeight;
      return `${x},${y}`;
    });

    return `M ${points.join(" L ")}`;
  });

  // Calculate area path (for gradient fill)
  let areaPath = $derived(() => {
    const path = chartPath();
    if (!path) return "";
    const width = 800;
    const height = 150;
    return `${path} L ${width - 20},${height - 20} L 20,${height - 20} Z`;
  });
</script>

<svelte:head>
  <title>Dashboard | Anirank Admin</title>
</svelte:head>

<div>
  {#if data.error}
    <div
      class="mb-8 p-4 bg-rose-500/10 border border-rose-500/20 rounded-2xl text-rose-400 flex items-center gap-3"
    >
      <span class="material-symbols-outlined">error</span>
      <p class="text-sm font-medium">{data.error}</p>
    </div>
  {/if}

  <!-- Stats Grid: Focused on Pending & Moderation -->
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 mb-8">
    {#each displayStats as stat}
      <a
        href={stat.link || "#"}
        class="bg-anirank-card border border-white/5 rounded-xl p-4 relative overflow-hidden group transition-all hover:border-white/10 hover:bg-white/[0.03]"
      >
        <div
          class="absolute top-2 right-2 transition-transform duration-500 group-hover:scale-110"
        >
          <span
            class="material-symbols-outlined text-2xl opacity-10 {stat.color.split(
              ' ',
            )[1]}">{stat.icon}</span
          >
        </div>
        <p
          class="text-[10px] font-bold uppercase tracking-wider text-gray-500 mb-1 relative z-10"
        >
          {stat.name}
        </p>
        <div class="flex items-baseline gap-2 relative z-10">
          <h3 class="text-2xl font-black text-white tracking-tight">
            {stat.value}
          </h3>
        </div>
      </a>
    {/each}
  </div>

  <!-- Historical Traffic Chart -->
  <div class="bg-anirank-card border border-white/5 rounded-2xl p-6 mb-8">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h3 class="text-xl font-bold text-white">Historical Traffic</h3>
        <p class="text-xs text-gray-400">
          Daily views across the platform (Last 30 days)
        </p>
      </div>
      <div
        class="flex items-center gap-2 px-3 py-1 bg-white/5 rounded-full border border-white/5"
      >
        <div class="w-2 h-2 rounded-full bg-anirank-primary"></div>
        <span
          class="text-[10px] font-bold text-gray-400 uppercase tracking-wider"
          >Views</span
        >
      </div>
    </div>

    {#if chartData.length > 0}
      <div class="relative h-[150px] w-full">
        <svg
          viewBox="0 0 800 150"
          class="w-full h-full preserve-3d"
          preserveAspectRatio="none"
        >
          <defs>
            <linearGradient id="chartGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.2" />
              <stop offset="100%" stop-color="#3b82f6" stop-opacity="0" />
            </linearGradient>
          </defs>

          <!-- Grid Lines (Simple) -->
          <line
            x1="20"
            y1="20"
            x2="780"
            y2="20"
            stroke="white"
            stroke-opacity="0.05"
            stroke-dasharray="4"
          />
          <line
            x1="20"
            y1="85"
            x2="780"
            y2="85"
            stroke="white"
            stroke-opacity="0.05"
            stroke-dasharray="4"
          />
          <line
            x1="20"
            y1="130"
            x2="780"
            y2="130"
            stroke="white"
            stroke-opacity="0.1"
          />

          <!-- Area -->
          <path d={areaPath()} fill="url(#chartGradient)" />

          <!-- Line -->
          <path
            d={chartPath()}
            fill="none"
            stroke="#3b82f6"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="drop-shadow-[0_0_8px_rgba(59,130,246,0.4)]"
          />
        </svg>
      </div>
      <div class="flex justify-between mt-4 px-5">
        <span class="text-[10px] text-gray-500 font-bold uppercase"
          >{new Date(chartData[0].date).toLocaleDateString()}</span
        >
        <span class="text-[10px] text-gray-500 font-bold uppercase"
          >{new Date(
            chartData[chartData.length - 1].date,
          ).toLocaleDateString()}</span
        >
      </div>
    {:else}
      <div
        class="h-[150px] flex items-center justify-center border border-dashed border-white/10 rounded-xl"
      >
        <p class="text-sm text-gray-500">No traffic data available yet</p>
      </div>
    {/if}
  </div>

  <!-- Recent Activity / Quick Actions -->
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <div
      class="lg:col-span-2 bg-anirank-card border border-white/5 rounded-2xl p-6"
    >
      <h3 class="text-xl font-bold text-white mb-6">
        Recent Reports & Requests
      </h3>

      <div class="space-y-4">
        {#if (data.stats?.recent_reports?.length || 0) === 0 && (data.stats?.recent_requests?.length || 0) === 0}
          <div
            class="p-8 text-center bg-white/5 border border-dashed border-white/10 rounded-xl"
          >
            <p class="text-sm text-gray-500">No pending items to review</p>
          </div>
        {/if}

        {#each data.stats?.recent_reports || [] as report}
          <div
            class="p-4 rounded-xl bg-white/5 border border-white/5 flex items-start justify-between"
          >
            <div class="flex gap-4">
              <div
                class="w-10 h-10 rounded-full bg-rose-500/20 text-rose-400 flex items-center justify-center shrink-0"
              >
                <span class="material-symbols-outlined text-xl">report</span>
              </div>
              <div>
                <p class="text-sm font-medium text-white">
                  {report.title}
                </p>
                <p class="text-xs text-gray-400 mt-1">
                  Reported by @{report.user?.name || "anonymous"} • {new Date(
                    report.created_at,
                  ).toLocaleDateString()}
                </p>
              </div>
            </div>
            <a
              href="/admin/reports/{report.id}"
              class="text-xs font-semibold px-3 py-1 bg-white/10 hover:bg-white/20 rounded-lg text-white transition-colors"
              >Review</a
            >
          </div>
        {/each}

        {#each data.stats?.recent_requests || [] as request}
          <div
            class="p-4 rounded-xl bg-white/5 border border-white/5 flex items-start justify-between"
          >
            <div class="flex gap-4">
              <div
                class="w-10 h-10 rounded-full bg-blue-500/20 text-blue-400 flex items-center justify-center shrink-0"
              >
                <span class="material-symbols-outlined text-xl">help</span>
              </div>
              <div>
                <p class="text-sm font-medium text-white">
                  {request.title}
                </p>
                <p class="text-xs text-gray-400 mt-1">
                  Request by @{request.user?.name || "anonymous"} • {new Date(
                    request.created_at,
                  ).toLocaleDateString()}
                </p>
              </div>
            </div>
            <a
              href="/admin/requests/{request.id}"
              class="text-xs font-semibold px-3 py-1 bg-white/10 hover:bg-white/20 rounded-lg text-white transition-colors"
              >Review</a
            >
          </div>
        {/each}
      </div>

      <div class="mt-6 text-center">
        <a
          href="/admin/reports"
          class="text-sm font-medium text-anirank-primary hover:text-blue-400 transition-colors"
          >View all pending items &rarr;</a
        >
      </div>
    </div>

    <div class="bg-anirank-card border border-white/5 rounded-2xl p-6">
      <h3 class="text-xl font-bold text-white mb-6">Quick Actions</h3>
      <div class="space-y-3">
        <a
          href="/admin/songs/create"
          class="flex items-center gap-3 p-4 rounded-xl bg-white/5 hover:bg-white/10 transition-colors border border-transparent hover:border-white/10"
        >
          <div
            class="w-8 h-8 rounded-lg bg-emerald-500/20 text-emerald-400 flex items-center justify-center"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              /></svg
            >
          </div>
          <span class="font-medium text-sm text-gray-200">Add New Song</span>
        </a>
        <button
          class="w-full flex items-center gap-3 p-4 rounded-xl bg-white/5 hover:bg-white/10 transition-colors border border-transparent hover:border-white/10 text-left"
        >
          <div
            class="w-8 h-8 rounded-lg bg-amber-500/20 text-amber-400 flex items-center justify-center"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              /></svg
            >
          </div>
          <span class="font-medium text-sm text-gray-200"
            >Sync with Anilist</span
          >
        </button>

        <button
          onclick={handleSnapshot}
          disabled={isUpdating}
          class="w-full flex items-center gap-3 p-4 rounded-xl bg-white/5 hover:bg-white/10 transition-colors border border-transparent hover:border-white/10 text-left disabled:opacity-50 disabled:cursor-wait"
        >
          <div
            class="w-8 h-8 rounded-lg bg-indigo-500/20 text-indigo-400 flex items-center justify-center"
          >
            {#if isUpdating}
              <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
            {:else}
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
                /></svg
              >
            {/if}
          </div>
          <span class="font-medium text-sm text-gray-200"
            >{isUpdating ? "Updating..." : "Update Ranking Snapshot"}</span
          >
        </button>
      </div>
    </div>
  </div>
</div>
