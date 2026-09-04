<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import { authState, getAuthToken } from "$lib/state/auth.svelte";
  import { PUBLIC_API_URL } from "$lib/api";
  import api from "$lib/api";
  import SEO from "$lib/components/SEO.svelte";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";
  import Users from "lucide-svelte/icons/users";
  import Play from "lucide-svelte/icons/play";
  import SkipForward from "lucide-svelte/icons/skip-forward";
  import Search from "lucide-svelte/icons/search";
  import X from "lucide-svelte/icons/x";
  import Star from "lucide-svelte/icons/star";
  import MessageSquare from "lucide-svelte/icons/message-square";
  import ChevronDown from "lucide-svelte/icons/chevron-down";
  import ChevronUp from "lucide-svelte/icons/chevron-up";
  import ChevronLeft from "lucide-svelte/icons/chevron-left";
  import ChevronRight from "lucide-svelte/icons/chevron-right";
  import ListMusic from "lucide-svelte/icons/list-music";
  import {
    applyLobbyStateUpdate,
    applyRatingUpdate,
    canAddToQueue,
    fromCanonicalScore,
    isSeasonalPool,
    toCanonicalScore,
    type RatePlayer,
    type RateRoomState,
    type SourceMode,
    type SeasonalPoolThemeType,
  } from "$lib/rate/room-state";
  import { getFormattedScore, getSongName } from "$lib/song-utils";
  import { configState } from "$lib/state/config.svelte";

  const roomId = page.params.roomId;

  let ws = $state<WebSocket | null>(null);
  let roomState = $state<RateRoomState | null>(null);
  let status = $derived(roomState?.status || "lobby");
  let config = $derived(roomState?.config);
  let players = $derived(roomState?.players || []);
  let spectators = $derived(roomState?.spectators || []);
  let queue = $derived(roomState?.queue || []);
  let ratingData = $derived(roomState?.rating_data);
  let mySessionId = $derived(roomState?.my_session_id || "");
  let me = $derived(players.find((p) => p.session_id === mySessionId) as RatePlayer | undefined);
  let isHost = $derived(!!me?.is_host);
  let playersVersion = $state(0);

  let deviceId = $state("");
  let guestNickname = $state("");
  let isSpectator = $state(false);
  let chatInput = $state("");
  let chatMessages = $state<Array<{ sender: string; text: string; type: string }>>([]);
  let errorBanner = $state("");
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  let mediaEl = $state<HTMLVideoElement | null>(null);
  let audioUrl = $derived(roomState?.audio_url || "");
  let currentSong = $derived(roomState?.current_song as any);

  // Anime search modal
  let showSearchModal = $state(false);
  let searchQuery = $state("");
  let animeResults = $state<any[]>([]);
  let searchLoading = $state(false);
  let searchTimer: ReturnType<typeof setTimeout> | null = null;
  let selectedAnime = $state<any | null>(null);
  let animeSongs = $state<any[]>([]);
  let songsLoading = $state(false);
  let themeFilter = $state<"ALL" | "OP" | "ED">("ALL");

  let filteredSongs = $derived.by(() => {
    if (themeFilter === "ALL") return animeSongs;
    return animeSongs.filter((s) => (s.type || "").toUpperCase() === themeFilter);
  });

  let queueCollapsed = $state(false);
  let chatCollapsed = $state(false);

  let queuePanelVisible = $derived(status !== "lobby" && status !== "finished");
  let centerColClass = $derived.by(() => {
    if (!queuePanelVisible) return "lg:col-span-9";
    if (queueCollapsed) return "lg:col-span-8";
    return "lg:col-span-6";
  });
  let queueColClass = $derived(queueCollapsed ? "lg:col-span-1" : "lg:col-span-3");

  let scoreFormat = $derived(authState.user?.score_format || "POINT_10_DECIMAL");
  let draftScore = $state(0);
  let submitting = $state(false);
  let canSubmitRating = $derived(draftScore > 0);

  let displayDraft = $derived(fromCanonicalScore(draftScore, scoreFormat));
  let sessionAvgDisplay = $derived(
    ratingData?.session_avg != null
      ? getFormattedScore(ratingData.session_avg, scoreFormat)
      : "—"
  );

  let ratedProgress = $derived.by(() => {
    const total = ratingData?.player_count || 0;
    const rated = ratingData?.rated_count || 0;
    if (total <= 0) return 0;
    return Math.min(100, Math.round((rated / total) * 100));
  });

  let statusLabel = $derived(
    status === "rating"
      ? "Rating"
      : status === "waiting"
        ? "Waiting"
        : status === "lobby"
          ? "Lobby"
          : status === "finished"
            ? "Finished"
            : status
  );

  let queuePermission = $derived(
    canAddToQueue(
      config || {
        queue_mode: "host_only",
        queue_limit_per_user: 3,
        reveal_mode: "blind",
        max_players: 16,
        auto_advance: "never",
        name: "",
        private: false,
      },
      me,
      queue
    )
  );

  let canOpenSearch = $derived(
    (status === "waiting" || status === "rating") &&
      !isSeasonalPool(config) &&
      (isHost || (config?.queue_mode !== "disabled" && queuePermission.ok))
  );

  let seasonalActive = $derived(isSeasonalPool(config));
  let poolLabel = $derived.by(() => {
    if (!seasonalActive || !config) return "";
    const type =
      config.pool_theme_type && config.pool_theme_type !== "all" ? ` · ${config.pool_theme_type}` : "";
    return `${config.pool_season || ""} ${config.pool_year || ""}${type}`.trim();
  });

  let editSourceMode = $state<SourceMode>("manual");
  let editPoolYear = $state("");
  let editPoolSeason = $state("");
  let editPoolThemeType = $state<SeasonalPoolThemeType>("all");
  let editPoolLimit = $state(30);
  let sortedYears = $derived(
    [...configState.years].sort((a, b) => Number(b.slug) - Number(a.slug) || b.name.localeCompare(a.name))
  );

  $effect(() => {
    if (status === "lobby" && config) {
      editSourceMode = config.source_mode || "manual";
      editPoolYear = config.pool_year || "";
      editPoolSeason = config.pool_season || "";
      editPoolThemeType = (config.pool_theme_type as SeasonalPoolThemeType) || "all";
      editPoolLimit = config.pool_limit || 30;
    }
  });

  let showSessionControls = $derived(status === "waiting" || status === "rating");

  onMount(() => {
    isSpectator = page.url.searchParams.get("spectator") === "true";
    if (typeof window !== "undefined") {
      deviceId = localStorage.getItem("rate_device_id") || crypto.randomUUID();
      localStorage.setItem("rate_device_id", deviceId);
      guestNickname = localStorage.getItem("rate_nickname") || "Guest";
    }
    connectWebSocket();
  });

  onDestroy(() => {
    closeWebSocket();
    if (reconnectTimer) clearTimeout(reconnectTimer);
    if (searchTimer) clearTimeout(searchTimer);
  });

  $effect(() => {
    if (audioUrl && mediaEl) {
      mediaEl.load();
      mediaEl.play().catch(() => {});
    }
  });

  function connectWebSocket() {
    closeWebSocket();
    let base = PUBLIC_API_URL;
    if (base.endsWith("/api")) base = base.slice(0, -4);
    else if (base.endsWith("/api/")) base = base.slice(0, -5);
    const wsProtocol = base.startsWith("https") ? "wss://" : "ws://";
    const host = base.replace(/^https?:\/\//, "");
    const wsUrl = `${wsProtocol}${host}/api/rate/ws/${roomId}?token=${encodeURIComponent(getAuthToken() || "")}&device_id=${encodeURIComponent(deviceId)}&nickname=${encodeURIComponent(guestNickname)}&spectator=${isSpectator}`;

    ws = new WebSocket(wsUrl);
    ws.onmessage = (event) => {
      try {
        handleMessage(JSON.parse(event.data));
      } catch (e) {
        console.warn("[RATE] bad message", e);
      }
    };
    ws.onclose = () => {
      reconnectTimer = setTimeout(connectWebSocket, 3000);
    };
    ws.onerror = () => {
      errorBanner = "Connection error";
    };
  }

  function closeWebSocket() {
    if (ws) {
      ws.onclose = null;
      ws.close();
      ws = null;
    }
  }

  function handleMessage(msg: { type: string; payload: any }) {
    switch (msg.type) {
      case "lobby_state_update":
        roomState = applyLobbyStateUpdate(roomState, msg.payload);
        playersVersion += 1;
        break;
      case "rating_update":
        roomState = applyRatingUpdate(roomState, msg.payload);
        playersVersion += 1;
        submitting = false;
        break;
      case "song_started":
        if (msg.payload?.audio_url && roomState) {
          roomState = {
            ...roomState,
            audio_url: msg.payload.audio_url,
            current_song: msg.payload.song,
            status: "rating",
          };
        }
        draftScore = 0;
        submitting = false;
        break;
      case "chat_message":
        chatMessages = [...chatMessages.slice(-80), msg.payload];
        break;
      case "error":
        errorBanner = typeof msg.payload === "string" ? msg.payload : "Error";
        setTimeout(() => (errorBanner = ""), 4000);
        break;
      case "room_closed":
        goto("/rate");
        break;
    }
  }

  function send(type: string, payload: Record<string, unknown> = {}) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type, payload }));
    }
  }

  function submitRating() {
    if (!authState.isAuthenticated) {
      errorBanner = "Login required to rate";
      return;
    }
    if (draftScore <= 0) {
      errorBanner = "Pick a score greater than 0";
      return;
    }
    submitting = true;
    send("submit_rating", { score: draftScore });
  }

  function openSearchModal() {
    showSearchModal = true;
    searchQuery = "";
    animeResults = [];
    selectedAnime = null;
    animeSongs = [];
    themeFilter = "ALL";
  }

  function closeSearchModal() {
    showSearchModal = false;
    selectedAnime = null;
    animeSongs = [];
    animeResults = [];
    searchQuery = "";
  }

  function runAnimeSearch() {
    if (searchTimer) clearTimeout(searchTimer);
    searchTimer = setTimeout(async () => {
      const q = searchQuery.trim();
      if (q.length < 3) {
        animeResults = [];
        return;
      }
      searchLoading = true;
      try {
        const res = await api.get("/search", { params: { q } });
        animeResults = res.data?.data?.animes || [];
      } catch {
        animeResults = [];
      } finally {
        searchLoading = false;
      }
    }, 300);
  }

  async function selectAnime(anime: any) {
    selectedAnime = anime;
    animeSongs = [];
    songsLoading = true;
    themeFilter = "ALL";
    try {
      const res = await api.get(`/animes/${anime.slug}`);
      animeSongs = res.data?.data?.songs || [];
    } catch {
      errorBanner = "Failed to load anime themes";
      animeSongs = [];
    } finally {
      songsLoading = false;
    }
  }

  function songLabel(song: any) {
    const type = (song.type || "").toUpperCase();
    const num = song.theme_num || "";
    const name = getSongName(song);
    return `${type}${num} · ${name}`;
  }

  function addSong(song: any) {
    const uuid = song.id || song.uuid;
    if (!uuid) return;
    if (!queuePermission.ok && !isHost) {
      errorBanner = queuePermission.reason || "Cannot add to queue";
      return;
    }
    if (config?.queue_mode === "disabled") {
      if (isHost) playNow(song);
      return;
    }
    if (!queuePermission.ok) {
      errorBanner = queuePermission.reason || "Cannot add to queue";
      return;
    }
    send("queue_add", { song_uuid: uuid });
    closeSearchModal();
  }

  function playNow(song: any) {
    const uuid = song.id || song.uuid;
    if (!uuid || !isHost) return;
    send("set_song", { song_uuid: uuid });
    closeSearchModal();
  }

  function sendChat() {
    if (!chatInput.trim()) return;
    send("send_chat_message", { text: chatInput.trim() });
    chatInput = "";
  }

  function saveLobbyConfig() {
    if (!isHost || status !== "lobby") return;
    if (editSourceMode === "seasonal_pool" && (!editPoolYear || !editPoolSeason)) {
      errorBanner = "Pick year and season for seasonal pool";
      return;
    }
    send("update_lobby_config", {
      ...(config || {}),
      source_mode: editSourceMode,
      queue_mode: editSourceMode === "seasonal_pool" ? "disabled" : config?.queue_mode || "host_only",
      pool_year: editSourceMode === "seasonal_pool" ? editPoolYear : "",
      pool_season: editSourceMode === "seasonal_pool" ? editPoolSeason : "",
      pool_theme_type: editSourceMode === "seasonal_pool" ? editPoolThemeType : "all",
      pool_limit: editSourceMode === "seasonal_pool" ? editPoolLimit : 0,
    });
  }

  function playerRated(p: RatePlayer) {
    return !!ratingData?.ratings?.[p.session_id]?.rated;
  }

  function playerScoreLabel(p: RatePlayer) {
    const entry = ratingData?.ratings?.[p.session_id];
    if (!entry?.rated) return "—";
    if (config?.reveal_mode === "live" && entry.score != null) {
      return getFormattedScore(entry.score, scoreFormat);
    }
    if (p.session_id === mySessionId && ratingData?.my_score != null) {
      return getFormattedScore(ratingData.my_score, scoreFormat);
    }
    return "✓";
  }
</script>

<SEO title={`Rate Party ${roomId}`} description="Live group rating session" />

<!-- Room sub-header (Stitch RoomSubHeader → DESIGN tokens) -->
<div class="w-full bg-surface-low py-2.5 px-4 lg:px-8" data-purpose="room-context-header">
  <div class="max-w-7xl mx-auto flex items-center justify-between gap-3 text-sm">
    <button
      onclick={() => goto("/rate")}
      class="inline-flex items-center gap-2 text-primary hover:text-primary-container font-medium transition-colors cursor-pointer group"
      aria-label="Back to lobbies"
    >
      <ArrowLeft size={16} class="group-hover:-translate-x-0.5 transition-transform" />
      <span>Back to lobbies</span>
    </button>
    <div class="flex items-center gap-2 text-on-surface-variant text-xs sm:text-sm">
      <span>
        Room <strong class="text-on-surface font-semibold font-mono">{roomId}</strong>
      </span>
      <span class="text-outline-variant" aria-hidden="true">·</span>
      <span class="text-primary font-medium">{statusLabel}</span>
    </div>
  </div>
</div>

<main class="flex-1 max-w-7xl w-full mx-auto p-4 lg:p-6 space-y-4">
  {#if errorBanner}
    <div class="p-3 bg-red-50 text-red-700 text-sm rounded-sm" role="alert">{errorBanner}</div>
  {/if}

  <div class="grid grid-cols-1 lg:grid-cols-12 gap-5 items-start">
    <!-- LEFT: Players -->
    <aside class="lg:col-span-3 space-y-4 order-2 lg:order-1 min-w-0" data-purpose="players-sidebar">
      <div class="bg-surface-container rounded-sm p-4">
        <div class="flex items-center gap-2 text-xs font-bold tracking-wider text-primary uppercase pb-3 mb-3 bg-surface-low -mx-4 -mt-4 px-4 pt-4 rounded-t-sm">
          <Users size={16} class="text-primary" aria-hidden="true" />
          <span>Players</span>
        </div>

        <ul class="space-y-2">
          {#each players as player (player.session_id + "-" + playersVersion)}
            <li class="flex items-center justify-between gap-2 p-2.5 rounded-sm bg-surface-highest">
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <span class="text-xs font-semibold text-on-surface truncate">{player.nickname}</span>
                  {#if player.is_host}
                    <span class="text-[10px] font-extrabold uppercase px-1.5 py-0.5 rounded-sm bg-primary text-white tracking-wide shrink-0">
                      Host
                    </span>
                  {/if}
                  {#if player.offline}
                    <span class="text-[10px] text-on-surface-variant shrink-0">offline</span>
                  {/if}
                </div>
                {#if !player.user_uuid}
                  <p class="text-[10px] text-on-surface-variant mt-0.5">Guest (cannot rate)</p>
                {/if}
              </div>
              {#if status === "rating"}
                <span class="text-xs text-on-surface-variant font-mono shrink-0">{playerScoreLabel(player)}</span>
              {:else}
                <span class="text-xs text-on-surface-variant font-mono shrink-0">—</span>
              {/if}
            </li>
          {/each}
        </ul>

        {#if spectators.length}
          <p class="text-xs text-on-surface-variant pt-3">Spectators: {spectators.length}</p>
        {/if}

        <div class="mt-6 pt-4 bg-surface-low -mx-4 px-4 pb-1 rounded-b-sm">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-on-surface-variant font-medium">Session AVG</span>
            <span class="text-lg font-black text-primary tabular-nums">{sessionAvgDisplay}</span>
          </div>
          <div
            class="w-full bg-surface-highest h-1.5 rounded-sm overflow-hidden mb-2"
            role="progressbar"
            aria-valuenow={ratedProgress}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label="Players rated progress"
          >
            <div class="bg-primary h-full rounded-sm transition-all" style="width: {ratedProgress}%"></div>
          </div>
          <p class="text-xs text-on-surface-variant pb-3">
            {ratingData?.rated_count ?? 0}/{ratingData?.player_count ?? 0} rated
          </p>
        </div>
      </div>

      <div
        class="bg-surface-container rounded-sm flex flex-col overflow-hidden min-w-0 {chatCollapsed
          ? ''
          : 'min-h-[240px] h-[300px] lg:h-[340px]'}"
        data-purpose="chat-box"
      >
        <button
          type="button"
          class="flex items-center justify-between gap-2 w-full shrink-0 cursor-pointer bg-surface-low px-4 py-3"
          onclick={() => (chatCollapsed = !chatCollapsed)}
          aria-expanded={!chatCollapsed}
          aria-controls="rate-chat-panel"
        >
          <div class="flex items-center gap-2 text-xs font-bold tracking-wider text-primary uppercase min-w-0">
            <MessageSquare size={16} class="shrink-0" aria-hidden="true" />
            <span>Chat</span>
          </div>
          {#if chatCollapsed}
            <ChevronDown size={16} class="text-on-surface-variant shrink-0" aria-hidden="true" />
          {:else}
            <ChevronUp size={16} class="text-on-surface-variant shrink-0" aria-hidden="true" />
          {/if}
        </button>

        {#if !chatCollapsed}
          <div id="rate-chat-panel" class="flex-1 overflow-y-auto overflow-x-hidden space-y-2 text-xs px-4 py-3 min-h-0 min-w-0">
            {#each chatMessages as m}
              <p
                class="text-[11px] leading-relaxed break-words {m.type === 'system'
                  ? 'text-on-surface-variant italic'
                  : 'text-on-surface'}"
              >
                <strong class="text-primary not-italic font-semibold">{m.sender}:</strong>
                {m.text}
              </p>
            {/each}
          </div>
          <form
            class="shrink-0 flex items-center gap-2 px-4 py-3 bg-surface-low min-w-0"
            onsubmit={(e) => {
              e.preventDefault();
              sendChat();
            }}
          >
            <input
              bind:value={chatInput}
              class="min-w-0 flex-1 px-3 py-2 text-xs rounded-sm bg-surface-highest border border-outline-variant text-on-surface placeholder:text-on-surface-variant focus:outline-hidden focus:border-primary transition-colors"
              placeholder="Message…"
              aria-label="Chat message"
            />
            <button
              class="shrink-0 px-3.5 py-2 rounded-sm bg-primary hover:bg-primary-container text-white font-semibold text-xs tracking-wide transition-colors cursor-pointer"
              type="submit"
            >
              Send
            </button>
          </form>
        {/if}
      </div>
    </aside>

    <!-- CENTER: Playback + rating -->
    <section class="{centerColClass} space-y-4 order-1 lg:order-2 transition-[grid-column] duration-200" data-purpose="playback-main-area">
      <div class="bg-surface-container rounded-sm p-5">
        {#if status === "lobby"}
          <div class="text-center space-y-4 py-8">
            <Star size={40} class="mx-auto text-primary" aria-hidden="true" />
            <h1 class="text-2xl font-black text-on-surface tracking-tight">{config?.name || "Rate Party"}</h1>
            <p class="text-sm text-on-surface-variant max-w-md mx-auto">
              {#if seasonalActive}
                Seasonal pool: <span class="font-semibold text-on-surface capitalize">{poolLabel}</span>
                · Manual adds locked until host switches to manual
              {:else}
                Queue: {config?.queue_mode}
                {#if config?.queue_mode === "everyone"} (max {config.queue_limit_per_user}/user){/if}
              {/if}
              · Reveal: {config?.reveal_mode}
            </p>

            {#if isHost}
              <div class="text-left max-w-md mx-auto space-y-3 bg-surface-highest rounded-sm p-4">
                <p class="text-xs font-bold uppercase tracking-wide text-on-surface-variant">Lobby settings</p>
                <label class="block space-y-1">
                  <span class="text-xs font-bold text-on-surface-variant">Theme source</span>
                  <select bind:value={editSourceMode} class="w-full h-10 bg-surface border border-outline-variant rounded-sm px-3 text-sm">
                    <option value="manual">Manual (search &amp; queue)</option>
                    <option value="seasonal_pool">Seasonal pool</option>
                  </select>
                </label>
                {#if editSourceMode === "seasonal_pool"}
                  <div class="grid grid-cols-2 gap-2">
                    <label class="block space-y-1">
                      <span class="text-xs font-bold text-on-surface-variant">Year</span>
                      <select bind:value={editPoolYear} class="w-full h-10 bg-surface border border-outline-variant rounded-sm px-2 text-sm">
                        <option value="">Year</option>
                        {#each sortedYears as y}
                          <option value={y.slug}>{y.name}</option>
                        {/each}
                      </select>
                    </label>
                    <label class="block space-y-1">
                      <span class="text-xs font-bold text-on-surface-variant">Season</span>
                      <select bind:value={editPoolSeason} class="w-full h-10 bg-surface border border-outline-variant rounded-sm px-2 text-sm">
                        <option value="">Season</option>
                        {#each configState.seasons as s}
                          <option value={s.slug}>{s.name}</option>
                        {/each}
                      </select>
                    </label>
                  </div>
                  <label class="block space-y-1">
                    <span class="text-xs font-bold text-on-surface-variant">Type</span>
                    <select bind:value={editPoolThemeType} class="w-full h-10 bg-surface border border-outline-variant rounded-sm px-2 text-sm">
                      <option value="all">ALL</option>
                      <option value="OP">OP</option>
                      <option value="ED">ED</option>
                    </select>
                  </label>
                  <label class="block space-y-1">
                    <span class="text-xs font-bold text-on-surface-variant">Pool size</span>
                    <input type="number" min="5" max="50" bind:value={editPoolLimit} class="w-full h-10 bg-surface border border-outline-variant rounded-sm px-2 text-sm" />
                  </label>
                {/if}
                <button
                  type="button"
                  onclick={saveLobbyConfig}
                  class="w-full h-10 bg-surface-container text-primary font-bold text-sm rounded-sm cursor-pointer"
                >
                  Save settings
                </button>
              </div>

              <button
                onclick={() => send("start_session")}
                class="h-12 px-8 bg-primary hover:bg-primary-container text-white font-bold rounded-sm cursor-pointer transition-colors"
              >
                Start Session
              </button>
            {:else}
              <p class="text-sm text-on-surface-variant">Waiting for host to start…</p>
            {/if}
          </div>
        {:else if status === "finished"}
          <div class="text-center space-y-4 py-10">
            <h1 class="text-2xl font-black text-on-surface">Session ended</h1>
            <p class="text-sm text-on-surface-variant">Songs rated this session: {roomState?.songs_rated || 0}</p>
            {#if isHost}
              <button
                onclick={() => send("reset_to_lobby")}
                class="h-11 px-6 bg-primary hover:bg-primary-container text-white font-bold rounded-sm cursor-pointer transition-colors"
              >
                Back to lobby
              </button>
            {/if}
          </div>
        {:else}
          {#if currentSong}
            <div class="mb-3">
              <div class="text-[10px] font-bold tracking-widest uppercase text-primary mb-1">Now rating</div>
              <h1 class="text-2xl sm:text-3xl font-extrabold text-on-surface tracking-tight leading-tight">
                {getSongName(currentSong) || currentSong.name || "Theme"}
              </h1>
              {#if currentSong.anime}
                <p class="text-sm text-on-surface-variant mt-1">{currentSong.anime.title}</p>
              {/if}
            </div>

            <video
              bind:this={mediaEl}
              class="relative w-full aspect-video rounded-sm overflow-hidden bg-on-surface"
              controls
              playsinline
              src={audioUrl || undefined}
            >
              <track kind="captions" />
            </video>

            {#if status === "rating" && !me?.is_spectator}
              <div class="mt-4 p-4 rounded-sm bg-surface-highest" data-purpose="rating-form">
                {#if !authState.isAuthenticated}
                  <p class="text-sm text-on-surface-variant">
                    Log in to rate this song (saves to your global ranking).
                  </p>
                {:else if playerRated(me!)}
                  <div class="flex items-center justify-between">
                    <span class="text-xs font-semibold text-on-surface-variant">Your score</span>
                    <span class="text-2xl font-black text-primary tabular-nums">
                      {getFormattedScore(ratingData?.my_score ?? draftScore, scoreFormat)}
                    </span>
                  </div>
                {:else}
                  <div class="flex items-center justify-between mb-3">
                    <span class="text-xs font-semibold text-on-surface-variant">Your score</span>
                    <span class="text-2xl font-black text-primary tabular-nums tracking-tight">
                      {scoreFormat === "POINT_10_DECIMAL" || scoreFormat === "POINT_5"
                        ? displayDraft.toFixed(1)
                        : Math.round(displayDraft)}
                    </span>
                  </div>

                  {#if scoreFormat === "POINT_10" || scoreFormat === "POINT_10_DECIMAL"}
                    <div class="grid grid-cols-5 gap-2 mb-3">
                      {#each [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] as n}
                        <button
                          onclick={() => {
                            draftScore = toCanonicalScore(n, scoreFormat);
                          }}
                          class="h-10 rounded-sm text-sm font-bold bg-surface-container text-on-surface hover:bg-primary hover:text-white cursor-pointer transition-colors"
                          aria-label="Set score to {n}"
                        >
                          {n}
                        </button>
                      {/each}
                    </div>
                    {#if scoreFormat === "POINT_10_DECIMAL"}
                      <div class="py-2">
                        <input
                          type="range"
                          min="0"
                          max="100"
                          step="1"
                          bind:value={draftScore}
                          class="w-full h-2 rounded-sm cursor-pointer accent-primary"
                          aria-label="Score slider"
                        />
                      </div>
                    {/if}
                  {:else if scoreFormat === "POINT_100"}
                    <div class="py-2">
                      <input
                        type="range"
                        min="0"
                        max="100"
                        step="1"
                        bind:value={draftScore}
                        class="w-full h-2 rounded-sm cursor-pointer accent-primary"
                        aria-label="Score slider"
                      />
                    </div>
                  {:else}
                    <div class="grid grid-cols-5 gap-2 mb-3">
                      {#each [0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5] as n}
                        <button
                          onclick={() => (draftScore = toCanonicalScore(n, "POINT_5"))}
                          class="h-10 rounded-sm text-sm font-bold bg-surface-container text-on-surface hover:bg-primary hover:text-white cursor-pointer transition-colors"
                          aria-label="Set score to {n}"
                        >
                          {n}
                        </button>
                      {/each}
                    </div>
                  {/if}

                  <button
                    onclick={submitRating}
                    disabled={submitting || !canSubmitRating}
                    class="w-full mt-3 py-2.5 px-4 rounded-sm font-bold text-sm text-white bg-primary hover:bg-primary-container transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {canSubmitRating ? "Submit rating" : "Pick a score to submit"}
                  </button>
                {/if}
              </div>
            {/if}
          {:else}
            <div class="text-center py-16 space-y-3">
              <Play size={36} class="mx-auto text-primary" aria-hidden="true" />
              <p class="font-bold text-on-surface">Waiting for a song</p>
              <p class="text-sm text-on-surface-variant">
                {#if seasonalActive}
                  Seasonal pool ready — host presses Next to play the next theme.
                {:else if canOpenSearch}
                  Search an anime to pick a theme.
                {:else}
                  Host will pick the next theme.
                {/if}
              </p>
            </div>
          {/if}

          {#if showSessionControls}
            <div class="mt-4 flex items-center justify-between gap-3 flex-wrap text-xs">
              <div class="flex items-center gap-2 flex-wrap">
                {#if canOpenSearch}
                  <button
                    onclick={openSearchModal}
                    class="px-3.5 py-2 rounded-sm bg-primary hover:bg-primary-container text-white font-medium flex items-center gap-1.5 transition-colors cursor-pointer"
                    type="button"
                  >
                    <Search size={14} aria-hidden="true" />
                    <span>Search anime</span>
                  </button>
                {/if}
                {#if isHost}
                  <button
                    onclick={() => send("next")}
                    class="px-3 py-2 rounded-sm bg-surface-highest text-primary hover:bg-surface-low font-medium flex items-center gap-1.5 transition-colors cursor-pointer"
                    type="button"
                  >
                    <SkipForward size={14} aria-hidden="true" />
                    <span>{seasonalActive ? "Next from pool" : "Next"}</span>
                  </button>
                {/if}
              </div>
              {#if isHost}
                <button
                  onclick={() => send("end_session")}
                  class="px-3 py-2 rounded-sm bg-surface-highest text-on-surface-variant hover:text-on-surface font-medium transition-colors cursor-pointer"
                  type="button"
                >
                  End session
                </button>
              {/if}
            </div>
          {/if}
        {/if}
      </div>
    </section>

    <!-- RIGHT: Queue (desktop can collapse horizontally to widen video) -->
    {#if queuePanelVisible}
      <aside class="{queueColClass} space-y-4 order-3 transition-[grid-column] duration-200" data-purpose="queue-sidebar">
        {#if queueCollapsed}
          <!-- Desktop rail -->
          <div class="hidden lg:flex bg-surface-container rounded-sm p-2 flex-col items-center gap-3 min-h-[12rem]">
            <button
              type="button"
              class="w-full flex flex-col items-center gap-2 py-3 px-1 rounded-sm bg-surface-low hover:bg-surface-highest text-primary cursor-pointer transition-colors"
              onclick={() => (queueCollapsed = false)}
              aria-expanded="false"
              aria-controls="rate-queue-panel"
              title="Expand queue"
            >
              <ChevronLeft size={18} aria-hidden="true" />
              <ListMusic size={18} aria-hidden="true" />
              <span class="text-[10px] font-bold uppercase tracking-wider [writing-mode:vertical-rl] rotate-180">
                Queue ({queue.length})
              </span>
            </button>
          </div>
          <!-- Mobile collapsed header -->
          <div class="lg:hidden bg-surface-container rounded-sm p-4">
            <button
              type="button"
              class="flex items-center justify-between w-full cursor-pointer"
              onclick={() => (queueCollapsed = false)}
              aria-expanded="false"
              aria-controls="rate-queue-panel"
            >
              <div class="flex items-center gap-2 text-xs font-bold tracking-wider text-primary uppercase">
                <ListMusic size={16} aria-hidden="true" />
                <span>Queue ({queue.length})</span>
              </div>
              <ChevronDown size={16} class="text-on-surface-variant" aria-hidden="true" />
            </button>
          </div>
        {:else}
          <div class="bg-surface-container rounded-sm p-4 flex flex-col">
            <div class="flex items-center justify-between pb-2.5 mb-2.5 bg-surface-low -mx-4 -mt-4 px-4 pt-4 rounded-t-sm">
              <div class="flex items-center gap-2 text-xs font-bold tracking-wider text-primary uppercase">
                <ListMusic size={16} aria-hidden="true" />
                <span>Queue ({queue.length})</span>
              </div>
              <button
                type="button"
                class="text-on-surface-variant hover:text-primary p-1 cursor-pointer rounded-sm transition-colors"
                onclick={() => (queueCollapsed = true)}
                aria-expanded="true"
                aria-controls="rate-queue-panel"
                title="Collapse queue"
              >
                <span class="lg:hidden"><ChevronUp size={16} aria-hidden="true" /></span>
                <span class="hidden lg:inline-flex"><ChevronRight size={16} aria-hidden="true" /></span>
              </button>
            </div>

            <div id="rate-queue-panel" class="space-y-2 max-h-56 lg:max-h-[28rem] overflow-y-auto pr-1">
              {#if seasonalActive}
                <p class="text-xs text-on-surface-variant">
                  Seasonal pool ({poolLabel}) — manual adds locked. Host can switch to manual in lobby.
                </p>
              {:else if config?.queue_mode === "disabled"}
                <p class="text-xs text-on-surface-variant">Queue disabled — host plays songs directly.</p>
              {/if}

              {#if seasonalActive || config?.queue_mode !== "disabled"}
                {#each queue as item (item.item_id)}
                  <div class="flex items-start justify-between p-2 rounded-sm bg-surface-highest">
                    <div class="min-w-0 pr-2">
                      <h4 class="text-xs font-bold text-on-surface truncate">{item.song_name}</h4>
                      <p class="text-[10px] text-on-surface-variant truncate mt-0.5">
                        {item.anime_title} · {item.added_by_nickname}
                      </p>
                    </div>
                    {#if isHost || (!seasonalActive && item.added_by_session_id === mySessionId)}
                      <button
                        onclick={() => send("queue_remove", { item_id: item.item_id })}
                        class="text-on-surface-variant hover:text-red-600 p-0.5 cursor-pointer shrink-0"
                        aria-label="Remove {item.song_name} from queue"
                        type="button"
                      >
                        <X size={14} />
                      </button>
                    {/if}
                  </div>
                {:else}
                  <p class="text-xs text-on-surface-variant">
                    {seasonalActive ? "Pool loading or empty" : "Queue is empty"}
                  </p>
                {/each}
              {/if}
            </div>
          </div>
        {/if}
      </aside>
    {/if}
  </div>
</main>

{#if showSearchModal}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
    onclick={closeSearchModal}
    role="presentation"
  >
    <div
      class="bg-surface-highest w-full max-w-2xl max-h-[85vh] rounded-sm flex flex-col overflow-hidden"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      aria-label="Search anime"
    >
      <div class="flex items-center justify-between gap-3 p-4 bg-surface-container">
        <h2 class="text-lg font-black text-on-surface">Search anime</h2>
        <button
          onclick={closeSearchModal}
          class="text-on-surface-variant hover:text-on-surface cursor-pointer min-h-11 min-w-11 flex items-center justify-center"
          aria-label="Close search"
          type="button"
        >
          <X size={20} />
        </button>
      </div>

      <div class="p-4 space-y-3 bg-surface-low">
        <div class="relative">
          <Search
            size={16}
            class="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant"
            aria-hidden="true"
          />
          <input
            bind:value={searchQuery}
            oninput={runAnimeSearch}
            placeholder="Type an anime title…"
            class="w-full h-12 pl-10 pr-3 bg-surface border border-outline-variant rounded-sm text-sm text-on-surface focus:outline-hidden focus:border-primary"
            autofocus
            aria-label="Anime search"
          />
        </div>
        {#if selectedAnime}
          <div class="flex items-center justify-between gap-3 flex-wrap">
            <button
              type="button"
              class="text-xs font-bold text-primary cursor-pointer min-h-11"
              onclick={() => {
                selectedAnime = null;
                animeSongs = [];
              }}
            >
              ← Back to results
            </button>
            <label class="flex items-center gap-2 text-xs text-on-surface-variant">
              <span class="font-bold uppercase tracking-wide">Type</span>
              <select
                bind:value={themeFilter}
                class="h-9 bg-surface border border-outline-variant rounded-sm px-2 text-sm text-on-surface"
                aria-label="Filter theme type"
              >
                <option value="ALL">ALL</option>
                <option value="OP">OP</option>
                <option value="ED">ED</option>
              </select>
            </label>
          </div>
        {/if}
      </div>

      <div class="flex-1 overflow-y-auto p-4 space-y-2 min-h-[240px]">
        {#if !selectedAnime}
          {#if searchLoading}
            <p class="text-sm text-on-surface-variant">Searching…</p>
          {:else if searchQuery.trim().length < 3}
            <p class="text-sm text-on-surface-variant">Type at least 3 characters to find animes.</p>
          {:else if animeResults.length === 0}
            <p class="text-sm text-on-surface-variant">No animes found.</p>
          {:else}
            {#each animeResults as anime}
              <button
                type="button"
                class="w-full text-left bg-surface-container hover:bg-surface-low rounded-sm px-4 py-3 transition-colors cursor-pointer"
                onclick={() => selectAnime(anime)}
              >
                <p class="font-bold text-on-surface">{anime.title}</p>
                {#if anime.slug}
                  <p class="text-xs text-on-surface-variant">{anime.slug}</p>
                {/if}
              </button>
            {/each}
          {/if}
        {:else}
          <p class="text-sm font-bold text-on-surface mb-2">{selectedAnime.title}</p>
          {#if songsLoading}
            <p class="text-sm text-on-surface-variant">Loading themes…</p>
          {:else if filteredSongs.length === 0}
            <p class="text-sm text-on-surface-variant">No themes for this filter.</p>
          {:else}
            {#each filteredSongs as song}
              <div class="flex items-center justify-between gap-3 bg-surface-container rounded-sm px-3 py-2.5">
                <div class="min-w-0">
                  <p class="text-sm font-bold text-on-surface truncate">{songLabel(song)}</p>
                </div>
                <div class="flex gap-2 shrink-0">
                  {#if isHost}
                    <button
                      onclick={() => playNow(song)}
                      class="h-9 px-3 text-xs font-bold text-white bg-primary hover:bg-primary-container rounded-sm cursor-pointer transition-colors"
                      type="button"
                    >
                      Play
                    </button>
                  {/if}
                  {#if config?.queue_mode !== "disabled" && queuePermission.ok}
                    <button
                      onclick={() => addSong(song)}
                      class="h-9 px-3 text-xs font-bold text-primary bg-surface-highest hover:bg-surface-low rounded-sm cursor-pointer transition-colors"
                      type="button"
                    >
                      + Queue
                    </button>
                  {/if}
                </div>
              </div>
            {/each}
          {/if}
        {/if}
      </div>
    </div>
  </div>
{/if}
