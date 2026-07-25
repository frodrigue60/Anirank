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
  import Eye from "lucide-svelte/icons/eye";
  import MessageSquare from "lucide-svelte/icons/message-square";
  import ChevronLeft from "lucide-svelte/icons/chevron-left";
  import ChevronRight from "lucide-svelte/icons/chevron-right";
  import Settings from "lucide-svelte/icons/settings";
  import {
    isSaveGameType as isSaveGameTypeHelper,
    saveGameModeLabel,
    formatSaveVoteSeconds,
    getSaveActiveCandidateIndex as getSaveActiveCandidateIndexHelper,
    shouldShowSavedSelection,
    canSelectSaveCandidate,
    savePhaseTimerLabel,
  } from "$lib/amq/save-mode";
  import {
    prefetchSaveCandidateMedia,
    warmSaveVideoElement,
    warmSaveVideoElements,
    resetSaveMediaPrefetchCache,
  } from "$lib/amq/save-video-prefetch";
  import { watchSaveVideoPlayback, type SaveVideoPlaybackContext } from "$lib/amq/save-video-playback";
  import {
    applySaveLobbyStateUpdate,
    applySavePhaseChange,
    applySaveRoundResults,
    applySaveRoundStart,
    isSaveRoundData,
  } from "$lib/amq/room-state";

  const roomId = page.params.roomId;

  let ws = $state<WebSocket | null>(null);
  let roomState = $state<any>(null);
  let status = $derived(roomState?.status || "lobby");

  // Sidebar visibility states
  let showLeftSidebar = $state(true);
  let showRightSidebar = $state(true);

  let centerColSpanClass = $derived.by(() => {
    let span = 12;
    if (showLeftSidebar) span -= 3;
    if (showRightSidebar) span -= 3;
    
    if (span === 6) return "lg:col-span-6";
    if (span === 9) return "lg:col-span-9";
    return "lg:col-span-12";
  });
  let players = $derived(roomState?.players || []);
  let spectators = $derived(roomState?.spectators || []);
  let sortedPlayers = $derived([...players].sort((a: any, b: any) => b.score - a.score));
  // Increments on every lobby_state_update to force Svelte to re-diff keyed #each items
  // when only inner fields change (e.g., is_host, is_ready) without a session_id change.
  let playersVersion = $state(0);
  let isSpectator = $state(false);
  let localTimer = $state(0);
  let config = $derived(roomState?.config || {});

  let currentRoundData = $state<any>(null);
  let activeRound = $derived(currentRoundData || roomState?.round_data || null);
  let roundResult = $state<any>(null);
  let selectedCandidate = $state("");
  let saveRoundResults = $state<any>(null);
  let saveVideoEls = $state<Record<string, HTMLVideoElement>>({});
  /** Tracks which candidates already received their one-time start seek this round. */
  let saveVideoStarted = $state<Record<string, boolean>>({});
  /** After reconnect, realign the active clip once to server timer. */
  let saveReconnectPending = $state(false);

  let isSaveMode = $derived(isSaveGameTypeHelper(config.game_type));
  let saveRoundHistory = $derived(roomState?.save_round_history || []);
  let savePhase = $derived(activeRound?.round_phase || "");

  let guessInput = $state("");
  let searchResults = $state<any[]>([]);
  let isLocked = $state(false);
  let selectedGuess = $state<any>(null);

  // Chat variables
  let chatInput = $state("");
  let chatMessages = $state<Array<{ sender: string; text: string; type: "system" | "user"; timestamp: string }>>([]);

  // Audio/Video player variables
  let videoElement = $state<HTMLVideoElement | null>(null);
  let volume = $state(0.5);
  let activeVolume = $state(0.5);
  let isMuted = $state(false);
  let isPlaying = $state(false);
  let onlyAudio = $state(false);
  let nextAudioUrl = $state("");

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
  let isRedirecting = false;
  let reconnectTimer: any = null;
  let connectionGeneration = 0;

  // Lobby config editing
  let showEditConfigModal = $state(false);
  let editRoomName = $state("");
  let editMaxRounds = $state(10);
  let editGuessTime = $state(20);
  let editRevealTime = $state(10);
  let editPreviewSeconds = $state(12);
  let editVoteSeconds = $state(10);
  let editThemeType = $state("both");
  let editGameType = $state("type-in");
  let editThemeDistribution = $state("random");
  let editPersonalizedPool = $state(false);
  let editIsPrivate = $state(false);

  let editIsSaveMode = $derived(editGameType === "save-4" || editGameType === "save-6");

  function openEditConfigModal() {
    if (!config) return;
    editRoomName = config.name || "";
    editMaxRounds = config.max_rounds || 10;
    editGuessTime = config.guess_time || 20;
    editRevealTime = config.reveal_time || 10;
    editPreviewSeconds = config.preview_seconds || 12;
    editVoteSeconds = config.vote_seconds ?? 10;
    editThemeType = config.theme_type || "both";
    editGameType = config.game_type || "type-in";
    editThemeDistribution = config.theme_distribution || "random";
    editPersonalizedPool = config.personalized_pool || false;
    editIsPrivate = config.private || false;
    showEditConfigModal = true;
  }

  function saveLobbyConfig() {
    const updated = {
      name: editRoomName.trim() || `${selfPlayer?.nickname || "Host"}'s Lobby`,
      max_rounds: Number(editMaxRounds),
      guess_time: Number(editGuessTime),
      reveal_time: Number(editRevealTime),
      preview_seconds: Number(editPreviewSeconds),
      vote_seconds: Number(editVoteSeconds),
      theme_type: editThemeType,
      game_type: editGameType,
      theme_distribution: editIsSaveMode ? editThemeDistribution : "random",
      personalized_pool: editIsSaveMode ? false : editPersonalizedPool,
      private: editIsPrivate,
    };
    sendWSMessage("update_lobby_config", updated);
    showEditConfigModal = false;
  }

  let showNicknamePrompt = $state(false);
  let nicknameError = $state("");
  let connectionInitiated = false;

  function generateUUID(): string {
    if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
      return crypto.randomUUID();
    }
    return Math.random().toString(36).substring(2, 15) + 
           Math.random().toString(36).substring(2, 15) + 
           Date.now().toString(36);
  }

  $effect(() => {
    if (!authState.loading && !connectionInitiated) {
      connectionInitiated = true;

      deviceId = localStorage.getItem("amq_device_id") || "";
      if (!deviceId) {
        deviceId = generateUUID();
        localStorage.setItem("amq_device_id", deviceId);
      }
      onlyAudio = localStorage.getItem("amq_only_audio") === "true";
      const urlParams = new URLSearchParams(window.location.search);
      isSpectator = urlParams.get("spectator") === "true";

      if (authState.isAuthenticated) {
        connectWebSocket();
      } else {
        const savedNickname = localStorage.getItem("amq_nickname") || "";
        if (!savedNickname) {
          showNicknamePrompt = true;
        } else {
          guestNickname = savedNickname;
          connectWebSocket();
        }
      }
    }
  });

  function confirmNickname() {
    nicknameError = "";
    if (!guestNickname.trim()) {
      nicknameError = "Nickname is required.";
      return;
    }
    localStorage.setItem("amq_nickname", guestNickname.trim());
    showNicknamePrompt = false;

    if (isSpectator) {
      const url = new URL(window.location.href);
      url.searchParams.set("spectator", "true");
      window.history.replaceState({}, "", url.toString());
    }

    connectWebSocket();
  }

  onDestroy(() => {
    closeWebSocket();
  });

  function toggleOnlyAudio() {
    onlyAudio = !onlyAudio;
    localStorage.setItem("amq_only_audio", onlyAudio.toString());
  }

  function handleLoadedMetadata() {
    if (!videoElement || !activeRound) return;
    
    const startPercent = activeRound.start_percent ?? activeRound.candidates?.[activeRound.preview_index ?? 0]?.start_percent ?? 0;
    const duration = videoElement.duration;
    
    if (isNaN(duration) || duration <= 0) return;

    let targetTime = duration * startPercent;
    const stepSeconds = activeRound.preview_seconds ?? config.preview_seconds ?? 12;
    
    if (isSaveMode && status === "playing") {
      const elapsed = Math.max(0, stepSeconds - localTimer);
      targetTime += elapsed;
    } else if (status === "playing") {
      const guessTime = activeRound.guess_time ?? 20;
      const elapsed = Math.max(0, guessTime - localTimer);
      targetTime += elapsed;
    } else if (status === "reveal") {
      const guessTime = activeRound.guess_time ?? 20;
      targetTime += guessTime;
    }
    
    if (targetTime < duration) {
      videoElement.currentTime = targetTime;
    } else {
      videoElement.currentTime = 0;
    }
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
    // Clear any pending reconnection timer
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }

    // Close previous connection cleanly before creating a new one
    if (ws) {
      ws.onclose = null;
      ws.onerror = null;
      ws.onmessage = null;
      ws.close();
      ws = null;
    }

    wsError = "";
    isReconnecting = false;

    // Increment generation so stale handlers are ignored
    const thisGeneration = ++connectionGeneration;
    
    const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    // Remove /api suffix from PUBLIC_API_URL to get backend root
    const apiBaseUrl = PUBLIC_API_URL.replace(/\/api$/, "").replace(/^https?:/, "");
    const wsUrl = `${wsProtocol}${apiBaseUrl}/api/amq/ws/${roomId}?token=${encodeURIComponent(getAuthToken() || "")}&device_id=${encodeURIComponent(deviceId)}&nickname=${encodeURIComponent(guestNickname)}&spectator=${isSpectator}`;

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      if (thisGeneration !== connectionGeneration) return;
      wsError = "";
      if (roomState?.status === "playing" && isSaveRoundData(roomState.round_data)) {
        saveVideoStarted = {};
        saveReconnectPending = true;
      }
    };

    ws.onmessage = (event) => {
      if (thisGeneration !== connectionGeneration) return;
      const msg = JSON.parse(event.data);

      switch (msg.type) {
        case "lobby_state_update":
          roomState = msg.payload;
          localTimer = msg.payload.timer_left ?? 0;
          playersVersion++;
          if (isSaveRoundData(roomState.round_data)) {
            const saveSync = applySaveLobbyStateUpdate(roomState, {
              isAuthenticated: authState.isAuthenticated,
              userUuid: authState.user?.uuid,
              deviceId,
            });
            currentRoundData = saveSync.currentRoundData;
            selectedCandidate = saveSync.selectedCandidate;
          }
          if (roomState.status === "reveal" && roomState.reveal_data && !roundResult) {
            roundResult = roomState.reveal_data;
          }
          break;
        case "round_start":
          if (isSaveRoundData(msg.payload)) {
            const saveStart = applySaveRoundStart(msg.payload);
            currentRoundData = saveStart.currentRoundData;
            localTimer = saveStart.localTimer;
            prefetchSaveCandidateMedia(msg.payload.candidates);
          } else {
            currentRoundData = msg.payload;
            localTimer = msg.payload.preview_seconds ?? msg.payload.guess_time ?? 20;
          }
          roundResult = null;
          saveRoundResults = null;
          guessInput = "";
          searchResults = [];
          isLocked = false;
          selectedGuess = null;
          selectedCandidate = "";
          saveVideoEls = {};
          saveVideoStarted = {};
          saveReconnectPending = false;
          resetSaveMediaPrefetchCache();
          isPlaying = true;
          nextAudioUrl = "";
          
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
        case "phase_change":
          if (isSaveRoundData(currentRoundData) || isSaveRoundData(msg.payload)) {
            const savePhaseUpdate = applySavePhaseChange(
              currentRoundData,
              msg.payload,
              config.preview_seconds ?? 12
            );
            currentRoundData = savePhaseUpdate.currentRoundData;
            if (savePhaseUpdate.saveRoundResults) {
              saveRoundResults = savePhaseUpdate.saveRoundResults;
            }
            localTimer = savePhaseUpdate.localTimer;
          } else if (currentRoundData) {
            currentRoundData = { ...currentRoundData, ...msg.payload };
            localTimer = currentRoundData?.preview_seconds ?? config.preview_seconds ?? 12;
          }
          setTimeout(() => {
            if (videoElement) {
              if (videoElement.readyState >= 1) {
                handleLoadedMetadata();
              }
              videoElement.play().catch(e => console.warn("Autoplay blocked:", e));
            }
          }, 100);
          break;
        case "round_results":
          if (isSaveRoundData(currentRoundData)) {
            const saveResults = applySaveRoundResults(currentRoundData, msg.payload);
            currentRoundData = saveResults.currentRoundData;
            saveRoundResults = saveResults.saveRoundResults;
            selectedCandidate = "";
          } else {
            saveRoundResults = msg.payload;
          }
          break;
        case "round_ended":
          roundResult = msg.payload;
          isPlaying = false;
          nextAudioUrl = msg.payload.next_audio_url || ""; // Set prefetch url for next round
          // Stop theme if playing
          if (videoElement) {
            fadeInVolume();
            videoElement.play().catch(e => console.warn("Failed to play in reveal:", e));
          }
          break;
        case "chat_message":
          chatMessages = [...chatMessages, msg.payload].slice(-100);
          break;
        case "room_closed":
          isRedirecting = true;
          closeWebSocket();
          alert("The room has been closed by the host.");
          goto("/amq");
          break;
        case "error":
          wsError = msg.payload;
          if (msg.payload === "room not found" || !roomState) {
            isRedirecting = true;
            closeWebSocket();
            alert(msg.payload || "An error occurred.");
            goto("/amq");
          }
          break;
      }
    };

    ws.onclose = (event) => {
      if (thisGeneration !== connectionGeneration) return;
      console.warn("[AMQ] WebSocket closed:", event);
      if (isRedirecting) return;
      if (status !== "finished" && !event.wasClean) {
        if (!roomState) {
          isRedirecting = true;
          alert("Could not connect to the room. It may no longer exist.");
          goto("/amq");
          return;
        }
        wsError = "Connection lost. Reconnecting...";
        isReconnecting = true;
        reconnectTimer = setTimeout(connectWebSocket, 3000);
      }
    };

    ws.onerror = (err) => {
      if (thisGeneration !== connectionGeneration) return;
      console.error("[AMQ] WebSocket error:", err);
      if (isRedirecting) return;
      if (!roomState) {
        isRedirecting = true;
        alert("Failed to connect to the game server. Room may not exist.");
        goto("/amq");
      } else {
        wsError = "Failed to connect to the game server.";
      }
    };
  }

  function closeWebSocket() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    connectionGeneration++;
    if (ws) {
      ws.onclose = null;
      ws.onerror = null;
      ws.onmessage = null;
      ws.close();
      ws = null;
    }
  }

  // Volume control reactivity
  $effect(() => {
    const vol = isMuted ? 0 : activeVolume;
    if (videoElement) {
      videoElement.volume = vol;
    }
    if (isSaveMode) {
      for (const vid of Object.values(saveVideoEls)) {
        vid.volume = vol;
      }
    }
  });

  function getSavePreviewCandidate() {
    if (!activeRound?.candidates?.length) return null;
    const idx = getSaveActiveCandidateIndex();
    return activeRound.candidates[idx] || null;
  }

  function saveVideoRef(node: HTMLVideoElement, songUuid: string) {
    saveVideoEls = { ...saveVideoEls, [songUuid]: node };
    warmSaveVideoElement(node);
    return {
      destroy() {
        if (saveVideoEls[songUuid] === node) {
          const next = { ...saveVideoEls };
          delete next[songUuid];
          saveVideoEls = next;
        }
      },
    };
  }

  function getSaveActiveCandidateIndex(): number {
    return getSaveActiveCandidateIndexHelper(activeRound, savePhase, saveRoundResults);
  }

  function buildSaveVideoPlaybackContext(
    candidate: { song_uuid: string; start_percent?: number },
    mode: SaveVideoPlaybackContext["mode"]
  ): SaveVideoPlaybackContext {
    const stepSeconds = activeRound?.preview_seconds ?? config.preview_seconds ?? 12;
    return {
      startPercent: candidate.start_percent ?? 0,
      mode,
      stepSeconds,
      elapsedSeconds: mode === "realign" ? Math.max(0, stepSeconds - localTimer) : undefined,
    };
  }

  // Play/pause tiles; retry seek/play until the active video is ready.
  $effect(() => {
    if (!isSaveMode || status !== "playing" || !activeRound?.candidates?.length) return;
    const activeIdx = getSaveActiveCandidateIndex();
    const _phase = savePhase;
    const _preview = activeRound.preview_index;
    const _winner = activeRound.winner_play_index;
    const _timer = localTimer;
    const cleanups: Array<() => void> = [];

    if (savePhase === "vote_select") {
      activeRound.candidates.forEach((candidate: any) => {
        const vid = saveVideoEls[candidate.song_uuid];
        if (vid) vid.pause();
      });
      return;
    }

    if (activeIdx < 0) return;

    activeRound.candidates.forEach((candidate: any, idx: number) => {
      const vid = saveVideoEls[candidate.song_uuid];
      if (!vid) return;

      if (idx === activeIdx) {
        const playbackMode: SaveVideoPlaybackContext["mode"] = saveReconnectPending ? "realign" : "initial";
        const alreadyStarted = saveVideoStarted[candidate.song_uuid];

        if (!alreadyStarted || saveReconnectPending) {
          cleanups.push(
            watchSaveVideoPlayback(
              vid,
              () => {
                if (getSaveActiveCandidateIndex() !== idx) return null;
                return buildSaveVideoPlaybackContext(candidate, playbackMode);
              },
              () => {
                saveVideoStarted = { ...saveVideoStarted, [candidate.song_uuid]: true };
                if (saveReconnectPending) saveReconnectPending = false;
              }
            )
          );
        } else {
          warmSaveVideoElement(vid);
          vid.play().catch(() => {});
        }
      } else {
        vid.pause();
      }
    });

    return () => {
      for (const cleanup of cleanups) cleanup();
    };
  });

  $effect(() => {
    if (!isSaveMode || status !== "playing" || !activeRound?.candidates?.length) return;
    warmSaveVideoElements(saveVideoEls);
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

  function selectSaveCandidate(songUuid: string) {
    if (!canSelectSaveCandidate(savePhase) || !selfPlayer || selfPlayer.is_spectator) return;
    if (selectedCandidate === songUuid) {
      selectedCandidate = "";
      sendWSMessage("select_candidate", { song_uuid: "" });
    } else {
      selectedCandidate = songUuid;
      sendWSMessage("select_candidate", { song_uuid: songUuid });
    }
  }

  function sendWSMessage(type: string, payload: any) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type, payload }));
    }
  }

  function sendChatMessage() {
    if (!chatInput.trim()) return;
    sendWSMessage("send_chat_message", { text: chatInput.trim() });
    chatInput = "";
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

  function skipWinnerPlayback() {
    sendWSMessage("skip_winner_playback", null);
  }

  // Return to Lobby after finished
  function resetToLobby() {
    sendWSMessage("reset_to_lobby", null);
  }

  function closeRoom() {
    if (confirm("Are you sure you want to close this room? This will disconnect all players and viewers.")) {
      sendWSMessage("close_room", null);
    }
  }

  function getSelfPlayer() {
    const all = [...(roomState?.players || []), ...(roomState?.spectators || [])];
    return all.find((p: any) => p.session_id === all.find((sp: any) => sp.session_id === p.session_id)?.session_id); // Session ID matching
  }

  let selfPlayer = $derived.by(() => {
    if (!roomState) return null;
    const all = [...(roomState.players || []), ...(roomState.spectators || [])];
    return all.find((p: any) => {
      if (authState.isAuthenticated) {
        return p.user_uuid === authState.user?.uuid;
      }
      return p.device_id === deviceId;
    });
  });

  // Auto-scroll chat to bottom when messages update
  $effect(() => {
    if (chatMessages.length) {
      setTimeout(() => {
        const container = document.getElementById("chat-messages-container");
        if (container) {
          container.scrollTop = container.scrollHeight;
        }
      }, 50);
    }
  });
</script>

<SEO
  title="Game Room - Anime Music Quiz"
  description="Play Anime Music Quiz. Guess anime songs and see results."
/>

<main class="max-w-[1440px] mx-auto px-6 py-10 space-y-8">
  <!-- Top Navigation & Code -->
  <nav class="flex items-center justify-between flex-wrap gap-4">
    <div class="flex items-center gap-4">
      <button
        onclick={() => goto("/amq")}
        class="h-10 text-primary hover:text-primary-container font-bold text-sm flex items-center gap-2 transition-colors cursor-pointer"
      >
        <ArrowLeft size={16} />
        Back to Lobbies
      </button>

      {#if !showLeftSidebar || !showRightSidebar}
        <div class="h-4 w-px bg-outline-variant hidden md:block"></div>
        <div class="flex items-center gap-2">
          {#if !showLeftSidebar}
            <button
              onclick={() => showLeftSidebar = true}
              class="h-8 bg-surface-low border border-outline-variant px-3 rounded-sm text-xs font-bold text-on-surface-variant hover:text-primary hover:border-primary/50 transition-all cursor-pointer flex items-center gap-1.5"
              title="Show Players"
            >
              <Users size={12} />
              Show Players
            </button>
          {/if}
          {#if !showRightSidebar}
            <button
              onclick={() => showRightSidebar = true}
              class="h-8 bg-surface-low border border-outline-variant px-3 rounded-sm text-xs font-bold text-on-surface-variant hover:text-primary hover:border-primary/50 transition-all cursor-pointer flex items-center gap-1.5"
              title="Show Chat"
            >
              <MessageSquare size={12} />
              Show Chat
            </button>
          {/if}
        </div>
      {/if}

      {#if selfPlayer?.is_host}
        <div class="h-4 w-px bg-outline-variant hidden md:block"></div>
        <button
          onclick={closeRoom}
          class="h-8 bg-red-500/10 hover:bg-red-500/20 text-red-500 border border-red-500/20 px-3 rounded-sm text-xs font-bold transition-all cursor-pointer flex items-center gap-1.5"
          title="Close Room (Disconnect all)"
        >
          <XCircle size={12} />
          Close Room
        </button>
      {/if}
    </div>

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
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- 1. Left Sidebar: Room Members (1 column) -->
      {#if showLeftSidebar}
        <section class="space-y-6 lg:col-span-3">
          <!-- Players List -->
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h3 class="text-xs font-black text-on-surface-variant uppercase tracking-widest flex items-center gap-2">
                <Users size={14} />
                Players ({players.length})
              </h3>
              <button
                onclick={() => showLeftSidebar = false}
                class="text-on-surface-variant hover:text-primary transition-colors cursor-pointer p-0.5 rounded-sm"
                title="Hide Players"
              >
                <ChevronLeft size={16} />
              </button>
            </div>
          <div class="bg-surface-low border border-outline-variant rounded-sm divide-y divide-outline-variant">
            {#if players.length === 0}
              <div class="p-4 text-xs text-on-surface-variant text-center">No players</div>
            {:else}
              {#each sortedPlayers as player (player.session_id + '-' + playersVersion)}
                <div class="px-4 py-3 flex items-center justify-between text-sm {player.offline ? 'opacity-50' : ''}">
                  <div class="flex items-center gap-2 truncate">
                    <!-- Status Indicator Dot -->
                    <span
                      class="w-2 h-2 rounded-full shrink-0
                        {player.offline ? 'bg-gray-500' :
                         status === 'lobby' ? (player.is_ready ? 'bg-green-500' : 'bg-yellow-500') :
                         status === 'playing' ? (player.locked ? 'bg-green-500 animate-pulse' : 'bg-yellow-500') :
                         status === 'reveal' ? (player.last_guess_correct ? 'bg-green-500' : 'bg-red-500') : 'bg-primary'}"
                    ></span>
                    <span class="font-bold text-on-surface truncate" title={player.nickname}>
                      {player.nickname}
                    </span>
                    {#if player.is_host}
                      <span class="text-[9px] bg-primary text-white px-1 py-0.2 rounded-xs font-black shrink-0">H</span>
                    {/if}
                  </div>
                  <div class="flex items-center gap-3 shrink-0">
                    <!-- Transfer Host Button -->
                    {#if selfPlayer?.is_host && !player.is_host && !player.offline}
                      <button
                        onclick={() => sendWSMessage("transfer_host", { target_session_id: player.session_id })}
                        class="text-[10px] bg-surface-highest hover:bg-primary hover:text-white border border-outline-variant px-2 py-0.5 rounded-xs font-bold transition-all cursor-pointer"
                        title="Make Host"
                      >
                        Make Host
                      </button>
                    {/if}
                    <div class="font-black text-xs text-primary">
                      {player.score} Pts
                    </div>
                  </div>
                </div>
              {/each}
            {/if}
          </div>
        </div>

        <!-- Viewers/Spectators List -->
        {#if spectators.length > 0}
          <div class="space-y-3">
            <h3 class="text-xs font-black text-on-surface-variant uppercase tracking-widest flex items-center gap-2">
              <Eye size={14} />
              Watching ({spectators.length})
            </h3>
            <div class="bg-surface-low border border-outline-variant rounded-sm divide-y divide-outline-variant max-h-48 overflow-y-auto">
              {#each spectators as spec (spec.session_id)}
                <div class="px-4 py-2.5 flex items-center justify-between text-xs {spec.offline ? 'opacity-50' : ''}">
                  <div class="flex items-center gap-2 truncate">
                    <span class="w-1.5 h-1.5 rounded-full shrink-0 {spec.offline ? 'bg-gray-500' : 'bg-primary'}"></span>
                    <span class="font-bold text-on-surface truncate" title={spec.nickname}>{spec.nickname}</span>
                  </div>
                  {#if spec.offline}
                    <span class="text-[9px] text-on-surface-variant shrink-0">(offline)</span>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </section>
    {/if}

    <!-- 2. Main Game Board (Center 3 columns) -->
    <section class="{centerColSpanClass} space-y-6">
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
                <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Game Mode</div>
                <div class="font-bold text-on-surface">{saveGameModeLabel(config.game_type)}</div>
              </div>
              <div class="space-y-1">
                <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Theme Type</div>
                <div class="font-bold text-on-surface uppercase">{config.theme_type}</div>
              </div>
              {#if isSaveMode}
                <div class="space-y-1">
                  <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Preview Time</div>
                  <div class="font-bold text-on-surface">{config.preview_seconds || 12} seconds</div>
                </div>
                <div class="space-y-1">
                  <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Vote Time</div>
                  <div class="font-bold text-on-surface">{formatSaveVoteSeconds(config.vote_seconds)}</div>
                </div>
              {:else}
                <div class="space-y-1">
                  <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Guess Time</div>
                  <div class="font-bold text-on-surface">{config.guess_time} seconds</div>
                </div>
                <div class="space-y-1">
                  <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">AniList Sync</div>
                  <div class="font-bold text-on-surface">{config.personalized_pool ? "Enabled" : "Disabled"}</div>
                </div>
              {/if}
              <div class="space-y-1">
                <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">Visibility</div>
                <div class="font-bold text-on-surface">{config.private ? "Private" : "Public"}</div>
              </div>
            </div>

            <!-- Host / Player ready triggers -->
            <div class="flex gap-4 pt-6 border-t border-outline-variant">
              {#if selfPlayer}
                {#if selfPlayer.is_spectator}
                  <div class="flex-1 p-3 bg-surface-low border border-outline-variant text-on-surface-variant rounded-sm text-center text-xs font-bold flex items-center justify-center gap-2">
                    <Eye size={16} class="text-primary" />
                    Watching as a spectator
                  </div>
                {:else}
                  <button
                    onclick={toggleReady}
                    class="flex-1 h-12 {selfPlayer.is_ready ? 'bg-green-600 hover:bg-green-700' : 'bg-primary hover:bg-primary-container'} text-white rounded-sm font-bold text-sm transition-colors cursor-pointer"
                  >
                    {selfPlayer.is_ready ? "You are Ready" : "Toggle Ready"}
                  </button>

                  {#if selfPlayer.is_host}
                    <button
                      onclick={openEditConfigModal}
                      class="h-12 bg-surface-highest hover:bg-surface-container text-on-surface border border-outline-variant px-5 rounded-sm font-bold text-sm flex items-center justify-center gap-2 transition-colors cursor-pointer"
                      title="Configure Room Settings"
                    >
                      <Settings size={16} />
                      Configure
                    </button>
                    <button
                      onclick={startGame}
                      class="flex-1 h-12 bg-primary hover:bg-primary-container text-white rounded-sm font-bold text-sm flex items-center justify-center gap-2 transition-colors cursor-pointer"
                    >
                      <Play size={16} />
                      Start Game
                    </button>
                  {/if}
                {/if}
              {/if}
            </div>
          </div>

        <!-- 2. SAVE MODE IN PROGRESS -->
        {:else if status === "playing" && isSaveMode}
          <div class="space-y-4">
            <div class="bg-surface-container px-5 py-4 rounded-md border border-white/5 flex flex-wrap items-center justify-between gap-3">
              <div class="space-y-1 min-w-0 flex-1">
                <div class="text-[10px] uppercase font-black text-primary tracking-widest">
                  Round {activeRound?.current_round}/{activeRound?.max_rounds}
                  {#if activeRound?.round_theme_type}
                    <span class="text-on-surface-variant"> · {activeRound.round_theme_type}</span>
                  {/if}
                </div>
                <h2 class="text-lg font-black text-on-surface tracking-tight leading-snug">{activeRound?.theme_label || "Save Round"}</h2>
                {#if savePhase === "vote_select"}
                  <p class="text-[11px] text-on-surface-variant">
                    Final vote — pick your save before time runs out.
                  </p>
                {:else if savePhase === "preview_select" && getSavePreviewCandidate()}
                  <p class="text-[11px] text-on-surface-variant truncate">
                    Now playing: <span class="font-bold text-on-surface">{getSavePreviewCandidate()?.theme_label}</span>
                    — {getSavePreviewCandidate()?.anime_title}
                  </p>
                {/if}
              </div>
              <div class="flex items-center gap-3 shrink-0">
                {#if activeRound?.is_fallback}
                  <span class="text-[10px] uppercase font-black tracking-widest bg-amber-500/15 text-amber-600 border border-amber-500/30 px-3 py-1 rounded-sm">
                    Fallback
                  </span>
                {/if}
                <div class="flex items-center gap-2 bg-surface-low border border-outline-variant rounded-sm px-2 py-1 text-white">
                  <button onclick={() => isMuted = !isMuted} class="cursor-pointer text-on-surface">
                    {#if isMuted || volume === 0}
                      <VolumeX size={16} />
                    {:else}
                      <Volume2 size={16} />
                    {/if}
                  </button>
                  <input type="range" min="0" max="1" step="0.05" bind:value={volume} class="w-16 accent-primary cursor-pointer h-1" />
                </div>
                <div class="text-right min-w-[72px]">
                  <div class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest">
                    {savePhaseTimerLabel(savePhase)}
                  </div>
                  <div class="text-2xl font-black text-primary leading-none">{localTimer}s</div>
                </div>
                {#if savePhase === "winner_playback" && selfPlayer?.is_host}
                  <button
                    onclick={skipWinnerPlayback}
                    class="bg-primary hover:bg-primary-container text-white px-3 py-2 rounded-sm font-bold text-[10px] flex items-center gap-1 transition-colors cursor-pointer"
                  >
                    <SkipForward size={12} />
                    Skip winners
                  </button>
                {/if}
              </div>
            </div>

            {#if activeRound?.candidates?.length}
              <div class="grid grid-cols-2 {activeRound.candidates.length > 4 ? 'lg:grid-cols-3' : ''} gap-3">
                {#each activeRound.candidates as candidate, idx (candidate.song_uuid)}
                  {@const isActiveWinner = savePhase === "winner_playback" && getSaveActiveCandidateIndex() === idx}
                  {@const isSelected = shouldShowSavedSelection(savePhase, selectedCandidate, candidate.song_uuid)}
                  {@const isWinner = candidate.is_winner && savePhase === "winner_playback"}
                  <button
                    type="button"
                    onclick={() => selectSaveCandidate(candidate.song_uuid)}
                    disabled={!canSelectSaveCandidate(savePhase) || selfPlayer?.is_spectator}
                    class="relative aspect-video rounded-sm border overflow-hidden transition-all cursor-pointer disabled:cursor-default text-left
                      {isWinner ? 'border-primary ring-2 ring-primary/50' : 'border-outline-variant'}
                      {isActiveWinner ? 'ring-2 ring-primary border-primary' : ''}
                      {isSelected ? 'ring-2 ring-green-500/80 border-green-500/60' : ''}
                      {canSelectSaveCandidate(savePhase) && !selfPlayer?.is_spectator && !isSelected ? 'hover:border-primary/40' : ''}"
                  >
                    <video
                      use:saveVideoRef={candidate.song_uuid}
                      src={candidate.audio_url}
                      preload="auto"
                      playsinline
                      class="absolute inset-0 w-full h-full object-cover bg-[#09070e] {onlyAudio ? 'opacity-0' : 'opacity-100'}"
                    >
                      <track kind="captions" />
                    </video>

                    {#if onlyAudio}
                      <div class="absolute inset-0 bg-[#09070e] flex items-center justify-center pointer-events-none">
                        <Music size={28} class="text-primary/50" />
                      </div>
                    {/if}

                    <div class="absolute inset-x-0 bottom-0 bg-[#09070e]/90 px-2.5 py-2 pointer-events-none">
                      <div class="text-[9px] font-black text-primary uppercase tracking-widest">{candidate.theme_label}</div>
                      <div class="text-[11px] font-bold text-white truncate" title={candidate.anime_title}>
                        {candidate.anime_title || "Unknown Anime"}
                      </div>
                    </div>

                    {#if isSelected}
                      <div class="absolute top-2 right-2 text-[9px] font-black uppercase bg-green-600 text-white px-2 py-0.5 rounded-sm pointer-events-none">
                        Saved
                      </div>
                    {/if}

                    {#if savePhase === "winner_playback"}
                      <div class="absolute top-2 right-2 text-[10px] font-black bg-black/75 text-white px-2 py-0.5 rounded-sm pointer-events-none">
                        {candidate.vote_count ?? 0} votes
                      </div>
                    {/if}
                  </button>
                {/each}
              </div>
            {/if}

            {#if selfPlayer?.is_spectator}
              <div class="p-4 bg-surface-low border border-outline-variant text-on-surface-variant text-sm font-bold rounded-sm text-center flex items-center justify-center gap-2">
                <Eye size={18} class="text-primary" />
                Watching as a spectator. Vote counts appear after preview ends.
              </div>
            {:else if savePhase === "preview_select"}
              <p class="text-xs text-on-surface-variant text-center">Click a video to save it. Nothing is selected until you choose one.</p>
            {/if}
          </div>

        <!-- 3. QUIZ MODE IN PROGRESS (Playing or Reveal) -->
        {:else if status === "playing" || status === "reveal"}
          <!-- Audio/Video Player Layer -->
          <div class="bg-surface-low rounded-md overflow-hidden border border-outline-variant flex flex-col items-center justify-center relative aspect-video w-full max-h-[360px] lg:max-h-[380px]">
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
                preload="auto"
                onloadedmetadata={handleLoadedMetadata}
                class="w-full h-full object-contain bg-[#09070e] {(status === 'playing' || onlyAudio) ? 'opacity-0 absolute pointer-events-none w-0 h-0' : 'opacity-100 block'}"
              >
                <track kind="captions" />
              </video>

              <!-- Prefetch next round video/audio in background -->
              {#if nextAudioUrl}
                <video src={nextAudioUrl} preload="auto" class="hidden" muted playsinline></video>
              {/if}

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
              {#if selfPlayer.is_spectator}
                <div class="p-4 bg-surface-low border border-outline-variant text-on-surface-variant text-sm font-bold rounded-sm text-center flex items-center justify-center gap-2">
                  <Eye size={18} class="text-primary animate-pulse" />
                  Watching as a spectator. No guesses can be made.
                </div>
              {:else if isLocked}
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
                    <label class="block class text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Select the correct Anime</label>
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
                    by {roundResult.song.artists?.map((a: any) => a.name).join(", ") || "Unknown Artist"}
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
              <p class="text-sm text-on-surface-variant">{isSaveMode ? "Review how each round was saved." : "Check out the final rankings below."}</p>
            </header>

            {#if isSaveMode && saveRoundHistory.length > 0}
              <div class="space-y-4 max-w-3xl mx-auto text-left">
                {#each saveRoundHistory as round (round.round_number)}
                  <div class="bg-surface-low border border-outline-variant rounded-sm p-4 space-y-3">
                    <div class="flex flex-wrap items-center justify-between gap-2 border-b border-outline-variant pb-2">
                      <div>
                        <div class="text-[10px] uppercase font-black text-primary tracking-widest">Round {round.round_number}</div>
                        <h3 class="text-sm font-black text-on-surface">{round.theme_label}</h3>
                      </div>
                      <div class="flex items-center gap-2">
                        {#if round.round_theme_type}
                          <span class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant">{round.round_theme_type}</span>
                        {/if}
                        {#if round.is_fallback}
                          <span class="text-[10px] uppercase font-black tracking-widest bg-amber-500/15 text-amber-600 border border-amber-500/30 px-2 py-0.5 rounded-sm">Fallback</span>
                        {/if}
                      </div>
                    </div>
                    <div class="space-y-2">
                      {#each round.candidates || [] as candidate (candidate.song_uuid)}
                        <div class="flex items-center justify-between gap-3 text-xs py-1.5 border-b border-outline-variant/40 last:border-0">
                          <div class="min-w-0">
                            <span class="font-bold text-on-surface">{candidate.theme_label}</span>
                            <span class="text-on-surface-variant"> — {candidate.anime_title}</span>
                          </div>
                          <div class="flex items-center gap-2 shrink-0">
                            <span class="font-black text-on-surface-variant">{candidate.vote_count ?? 0} votes</span>
                            {#if candidate.is_winner}
                              <span class="text-[10px] font-black uppercase tracking-widest text-primary">Winner</span>
                            {/if}
                          </div>
                        </div>
                      {/each}
                    </div>
                  </div>
                {/each}
              </div>
            {:else if roomState?.played_songs && roomState.played_songs.length > 0}
              <div class="bg-surface-low border border-outline-variant rounded-sm p-4 text-left space-y-3 max-w-2xl mx-auto">
                <h3 class="text-xs font-black text-on-surface-variant uppercase tracking-widest border-b border-outline-variant pb-2">
                  Song List ({roomState.played_songs.length})
                </h3>
                <div class="max-h-60 overflow-y-auto divide-y divide-outline-variant/50 text-xs">
                  {#each roomState.played_songs as song, idx}
                    <div class="py-2 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 sm:gap-4">
                      <div class="truncate text-on-surface">
                        <span class="font-bold text-on-surface-variant mr-1">#{idx + 1}</span>
                        <span class="font-bold">{song.song_romaji}</span>
                        <span class="text-on-surface-variant text-[11px]"> by {song.artists?.map(a => a.name).join(", ") || "Unknown Artist"}</span>
                      </div>
                      <div class="text-[10px] text-primary font-bold shrink-0 uppercase tracking-wider truncate max-w-xs" title={song.anime?.title}>
                        {song.anime?.title}
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}

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

      <!-- 3. Right Sidebar: Chat & Room Log (1 column) -->
      {#if showRightSidebar}
        <section class="lg:col-span-3 flex flex-col bg-surface-container rounded-md border border-white/5 overflow-hidden h-[550px] lg:h-[600px]">
          <header class="bg-surface-low border-b border-outline-variant p-4 flex items-center justify-between shrink-0">
            <h3 class="text-xs font-black text-on-surface-variant uppercase tracking-widest">
              Chat & Logs
            </h3>
            <button
              onclick={() => showRightSidebar = false}
              class="text-on-surface-variant hover:text-primary transition-colors cursor-pointer p-0.5 rounded-sm"
              title="Hide Chat"
            >
              <ChevronRight size={16} />
            </button>
          </header>

        <!-- Message List -->
        <div class="flex-1 overflow-y-auto p-4 space-y-1 flex flex-col justify-end" id="chat-messages-container">
          <div class="space-y-1 overflow-y-auto max-h-full">
            {#if chatMessages.length === 0}
              <div class="text-center text-xs text-on-surface-variant py-10 font-mono">
                No logs or messages yet.
              </div>
            {:else}
              {#each chatMessages as msg}
                <div class="font-mono text-[11px] leading-normal break-words py-0.5">
                  {#if msg.type === "system"}
                    <span class="text-amber-600/90 font-bold">{msg.text}</span>
                  {:else}
                    <span class="font-bold text-primary">{msg.sender}:</span>
                    <span class="text-on-surface">{msg.text}</span>
                  {/if}
                </div>
              {/each}
            {/if}
          </div>
        </div>

        <!-- Chat Input Form -->
        <form
          onsubmit={(e) => { e.preventDefault(); sendChatMessage(); }}
          class="p-3 bg-surface-low border-t border-outline-variant flex gap-2 shrink-0"
        >
          <input
            type="text"
            bind:value={chatInput}
            placeholder="Type a message..."
            class="flex-1 h-9 bg-surface-highest border border-outline-variant rounded-sm px-3 text-xs text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-2 focus:ring-primary/10 transition-all"
          />
          <button
            type="submit"
            class="h-9 bg-primary hover:bg-primary-container text-white px-4 rounded-sm font-bold text-xs transition-colors cursor-pointer shrink-0"
          >
            Send
          </button>
        </form>
      </section>
    {/if}
  </div>
  {/if}

  {#if showNicknamePrompt}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-[#09070e]/80 p-4 backdrop-blur-xs">
      <div class="bg-surface-highest max-w-sm w-full rounded-md shadow-2xl p-8 border border-outline-variant space-y-6 text-center">
        <div class="space-y-2">
          <h3 class="text-xl font-black text-on-surface tracking-tight">Join Game Room</h3>
          <p class="text-xs text-on-surface-variant">Enter a nickname to identify yourself in the lobby.</p>
        </div>

        <div class="space-y-4">
          <input
            type="text"
            bind:value={guestNickname}
            placeholder="Type your nickname..."
            maxlength="15"
            class="w-full h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-center text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all font-bold"
            onkeydown={(e) => { if (e.key === "Enter") confirmNickname(); }}
          />

          {#if nicknameError}
            <p class="text-xs text-red-500 font-bold">{nicknameError}</p>
          {/if}
        </div>

        <div class="flex gap-3 pt-2">
          <button
            onclick={() => { isSpectator = true; confirmNickname(); }}
            class="flex-1 h-11 bg-surface hover:bg-surface-low border border-outline-variant text-on-surface rounded-sm font-bold text-xs transition-colors cursor-pointer flex items-center justify-center gap-1"
          >
            <Eye size={14} />
            Watch
          </button>
          <button
            onclick={() => { isSpectator = false; confirmNickname(); }}
            class="flex-1 h-11 bg-primary hover:bg-primary-container text-white rounded-sm font-bold text-xs transition-colors cursor-pointer"
          >
            Play Game
          </button>
        </div>
      </div>
    </div>
  {/if}

  <!-- Edit Config Modal (Host only, only when in lobby status) -->
  {#if showEditConfigModal && selfPlayer?.is_host && status === "lobby"}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-[#09070e] p-4">
      <div class="bg-surface-highest max-w-lg w-full rounded-md shadow-2xl relative z-10 p-8 space-y-6 max-h-[90vh] overflow-y-auto border border-outline-variant">
        <h3 class="text-2xl font-black text-on-surface tracking-tight flex items-center gap-2">
          <Settings size={22} class="text-primary" />
          Configure Room Settings
        </h3>

        <div class="space-y-4">
          <!-- Room Name -->
          <div class="flex flex-col gap-2">
            <label for="edit-room-name" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Lobby Name</label>
            <input
              id="edit-room-name"
              type="text"
              bind:value={editRoomName}
              placeholder="Enter room name..."
              class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all"
            />
          </div>

          <!-- Rounds and Timers -->
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-2">
              <label for="edit-rounds" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Rounds</label>
              <input
                id="edit-rounds"
                type="number"
                min="5"
                max={editIsSaveMode ? "30" : "50"}
                bind:value={editMaxRounds}
                class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all"
              />
            </div>
            {#if editIsSaveMode}
              <div class="flex flex-col gap-2">
                <label for="edit-preview-seconds" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Preview (s)</label>
                <input
                  id="edit-preview-seconds"
                  type="number"
                  min="10"
                  max="15"
                  bind:value={editPreviewSeconds}
                  class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all"
                />
              </div>
              <div class="flex flex-col gap-2 col-span-2">
                <label for="edit-vote-seconds" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Vote (s)</label>
                <input
                  id="edit-vote-seconds"
                  type="number"
                  min="0"
                  max="60"
                  bind:value={editVoteSeconds}
                  class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all"
                />
                <span class="text-[11px] text-on-surface-variant ml-1">0 = tally immediately after previews</span>
              </div>
            {:else}
              <div class="flex flex-col gap-2">
                <label for="edit-guess-time" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Guess (s)</label>
                <input
                  id="edit-guess-time"
                  type="number"
                  min="10"
                  max="60"
                  bind:value={editGuessTime}
                  class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all"
                />
              </div>
              <div class="flex flex-col gap-2 col-span-2">
                <label for="edit-reveal-time" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Reveal (s)</label>
                <input
                  id="edit-reveal-time"
                  type="number"
                  min="5"
                  max="30"
                  bind:value={editRevealTime}
                  class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all"
                />
              </div>
            {/if}
          </div>

          <!-- Theme Pool and Game Type -->
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-2">
              <label for="edit-theme-type" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1 font-black">Themes</label>
              <select
                id="edit-theme-type"
                bind:value={editThemeType}
                class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
              >
                <option value="both">Openings & Endings</option>
                <option value="OP">Openings Only</option>
                <option value="ED">Endings Only</option>
              </select>
            </div>
            <div class="flex flex-col gap-2">
              <label for="edit-game-type" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Game Mode</label>
              <select
                id="edit-game-type"
                bind:value={editGameType}
                class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
              >
                <option value="type-in">Type-In (Autocomplete)</option>
                <option value="multiple-choice">Multiple Choice</option>
                <option value="save-4">Save 1 of 4</option>
                <option value="save-6">Save 1 of 6</option>
              </select>
            </div>
          </div>

          {#if editIsSaveMode}
            <div class="flex flex-col gap-2">
              <label for="edit-theme-distribution" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Theme Pool</label>
              <select
                id="edit-theme-distribution"
                bind:value={editThemeDistribution}
                class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
              >
                <option value="random">Random themes</option>
                <option value="balanced">Balanced themes</option>
              </select>
            </div>
          {/if}

          {#if authState.isAuthenticated && !editIsSaveMode}
            <div class="flex items-center gap-3 p-3 bg-surface rounded-sm border border-outline-variant">
              <input
                id="edit-personalized"
                type="checkbox"
                bind:checked={editPersonalizedPool}
                class="w-5 h-5 accent-primary cursor-pointer"
              />
              <div class="flex flex-col">
                <label for="edit-personalized" class="text-sm font-bold text-on-surface cursor-pointer">AniList Intersect Pool</label>
                <span class="text-[11px] text-on-surface-variant">Only draw songs from watched lists of synced users.</span>
              </div>
            </div>
          {/if}

          <!-- Private Lobby -->
          <div class="flex items-center gap-3 p-3 bg-surface rounded-sm border border-outline-variant">
            <input
              id="edit-private"
              type="checkbox"
              bind:checked={editIsPrivate}
              class="w-5 h-5 accent-primary cursor-pointer"
            />
            <div class="flex flex-col">
              <label for="edit-private" class="text-sm font-bold text-on-surface cursor-pointer">Private Lobby</label>
              <span class="text-[11px] text-on-surface-variant">Lobby will not appear in the public browser. Joined by room code only.</span>
            </div>
          </div>
        </div>

        <div class="flex gap-4 pt-4">
          <button
            onclick={() => showEditConfigModal = false}
            class="flex-1 h-12 bg-surface hover:bg-surface-low border border-outline-variant text-on-surface rounded-sm font-bold text-sm transition-colors cursor-pointer"
          >
            Cancel
          </button>
          <button
            onclick={saveLobbyConfig}
            class="flex-1 h-12 bg-primary hover:bg-primary-container text-white rounded-sm font-bold text-sm transition-colors cursor-pointer"
          >
            Save Preferences
          </button>
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
</style>
