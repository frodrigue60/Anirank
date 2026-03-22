<script lang="ts">
  import type { RankingUser } from "$lib/types/user";

  let {
    users = [] as RankingUser[],
    startIndex = 0,
    sort = "xp"
  } = $props();

  let scrollTrigger = $state<HTMLElement>();

  function formatNumber(num: number) {
    return num.toLocaleString();
  }
</script>

<div class="bg-surface-dark/30 border border-white/5 rounded-2xl overflow-hidden mb-8">
  <div class="grid grid-cols-[80px_1fr_100px_120px_120px_120px] gap-4 px-8 py-4 border-b border-white/5 text-[10px] font-black uppercase tracking-widest text-white/30 bg-surface-darker/50">
    <div class="text-center">Rank</div>
    <div>User</div>
    <div class="text-center">Level</div>
    <div class="text-center">XP</div>
    <div class="text-center">Ratings</div>
    <div class="text-center">Comments</div>
  </div>

  <div class="flex flex-col">
    {#if users.length > 0}
      {#each users as user, index}
        {@const rank = startIndex + index + 1}
        <div class="ranking-row grid grid-cols-[80px_1fr_100px_120px_120px_120px] gap-4 px-8 py-5 items-center transition-colors border-b border-white/5 last:border-0 hover:bg-white/5">
          <div class="text-center">
            <span class="text-2xl font-black leading-none {rank <= 3 ? 'text-primary' : 'text-white/90'}">
              {rank.toString().padStart(2, "0")}
            </span>
          </div>
          
          <div class="flex items-center gap-4 min-w-0">
            <div class="w-12 h-12 rounded-full overflow-hidden shrink-0 shadow-lg border border-white/10">
              <img
                alt={user.name}
                title={user.name}
                class="w-full h-full object-cover"
                src={user.avatar_url || "/default-avatar.png"}
              />
            </div>
            <div class="min-w-0 flex flex-col">
              <a href={`/users/${user.slug}`} class="text-lg font-bold text-white hover:text-primary transition-colors truncate">
                {user.name}
              </a>
              <span class="text-white/40 text-xs">Joined {new Date(user.created_at).toLocaleDateString()}</span>
            </div>
          </div>

          <div class="text-center">
            <span class="px-3 py-1 bg-primary/10 text-primary border border-primary/20 rounded-lg font-black text-sm">
              LVL {user.level}
            </span>
          </div>

          <div class="text-center">
            <span class="text-white font-bold">{formatNumber(user.xp)}</span>
          </div>

          <div class="text-center">
            <span class="text-white/80">{formatNumber(user.ratings_count)}</span>
          </div>

          <div class="text-center">
            <span class="text-white/80">{formatNumber(user.comments_count)}</span>
          </div>
        </div>
      {/each}

    {:else}
      <div class="flex flex-col items-center justify-center py-20 text-white/30">
        <span class="material-symbols-outlined text-6xl mb-4">group_off</span>
        <p class="text-lg font-bold">No users found</p>
      </div>
    {/if}
  </div>
</div>
