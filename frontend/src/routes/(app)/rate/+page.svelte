<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { authState } from "$lib/state/auth.svelte";
  import { configState } from "$lib/state/config.svelte";
  import api from "$lib/api";
  import SEO from "$lib/components/SEO.svelte";
  import Plus from "lucide-svelte/icons/plus";
  import Lock from "lucide-svelte/icons/lock";
  import Users from "lucide-svelte/icons/users";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import Eye from "lucide-svelte/icons/eye";
  import UserIcon from "lucide-svelte/icons/user";
  import Star from "lucide-svelte/icons/star";
  import {
    defaultRateConfig,
    type QueueMode,
    type RevealMode,
    type AutoAdvance,
    type SourceMode,
    type SeasonalPoolThemeType,
  } from "$lib/rate/room-state";

  interface RoomInfo {
    room_id: string;
    name: string;
    host_nickname: string;
    player_count: number;
    spectator_count: number;
    status: string;
    private: boolean;
    queue_mode: string;
    reveal_mode: string;
    queue_length: number;
    source_mode?: string;
    pool_year?: string;
    pool_season?: string;
    pool_theme_type?: string;
    pool_format?: string;
  }

  let rooms = $state<RoomInfo[]>([]);
  let loading = $state(false);
  let showCreateModal = $state(false);
  let showNicknameModal = $state(false);

  let guestNickname = $state("");
  let deviceId = $state("");
  let hasSavedNickname = $state(false);

  let roomName = $state("");
  let sourceMode = $state<SourceMode>("manual");
  let queueMode = $state<QueueMode>("host_only");
  let queueLimitPerUser = $state(3);
  let revealMode = $state<RevealMode>("blind");
  let maxPlayers = $state(16);
  let autoAdvance = $state<AutoAdvance>("never");
  let voteSkip = $state(false);
  let isPrivate = $state(false);
  let poolYear = $state("");
  let poolSeason = $state("");
  let poolThemeType = $state<SeasonalPoolThemeType>("all");
  let poolFormat = $state("all");
  let poolLimit = $state<number | "">("");

  let joinCode = $state("");
  let errorMsg = $state("");

  let sortedYears = $derived(
    [...configState.years].sort((a, b) => Number(b.slug) - Number(a.slug) || b.name.localeCompare(a.name))
  );
  let formatOptions = $derived(
    configState.formats.map((f) => ({ value: f.slug, label: f.name })),
  );

  function generateUUID(): string {
    if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
      return crypto.randomUUID();
    }
    return Math.random().toString(36).substring(2, 15) + Date.now().toString(36);
  }

  onMount(() => {
    if (typeof window !== "undefined") {
      let storedDeviceId = localStorage.getItem("rate_device_id");
      if (!storedDeviceId) {
        storedDeviceId = generateUUID();
        localStorage.setItem("rate_device_id", storedDeviceId);
      }
      deviceId = storedDeviceId;
      guestNickname = localStorage.getItem("rate_nickname") || "";
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
      const response = await api.get("/rate/rooms");
      if (response.data?.success) {
        rooms = response.data.data || [];
      }
    } catch (e) {
      console.error("Error loading rate lobbies", e);
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
    localStorage.setItem("rate_nickname", guestNickname.trim());
    hasSavedNickname = true;
    showNicknameModal = false;
    errorMsg = "";
    return true;
  }

  async function handleCreateRoom() {
    errorMsg = "";
    const nickname = authState.isAuthenticated ? authState.user?.name : guestNickname.trim();
    if (!nickname) {
      errorMsg = "Please enter a nickname first.";
      showNicknameModal = true;
      return;
    }
    if (!authState.isAuthenticated) saveNickname();
    if (!roomName.trim()) roomName = `${nickname}'s Rate Party`;

    if (sourceMode === "seasonal_pool") {
      if (!poolYear || !poolSeason) {
        errorMsg = "Pick a year and season for the seasonal pool.";
        return;
      }
    }

    try {
      const defaults = defaultRateConfig();
      const payload = {
        config: {
          ...defaults,
          name: roomName.trim(),
          private: isPrivate,
          queue_mode: sourceMode === "seasonal_pool" ? "disabled" : queueMode,
          queue_limit_per_user: queueLimitPerUser,
          reveal_mode: revealMode,
          max_players: maxPlayers,
          auto_advance: autoAdvance,
          vote_skip: voteSkip,
          source_mode: sourceMode,
          ...(sourceMode === "seasonal_pool"
            ? {
                pool_year: poolYear,
                pool_season: poolSeason,
                pool_theme_type: poolThemeType,
                pool_format: poolFormat,
                pool_limit: typeof poolLimit === "number" && poolLimit > 0 ? poolLimit : 0,
              }
            : {}),
        },
        guest_nickname: nickname,
        guest_device_id: deviceId,
      };
      const response = await api.post("/rate/rooms", payload);
      if (response.data?.success) {
        goto(`/rate/${response.data.data.room_id}`);
      }
    } catch (e: any) {
      errorMsg = e.response?.data?.message || "Failed to create lobby.";
    }
  }

  function handleJoinRoom(roomId: string, spectator = false) {
    const nickname = authState.isAuthenticated ? authState.user?.name : guestNickname.trim();
    if (!nickname) {
      errorMsg = "Please enter a nickname first.";
      showNicknameModal = true;
      return;
    }
    if (!authState.isAuthenticated) saveNickname();
    goto(`/rate/${roomId}${spectator ? "?spectator=true" : ""}`);
  }

  function handleJoinByCode() {
    if (!joinCode.trim()) return;
    handleJoinRoom(joinCode.trim().toUpperCase());
  }

  function queueModeLabel(mode: string) {
    if (mode === "everyone") return "Open queue";
    if (mode === "disabled") return "No queue";
    return "Host queue";
  }

  function sourceLabel(room: RoomInfo) {
    if (room.source_mode === "seasonal_pool") {
      const type = room.pool_theme_type && room.pool_theme_type !== "all" ? ` · ${room.pool_theme_type}` : "";
      const format = room.pool_format && room.pool_format !== "all" ? ` · ${room.pool_format}` : "";
      return `${room.pool_season || "season"} ${room.pool_year || ""}${type}${format}`.trim();
    }
    return queueModeLabel(room.queue_mode);
  }
</script>

<SEO
  title="Rate Party"
  description="Rate anime themes together in real-time. Join a group rating session and score songs with your friends."
/>

<main class="max-w-[1440px] mx-auto px-6 py-10 space-y-8">
  <header class="space-y-4">
    <div class="flex items-center gap-3">
      <span class="w-1.5 h-8 bg-primary rounded-full"></span>
      <h1 class="text-4xl font-black tracking-tight text-on-surface">RATE PARTY</h1>
    </div>
    <p class="text-on-surface-variant text-base max-w-xl leading-relaxed">
      Listen together and rate anime themes using your score preference. Scores always save to your
      global ranking. Create a lobby or join with a room code.
    </p>
    {#if !authState.isAuthenticated && guestNickname}
      <div class="flex items-center gap-2 text-xs text-on-surface-variant">
        <span>Watching as guest: <strong class="text-on-surface">{guestNickname}</strong></span>
        <button
          onclick={() => (showNicknameModal = true)}
          class="text-primary hover:underline font-bold cursor-pointer"
        >
          (Change Nickname)
        </button>
      </div>
    {/if}
  </header>

  {#if errorMsg}
    <div class="p-4 bg-red-50 text-red-700 text-sm rounded-sm">{errorMsg}</div>
  {/if}

  <section class="flex flex-col md:flex-row gap-4 justify-between items-stretch md:items-center">
    <div class="flex gap-3 flex-wrap">
      <button
        onclick={() => (showCreateModal = true)}
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
    </div>

    <div class="flex gap-2">
      <input
        type="text"
        bind:value={joinCode}
        placeholder="Enter Room Code..."
        class="h-12 bg-surface-highest border border-outline-variant rounded-sm px-4 text-sm text-on-surface uppercase focus:outline-hidden focus:border-primary/50 w-full md:w-60"
      />
      <button
        onclick={handleJoinByCode}
        class="h-12 bg-primary hover:bg-primary-container text-white px-6 rounded-sm font-bold text-sm transition-colors cursor-pointer"
      >
        Join
      </button>
    </div>
  </section>

  <section class="space-y-4">
    <h2 class="text-xl font-bold text-on-surface">Active Public Lobbies</h2>
    {#if loading && rooms.length === 0}
      <p class="text-on-surface-variant text-sm">Loading lobbies…</p>
    {:else if rooms.length === 0}
      <div class="bg-surface-low rounded-sm p-10 text-center space-y-2">
        <Star size={32} class="mx-auto text-primary" />
        <p class="text-on-surface font-bold">No public rate parties right now</p>
        <p class="text-on-surface-variant text-sm">Create one and invite your friends.</p>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {#each rooms as room (room.room_id)}
          <article class="bg-surface-container rounded-sm p-5 space-y-4 hover:bg-surface-low transition-colors">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h3 class="font-bold text-on-surface text-lg leading-tight">{room.name}</h3>
                <p class="text-xs text-on-surface-variant mt-1">
                  Host: {room.host_nickname} · {room.room_id}
                </p>
              </div>
              {#if room.private}
                <Lock size={16} class="text-on-surface-variant shrink-0" />
              {/if}
            </div>
            <div class="flex flex-wrap gap-2 text-xs">
              <span class="bg-surface-highest text-on-surface-variant px-2 py-1 rounded-sm capitalize">{room.status}</span>
              <span class="bg-surface-highest text-on-surface-variant px-2 py-1 rounded-sm">{sourceLabel(room)}</span>
              <span class="bg-surface-highest text-on-surface-variant px-2 py-1 rounded-sm">{room.reveal_mode === "live" ? "Live scores" : "Blind scores"}</span>
            </div>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3 text-sm text-on-surface-variant">
                <span class="flex items-center gap-1"><Users size={14} /> {room.player_count}</span>
                <span class="flex items-center gap-1"><Eye size={14} /> {room.spectator_count}</span>
                <span>Queue: {room.queue_length}</span>
              </div>
              <div class="flex gap-2">
                <button
                  onclick={() => handleJoinRoom(room.room_id, true)}
                  class="h-9 px-3 text-xs font-bold text-primary bg-surface-highest rounded-sm cursor-pointer"
                  aria-label="Spectate room"
                >
                  Spectate
                </button>
                <button
                  onclick={() => handleJoinRoom(room.room_id)}
                  class="h-9 px-4 text-xs font-bold text-white bg-primary rounded-sm cursor-pointer"
                >
                  Join
                </button>
              </div>
            </div>
          </article>
        {/each}
      </div>
    {/if}
  </section>
</main>

{#if showCreateModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
    <div class="bg-surface-highest w-full max-w-lg rounded-sm p-6 space-y-5 max-h-[90vh] overflow-y-auto">
      <h2 class="text-xl font-black text-on-surface">Create Rate Party</h2>

      <label class="block space-y-1">
        <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Room name</span>
        <input bind:value={roomName} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm" placeholder="My Rate Party" />
      </label>

      <label class="block space-y-1">
        <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Theme source</span>
        <select bind:value={sourceMode} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm">
          <option value="manual">Manual (search &amp; queue)</option>
          <option value="seasonal_pool">Seasonal pool</option>
        </select>
      </label>

      {#if sourceMode === "seasonal_pool"}
        <p class="text-xs text-on-surface-variant leading-relaxed">
          Themes load from the selected season on start. Manual adds stay locked until the host switches back to manual in the lobby.
        </p>
        <div class="grid grid-cols-2 gap-3">
          <label class="block space-y-1">
            <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Year</span>
            <select bind:value={poolYear} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm">
              <option value="">Select year</option>
              {#each sortedYears as y}
                <option value={y.slug}>{y.name}</option>
              {/each}
            </select>
          </label>
          <label class="block space-y-1">
            <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Season</span>
            <select bind:value={poolSeason} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm">
              <option value="">Select season</option>
              {#each configState.seasons as s}
                <option value={s.slug}>{s.name}</option>
              {/each}
            </select>
          </label>
        </div>
        <label class="block space-y-1">
          <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Theme type</span>
          <select bind:value={poolThemeType} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm">
            <option value="all">ALL</option>
            <option value="OP">OP</option>
            <option value="ED">ED</option>
          </select>
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Series format</span>
          <select bind:value={poolFormat} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm">
            <option value="all">All formats</option>
            {#each formatOptions as fmt}
              <option value={fmt.value}>{fmt.label}</option>
            {/each}
          </select>
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Pool size (optional)</span>
          <input
            type="number"
            min="5"
            max="50"
            bind:value={poolLimit}
            placeholder="All themes"
            class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm"
          />
          <span class="text-[11px] text-on-surface-variant">Leave empty to load the full season. Cap is 5–50 when set.</span>
        </label>
      {:else}
        <label class="block space-y-1">
          <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Queue mode</span>
          <select bind:value={queueMode} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm">
            <option value="host_only">Host only</option>
            <option value="everyone">Everyone (auth)</option>
            <option value="disabled">Disabled</option>
          </select>
        </label>

        {#if queueMode === "everyone"}
          <label class="block space-y-1">
            <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Songs per user</span>
            <input type="number" min="1" max="10" bind:value={queueLimitPerUser} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm" />
          </label>
        {/if}
      {/if}

      <label class="block space-y-1">
        <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Reveal mode</span>
        <select bind:value={revealMode} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm">
          <option value="blind">Blind (avg only + own score)</option>
          <option value="live">Live (avg + everyone's scores)</option>
        </select>
      </label>

      <label class="block space-y-1">
        <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Auto advance</span>
        <select bind:value={autoAdvance} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm">
          <option value="never">Host advances manually</option>
          <option value="all_rated">When everyone has rated</option>
        </select>
      </label>

      <label class="flex items-start gap-2 text-sm text-on-surface cursor-pointer">
        <input type="checkbox" bind:checked={voteSkip} class="mt-0.5 rounded-sm" />
        <span>
          <span class="font-semibold">Vote skip</span>
          <span class="block text-xs text-on-surface-variant mt-0.5">
            Majority of online players can vote to skip the current song (host can always Next).
          </span>
        </span>
      </label>

      <label class="block space-y-1">
        <span class="text-xs font-bold text-on-surface-variant uppercase tracking-wide">Max players</span>
        <input type="number" min="2" max="32" bind:value={maxPlayers} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm" />
      </label>

      <label class="flex items-center gap-2 text-sm text-on-surface cursor-pointer">
        <input type="checkbox" bind:checked={isPrivate} class="rounded-sm" />
        Private room (join by code only)
      </label>

      <div class="flex gap-3 justify-end pt-2">
        <button onclick={() => (showCreateModal = false)} class="h-11 px-4 rounded-sm text-sm font-bold text-on-surface-variant cursor-pointer">Cancel</button>
        <button onclick={handleCreateRoom} class="h-11 px-5 rounded-sm text-sm font-bold text-white bg-primary cursor-pointer">Create</button>
      </div>
    </div>
  </div>
{/if}

{#if showNicknameModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
    <div class="bg-surface-highest w-full max-w-sm rounded-sm p-6 space-y-4">
      <div class="flex items-center gap-2">
        <UserIcon size={20} class="text-primary" />
        <h2 class="text-lg font-bold text-on-surface">Guest nickname</h2>
      </div>
      <p class="text-sm text-on-surface-variant">Guests can watch; login is required to rate songs.</p>
      <input bind:value={guestNickname} class="w-full h-11 bg-surface border border-outline-variant rounded-sm px-3 text-sm" placeholder="Nickname" />
      <div class="flex gap-3 justify-end">
        <button onclick={() => (showNicknameModal = false)} class="h-10 px-4 text-sm font-bold text-on-surface-variant cursor-pointer">Cancel</button>
        <button onclick={saveNickname} class="h-10 px-4 text-sm font-bold text-white bg-primary rounded-sm cursor-pointer">Save</button>
      </div>
    </div>
  </div>
{/if}
