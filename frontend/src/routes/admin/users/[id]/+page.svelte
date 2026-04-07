<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { fade, slide } from "svelte/transition";

  let { data } = $props();
  const user = $derived(data.user);

  let isResetting = $state(false);
  let newPassword = $state<string | null>(null);
  let isDeleting = $state(false);

  async function handleResetPassword() {
    if (!confirm(`Are you sure you want to reset the password for ${user.name}?`)) return;

    isResetting = true;
    newPassword = null;
    try {
      const res = await api.post(`/admin/users/${user.uuid}/reset-password`);
      newPassword = res.data.password;
      toastState.addToast("New password generated!", "success");
    } catch (err) {
      console.error("Failed to reset password", err);
      toastState.addToast("Failed to reset password", "error");
    } finally {
      isResetting = false;
    }
  }

  async function copyPassword() {
    if (newPassword) {
      await navigator.clipboard.writeText(newPassword);
      toastState.addToast("Password copied to clipboard!", "success");
    }
  }

  async function handleDeleteUser() {
    if (!confirm("CRITICAL: Are you sure you want to PERMANENTLY delete this user? This cannot be undone.")) return;

    isDeleting = true;
    try {
      await api.delete(`/admin/users/${user.uuid}`);
      toastState.addToast("User deleted successfully", "success");
      goto("/admin/users");
    } catch (err) {
      console.error("Failed to delete user", err);
      toastState.addToast("Failed to delete user", "error");
    } finally {
      isDeleting = false;
    }
  }
</script>

<svelte:head>
  <title>{user.name} | User Details | Admin</title>
</svelte:head>

<div class="mb-8 flex items-center justify-between">
  <div class="flex items-center gap-4">
    <a
      href="/admin/users"
      class="p-2 text-on-surface-variant/70 hover:text-on-surface hover:bg-surface-highest rounded-lg transition-colors border border-transparent hover:border-outline-variant"
    >
      <span class="material-symbols-outlined text-xl">arrow_back</span>
    </a>
    <div>
      <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-0.5">User Details</h1>
      <p class="text-on-surface-variant/70 text-sm">Managing account and permissions for <span class="text-primary font-medium">{user.name}</span></p>
    </div>
  </div>

  <div class="flex items-center gap-2">
    <a
      href="/admin/users/{user.uuid}/edit"
      class="px-4 py-2 bg-primary/10 hover:bg-primary/20 text-primary border border-primary/20 rounded-xl transition-all font-medium flex items-center gap-2"
    >
      <span class="material-symbols-outlined text-[18px]">edit</span>
      Edit Profile
    </a>
  </div>
</div>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
  <!-- Profile Header & Main Info -->
  <div class="lg:col-span-2 space-y-8">
    
    <!-- Hero Section -->
    <div class="relative bg-surface-container border border-outline-variant rounded-3xl overflow-hidden group">
      <!-- Banner -->
      <div class="h-48 relative overflow-hidden">
        {#if user.banner_url}
          <img src={user.banner_url} alt="Banner" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-700" />
        {:else}
          <div class="w-full h-full bg-linear-to-br from-anirank-card to-white/5"></div>
        {/if}
        <div class="absolute inset-0 bg-linear-to-t from-anirank-card via-transparent to-transparent"></div>
      </div>

      <!-- User Identity -->
      <div class="px-8 pb-8 -mt-12 relative flex flex-col sm:flex-row items-end gap-6">
        <div class="relative shrink-0">
          <div class="w-32 h-32 rounded-3xl overflow-hidden border-4 border-anirank-bg shadow-2xl bg-surface-highest">
            {#if user.avatar_url}
              <img src={user.avatar_url} alt={user.name} class="w-full h-full object-cover" />
            {:else}
              <div class="w-full h-full flex items-center justify-center text-4xl font-bold text-on-surface-variant/40 uppercase">
                {user.name.charAt(0)}
              </div>
            {/if}
          </div>
          <div class="absolute -bottom-2 -right-2 bg-primary text-on-surface w-10 h-10 rounded-2xl flex items-center justify-center font-bold text-sm shadow-xl border-4 border-anirank-bg">
            {user.level}
          </div>
        </div>

        <div class="flex-1 pb-2">
          <div class="flex items-center gap-3 mb-1">
            <h2 class="text-3xl font-bold text-on-surface leading-none">{user.name}</h2>
            {#if user.roles?.some((r: any) => r.slug === 'admin' || r.slug === 'owner')}
               <span class="material-symbols-outlined text-primary text-xl" title="Verified Staff">verified</span>
            {/if}
          </div>
          <div class="flex flex-wrap items-center gap-4 text-sm text-on-surface-variant/70">
             <span class="flex items-center gap-1.5"><span class="material-symbols-outlined text-[16px]">alternate_email</span> {user.email}</span>
             <span class="flex items-center gap-1.5"><span class="material-symbols-outlined text-[16px]">fingerprint</span> {user.uuid}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="bg-surface-container border border-outline-variant rounded-2xl p-5 hover:border-outline-variant transition-colors">
        <div class="text-on-surface-variant/40 text-xs font-semibold uppercase tracking-wider mb-2">Total XP</div>
        <div class="text-2xl font-bold text-on-surface flex items-baseline gap-1">
          {user.xp.toLocaleString()}
          <span class="text-xs text-primary">pts</span>
        </div>
      </div>
      <div class="bg-surface-container border border-outline-variant rounded-2xl p-5 hover:border-outline-variant transition-colors">
        <div class="text-on-surface-variant/40 text-xs font-semibold uppercase tracking-wider mb-2">Followers</div>
        <div class="text-2xl font-bold text-on-surface">{user.followers_count || 0}</div>
      </div>
       <div class="bg-surface-container border border-outline-variant rounded-2xl p-5 hover:border-outline-variant transition-colors">
        <div class="text-on-surface-variant/40 text-xs font-semibold uppercase tracking-wider mb-2">Following</div>
        <div class="text-2xl font-bold text-on-surface">{user.following_count || 0}</div>
      </div>
      <div class="bg-surface-container border border-outline-variant rounded-2xl p-5 hover:border-outline-variant transition-colors">
        <div class="text-on-surface-variant/40 text-xs font-semibold uppercase tracking-wider mb-2">Ratings</div>
        <div class="text-2xl font-bold text-on-surface">{user.ratings_count || 0}</div>
      </div>
    </div>

    <!-- Details Card -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
       <!-- Account Information -->
       <div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden">
         <div class="px-6 py-4 border-b border-outline-variant bg-white/2">
           <h3 class="text-sm font-bold text-on-surface flex items-center gap-2">
             <span class="material-symbols-outlined text-primary text-[18px]">account_circle</span>
             Account Information
           </h3>
         </div>
         <div class="p-6 space-y-4">
            <div class="flex justify-between items-center bg-white/2 px-4 py-3 rounded-xl border border-outline-variant">
              <span class="text-on-surface-variant/70 text-sm">Joined Date</span>
              <span class="text-on-surface text-sm font-medium">{new Date(user.created_at).toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' })}</span>
            </div>
            <div class="flex justify-between items-center bg-white/2 px-4 py-3 rounded-xl border border-outline-variant">
              <span class="text-on-surface-variant/70 text-sm">Email Verified</span>
              {#if user.email_verified_at}
                <span class="text-emerald-400 text-xs font-bold uppercase tracking-widest flex items-center gap-1">
                  <span class="material-symbols-outlined text-[16px]">check_circle</span> Verified
                </span>
              {:else}
                 <span class="text-yellow-500/80 text-xs font-bold uppercase tracking-widest flex items-center gap-1">
                  <span class="material-symbols-outlined text-[16px]">pending</span> Unverified
                </span>
              {/if}
            </div>
             <div class="flex justify-between items-center bg-white/2 px-4 py-3 rounded-xl border border-outline-variant">
              <span class="text-on-surface-variant/70 text-sm">Score Format</span>
              <span class="text-on-surface text-sm font-medium uppercase">{user.score_format || 'Standard'}</span>
            </div>
         </div>
       </div>

       <!-- Linked Accounts -->
       <div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden">
         <div class="px-6 py-4 border-b border-outline-variant bg-white/2">
           <h3 class="text-sm font-bold text-on-surface flex items-center gap-2">
             <span class="material-symbols-outlined text-primary text-[18px]">hub</span>
             Linked Accounts
           </h3>
         </div>
         <div class="p-6 space-y-4">
            <!-- AniList -->
             <div class="flex justify-between items-center bg-white/2 px-4 py-3 rounded-xl border border-outline-variant">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-lg bg-[#3dbbee]/10 flex items-center justify-center">
                  <svg class="w-5 h-5 text-[#3dbbee]" viewBox="0 0 24 24" fill="currentColor"><path d="M24 17.53v2.42c0 .59-.44 1.05-1.02 1.05h-4.96c-.58 0-1.02-.46-1.02-1.05v-2.42c0-.59.44-1.05 1.02-1.05h4.96c.58 0 1.02.46 1.02 1.05zm-14.87-4.74h-4.96c-.58 0-1.02.46-1.02 1.05v7.16c0 .59.44 1.05 1.02 1.05h4.96c.58 0 1.02-.46 1.02-1.05v-7.16c0-.59-.44-1.05-1.02-1.05zm9.11-10.84h-4.96c-.58 0-1.02.46-1.02 1.05v13.06c0 .59.44 1.05 1.02 1.05h4.96c.58 0 1.02-.46 1.02-1.05v-13.06c0-.59-.44-1.05-1.02-1.05zM4.17 1.95c-.58 0-1.02.46-1.02 1.05v7.16c0 .59.44 1.05 1.02 1.05h4.96c.58 0 1.02-.46 1.02-1.05v-7.16c0-.59-.44-1.05-1.02-1.05z"/></svg>
                </div>
                <span class="text-on-surface text-sm font-medium">AniList</span>
              </div>
              {#if user.anilist_id}
                <span class="text-emerald-400 text-[10px] font-bold uppercase tracking-widest bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">Connected</span>
              {:else}
                <span class="text-on-surface-variant/40 text-[10px] font-bold uppercase tracking-widest bg-surface-highest px-2 py-0.5 rounded">Not Linked</span>
              {/if}
            </div>

            <!-- Google -->
            <div class="flex justify-between items-center bg-white/2 px-4 py-3 rounded-xl border border-outline-variant">
               <div class="flex items-center gap-3">
                 <div class="w-8 h-8 rounded-lg bg-surface-highest flex items-center justify-center">
                   <svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor"><path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/><path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/><path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/><path d="M12 5.38c1.62 0 3.06.56 4.21 1.66l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/></svg>
                 </div>
                 <span class="text-on-surface text-sm font-medium">Google</span>
               </div>
               {#if user.google_id}
                <span class="text-emerald-400 text-[10px] font-bold uppercase tracking-widest bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">Connected</span>
              {:else}
                <span class="text-on-surface-variant/40 text-[10px] font-bold uppercase tracking-widest bg-surface-highest px-2 py-0.5 rounded">Not Linked</span>
              {/if}
            </div>
         </div>
       </div>
    </div>
  </div>

  <!-- Sidebar: RBAC & Badges -->
  <div class="space-y-8">
     <!-- Roles Card -->
     <div class="bg-surface-container border border-outline-variant rounded-3xl p-6">
       <div class="flex items-center justify-between mb-6">
         <h3 class="text-base font-bold text-on-surface flex items-center gap-2">
           <span class="material-symbols-outlined text-rose-400">shield_person</span>
           Assigned Roles
         </h3>
       </div>
       <div class="flex flex-wrap gap-2">
         {#each user.roles || [] as role}
           <div class="px-3 py-1.5 bg-primary/10 border border-primary/20 rounded-xl text-primary text-xs font-bold uppercase tracking-wider">
             {role.name}
           </div>
         {:else}
           <div class="w-full py-4 text-center text-on-surface-variant/40 text-sm italic bg-white/2 rounded-2xl border border-outline-variant">
             No roles assigned
           </div>
         {/each}
       </div>
     </div>

     <!-- Badges Card -->
     <div class="bg-surface-container border border-outline-variant rounded-3xl p-6">
       <div class="flex items-center justify-between mb-6">
         <h3 class="text-base font-bold text-on-surface flex items-center gap-2">
           <span class="material-symbols-outlined text-yellow-400">military_tech</span>
           User Badges
         </h3>
         <span class="text-xs text-on-surface-variant/40 font-mono">{user.badges?.length || 0} Total</span>
       </div>
       <div class="grid grid-cols-4 sm:grid-cols-6 lg:grid-cols-4 gap-3">
         {#each user.badges || [] as badge}
           <div class="aspect-square bg-surface-highest rounded-2xl flex items-center justify-center p-2 border border-outline-variant hover:border-outline-variant transition-all hover:-translate-y-1 relative group" title={badge.name}>
             {#if badge.icon_url}
               <img src={badge.icon_url} alt={badge.name} class="w-full h-full object-contain" />
             {:else}
               <span class="material-symbols-outlined text-gray-600 text-xl">image_not_supported</span>
             {/if}
             <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2 py-1 bg-black rounded text-[10px] text-on-surface opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-10 shadow-2xl">
               {badge.name}
             </div>
           </div>
         {:else}
           <div class="col-span-full py-4 text-center text-on-surface-variant/40 text-sm italic bg-white/2 rounded-2xl border border-outline-variant">
             No badges earned
           </div>
         {/each}
       </div>
     </div>

     <!-- Security & Danger Zone -->
     <div class="bg-rose-500/5 border border-rose-500/10 rounded-3xl p-6">
       <h3 class="text-base font-bold text-rose-400 mb-4 flex items-center gap-2">
         <span class="material-symbols-outlined">security</span>
         Control Panel
       </h3>
       
       <div class="space-y-3">
         {#if newPassword}
           <div class="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-2xl mb-4" in:slide>
             <p class="text-emerald-400 text-[10px] font-bold uppercase tracking-widest mb-2">New Temporary Password:</p>
             <div class="flex items-center gap-2">
                <code class="flex-1 bg-black/40 px-3 py-2 rounded-xl text-on-surface font-mono text-lg tracking-wider border border-outline-variant">{newPassword}</code>
                <button onclick={copyPassword} class="p-2 bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-400 rounded-xl transition-colors">
                  <span class="material-symbols-outlined">content_copy</span>
                </button>
             </div>
           </div>
         {/if}

         <button 
           onclick={handleResetPassword}
           disabled={isResetting}
           class="w-full px-4 py-3 bg-surface-highest hover:bg-surface-highest text-on-surface rounded-2xl transition-all border border-outline-variant text-sm font-semibold flex items-center gap-3 disabled:opacity-50"
         >
           <span class="material-symbols-outlined text-rose-400">lock_reset</span>
           {isResetting ? 'Resetting...' : 'Reset Password'}
         </button>

         <button 
           onclick={handleDeleteUser}
           disabled={isDeleting}
           class="w-full px-4 py-3 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 rounded-2xl transition-all border border-rose-500/20 text-sm font-semibold flex items-center gap-3 disabled:opacity-50"
         >
           <span class="material-symbols-outlined">person_remove</span>
           {isDeleting ? 'Deleting...' : 'Terminate Account'}
         </button>
       </div>
     </div>
  </div>
</div>
