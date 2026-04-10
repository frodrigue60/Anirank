<script lang="ts">
  import { onMount } from "svelte";
  import api from "$lib/api";
  import { Bell, Users, MessageSquare, Info, ShieldCheck, Loader2 } from "lucide-svelte";

  let loading = $state(true);
  let saving = $state(false);
  let settings = $state({
    social_follow: true,
    comment_reply: true,
    user_request_feedback: true,
  });

  async function fetchSettings() {
    try {
      const response = await api.get("/settings/notifications");
      if (response.data.data?.settings) {
        // Parse JSON if it's a string or just use it if it's an object
        const raw = response.data.data.settings;
        const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw;
        settings = { ...settings, ...parsed };
      }
    } catch (error) {
      console.error("Error fetching notification settings:", error);
    } finally {
      loading = false;
    }
  }

  async function saveSettings() {
    saving = true;
    try {
      await api.put("/settings/notifications", { settings });
      // Show success toast or feedback (optional but recommended)
    } catch (error) {
      console.error("Error saving notification settings:", error);
    } finally {
      saving = false;
    }
  }

  function toggle(key: keyof typeof settings) {
    settings[key] = !settings[key];
  }

  onMount(fetchSettings);

  const categories = $derived([
    {
      id: "social",
      title: "Social",
      icon: Users,
      items: [
        {
          id: "social_follow",
          label: "When someone follows me",
          description: "Get notified when a new user starts following your activity.",
          checked: settings.social_follow,
        },
        {
          id: "comment_reply",
          label: "When someone replies to my comment",
          description: "Get notified when someone interacts with your comments on themes or playlists.",
          checked: settings.comment_reply,
        },
      ],
    },
    {
      id: "submissions",
      title: "Submissions",
      icon: ShieldCheck,
      items: [
        {
          id: "user_request_feedback",
          label: "When a song request status is updated",
          description: "Get updates on your media submissions (Approve/Reject/Feedback).",
          checked: settings.user_request_feedback,
        },
      ],
    },
  ]);
</script>

<div class="space-y-12 animate-in fade-in slide-in-from-bottom-4 duration-700">
  <!-- Header Section -->
  <div>
    <h1 class="text-4xl font-black text-white tracking-tighter mb-2">
      Notification Preferences
    </h1>
    <p class="text-on-surface-variant max-w-2xl font-medium leading-relaxed">
      Manage which activities you'd like to be notified about. These settings apply to your real-time notification feed.
    </p>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-24">
      <Loader2 class="w-8 h-8 text-primary animate-spin" />
    </div>
  {:else}
    <div class="space-y-10">
      {#each categories as category}
        <section class="space-y-6">
          <div class="flex items-center gap-3 px-1">
            <div class="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center border border-primary/20">
              <category.icon size={16} class="text-primary" />
            </div>
            <h2 class="text-lg font-black text-white uppercase tracking-wider">{category.title}</h2>
          </div>

          <div class="grid gap-4">
            {#each category.items as item}
              <button
                onclick={() => toggle(item.id as any)}
                class="group w-full text-left bg-surface-container/50 border border-outline-variant/10 p-6 rounded-2xl transition-all duration-300 hover:bg-surface-highest hover:border-primary/20 hover:shadow-2xl hover:shadow-primary/5 flex items-center justify-between"
              >
                <div class="flex-1 min-w-0">
                  <h3 class="text-base font-bold text-on-surface mb-1 group-hover:text-primary transition-colors">
                    {item.label}
                  </h3>
                  <p class="text-sm text-on-surface-variant font-medium">
                    {item.description}
                  </p>
                </div>

                <div
                  class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none {item.checked ? 'bg-primary' : 'bg-surface-highest-container'}"
                >
                  <span
                    class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out {item.checked ? 'translate-x-5' : 'translate-x-0'}"
                  ></span>
                </div>
              </button>
            {/each}
          </div>
        </section>
      {/each}

      <!-- Sticky Footer/Action Bar -->
      <div class="pt-8 flex items-center justify-end border-t border-outline-variant/10">
        <button
          onclick={saveSettings}
          disabled={saving}
          class="flex items-center gap-2 bg-primary hover:bg-primary-hover text-white px-8 py-3 rounded-xl font-bold transition-all active:scale-95 disabled:opacity-50 disabled:active:scale-100 shadow-lg shadow-primary/20"
        >
          {#if saving}
            <Loader2 size={18} class="animate-spin" />
          {/if}
          {saving ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </div>
  {/if}

  <!-- Footer Note -->
  <div class="bg-surface-container/30 border border-yellow-500/10 p-6 rounded-2xl flex gap-4">
    <div class="w-10 h-10 rounded-xl bg-yellow-500/10 flex items-center justify-center border border-yellow-500/20 shrink-0">
      <Bell size={20} class="text-yellow-500" />
    </div>
    <div class="space-y-1">
      <h4 class="text-sm font-bold text-on-surface">Universal Fallback</h4>
      <p class="text-xs text-on-surface-variant font-medium leading-relaxed">
        Critical system messages, security alerts, and administrative notices will always be sent regardless of your preferences.
      </p>
    </div>
  </div>
</div>
