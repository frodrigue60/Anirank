<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { authState } from "$lib/state/auth.svelte";
  import api from "$lib/api";
  import SEO from "$lib/components/SEO.svelte";
  import Plus from "lucide-svelte/icons/plus";
  import Lock from "lucide-svelte/icons/lock";
  import Users from "lucide-svelte/icons/users";
  import Play from "lucide-svelte/icons/play";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import Eye from "lucide-svelte/icons/eye";
  import UserIcon from "lucide-svelte/icons/user";

  interface RoomInfo {
    room_id: string;
    name: string;
    host_nickname: string;
    player_count: number;
    spectator_count: number;
    max_rounds: number;
    status: string;
    private: boolean;
    theme_type: string;
    game_type: string;
  }

  let rooms = $state<RoomInfo[]>([]);
  let loading = $state(false);
  let showCreateModal = $state(false);
  let showNicknameModal = $state(false);

  // Guest details (for unauthenticated users)
  let guestNickname = $state("");
  let deviceId = $state("");
  let hasSavedNickname = $state(false);

  // Room config form
  let roomName = $state("");
  let maxRounds = $state(10);
  let guessTime = $state(20);
  let revealTime = $state(10);
  let themeType = $state("both"); // "OP", "ED", "both"
  let gameType = $state("type-in"); // "type-in", "multiple-choice"
  let personalizedPool = $state(false);
  let isPrivate = $state(false);

  let joinCode = $state("");
  let errorMsg = $state("");

  onMount(() => {
    // Guest device ID initialization
    if (typeof window !== "undefined") {
      let storedDeviceId = localStorage.getItem("amq_device_id");
      if (!storedDeviceId) {
        storedDeviceId = crypto.randomUUID();
        localStorage.setItem("amq_device_id", storedDeviceId);
      }
      deviceId = storedDeviceId;
      guestNickname = localStorage.getItem("amq_nickname") || "";
      hasSavedNickname = !!guestNickname;
    }
    fetchRooms();
  });

  $effect(() => {
    if (!authState.loading && !authState.isAuthenticated && !hasSavedNickname) {
      showNicknameModal = true;
    }
  });

  async function fetchRooms() {
    loading = true;
    errorMsg = "";
    try {
      const response = await api.get("/amq/rooms");
      if (response.data?.success) {
        rooms = response.data.data;
      }
    } catch (e) {
      console.error("Error loading lobbies", e);
      errorMsg = "Failed to load active lobbies.";
    } finally {
      loading = false;
    }
  }

  function saveNickname() {
    if (!guestNickname.trim()) {
      errorMsg = "Nickname cannot be empty.";
      return false;
    }
    localStorage.setItem("amq_nickname", guestNickname.trim());
    hasSavedNickname = true;
    showNicknameModal = false;
    errorMsg = "";
    return true;
  }

  function cancelNicknameChange() {
    guestNickname = localStorage.getItem("amq_nickname") || "";
    showNicknameModal = false;
    errorMsg = "";
  }

  async function handleCreateRoom() {
    errorMsg = "";

    // Validation
    const nickname = authState.isAuthenticated ? authState.user?.name : guestNickname.trim();
    if (!nickname) {
      errorMsg = "Please enter a nickname first.";
      showNicknameModal = true;
      return;
    }

    if (!authState.isAuthenticated) {
      saveNickname();
    }

    if (!roomName.trim()) {
      roomName = `${nickname}'s Lobby`;
    }

    try {
      const payload = {
        config: {
          name: roomName.trim(),
          max_rounds: maxRounds,
          guess_time: guessTime,
          reveal_time: revealTime,
          theme_type: themeType,
          game_type: gameType,
          personalized_pool: personalizedPool,
          private: isPrivate,
        },
        guest_nickname: nickname,
        guest_device_id: deviceId,
      };

      const response = await api.post("/amq/rooms", payload);
      if (response.data?.success) {
        const roomId = response.data.data.room_id;
        goto(`/amq/${roomId}`);
      }
    } catch (e: any) {
      console.error("Error creating room", e);
      errorMsg = e.response?.data?.message || "Failed to create lobby.";
    }
  }

  function handleJoinRoom(roomId: string, spectator: boolean = false) {
    errorMsg = "";
    const nickname = authState.isAuthenticated ? authState.user?.name : guestNickname.trim();
    if (!nickname) {
      errorMsg = "Please enter a nickname first.";
      showNicknameModal = true;
      return;
    }

    if (!authState.isAuthenticated) {
      saveNickname();
    }

    goto(`/amq/${roomId}${spectator ? "?spectator=true" : ""}`);
  }

  function handleJoinByCode() {
    if (!joinCode.trim()) return;
    handleJoinRoom(joinCode.trim().toUpperCase());
  }
</script>

<SEO
  title="Anime Music Quiz"
  description="Join real-time anime music quiz lobbies. Test your knowledge on openings and endings with friends or guests."
/>

<main class="max-w-[1440px] mx-auto px-6 py-10 space-y-8">
  <!-- Editorial Header -->
  <header class="space-y-4">
    <div class="flex items-center gap-3">
      <span class="w-1.5 h-8 bg-primary rounded-full"></span>
      <h1 class="text-4xl font-black tracking-tight text-on-surface">ANIME MUSIC QUIZ</h1>
    </div>
    <p class="text-on-surface-variant text-base max-w-xl leading-relaxed">
      Challenge your anime music knowledge in real-time. Guess openings (OPs) and endings (EDs) 
      from your AniList synced list or the general catalog. Join a public lobby below or create your own.
    </p>
    {#if !authState.isAuthenticated && guestNickname}
      <div class="flex items-center gap-2 text-xs text-on-surface-variant">
        <span>Playing as guest: <strong class="text-on-surface">{guestNickname}</strong></span>
        <button 
          onclick={() => showNicknameModal = true} 
          class="text-primary hover:underline font-bold cursor-pointer"
        >
          (Change Nickname)
        </button>
      </div>
    {/if}
  </header>

  <!-- Error Alert -->
  {#if errorMsg}
    <div class="p-4 bg-red-50 text-red-700 text-sm rounded-sm border border-red-200">
      {errorMsg}
    </div>
  {/if}

  <!-- Actions and Join Code Section -->
  <section class="flex flex-col md:flex-row gap-4 justify-between items-stretch md:items-center">
    <div class="flex gap-3">
      <button
        onclick={() => showCreateModal = true}
        class="h-12 bg-primary hover:bg-primary-container text-white px-6 rounded-sm font-bold text-sm flex items-center gap-2 transition-colors cursor-pointer"
      >
        <Plus size={18} />
        Create Lobby
      </button>
      <button
        onclick={fetchRooms}
        disabled={loading}
        class="h-12 bg-surface-highest text-primary border border-outline-variant hover:bg-surface-container px-4 rounded-sm font-bold text-sm flex items-center gap-2 transition-all cursor-pointer disabled:opacity-50"
      >
        <RefreshCw size={16} class={loading ? "animate-spin" : ""} />
        Refresh
      </button>
      {#if !authState.isAuthenticated && hasSavedNickname}
        <button
          onclick={() => showNicknameModal = true}
          class="h-12 bg-surface-highest text-on-surface border border-outline-variant hover:bg-surface-container px-4 rounded-sm font-bold text-sm flex items-center gap-2 transition-all cursor-pointer"
        >
          <UserIcon size={16} />
          Change Nickname
        </button>
      {/if}
    </div>

    <div class="flex gap-2">
      <input
        type="text"
        bind:value={joinCode}
        placeholder="Enter 8-digit Room Code..."
        class="h-12 bg-surface-highest border border-outline-variant rounded-sm px-4 text-sm text-on-surface uppercase focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all w-full md:w-60"
      />
      <button
        onclick={handleJoinByCode}
        class="h-12 bg-primary hover:bg-primary-container text-white px-6 rounded-sm font-bold text-sm transition-colors cursor-pointer"
      >
        Join Code
      </button>
    </div>
  </section>

  <!-- Lobbies Grid -->
  <section class="space-y-4">
    <h3 class="text-xl font-bold text-on-surface">Active Public Lobbies</h3>
    
    {#if loading && rooms.length === 0}
      <div class="text-center py-20 text-on-surface-variant text-sm">
        Loading active rooms...
      </div>
    {:else if rooms.length === 0}
      <div class="bg-surface-low border border-dashed border-outline-variant p-20 rounded-md text-center space-y-2">
        <Users size={48} class="mx-auto text-on-surface-variant opacity-60" />
        <h4 class="text-lg font-bold text-on-surface">No active lobbies</h4>
        <p class="text-sm text-on-surface-variant max-w-xs mx-auto">Create a lobby and invite your friends to start guessing!</p>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each rooms as room}
          <div class="bg-surface-container p-6 rounded-md flex flex-col justify-between hover:bg-surface-low transition-all relative">
            <div class="space-y-4">
              <div class="flex justify-between items-start">
                <h4 class="font-black text-lg text-on-surface tracking-tight truncate pr-6">{room.name}</h4>
                {#if room.private}
                  <span class="text-on-surface-variant mt-1"><Lock size={16} /></span>
                {/if}
              </div>

              <div class="grid grid-cols-2 gap-y-2 text-xs text-on-surface-variant">
                <div>Host: <span class="font-semibold text-on-surface">{room.host_nickname}</span></div>
                <div>Players: <span class="font-semibold text-on-surface">{room.player_count}</span></div>
                <div>Watching: <span class="font-semibold text-on-surface">{room.spectator_count || 0}</span></div>
                <div>Rounds: <span class="font-semibold text-on-surface">{room.max_rounds}</span></div>
                <div>Type: <span class="font-semibold text-on-surface capitalize">{room.game_type}</span></div>
                <div>Themes: <span class="font-semibold text-on-surface uppercase">{room.theme_type}</span></div>
                <div class="col-span-2">Status: <span class="font-semibold text-primary capitalize">{room.status}</span></div>
              </div>
            </div>

            <div class="mt-6 flex gap-2">
              <button
                onclick={() => handleJoinRoom(room.room_id, false)}
                class="flex-1 h-10 bg-primary hover:bg-primary-container text-white rounded-sm font-bold text-xs flex items-center justify-center gap-2 transition-colors cursor-pointer"
              >
                <Play size={12} />
                Play
              </button>
              <button
                onclick={() => handleJoinRoom(room.room_id, true)}
                class="h-10 bg-surface-highest border border-outline-variant hover:bg-surface-container text-on-surface rounded-sm font-bold text-xs flex items-center justify-center px-4 gap-2 transition-colors cursor-pointer"
                title="Watch as spectator"
              >
                <Eye size={14} />
                Watch
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </section>

  <!-- Create Room Modal -->
  {#if showCreateModal}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-[#09070e]/40 p-4">
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="fixed inset-0" onclick={() => showCreateModal = false}></div>
      
      <div class="bg-surface-highest max-w-lg w-full rounded-md shadow-2xl relative z-10 p-8 space-y-6 max-h-[90vh] overflow-y-auto border border-outline-variant">
        <h3 class="text-2xl font-black text-on-surface tracking-tight flex items-center gap-2">
          Create Game Room
        </h3>

        <div class="space-y-4">
          <!-- Room Name -->
          <div class="flex flex-col gap-2">
            <label for="room-name" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Lobby Name</label>
            <input
              id="room-name"
              type="text"
              bind:value={roomName}
              placeholder="Enter room name..."
              class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all"
            />
          </div>

          <!-- Rounds and Timers -->
          <div class="grid grid-cols-3 gap-4">
            <div class="flex flex-col gap-2">
              <label for="rounds" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Rounds</label>
              <input
                id="rounds"
                type="number"
                min="5"
                max="50"
                bind:value={maxRounds}
                class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all"
              />
            </div>
            <div class="flex flex-col gap-2">
              <label for="guess-time" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Guess (s)</label>
              <input
                id="guess-time"
                type="number"
                min="10"
                max="60"
                bind:value={guessTime}
                class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all"
              />
            </div>
            <div class="flex flex-col gap-2">
              <label for="reveal-time" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Reveal (s)</label>
              <input
                id="reveal-time"
                type="number"
                min="5"
                max="30"
                bind:value={revealTime}
                class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all"
              />
            </div>
          </div>

          <!-- Theme Pool and Game Type -->
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-2">
              <label for="theme-type" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1 font-black">Themes</label>
              <select
                id="theme-type"
                bind:value={themeType}
                class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
              >
                <option value="both">Openings & Endings</option>
                <option value="OP">Openings Only</option>
                <option value="ED">Endings Only</option>
              </select>
            </div>
            <div class="flex flex-col gap-2">
              <label for="game-type" class="text-[10px] uppercase font-black text-on-surface-variant tracking-widest ml-1">Game Mode</label>
              <select
                id="game-type"
                bind:value={gameType}
                class="h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
              >
                <option value="type-in">Type-In (Autocomplete)</option>
                <option value="multiple-choice">Multiple Choice</option>
              </select>
            </div>
          </div>

          <!-- Personalized Pools Checkbox -->
          {#if authState.isAuthenticated}
            <div class="flex items-center gap-3 p-3 bg-surface rounded-sm border border-outline-variant">
              <input
                id="personalized"
                type="checkbox"
                bind:checked={personalizedPool}
                class="w-5 h-5 accent-primary cursor-pointer"
              />
              <div class="flex flex-col">
                <label for="personalized" class="text-sm font-bold text-on-surface cursor-pointer">AniList Intersect Pool</label>
                <span class="text-[11px] text-on-surface-variant">Only draw songs from watched lists of synced users.</span>
              </div>
            </div>
          {/if}

          <!-- Private Lobby -->
          <div class="flex items-center gap-3 p-3 bg-surface rounded-sm border border-outline-variant">
            <input
              id="private"
              type="checkbox"
              bind:checked={isPrivate}
              class="w-5 h-5 accent-primary cursor-pointer"
            />
            <div class="flex flex-col">
              <label for="private" class="text-sm font-bold text-on-surface cursor-pointer">Private Lobby</label>
              <span class="text-[11px] text-on-surface-variant">Lobby will not appear in the public browser. Joined by room code only.</span>
            </div>
          </div>
        </div>

        <div class="flex gap-4 pt-4">
          <button
            onclick={() => showCreateModal = false}
            class="flex-1 h-12 bg-surface hover:bg-surface-low border border-outline-variant text-on-surface rounded-sm font-bold text-sm transition-colors cursor-pointer"
          >
            Cancel
          </button>
          <button
            onclick={handleCreateRoom}
            class="flex-1 h-12 bg-primary hover:bg-primary-container text-white rounded-sm font-bold text-sm transition-colors cursor-pointer"
          >
            Create Room
          </button>
        </div>
      </div>
    </div>
  {/if}

  <!-- Guest Nickname Modal -->
  {#if showNicknameModal}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-[#09070e] p-4">
      <div class="bg-surface-highest max-w-sm w-full rounded-md shadow-2xl relative z-10 p-8 space-y-6 border border-outline-variant text-center">
        <div class="space-y-2">
          <h3 class="text-2xl font-black text-on-surface tracking-tight">
            {hasSavedNickname ? "Change Nickname" : "Enter Guest Nickname"}
          </h3>
          <p class="text-xs text-on-surface-variant leading-relaxed">
            Set a temporary nickname to join or create lobbies. To save stats, earn badges, and gain XP, please log in.
          </p>
        </div>
        
        <div class="space-y-4">
          <input
            id="guest-nick-modal"
            type="text"
            bind:value={guestNickname}
            placeholder="Type guest nickname..."
            onkeydown={(e) => e.key === "Enter" && saveNickname()}
            class="w-full h-12 bg-surface border border-outline-variant rounded-sm px-4 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all text-center font-bold"
          />
          
          {#if errorMsg && !guestNickname.trim()}
            <p class="text-red-500 text-xs font-semibold">{errorMsg}</p>
          {/if}

          <div class="flex gap-4">
            {#if hasSavedNickname}
              <button
                onclick={cancelNicknameChange}
                class="flex-1 h-12 bg-surface hover:bg-surface-low border border-outline-variant text-on-surface rounded-sm font-bold text-sm transition-colors cursor-pointer"
              >
                Cancel
              </button>
            {/if}
            <button
              onclick={saveNickname}
              class="flex-1 h-12 bg-primary hover:bg-primary-container text-white rounded-sm font-bold text-sm transition-colors cursor-pointer {!hasSavedNickname ? 'w-full' : ''}"
            >
              {hasSavedNickname ? "Save" : "Enter Lobby"}
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
</style>
