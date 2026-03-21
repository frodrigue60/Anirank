<script lang="ts">
  import { fade } from "svelte/transition";
  import api from "$lib/api";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";

  let { data } = $props();

  let following = $state(data.following || []);
  let currentPage = $state(data.pagination?.current_page || 1);
  let lastPage = $state(data.pagination?.last_page || 1);
  let loading = $state(false);

  // Sync if data changes (e.g. navigation)
  $effect(() => {
    if (data.following && currentPage === 1) {
      following = data.following;
      currentPage = data.pagination.current_page;
      lastPage = data.pagination.last_page;
    }
  });

  async function loadMore() {
    if (loading || currentPage >= lastPage) return;

    loading = true;
    try {
      const nextPage = currentPage + 1;
      const response = await api.get(`/users/${data.profile.slug}/following`, {
        params: { page: nextPage }
      });

      if (response.data.data) {
        following = [...following, ...response.data.data];
        currentPage = response.data.current_page;
        lastPage = response.data.last_page;
      }
    } catch (e) {
      console.error("Error loading more following users", e);
    } finally {
      loading = false;
    }
  }
</script>

<div class="space-y-8" in:fade>
  <div class="flex items-center justify-between">
    <div>
      <h2 class="text-2xl font-black text-white tracking-tighter">Following</h2>
      <p class="text-slate-400 text-sm mt-1">Users followed by this profile</p>
    </div>
    <div class="px-4 py-2 bg-white/5 rounded-lg border border-white/10">
      <span class="text-primary font-bold">{data.pagination.total}</span>
      <span class="text-slate-400 text-sm ml-1">Following</span>
    </div>
  </div>

  {#if following.length > 0}
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      {#each following as user}
        <a
          href={`/users/${user.slug}`}
          class="group bg-white/5 rounded-2xl p-4 border border-white/5 hover:border-primary/30 transition-all hover:bg-white/10 flex items-center gap-4"
        >
          <div class="size-14 rounded-xl overflow-hidden bg-background-dark border border-white/10 group-hover:border-primary/50 transition-colors">
            <img
              src={user.avatar_url || "/default-avatar.png"}
              alt={user.name}
              class="w-full h-full object-cover"
            />
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="font-bold text-white group-hover:text-primary transition-colors truncate">
              {user.name}
            </h3>
            <p class="text-slate-500 text-xs truncate">@{user.slug}</p>
          </div>
          <span class="material-symbols-outlined text-slate-600 group-hover:text-primary transition-colors">
            chevron_right
          </span>
        </a>
      {/each}
    </div>

    <InfiniteScroll
      hasMore={currentPage < lastPage}
      {loading}
      onLoadMore={loadMore}
    />
  {:else}
    <div class="py-20 text-center bg-white/5 rounded-3xl border border-dashed border-white/10">
      <div class="size-20 bg-white/5 rounded-full flex items-center justify-center mx-auto mb-4 text-slate-500">
        <span class="material-symbols-outlined text-4xl">person_search</span>
      </div>
      <h3 class="text-xl font-bold text-white">Not following anyone</h3>
      <p class="text-slate-400 max-w-xs mx-auto mt-2">
        When this user follows someone, they will appear here.
      </p>
    </div>
  {/if}
</div>
