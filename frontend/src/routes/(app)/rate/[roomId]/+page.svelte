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
  import {
    applyLobbyStateUpdate,
    applyRatingUpdate,
    canAddToQueue,
    fromCanonicalScore,
    toCanonicalScore,
    type RatePlayer,
    type RateRoomState,
  } from "$lib/rate/room-state";
  import { getFormattedScore, getSongName } from "$lib/song-utils";

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

  let searchQuery = $state("");
  let searchResults = $state<any[]>([]);
  let searchLoading = $state(false);
  let searchTimer: ReturnType<typeof setTimeout> | null = null;

  let scoreFormat = $derived(authState.user?.score_format || "POINT_10_DECIMAL");
  let draftScore = $state(70); // canonical 0-100
  let submitting = $state(false);

  let displayDraft = $derived(fromCanonicalScore(draftScore, scoreFormat));
  let sessionAvgDisplay = $derived(
    ratingData?.session_avg != null
      ? getFormattedScore(ratingData.session_avg, scoreFormat)
      : "—"
  );

  let queuePermission = $derived(canAddToQueue(config || { queue_mode: "host_only", queue_limit_per_user: 3, reveal_mode: "blind", max_players: 16, auto_advance: "never", name: "", private: false }, me, queue));

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
        const msg = JSON.parse(event.data);
        handleMessage(msg);
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
        draftScore = 70;
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
    submitting = true;
    send("submit_rating", { score: draftScore });
  }

  function runSearch() {
    if (searchTimer) clearTimeout(searchTimer);
    searchTimer = setTimeout(async () => {
      const q = searchQuery.trim();
      if (q.length < 3) {
        searchResults = [];
        return;
      }
      searchLoading = true;
      try {
        const res = await api.get("/search", { params: { q } });
        searchResults = res.data?.data?.songs || [];
      } catch {
        searchResults = [];
      } finally {
        searchLoading = false;
      }
    }, 300);
  }

  function addSong(song: any) {
    const uuid = song.uuid || song.id;
    if (!uuid) return;
    if (!queuePermission.ok) {
      errorBanner = queuePermission.reason || "Cannot add to queue";
      return;
    }
    send("queue_add", { song_uuid: uuid });
    searchQuery = "";
    searchResults = [];
  }

  function playNow(song: any) {
    const uuid = song.uuid || song.id;
    if (!uuid || !isHost) return;
    send("set_song", { song_uuid: uuid });
    searchQuery = "";
    searchResults = [];
  }

  function sendChat() {
    if (!chatInput.trim()) return;
    send("send_chat_message", { text: chatInput.trim() });
    chatInput = "";
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

<main class="max-w-[1440px] mx-auto px-4 md:px-6 py-6 space-y-4">
  <div class="flex items-center justify-between gap-3 flex-wrap">
    <button
      onclick={() => goto("/rate")}
      class="flex items-center gap-2 text-sm font-bold text-primary cursor-pointer"
    >
      <ArrowLeft size={16} /> Back to lobbies
    </button>
    <div class="text-sm text-on-surface-variant">
      Room <span class="font-mono font-bold text-on-surface">{roomId}</span>
      · <span class="capitalize">{status}</span>
    </div>
  </div>

  {#if errorBanner}
    <div class="p-3 bg-red-50 text-red-700 text-sm rounded-sm">{errorBanner}</div>
  {/if}

  <div class="grid grid-cols-1 lg:grid-cols-12 gap-4">
    <!-- Players -->
    <aside class="lg:col-span-3 bg-surface-low rounded-sm p-4 space-y-3">
      <h2 class="text-sm font-black uppercase tracking-wide text-on-surface flex items-center gap-2">
        <Users size={16} /> Players
      </h2>
      <ul class="space-y-2">
        {#each players as player (player.session_id + "-" + playersVersion)}
          <li class="flex items-center justify-between gap-2 bg-surface-container rounded-sm px-3 py-2">
            <div class="min-w-0">
              <p class="text-sm font-bold text-on-surface truncate">
                {player.nickname}
                {#if player.is_host}<span class="text-primary text-xs ml-1">HOST</span>{/if}
                {#if player.offline}<span class="text-on-surface-variant text-xs ml-1">offline</span>{/if}
              </p>
              {#if !player.user_uuid}
                <p class="text-[10px] text-on-surface-variant">Guest (cannot rate)</p>
              {/if}
            </div>
            {#if status === "rating"}
              <span class="text-xs font-bold text-on-surface-variant shrink-0">{playerScoreLabel(player)}</span>
            {/if}
          </li>
        {/each}
      </ul>
      {#if spectators.length}
        <p class="text-xs text-on-surface-variant pt-2">Spectators: {spectators.length}</p>
      {/if}
      {#if ratingData && status === "rating"}
        <div class="pt-2 space-y-1">
          <p class="text-xs text-on-surface-variant">Session AVG</p>
          <p class="text-2xl font-black text-primary">{sessionAvgDisplay}</p>
          <p class="text-xs text-on-surface-variant">{ratingData.rated_count}/{ratingData.player_count} rated</p>
        </div>
      {/if}
    </aside>

    <!-- Center -->
    <section class="lg:col-span-6 bg-surface-container rounded-sm p-4 md:p-6 space-y-4">
      {#if status === "lobby"}
        <div class="text-center space-y-4 py-10">
          <Star size={40} class="mx-auto text-primary" />
          <h1 class="text-2xl font-black text-on-surface">{config?.name || "Rate Party"}</h1>
          <p class="text-sm text-on-surface-variant max-w-md mx-auto">
            Queue: {config?.queue_mode}
            {#if config?.queue_mode === "everyone"} (max {config.queue_limit_per_user}/user){/if}
            · Reveal: {config?.reveal_mode}
          </p>
          {#if isHost}
            <button
              onclick={() => send("start_session")}
              class="h-12 px-8 bg-primary text-white font-bold rounded-sm cursor-pointer"
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
            <button onclick={() => send("reset_to_lobby")} class="h-11 px-6 bg-primary text-white font-bold rounded-sm cursor-pointer">
              Back to lobby
            </button>
          {/if}
        </div>
      {:else}
        {#if currentSong}
          <div class="space-y-2">
            <p class="text-xs font-bold uppercase tracking-wide text-on-surface-variant">Now rating</p>
            <h1 class="text-2xl font-black text-on-surface">{getSongName(currentSong) || currentSong.name || "Theme"}</h1>
            {#if currentSong.anime}
              <p class="text-sm text-on-surface-variant">{currentSong.anime.title}</p>
            {/if}
          </div>

          <video
            bind:this={mediaEl}
            class="w-full aspect-video bg-on-surface rounded-sm"
            controls
            playsinline
            src={audioUrl || undefined}
          >
            <track kind="captions" />
          </video>

          {#if status === "rating" && !me?.is_spectator}
            <div class="bg-surface-highest rounded-sm p-4 space-y-3">
              {#if !authState.isAuthenticated}
                <p class="text-sm text-on-surface-variant">Log in to rate this song (saves to your global ranking).</p>
              {:else if playerRated(me!)}
                <p class="text-sm font-bold text-primary">
                  Your score: {getFormattedScore(ratingData?.my_score ?? draftScore, scoreFormat)}
                </p>
              {:else}
                <div class="flex items-center justify-between">
                  <span class="text-sm font-bold text-on-surface">Your score</span>
                  <span class="text-xl font-black text-primary">
                    {scoreFormat === "POINT_10_DECIMAL" || scoreFormat === "POINT_5"
                      ? displayDraft.toFixed(1)
                      : Math.round(displayDraft)}
                  </span>
                </div>
                {#if scoreFormat === "POINT_10" || scoreFormat === "POINT_10_DECIMAL"}
                  <div class="grid grid-cols-5 gap-2">
                    {#each [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] as n}
                      <button
                        onclick={() => {
                          draftScore = toCanonicalScore(n, scoreFormat);
                        }}
                        class="h-10 rounded-sm text-sm font-bold bg-surface-container hover:bg-primary hover:text-white cursor-pointer"
                      >
                        {n}
                      </button>
                    {/each}
                  </div>
                  {#if scoreFormat === "POINT_10_DECIMAL"}
                    <input
                      type="range"
                      min="0"
                      max="100"
                      step="1"
                      bind:value={draftScore}
                      class="w-full"
                      aria-label="Score slider"
                    />
                  {/if}
                {:else if scoreFormat === "POINT_100"}
                  <input type="range" min="0" max="100" step="1" bind:value={draftScore} class="w-full" aria-label="Score slider" />
                {:else}
                  <div class="grid grid-cols-5 gap-2">
                    {#each [0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5] as n}
                      <button
                        onclick={() => (draftScore = toCanonicalScore(n, "POINT_5"))}
                        class="h-10 rounded-sm text-sm font-bold bg-surface-container hover:bg-primary hover:text-white cursor-pointer"
                      >
                        {n}
                      </button>
                    {/each}
                  </div>
                {/if}
                <button
                  onclick={submitRating}
                  disabled={submitting}
                  class="w-full h-12 bg-primary text-white font-bold rounded-sm cursor-pointer disabled:opacity-50"
                >
                  Submit rating
                </button>
              {/if}
            </div>
          {/if}
        {:else}
          <div class="text-center py-16 space-y-2">
            <Play size={36} class="mx-auto text-primary" />
            <p class="font-bold text-on-surface">Waiting for a song</p>
            <p class="text-sm text-on-surface-variant">
              {#if isHost}Search below to play a theme or advance the queue.{:else}Host will pick the next theme.{/if}
            </p>
          </div>
        {/if}

        {#if isHost && (status === "waiting" || status === "rating")}
          <div class="flex flex-wrap gap-2">
            <button
              onclick={() => send("next")}
              class="h-10 px-4 bg-surface-highest text-primary font-bold text-sm rounded-sm flex items-center gap-2 cursor-pointer"
            >
              <SkipForward size={16} /> Next
            </button>
            <button
              onclick={() => send("end_session")}
              class="h-10 px-4 bg-surface-highest text-on-surface-variant font-bold text-sm rounded-sm cursor-pointer"
            >
              End session
            </button>
          </div>
        {/if}
      {/if}
    </section>

    <!-- Queue / Search / Chat -->
    <aside class="lg:col-span-3 space-y-4">
      {#if status !== "lobby" && status !== "finished" && config?.queue_mode !== "disabled"}
        <div class="bg-surface-low rounded-sm p-4 space-y-3">
          <h2 class="text-sm font-black uppercase tracking-wide text-on-surface">Queue ({queue.length})</h2>
          {#if queuePermission.ok || isHost}
            <div class="relative">
              <Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant" />
              <input
                bind:value={searchQuery}
                oninput={runSearch}
                placeholder="Search themes…"
                class="w-full h-10 pl-9 pr-3 bg-surface-highest border border-outline-variant rounded-sm text-sm"
              />
            </div>
            {#if searchLoading}
              <p class="text-xs text-on-surface-variant">Searching…</p>
            {/if}
            {#if searchResults.length}
              <ul class="max-h-40 overflow-y-auto space-y-1">
                {#each searchResults as song}
                  <li class="flex items-center justify-between gap-2 bg-surface-container rounded-sm px-2 py-1.5">
                    <span class="text-xs text-on-surface truncate">{song.song_romaji || song.name || "Song"}</span>
                    <div class="flex gap-1 shrink-0">
                      {#if isHost}
                        <button onclick={() => playNow(song)} class="text-[10px] font-bold text-primary cursor-pointer" aria-label="Play now">Play</button>
                      {/if}
                      {#if queuePermission.ok}
                        <button onclick={() => addSong(song)} class="text-[10px] font-bold text-on-surface-variant cursor-pointer" aria-label="Add to queue">+Q</button>
                      {/if}
                    </div>
                  </li>
                {/each}
              </ul>
            {/if}
          {:else if queuePermission.reason}
            <p class="text-xs text-on-surface-variant">{queuePermission.reason}</p>
          {/if}

          <ul class="space-y-2 max-h-56 overflow-y-auto">
            {#each queue as item (item.item_id)}
              <li class="bg-surface-container rounded-sm px-3 py-2 flex justify-between gap-2">
                <div class="min-w-0">
                  <p class="text-xs font-bold text-on-surface truncate">{item.song_name}</p>
                  <p class="text-[10px] text-on-surface-variant truncate">{item.anime_title} · {item.added_by_nickname}</p>
                </div>
                {#if isHost || item.added_by_session_id === mySessionId}
                  <button
                    onclick={() => send("queue_remove", { item_id: item.item_id })}
                    class="text-on-surface-variant cursor-pointer"
                    aria-label="Remove from queue"
                  >
                    <X size={14} />
                  </button>
                {/if}
              </li>
            {/each}
          </ul>
        </div>
      {:else if isHost && (status === "waiting" || status === "rating")}
        <div class="bg-surface-low rounded-sm p-4 space-y-3">
          <h2 class="text-sm font-black uppercase tracking-wide text-on-surface">Pick a song</h2>
          <div class="relative">
            <Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant" />
            <input
              bind:value={searchQuery}
              oninput={runSearch}
              placeholder="Search themes…"
              class="w-full h-10 pl-9 pr-3 bg-surface-highest border border-outline-variant rounded-sm text-sm"
            />
          </div>
          {#if searchResults.length}
            <ul class="max-h-48 overflow-y-auto space-y-1">
              {#each searchResults as song}
                <li class="flex items-center justify-between gap-2 bg-surface-container rounded-sm px-2 py-1.5">
                  <span class="text-xs text-on-surface truncate">{song.song_romaji || song.name || "Song"}</span>
                  <button onclick={() => playNow(song)} class="text-[10px] font-bold text-primary cursor-pointer">Play</button>
                </li>
              {/each}
            </ul>
          {/if}
        </div>
      {/if}

      <div class="bg-surface-low rounded-sm p-4 space-y-3">
        <h2 class="text-sm font-black uppercase tracking-wide text-on-surface flex items-center gap-2">
          <MessageSquare size={14} /> Chat
        </h2>
        <div class="h-40 overflow-y-auto space-y-1 text-xs">
          {#each chatMessages as m}
            <p class={m.type === "system" ? "text-on-surface-variant italic" : "text-on-surface"}>
              <strong>{m.sender}:</strong> {m.text}
            </p>
          {/each}
        </div>
        <div class="flex gap-2">
          <input
            bind:value={chatInput}
            onkeydown={(e) => e.key === "Enter" && sendChat()}
            class="flex-1 h-9 bg-surface-highest border border-outline-variant rounded-sm px-2 text-xs"
            placeholder="Message…"
          />
          <button onclick={sendChat} class="h-9 px-3 bg-primary text-white text-xs font-bold rounded-sm cursor-pointer">Send</button>
        </div>
      </div>
    </aside>
  </div>
</main>
