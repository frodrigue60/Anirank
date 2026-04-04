<script lang="ts">
  import { fade } from "svelte/transition";
  import api from "$lib/api";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";

  let { data } = $props();

  // svelte-ignore state_referenced_locally
  let following = $state(data.following || []);
  // svelte-ignore state_referenced_locally
  let currentPage = $state(data.pagination?.current_page || 1);
  // svelte-ignore state_referenced_locally
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
        params: { page: nextPage },
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
      <h2 class="text-2xl font-black text-on-surface tracking-tighter">Following</h2>
      <p class="text-on-surface-variant text-sm mt-1">Users followed by this profile</p>
    </div>
    <div class="px-4 py-2 bg-surface-highest rounded-lg border border-on-surface-variant/10">
      <span class="text-primary font-bold">{data.pagination.total}</span>
      <span class="text-on-surface-variant text-sm ml-1 text-[10px] font-black uppercase tracking-widest">Following</span>
    </div>
  </div>

  {#if following.length > 0}
    <div
      class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4"
    >
      {#each following as user}
        <a
          href={`/users/${user.slug}`}
          class="group bg-surface-container rounded-2xl p-4 border border-on-surface-variant/10 hover:border-primary/30 transition-all hover:bg-surface-highest flex items-center gap-4 shadow-sm"
        >
          <div
            class="size-14 rounded-xl overflow-hidden bg-surface-low border border-on-surface-variant/10 group-hover:border-primary/50 transition-colors shadow-inner"
          >
            <img
              src={user.avatar_url || "/images/placeholders/default.jpg"}
              alt={user.name}
              class="w-full h-full object-cover"
            />
          </div>
          <div class="flex-1 min-w-0">
            <h3
              class="font-bold text-on-surface group-hover:text-primary transition-colors truncate"
            >
              {user.name}
            </h3>
            <p class="text-on-surface-variant text-xs truncate font-medium">@{user.slug}</p>
          </div>
          <span
            class="material-symbols-outlined text-on-surface-variant/30 group-hover:text-primary transition-colors"
          >
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
    <div
      class="py-20 text-center bg-surface-container rounded-3xl border border-on-surface-variant/10"
    >
      <div
        class="size-20 bg-surface-highest rounded-full flex items-center justify-center mx-auto mb-4 text-on-surface-variant/40"
      >
        <span class="material-symbols-outlined text-4xl">person_search</span>
      </div>
      <h3 class="text-xl font-bold text-on-surface tracking-tight">Not following anyone</h3>
      <p class="text-on-surface-variant max-w-xs mx-auto mt-2 font-medium">
        When this user follows someone, they will appear here.
      </p>
    </div>
  {/if}
</div>
