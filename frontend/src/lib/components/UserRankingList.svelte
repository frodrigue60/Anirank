<script lang="ts">
  import type { RankingUser } from "$lib/types/user";
  import Users from "lucide-svelte/icons/users";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";

  let { users = [] as RankingUser[], startIndex = 0, sort = "xp" } = $props();

  let scrollTrigger = $state<HTMLElement>();

  function formatNumber(num: number | undefined) {
    return (num ?? 0).toLocaleString();
  }
</script>

<div
  class="bg-surface-container border border-white/5 rounded-md overflow-x-auto mb-8"
>
  <table class="w-full border-collapse min-w-[800px]">
    <thead>
      <tr
        class="border-b border-white/5 text-[10px] bg-surface-highest font-black uppercase tracking-widest text-on-surface-variant"
      >
        <th class="w-20 py-4 pl-8 pr-2 text-center font-black">Rank</th>
        <th class="py-4 px-2 text-left font-black">User</th>
        <th class="w-[100px] py-4 px-2 text-center font-black">Level</th>
        <th class="w-[120px] py-4 px-2 text-center font-black">XP</th>
        <th class="w-[120px] py-4 px-2 text-center font-black">Ratings</th>
        <th class="w-[120px] py-4 pl-2 pr-8 text-center font-black">Comments</th
        >
      </tr>
    </thead>

    <tbody class="divide-y divide-white/5">
      {#if users.length > 0}
        {#each users as user, index}
          {@const rank = startIndex + index + 1}
          <tr class="hover:bg-surface-highest transition-colors">
            <td class="py-5 pl-8 pr-2 text-center">
              <span
                class="text-2xl font-black leading-none {rank <= 3
                  ? 'text-primary'
                  : 'text-white/90'}"
              >
                {rank.toString().padStart(2, "0")}
              </span>
            </td>

            <td class="py-5 px-2">
              <div class="flex items-center gap-4 min-w-0">
                <div
                  class="w-12 h-12 rounded-full overflow-hidden shrink-0 shadow-lg border border-white/10"
                >
                  <OptimizedImage
                    src={user.avatar_url}
                    sources={user.avatar_sources}
                    alt={user.name}
                    class="w-full h-full object-cover"
                    sizes="48px"
                  />
                </div>
                <div class="min-w-0 flex flex-col">
                  <a
                    href={`/users/${user.slug}`}
                    class="text-lg font-bold text-on-surface hover:text-primary transition-colors truncate"
                  >
                    {user.name}
                  </a>
                  <span class="text-on-surface-variant text-xs"
                    >Joined {new Date(
                      user.created_at,
                    ).toLocaleDateString()}</span
                  >
                </div>
              </div>
            </td>

            <td class="py-5 px-2 text-center">
              <span
                class="px-3 py-1 bg-primary/10 text-primary border border-primary/20 rounded-lg font-black text-sm"
              >
                LVL {user.level}
              </span>
            </td>

            <td class="py-5 px-2 text-center">
              <span class="text-on-surface font-bold"
                >{formatNumber(user.xp)}</span
              >
            </td>

            <td class="py-5 px-2 text-center">
              <span class="text-on-surface-variant"
                >{formatNumber(user.ratings_count)}</span
              >
            </td>

            <td class="py-5 pl-2 pr-8 text-center">
              <span class="text-on-surface-variant"
                >{formatNumber(user.comments_count)}</span
              >
            </td>
          </tr>
        {/each}
      {:else}
        <tr>
          <td colspan="6" class="py-20 text-on-surface-variant text-center">
            <div class="flex flex-col items-center justify-center">
              <Users size={60} class="mb-4" />

              <p class="text-lg font-bold">No users found</p>
            </div>
          </td>
        </tr>
      {/if}
    </tbody>
  </table>
</div>
