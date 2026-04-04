<script lang="ts">
  import { fade, fly } from "svelte/transition";
  import api from "$lib/api";
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let notifications = $state(data.notifications);
  // svelte-ignore state_referenced_locally
  let unreadCount = $state(data.unreadCount);
  // svelte-ignore state_referenced_locally
  let filterType = $state(data.filterType || "");

  $effect(() => {
    notifications = data.notifications;
    unreadCount = data.unreadCount;
    filterType = data.filterType || "";
  });

  async function markAsRead(id: string) {
    try {
      await api.put(`/notifications/${id}/read`);
      notifications = notifications.map((n: any) =>
        n.id === id ? { ...n, read_at: new Date().toISOString() } : n,
      );
      unreadCount = Math.max(0, unreadCount - 1);
    } catch (error) {
      console.error("Error marking notification as read:", error);
    }
  }

  async function markAllAsRead() {
    try {
      await api.post("/notifications/read-all");
      notifications = notifications.map((n: any) => ({
        ...n,
        read_at: new Date().toISOString(),
      }));
      unreadCount = 0;
    } catch (error) {
      console.error("Error marking all as read:", error);
    }
  }

  function formatTime(dateStr: string) {
    const date = new Date(dateStr);
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diffInSeconds < 60) return "just now";

    const diffInMinutes = Math.floor(diffInSeconds / 60);
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`;

    const diffInHours = Math.floor(diffInMinutes / 60);
    if (diffInHours < 24) return `${diffInHours}h ago`;

    const diffInDays = Math.floor(diffInHours / 24);
    if (diffInDays === 1) return "yesterday";
    if (diffInDays < 7) return `${diffInDays} days ago`;

    return date.toLocaleDateString();
  }

  function getNotificationIcon(type: string) {
    switch (type) {
      case "follow":
        return "person_add";
      case "reply":
        return "reply";
      case "like":
        return "favorite";
      default:
        return "notifications";
    }
  }

  function getNotificationColor(type: string) {
    switch (type) {
      case "follow":
        return "text-blue-400 bg-blue-400/10";
      case "reply":
        return "text-green-400 bg-green-400/10";
      case "like":
        return "text-red-400 bg-red-400/10";
      default:
        return "text-primary bg-primary/10";
    }
  }

  const filters = [
    { label: "All", value: "" },
    { label: "Airing", value: "airing", disabled: true },
    { label: "Activity", value: "activity" },
    { label: "Forum", value: "forum", disabled: true },
    { label: "Follows", value: "follow" },
    { label: "Media", value: "media", disabled: true },
  ];

  function setFilter(value: string) {
    const url = new URL(page.url);
    if (value) {
      url.searchParams.set("type", value);
    } else {
      url.searchParams.delete("type");
    }
    url.searchParams.delete("page");
    goto(url.toString());
  }
</script>

<div class="min-h-screen my-8">
  <div class="max-w-6xl mx-auto px-6">
    <div class="grid grid-cols-12 gap-8">
      <!-- Sidebar -->
      <div
        class="shrink-0 space-y-6 h-fit bg-surface-container p-6 rounded-2xl col-span-12 lg:col-span-3"
      >
        <div>
          <div class="flex items-center justify-between mb-4 px-2">
            <h2
              class="text-sm font-bold text-on-surface-variant uppercase tracking-wider"
            >
              Notifications
            </h2>
            <button
              class="text-on-surface-variant hover:text-on-surface transition-colors"
            >
              <span class="material-symbols-outlined text-lg">settings</span>
            </button>
          </div>
          <div class="space-y-1">
            {#each filters as filter}
              <button
                onclick={() => !filter.disabled && setFilter(filter.value)}
                class="w-full text-left px-4 py-2.5 rounded-xl transition-all font-medium flex items-center justify-between {filterType ===
                filter.value
                  ? 'bg-surface-highest text-on-surface shadow-sm'
                  : 'text-on-surface-variant hover:bg-surface-highest hover:text-on-surface'} {filter.disabled
                  ? 'opacity-30 cursor-not-allowed'
                  : 'cursor-pointer'}"
              >
                <span>{filter.label}</span>
                {#if filter.value === "" && unreadCount > 0}
                  <span
                    class="px-2 py-0.5 rounded-full bg-primary text-[10px] text-on-surface font-bold"
                    >{unreadCount}</span
                  >
                {/if}
              </button>
            {/each}
          </div>
        </div>

        {#if unreadCount > 0}
          <button
            onclick={markAllAsRead}
            class="w-full py-3 rounded-xl bg-primary text-on-surface font-bold text-sm shadow-lg shadow-primary/20 hover:scale-[1.02] active:scale-95 transition-all"
          >
            Mark all as read
          </button>
        {/if}
      </div>

      <!-- Main Content -->
      <div
        class="flex-1 min-w-0 bg-surface-container rounded-2xl p-6 col-span-12 lg:col-span-9"
      >
        {#if notifications.length === 0}
          <div
            class="p-20 rounded-3xl text-center border border-white/5 bg-white/2"
            in:fade
          >
            <div
              class="w-20 h-20 rounded-2xl bg-white/5 flex items-center justify-center mx-auto mb-6 text-on-surface-variant"
            >
              <span class="material-symbols-outlined text-5xl"
                >notifications_off</span
              >
            </div>
            <h3 class="text-2xl font-bold text-on-surface mb-2">
              No notifications found
            </h3>
            <p class="text-on-surface-variant max-w-xs mx-auto">
              {filterType
                ? `You don't have any notifications for this filter.`
                : "When people interact with you, you'll see it here."}
            </p>
          </div>
        {:else}
          <div class="space-y-3">
            {#each notifications as notification, i (notification.id)}
              <div
                class="group relative p-5 rounded-2xl border transition-all {notification.read_at
                  ? 'bg-white/2 border-white/5 opacity-80'
                  : 'bg-white/5 border-primary/20 shadow-lg shadow-primary/5'}"
                in:fly={{ y: 20, delay: i * 30 }}
              >
                <div class="flex gap-5 items-center">
                  <div
                    class="shrink-0 w-12 h-12 rounded-xl flex items-center justify-center overflow-hidden {getNotificationColor(
                      notification.type,
                    )}"
                  >
                    {#if notification.type === "follow"}
                      <img
                        src={notification.data.follower_avatar ||
                          "/images/placeholders/default.jpg"}
                        alt=""
                        class="w-full h-full object-cover"
                      />
                    {:else if notification.type === "reply" || notification.type === "comment_reply"}
                      <img
                        src={notification.data.replied_by_avatar ||
                          "/images/placeholders/default.jpg"}
                        alt=""
                        class="w-full h-full object-cover"
                      />
                    {:else}
                      <span class="material-symbols-outlined text-2xl"
                        >{getNotificationIcon(notification.type)}</span
                      >
                    {/if}
                  </div>

                  <div class="flex-1 min-w-0">
                    <div class="flex items-center justify-between gap-4">
                      <p class="text-slate-200 text-sm leading-relaxed">
                        {#if notification.type === "follow"}
                          <a
                            href="/users/{notification.data.follower_slug}"
                            class="font-bold text-white hover:text-primary transition-colors"
                            >{notification.data.follower_name}</a
                          > started following you
                        {:else if notification.type === "reply" || notification.type === "comment_reply"}
                          <a
                            href="/users/{notification.data.replied_by_slug ||
                              '#'}"
                            class="font-bold text-white hover:text-primary transition-colors"
                            >{notification.data.replied_by_name || "Someone"}</a
                          >
                          replied to your comment
                          {#if notification.data.anime_name}
                            on <a
                              href="/songs/{notification.data
                                .anime_slug}/{notification.data
                                .song_slug}#comment-{notification.data
                                ?.comment_uuid || notification.data.comment_id}"
                              class="text-white hover:text-primary transition-colors font-bold"
                              >{notification.data.anime_name}
                              {notification.data.song_slug}</a
                            >
                          {/if}
                        {:else if notification.type === "like"}
                          Someone liked your content
                        {:else}
                          New notification
                        {/if}
                      </p>
                      <span class="text-xs text-slate-500 whitespace-nowrap"
                        >{formatTime(notification.created_at)}</span
                      >
                    </div>

                    {#if !notification.read_at}
                      <button
                        onclick={() => markAsRead(notification.id)}
                        class="text-[11px] text-primary hover:text-primary-light font-medium mt-1 transition-colors"
                      >
                        Mark as read
                      </button>
                    {/if}
                  </div>

                  {#if !notification.read_at}
                    <div
                      class="w-2.5 h-2.5 rounded-full bg-primary shadow-sm shadow-primary"
                    ></div>
                  {/if}

                  {#if notification.data.anime_cover}
                    <div
                      class="shrink-0 w-16 h-20 rounded-lg overflow-hidden border border-white/10 ml-2"
                    >
                      <img
                        src={notification.data.anime_cover}
                        alt=""
                        class="w-full h-full object-cover"
                      />
                    </div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>

          {#if data.lastPage > 1}
            <div class="flex justify-center mt-10 gap-2">
              {#each Array(data.lastPage) as _, i}
                <a
                  href="?page={i + 1}{filterType ? `&type=${filterType}` : ''}"
                  class="w-10 h-10 rounded-xl flex items-center justify-center font-bold transition-all {data.currentPage ===
                  i + 1
                    ? 'bg-primary text-on-surface'
                    : 'bg-white/5 text-slate-400 hover:bg-white/10 hover:text-on-surface'}"
                >
                  {i + 1}
                </a>
              {/each}
            </div>
          {/if}
        {/if}
      </div>
    </div>
  </div>
</div>
