<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import { authState, getAuthToken } from "$lib/state/auth.svelte";
  import { PUBLIC_API_URL } from "$lib/api";
  import api from "$lib/api";
  import SEO from "$lib/components/SEO.svelte";
  import Volume2 from "lucide-svelte/icons/volume-2";
  import VolumeX from "lucide-svelte/icons/volume-x";
  import Lock from "lucide-svelte/icons/lock";
  import Users from "lucide-svelte/icons/users";
  import Trophy from "lucide-svelte/icons/trophy";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";
  import Play from "lucide-svelte/icons/play";
  import SkipForward from "lucide-svelte/icons/skip-forward";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import CheckCircle from "lucide-svelte/icons/check-circle";
  import XCircle from "lucide-svelte/icons/x-circle";
  import Music from "lucide-svelte/icons/music";

  const roomId = page.params.roomId;

  let ws = $state<WebSocket | null>(null);
  let roomState = $state<any>(null);
  let status = $derived(roomState?.status || "lobby");
  let players = $derived(roomState?.players || []);
  let localTimer = $state(0);
  let config = $derived(roomState?.config || {});

  let currentRoundData = $state<any>(null);
  let activeRound = $derived(currentRoundData || roomState?.round_data || null);
  let roundResult = $state<any>(null);

  let guessInput = $state("");
  let searchResults = $state<any[]>([]);
  let isLocked = $state(false);
  let selectedGuess = $state<any>(null);

  // Audio/Video player variables
  let videoElement = $state<HTMLVideoElement | null>(null);
  let volume = $state(0.5);
  let activeVolume = $state(0.5);
  let isMuted = $state(false);
  let isPlaying = $state(false);
  let onlyAudio = $state(false);

  let timerInterval: any;
  $effect(() => {
    if (status === "playing" || status === "reveal") {
      clearInterval(timerInterval);
      timerInterval = setInterval(() => {
        if (localTimer > 0) {
          localTimer--;
        }
      }, 1000);
    } else {
      clearInterval(timerInterval);
    }

    return () => {
      clearInterval(timerInterval);
    };
  });

  // Guest details from localStorage
  let deviceId = $state("");
  let guestNickname = $state("");
  let wsError = $state("");
  let isReconnecting = $state(false);

  onMount(() => {
    deviceId = localStorage.getItem("amq_device_id") || "";
    guestNickname = localStorage.getItem("amq_nickname") || "Guest";
    onlyAudio = localStorage.getItem("amq_only_audio") === "true";
    connectWebSocket();
  });

  onDestroy(() => {
    closeWebSocket();
  });

  function toggleOnlyAudio() {
    onlyAudio = !onlyAudio;
    localStorage.setItem("amq_only_audio", onlyAudio.toString());
  }

  function handleLoadedMetadata() {
    if (!videoElement || !activeRound) return;
    
    const startPercent = activeRound.start_percent ?? 0;
    const duration = videoElement.duration;
    
    if (isNaN(duration) || duration <= 0) return;

    let targetTime = duration * startPercent;
    
    // If the round is in progress (e.g. on page refresh/reconnection),
    // offset the current time by the elapsed time in the guess phase.
    if (status === "playing") {
      const guessTime = activeRound.guess_time ?? 20;
      const elapsed = Math.max(0, guessTime - localTimer);
      targetTime += elapsed;
    } else if (status === "reveal") {
      // In reveal phase, offset by full guess time
      const guessTime = activeRound.guess_time ?? 20;
      targetTime += guessTime;
    }
    
    if (targetTime < duration) {
      videoElement.currentTime = targetTime;
    } else {
      videoElement.currentTime = 0;
    }
    console.log(`[AMQ] Seeking media to ${targetTime}s (StartPercent: ${startPercent}, Duration: ${duration}s)`);
  }

  let fadeInterval: any;
  function fadeInVolume() {
    clearInterval(fadeInterval);
    activeVolume = 0;
    const targetVolume = volume;
    const steps = 20; // 1s duration, 50ms intervals
    const increment = targetVolume / steps;

    fadeInterval = setInterval(() => {
      if (activeVolume < targetVolume) {
        activeVolume = Math.min(targetVolume, activeVolume + increment);
      } else {
        clearInterval(fadeInterval);
      }
    }, 50);
  }

  // Volume control manual overrides
  $effect(() => {
    const currentVol = volume;
    activeVolume = currentVol;
    clearInterval(fadeInterval);
  });

  function connectWebSocket() {
    wsError = "";
    isReconnecting = false;
    
    const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    // Remove /api suffix from PUBLIC_API_URL to get backend root
    const apiBaseUrl = PUBLIC_API_URL.replace(/\/api$/, "").replace(/^https?:/, "");
    const wsUrl = `${wsProtocol}${apiBaseUrl}/api/amq/ws/${roomId}?token=${encodeURIComponent(getAuthToken() || "")}&device_id=${encodeURIComponent(deviceId)}&nickname=${encodeURIComponent(guestNickname)}`;

    console.log("[AMQ] Connecting to WebSocket:", wsUrl);
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log("[AMQ] WebSocket connection established.");
      wsError = "";
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      console.log("[AMQ] Message received:", msg.type, msg.payload);

      switch (msg.type) {
        case "lobby_state_update":
          roomState = msg.payload;
          localTimer = msg.payload.timer_left ?? 0;
          break;
        case "round_start":
          currentRoundData = msg.payload;
          localTimer = msg.payload.guess_time;
          roundResult = null;
          guessInput = "";
          searchResults = [];
          isLocked = false;
          selectedGuess = null;
          isPlaying = true;
          
          // Autoplay video/audio
          setTimeout(() => {
            if (videoElement) {
              fadeInVolume();
              if (videoElement.readyState >= 1) {
                handleLoadedMetadata();
              }
              videoElement.play().catch(e => console.warn("Autoplay blocked:", e));
            }
          }, 100);
          break;
        case "round_ended":
          roundResult = msg.payload;
          isPlaying = false;
          // Stop theme if playing
          if (videoElement) {
            fadeInVolume();
            videoElement.play().catch(e => console.warn("Failed to play in reveal:", e));
          }
          break;
        case "error":
          wsError = msg.payload;
          break;
      }
    };

    ws.onclose = (event) => {
      console.warn("[AMQ] WebSocket closed:", event);
      if (status !== "finished" && !event.wasClean) {
        wsError = "Connection lost. Reconnecting...";
        isReconnecting = true;
        setTimeout(connectWebSocket, 3000); // Retry after 3s
      }
    };

    ws.onerror = (err) => {
      console.error("[AMQ] WebSocket error:", err);
      wsError = "Failed to connect to the game server.";
    };
  }

  function closeWebSocket() {
    if (ws) {
      ws.close();
      ws = null;
    }
  }

  // Volume control reactivity
  $effect(() => {
    if (videoElement) {
      videoElement.volume = isMuted ? 0 : activeVolume;
    }
  });

  // Autocomplete search
  let searchTimeout: any;
  function handleSearchInput() {
    clearTimeout(searchTimeout);
    if (!guessInput.trim()) {
      searchResults = [];
      return;
    }

    searchTimeout = setTimeout(async () => {
      try {
        const response = await api.get("/animes", {
          params: { name: guessInput.trim() },
        });
        if (response.data?.data) {
          searchResults = response.data.data;
        }
      } catch (e) {
        console.error("Autocomplete fetch failed:", e);
      }
    }, 200);
  }

  function selectAnime(anime: any) {
    guessInput = anime.title;
    searchResults = [];
    isLocked = true;
    selectedGuess = anime;

    sendWSMessage("submit_guess", {
      anime_slug: anime.slug,
    });
  }

  function selectMultipleChoice(option: any) {
    guessInput = option.title;
    isLocked = true;
    selectedGuess = option;

    sendWSMessage("submit_guess", {
      anime_slug: option.slug,
    });
  }

  function sendWSMessage(type: string, payload: any) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type, payload }));
    }
  }

  function toggleReady() {
    sendWSMessage("player_ready_toggle", null);
  }

  function startGame() {
    sendWSMessage("start_game", null);
  }

  function skipSummary() {
    sendWSMessage("skip_summary", null);
  }

  // Return to Lobby after finished
  function resetToLobby() {
    sendWSMessage("reset_to_lobby", null);
  }

  function getSelfPlayer() {
    return players.find((p: any) => p.session_id === roomState?.players?.find((sp: any) => sp.session_id === p.session_id)?.session_id); // Session ID matching
  }

  let selfPlayer = $derived.by(() => {
    if (!roomState || !players) return null;
    // Look up via local guest deviceID or Auth UUID
    return players.find((p: any) => {
      if (authState.isAuthenticated) {
        return p.user_uuid === authState.user?.uuid;
      }
      return p.device_id === deviceId;
    });
  });
</script>

<SEO
  title="Game Room - Anime Music Quiz"
  description="Play Anime Music Quiz. Guess anime songs and see results."
/>

<main class="max-w-[1200px] mx-auto px-6 py-10 space-y-8">
  <!-- Top Navigation & Code -->
  <nav class="flex items-center justify-between">
    <button
      onclick={() => goto("/amq")}
      class="h-10 text-primary hover:text-primary-container font-bold text-sm flex items-center gap-2 transition-colors cursor-pointer"
    >
      <ArrowLeft size={16} />
      Back to Lobbies
    </button>

    <div class="flex items-center gap-4 bg-surface-low px-4 py-2 rounded-sm border border-outline-variant">
      <span class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Room Code</span>
      <span class="font-black text-lg text-primary tracking-wider uppercase select-all">{roomId}</span>
    </div>
  </nav>

  <!-- Error or Reconnection States -->
  {#if wsError}
    <div class="p-4 bg-red-50 text-red-700 text-sm rounded-sm border border-red-200 flex items-center gap-2">
      {#if isReconnecting}
        <RefreshCw size={16} class="animate-spin" />
      {/if}
      {wsError}
    </div>
  {/if}

  {#if !roomState}
    <div class="text-center py-20 text-on-surface-variant text-sm">
      Connecting to room session...
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-4 gap-8">
      <!-- Main Game Board (Left 3 columns) -->
      <section class="lg:col-span-3 space-y-6">
        <!-- 1. LOBBY STATE (Waiting Room) -->
        {#if status === "lobby"}
          <div class="bg-surface-container p-8 rounded-md space-y-6 border border-white/5">
            <header class="border-b border-outline-variant pb-4">
              <h2 class="text-2xl font-black text-on-surface tracking-tight">{config.name}</h2>
              <p class="text-xs text-on-surface-variant mt-1">Waiting for host to start the game.</p>
            </header>

            <!-- Rules Grid -->
            <div class="grid grid-cols-2 md:grid-cols-3 gap-6 text-sm">
              <div class="space-y-1">
                <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Rounds</div>
                <div class="font-bold text-on-surface">{config.max_rounds} rounds</div>
              </div>
              <div class="space-y-1">
                <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Guess Time</div>
                <div class="font-bold text-on-surface">{config.guess_time} seconds</div>
              </div>
              <div class="space-y-1">
                <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Theme Type</div>
                <div class="font-bold text-on-surface uppercase">{config.theme_type}</div>
              </div>
              <div class="space-y-1">
                <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Game Mode</div>
                <div class="font-bold text-on-surface capitalize">{config.game_type}</div>
              </div>
              <div class="space-y-1">
                <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Visibility</div>
                <div class="font-bold text-on-surface">{config.private ? "Private" : "Public"}</div>
              </div>
              <div class="space-y-1">
                <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">AniList Sync</div>
                <div class="font-bold text-on-surface">{config.personalized_pool ? "Enabled" : "Disabled"}</div>
              </div>
            </div>

            <!-- Host / Player ready triggers -->
            <div class="flex gap-4 pt-6 border-t border-outline-variant">
              {#if selfPlayer}
                <button
                  onclick={toggleReady}
                  class="flex-1 h-12 {selfPlayer.is_ready ? 'bg-green-600 hover:bg-green-700' : 'bg-primary hover:bg-primary-container'} text-white rounded-sm font-bold text-sm transition-colors cursor-pointer"
                >
                  {selfPlayer.is_ready ? "You are Ready" : "Toggle Ready"}
                </button>

                {#if selfPlayer.is_host}
                  <button
                    onclick={startGame}
                    class="flex-1 h-12 bg-primary hover:bg-primary-container text-white rounded-sm font-bold text-sm flex items-center justify-center gap-2 transition-colors cursor-pointer"
                  >
                    <Play size={16} />
                    Start Game
                  </button>
                {/if}
              {/if}
            </div>
          </div>

        <!-- 2. GAME IN PROGRESS (Playing or Reveal) -->
        {:else if status === "playing" || status === "reveal"}
          <!-- Audio/Video Player Layer -->
          <div class="bg-surface-low rounded-md overflow-hidden border border-outline-variant flex flex-col items-center justify-center relative aspect-video">
            <!-- Round Badge -->
            {#if activeRound}
              <div class="absolute top-4 left-4 bg-[#09070e]/80 px-3 py-1.5 rounded-sm text-xs font-black text-white tracking-widest uppercase border border-white/10 z-10">
                Round {activeRound.current_round}/{activeRound.max_rounds}
              </div>
            {/if}

            {#if activeRound}
              <!-- Hidden Video during guessing, revealed on summary -->
              <video
                bind:this={videoElement}
                src={activeRound.audio_url}
                onloadedmetadata={handleLoadedMetadata}
                class="w-full h-full object-cover {(status === 'playing' || onlyAudio) ? 'opacity-0 absolute pointer-events-none w-0 h-0' : 'opacity-100 block'}"
              >
                <track kind="captions" />
              </video>

              <!-- Only Audio Reveal Phase Placeholder -->
              {#if status === "reveal" && onlyAudio && roundResult}
                <div class="absolute inset-0 bg-[#09070e] flex flex-col items-center justify-center space-y-4 z-0">
                  {#if roundResult.song.anime?.cover_url}
                    <img
                      src={roundResult.song.anime.cover_url}
                      alt={roundResult.song.anime.title || "Cover"}
                      class="w-28 rounded-sm object-cover aspect-[3/4] shadow-2xl border border-white/10"
                    />
                  {/if}
                  <div class="text-center space-y-1 px-4">
                    <h4 class="font-black text-white text-base tracking-tight leading-tight">{roundResult.song.song_romaji}</h4>
                    <p class="text-xs text-primary font-bold">{roundResult.song.artists?.map(a => a.name).join(", ")}</p>
                    <p class="text-[10px] text-white/50">{roundResult.song.anime?.title}</p>
                  </div>
                </div>
              {/if}

              <!-- Guessing phase overlay -->
              {#if status === "playing"}
                <div class="absolute inset-0 bg-[#09070e] flex flex-col items-center justify-center space-y-4">
                  <div class="w-20 h-20 bg-primary/20 rounded-full flex items-center justify-center animate-pulse">
                    <Music size={40} class="text-primary animate-spin" style="animation-duration: 8s" />
                  </div>
                  <h3 class="font-black text-white text-xl tracking-tight uppercase">Playing Theme...</h3>
                  
                  <!-- Guessing Timer Countdown -->
                  <div class="text-5xl font-black text-primary">{localTimer}</div>
                </div>
              {/if}
            {/if}

            <!-- Volume Control Bar -->
            <div class="absolute bottom-4 left-4 flex items-center gap-3 bg-[#09070e]/80 p-2 rounded-sm text-white">
              <button onclick={() => isMuted = !isMuted} class="cursor-pointer">
                {#if isMuted || volume === 0}
                  <VolumeX size={18} />
                {:else}
                  <Volume2 size={18} />
                {/if}
              </button>
              <input
                type="range"
                min="0"
                max="1"
                step="0.05"
                bind:value={volume}
                class="w-20 accent-primary cursor-pointer h-1 rounded-full bg-outline-variant"
              />
              <span class="w-px h-4 bg-white/20"></span>
              <button
                onclick={toggleOnlyAudio}
                class="text-[9px] uppercase font-black px-2 py-1 rounded-sm border {onlyAudio ? 'bg-primary border-primary text-white' : 'border-white/25 text-white/60 hover:text-white'} transition-colors cursor-pointer"
              >
                {onlyAudio ? "Audio Only" : "Video On"}
              </button>
            </div>

            <!-- Host skip summary button in reveal phase -->
            {#if status === "reveal" && selfPlayer?.is_host}
              <button
                onclick={skipSummary}
                class="absolute bottom-4 right-4 bg-primary hover:bg-primary-container text-white px-4 py-2 rounded-sm font-bold text-xs flex items-center gap-1 transition-colors cursor-pointer"
              >
                <SkipForward size={14} />
                Skip Reveal
              </button>
            {/if}
          </div>

          <!-- Guessing Phase Inputs -->
          {#if status === "playing" && selfPlayer}
            <div class="bg-surface-container p-6 rounded-md border border-white/5 space-y-4">
              {#if isLocked}
                <div class="p-4 bg-green-50 text-green-800 text-sm font-bold rounded-sm border border-green-200 flex items-center gap-2">
                  <CheckCircle size={18} />
                  Your answer is locked: "{guessInput}"
                </div>
              {:else if !activeRound}
                <div class="text-center py-6 text-on-surface-variant text-sm flex items-center justify-center gap-2">
                  <RefreshCw size={16} class="animate-spin text-primary" />
                  Loading songs and preparing game pool...
                </div>
              {:else}
                <!-- Autocomplete input mode -->
                {#if activeRound.game_type === "type-in"}
                  <div class="relative group">
                    <label for="autocomplete-guess" class="block text-[10px] uppercase font-black text-on-surface-variant tracking-widest mb-2 ml-1">Your Guess</label>
                    <input
                      id="autocomplete-guess"
                      type="text"
                      bind:value={guessInput}
                      oninput={handleSearchInput}
                      placeholder="Type anime title..."
                      class="w-full h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all"
                    />

                    <!-- Autocomplete Matches -->
                    {#if searchResults.length > 0}
                      <ul class="absolute z-50 left-0 right-0 mt-1 bg-surface-highest border border-outline-variant rounded-sm shadow-xl max-h-60 overflow-y-auto divide-y divide-outline-variant">
                        {#each searchResults as anime}
                          <li>
                            <button
                              onclick={() => selectAnime(anime)}
                              class="w-full text-left p-3 text-sm text-on-surface hover:bg-surface-low transition-colors cursor-pointer flex items-center justify-between"
                            >
                              <div>
                                <span class="font-bold">{anime.title}</span>
                                {#if anime.title_english && anime.title_english !== anime.title}
                                  <span class="text-xs text-on-surface-variant block">{anime.title_english}</span>
                                  {/if}
                              </div>
                              {#if anime.year}
                                <span class="text-xs text-on-surface-variant font-semibold">{anime.year.name}</span>
                              {/if}
                            </button>
                          </li>
                        {/each}
                      </ul>
                    {/if}
                  </div>

                <!-- Multiple choice input mode -->
                {:else if activeRound.game_type === "multiple-choice"}
                  <div class="space-y-3">
                    <label class="block text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Select the correct Anime</label>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {#each activeRound.options as option}
                        <button
                          onclick={() => selectMultipleChoice(option)}
                          class="h-14 bg-surface hover:bg-surface-low text-on-surface font-bold text-sm border border-outline-variant rounded-sm px-4 text-left transition-all cursor-pointer flex items-center justify-between"
                        >
                          <span>{option.title}</span>
                        </button>
                      {/each}
                    </div>
                  </div>
                {/if}
              {/if}
            </div>
          {/if}

          <!-- Reveal Phase summary -->
          {#if status === "reveal" && roundResult}
            <div class="bg-surface-container p-6 rounded-md border border-white/5 space-y-6">
              <!-- Correct song headers -->
              <header class="flex flex-col md:flex-row gap-6 items-start border-b border-outline-variant pb-6">
                {#if roundResult.song.anime && roundResult.song.anime.cover_url}
                  <img
                    src={roundResult.song.anime.cover_url}
                    alt={roundResult.song.anime.title}
                    class="w-24 rounded-sm object-cover aspect-[3/4]"
                  />
                {/if}
                <div class="space-y-2">
                  <div class="text-[10px] uppercase font-black text-primary tracking-widest font-black">
                    Round {activeRound?.current_round} — {roundResult.song.theme_type?.name || roundResult.song.type}{roundResult.song.theme_num}
                  </div>
                  <h3 class="text-2xl font-black text-on-surface tracking-tight leading-none">
                    {roundResult.song.song_romaji || "Unknown Theme"}
                  </h3>
                  <div class="text-sm font-bold text-on-surface-variant">
                    by {roundResult.song.artists.map((a: any) => a.name).join(", ")}
                  </div>
                  <div class="text-xs text-on-surface-variant pt-2">
                    Anime: <span class="font-bold text-on-surface">{roundResult.song.anime?.title}</span>
                    {#if roundResult.song.anime?.title_english}
                      <span class="block">English: {roundResult.song.anime.title_english}</span>
                    {/if}
                  </div>
                </div>
              </header>

              <!-- Reveal Countdown Timer -->
              <div class="flex justify-between items-center text-sm">
                <span class="text-on-surface-variant">Next round starting in:</span>
                <span class="font-black text-primary text-xl">{localTimer}s</span>
              </div>
            </div>
          {/if}

        <!-- 3. GAME COMPLETED (Leaderboard & Play Again) -->
        {:else if status === "finished"}
          <div class="bg-surface-container p-8 rounded-md border border-white/5 space-y-8 text-center">
            <header class="space-y-2">
              <Trophy size={60} class="mx-auto text-primary animate-bounce" />
              <h2 class="text-3xl font-black text-on-surface tracking-tight">GAME FINISHED</h2>
              <p class="text-sm text-on-surface-variant">Check out the final rankings below.</p>
            </header>

            <!-- Play again host reset -->
            {#if selfPlayer?.is_host}
              <div class="pt-6 border-t border-outline-variant max-w-sm mx-auto">
                <button
                  onclick={resetToLobby}
                  class="w-full h-12 bg-primary hover:bg-primary-container text-white rounded-sm font-bold text-sm transition-colors cursor-pointer"
                >
                  Play Again (Reset Lobby)
                </button>
              </div>
            {/if}
          </div>
        {/if}
      </section>

      <!-- Sidebar Room Members & Scores (Right 1 column) -->
      <section class="space-y-4">
        <h3 class="text-sm font-black text-on-surface-variant uppercase tracking-widest flex items-center gap-2">
          <Users size={16} />
          Lobby Players ({players.length})
        </h3>

        <div class="bg-surface-low border border-outline-variant rounded-md divide-y divide-outline-variant">
          {#each players.sort((a: any, b: any) => b.score - a.score) as player}
            <div class="p-4 flex items-center justify-between text-sm {player.offline ? 'opacity-50' : ''}">
              <div class="flex items-center gap-3">
                <div
                  class="w-8 h-8 rounded-sm flex items-center justify-center font-bold text-white text-xs"
                  style="background-color: {player.profile_color || '#683bc9'}"
                >
                  {player.nickname.slice(0, 2).toUpperCase()}
                </div>
                <div class="space-y-0.5">
                  <div class="font-bold text-on-surface flex items-center gap-1.5 leading-none">
                    {player.nickname}
                    {#if player.is_host}
                      <span class="text-[9px] bg-primary text-white px-1 py-0.5 rounded-sm scale-90">Host</span>
                    {/if}
                  </div>
                  <div class="text-[10px] text-on-surface-variant font-semibold">
                    {#if status === "lobby"}
                      {player.is_ready ? "Ready" : "Not Ready"}
                    {:else if status === "playing"}
                      {player.locked ? "Locked" : "Selecting..."}
                    {:else if status === "reveal"}
                      {#if player.last_guess_correct}
                        <span class="text-green-600 font-bold">CORRECT</span>
                      {:else}
                        <span class="text-red-600">INCORRECT</span>
                      {/if}
                    {/if}
                  </div>
                </div>
              </div>

              <!-- Score Badge -->
              <div class="flex items-center gap-2">
                {#if status === "reveal" && player.last_guess_correct}
                  <span class="text-green-600"><CheckCircle size={16} /></span>
                {:else if status === "reveal" && !player.last_guess_correct}
                  <span class="text-red-600"><XCircle size={16} /></span>
                {/if}
                <div class="bg-surface-container px-2.5 py-1 rounded-sm font-black text-xs text-on-surface">
                  {player.score} pts
                </div>
              </div>
            </div>
          {/each}
        </div>
      </section>
    </div>
  {/if}
</main>

<style>
</style>
