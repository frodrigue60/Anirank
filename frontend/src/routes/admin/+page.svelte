<script lang="ts">
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { getApiErrorMessage } from "$lib/api-errors";
  import Film from "lucide-svelte/icons/film";
  import Music from "lucide-svelte/icons/music";
  import Layers from "lucide-svelte/icons/layers";
  import Video from "lucide-svelte/icons/video";
  import User from "lucide-svelte/icons/user";
  import AlertTriangle from "lucide-svelte/icons/alert-triangle";
  import MessageSquare from "lucide-svelte/icons/message-square";
  import HelpCircle from "lucide-svelte/icons/help-circle";
  import Trophy from "lucide-svelte/icons/trophy";
  import Zap from "lucide-svelte/icons/zap";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import Eraser from "lucide-svelte/icons/eraser";
  import Plus from "lucide-svelte/icons/plus";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import TrendingUp from "lucide-svelte/icons/trending-up";
  import Loader2 from "lucide-svelte/icons/loader-2";

  let { data } = $props();

  const iconMap: Record<string, any> = {
    movie: Film,
    music_note: Music,
    layers: Layers,
    videocam: Video,
    person: User,
    report: AlertTriangle,
    forum: MessageSquare,
    help: HelpCircle,
    emoji_events: Trophy,
    bolt: Zap,
  };

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
      link: "/admin/videos?status=false",
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
  let isFlushing = $state(false);

  async function handleSnapshot() {
    isUpdating = true;
    try {
      await api.post("/admin/ranking/snapshot");
      toastState.addToast("Ranking snapshot completed successfully", "success");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(
        getApiErrorMessage(err, "Failed to update rankings"),
        "error",
      );
    } finally {
      isUpdating = false;
    }
  }

  async function handleFlushOG() {
    isFlushing = true;
    try {
      await api.post("/admin/og/flush");
      toastState.addToast("OG Image Cache flushed successfully", "success");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(
        getApiErrorMessage(err, "Failed to flush OG cache"),
        "error",
      );
    } finally {
      isFlushing = false;
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

  // --- Hover Logic ---
  let hoverIndex = $state(-1);
  let svgElement: SVGSVGElement | null = $state(null);

  function handleMouseMove(e: MouseEvent) {
    if (!svgElement || chartData.length < 2) return;
    const rect = svgElement.getBoundingClientRect();
    const x = ((e.clientX - rect.left) / rect.width) * 800;

    const padding = 20;
    const innerWidth = 800 - padding * 2;
    const index = Math.round(
      ((x - padding) / innerWidth) * (chartData.length - 1),
    );

    if (index >= 0 && index < chartData.length) {
      hoverIndex = index;
    } else {
      hoverIndex = -1;
    }
  }

  function handleMouseLeave() {
    hoverIndex = -1;
  }

  let hoverPoint = $derived(() => {
    if (hoverIndex === -1 || chartData.length === 0) return null;
    const m = chartData[hoverIndex];
    const width = 800;
    const height = 150;
    const padding = 20;
    const innerWidth = width - padding * 2;
    const innerHeight = height - padding * 2;

    const x = padding + hoverIndex * (innerWidth / (chartData.length - 1));
    const y = height - padding - (m.views_count / maxViews) * innerHeight;

    return {
      x,
      y,
      date: new Date(m.date).toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
      }),
      value: m.views_count.toLocaleString(),
    };
  });
</script>

<svelte:head>
  <title>Dashboard | Anirank Admin</title>
</svelte:head>

<div class="">
  {#if data.error}
    <div
      class="mb-8 p-4 bg-rose-500/10 border border-rose-500/20 rounded-2xl text-rose-400 flex items-center gap-3"
    >
      <AlertCircle size={20} />
      <p class="text-sm font-medium">{data.error}</p>
    </div>
  {/if}

  <!-- Stats Grid: Focused on Pending & Moderation -->
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 mb-8">
    {#each displayStats as stat}
      <a
        href={stat.link || "#"}
        class="bg-surface-container border border-outline-variant rounded-xl p-4 relative overflow-hidden group transition-all hover:border-outline-variant hover:bg-white/3"
      >
        <div
          class="absolute top-2 right-2 transition-transform duration-500 group-hover:scale-110"
        >
          <svelte:component
            this={iconMap[stat.icon]}
            size={24}
            class={stat.color.split(" ")[1]}
          />

        </div>
        <p
          class="text-[10px] font-bold uppercase tracking-wider text-on-surface mb-1 relative z-10"
        >
          {stat.name}
        </p>
        <div class="flex items-baseline gap-2 relative z-10">
          <h3 class="text-2xl font-black text-on-surface tracking-tight">
            {stat.value}
          </h3>
        </div>
      </a>
    {/each}
  </div>

  <!-- Historical Traffic Chart -->
  <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 mb-8">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h3 class="text-xl font-bold text-on-surface">Historical Traffic</h3>
        <p class="text-xs text-on-surface-variant/70">
          Daily views across the platform (Last 30 days)
        </p>
      </div>
      <div
        class="flex items-center gap-2 px-3 py-1 bg-surface-highest rounded-full border border-outline-variant"
      >
        <div class="w-2 h-2 rounded-full bg-primary"></div>
        <span
          class="text-[10px] font-bold text-on-surface-variant/70 uppercase tracking-wider"
          >Views</span
        >
      </div>
    </div>

    {#if chartData.length > 0}
      <div class="relative h-[150px] w-full group/chart">
        <svg
          bind:this={svgElement}
          onmousemove={handleMouseMove}
          onmouseleave={handleMouseLeave}
          viewBox="0 0 800 150"
          class="w-full h-full preserve-3d cursor-crosshair"
          preserveAspectRatio="none"
          role="img"
          aria-label="Historical traffic chart"
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

          {#if hoverPoint()}
            <!-- Vertical Indicator -->
            <line
              x1={hoverPoint()!.x}
              y1="20"
              x2={hoverPoint()!.x}
              y2="130"
              stroke="white"
              stroke-opacity="0.2"
              stroke-dasharray="4"
            />

            <!-- Hover Dot -->
            <circle
              cx={hoverPoint()!.x}
              cy={hoverPoint()!.y}
              r="6"
              fill="#3b82f6"
              stroke="white"
              stroke-width="2"
              class="drop-shadow-[0_0_8px_rgba(59,130,246,0.8)]"
            />
          {/if}
        </svg>

        <!-- Tooltip -->
        {#if hoverPoint()}
          <div
            class="absolute pointer-events-none bg-surface-container border border-gray-500 rounded-lg p-2 shadow-2xl z-50 transition-all duration-75"
            style="left: {hoverPoint()!.x > 400
              ? 'auto'
              : (hoverPoint()!.x / 800) * 100 + '%'}; 
                   right: {hoverPoint()!.x > 400
              ? 100 - (hoverPoint()!.x / 800) * 100 + '%'
              : 'auto'}; 
                   top: {hoverPoint()!.y - 60}px;
                   transform: translateX({hoverPoint()!.x > 400
              ? '10px'
              : '-10px'});"
          >
            <p
              class="text-[10px] font-bold text-on-surface-variant/70 uppercase leading-none mb-1"
            >
              {hoverPoint()!.date}
            </p>
            <p class="text-xs font-black text-on-surface leading-none">
              {hoverPoint()!.value}
              <span class="text-[10px] font-normal text-on-surface-variant/40 uppercase"
                >Views</span
              >
            </p>
          </div>
        {/if}
      </div>
      <div class="flex justify-between mt-4 px-5">
        <span class="text-[10px] text-on-surface-variant/40 font-bold uppercase"
          >{new Date(chartData[0].date).toLocaleDateString()}</span
        >
        <span class="text-[10px] text-on-surface-variant/40 font-bold uppercase"
          >{new Date(
            chartData[chartData.length - 1].date,
          ).toLocaleDateString()}</span
        >
      </div>
    {:else}
      <div
        class="h-[150px] flex items-center justify-center border border-dashed border-outline-variant rounded-xl"
      >
        <p class="text-sm text-on-surface-variant/40">No traffic data available yet</p>
      </div>
    {/if}
  </div>

  <!-- Recent Activity / Quick Actions -->
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <div
      class="lg:col-span-2 bg-surface-container border border-outline-variant rounded-2xl p-6"
    >
      <h3 class="text-xl font-bold text-on-surface mb-6">
        Recent Reports & Requests
      </h3>

      <div class="space-y-4">
        {#if (data.stats?.recent_reports?.length || 0) === 0 && (data.stats?.recent_requests?.length || 0) === 0}
          <div
            class="p-8 text-center bg-surface-highest border border-dashed border-outline-variant rounded-xl"
          >
            <p class="text-sm text-on-surface-variant/40">No pending items to review</p>
          </div>
        {/if}

        {#each data.stats?.recent_reports || [] as report}
          <div
            class="p-4 rounded-xl bg-surface-highest border border-outline-variant flex items-start justify-between"
          >
            <div class="flex gap-4">
              <div
                class="w-10 h-10 rounded-full bg-rose-500/20 text-rose-400 flex items-center justify-center shrink-0"
              >
                <AlertTriangle size={20} />
              </div>
              <div>
                <p class="text-sm font-medium text-on-surface">
                  {report.title}
                </p>
                <p class="text-xs text-on-surface-variant/70 mt-1">
                  Reported by @{report.user?.name || "anonymous"} • {new Date(
                    report.created_at,
                  ).toLocaleDateString()}
                </p>
              </div>
            </div>
            <a
              href="/admin/reports/{report.id}"
              class="text-xs font-semibold px-3 py-1 bg-surface-highest hover:bg-white/20 rounded-lg text-on-surface transition-colors"
              >Review</a
            >
          </div>
        {/each}

        {#each data.stats?.recent_requests || [] as request}
          <div
            class="p-4 rounded-xl bg-surface-highest border border-outline-variant flex items-start justify-between"
          >
            <div class="flex gap-4">
              <div
                class="w-10 h-10 rounded-full bg-blue-500/20 text-blue-400 flex items-center justify-center shrink-0"
              >
                <HelpCircle size={20} />
              </div>
              <div>
                <p class="text-sm font-medium text-on-surface">
                  {request.title}
                </p>
                <p class="text-xs text-on-surface-variant/70 mt-1">
                  Request by @{request.user?.name || "anonymous"} • {new Date(
                    request.created_at,
                  ).toLocaleDateString()}
                </p>
              </div>
            </div>
            <a
              href="/admin/requests/{request.id}"
              class="text-xs font-semibold px-3 py-1 bg-surface-highest hover:bg-white/20 rounded-lg text-on-surface transition-colors"
              >Review</a
            >
          </div>
        {/each}
      </div>

      <div class="mt-6 text-center">
        <a
          href="/admin/reports"
          class="text-sm font-medium text-primary hover:text-blue-400 transition-colors"
          >View all pending items &rarr;</a
        >
      </div>
    </div>

    <div class="bg-surface-container border border-outline-variant rounded-2xl p-6">
      <h3 class="text-xl font-bold text-on-surface mb-6">Quick Actions</h3>
      <div class="space-y-3">
        <a
          href="/admin/songs/create"
          class="flex items-center gap-3 p-4 rounded-xl bg-surface-highest hover:bg-surface-highest transition-colors border border-transparent hover:border-outline-variant"
        >
          <div
            class="w-8 h-8 rounded-lg bg-emerald-500/20 text-emerald-400 flex items-center justify-center"
          >
            <Plus size={16} />
          </div>
          <span class="font-medium text-sm text-on-surface">Add New Song</span>
        </a>
        <button
          class="w-full flex items-center gap-3 p-4 rounded-xl bg-surface-highest hover:bg-surface-highest transition-colors border border-transparent hover:border-outline-variant text-left"
        >
          <div
            class="w-8 h-8 rounded-lg bg-amber-500/20 text-amber-400 flex items-center justify-center"
          >
            <RefreshCw size={16} />
          </div>
          <span class="font-medium text-sm text-on-surface"
            >Sync with Anilist</span
          >
        </button>

        <button
          onclick={handleSnapshot}
          disabled={isUpdating}
          class="w-full flex items-center gap-3 p-4 rounded-xl bg-surface-highest hover:bg-surface-highest transition-colors border border-transparent hover:border-outline-variant text-left disabled:opacity-50 disabled:cursor-wait"
        >
          <div
            class="w-8 h-8 rounded-lg bg-indigo-500/20 text-indigo-400 flex items-center justify-center"
          >
            {#if isUpdating}
              <Loader2 size={16} class="animate-spin" />
            {:else}
              <TrendingUp size={16} />
            {/if}
          </div>
          <span class="font-medium text-sm text-on-surface"
            >{isUpdating ? "Updating..." : "Update Ranking Snapshot"}</span
          >
        </button>

        <button
          onclick={handleFlushOG}
          disabled={isFlushing}
          class="w-full flex items-center gap-3 p-4 rounded-xl bg-surface-highest hover:bg-surface-highest transition-colors border border-transparent hover:border-outline-variant text-left disabled:opacity-50 disabled:cursor-wait"
        >
          <div
            class="w-8 h-8 rounded-lg bg-rose-500/20 text-rose-400 flex items-center justify-center"
          >
            {#if isFlushing}
              <Loader2 size={16} class="animate-spin" />
            {:else}
              <Eraser size={20} />

            {/if}
          </div>
          <span class="font-medium text-sm text-on-surface"
            >{isFlushing ? "Flushing..." : "Flush OG Image Cache"}</span
          >
        </button>
      </div>
    </div>
  </div>
</div>
