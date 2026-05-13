<script lang="ts">
  import { onMount } from "svelte";
  import {
    getAdminPartners,
    createPartner,
    updatePartner,
    deletePartner,
  } from "$lib/api";
  import Plus from "lucide-svelte/icons/plus";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import Edit2 from "lucide-svelte/icons/edit-2";
  import Image from "lucide-svelte/icons/image";
  import CheckCircle from "lucide-svelte/icons/check-circle";
  import XCircle from "lucide-svelte/icons/x-circle";
  import ExternalLink from "lucide-svelte/icons/external-link";
  import Globe from "lucide-svelte/icons/globe";

  let partners = $state([]);
  let loading = $state(true);
  let error = $state(null);

  // Modal State
  let showModal = $state(false);
  let editingPartner = $state(null);
  let form = $state({
    name: "",
    url: "",
    description: "",
    type: "partner", // partner, source, alliance
    sort_order: 0,
    is_active: true,
  });

  let bannerFile = $state(null);
  let bannerPreview = $state(null);
  let saving = $state(false);

  async function loadPartners() {
    loading = true;
    try {
      partners = await getAdminPartners();
    } catch (e) {
      error = "Error loading partners";
      console.error(e);
    } finally {
      loading = false;
    }
  }

  onMount(loadPartners);

  function openCreate() {
    editingPartner = null;
    form = {
      name: "",
      url: "",
      description: "",
      type: "partner",
      sort_order: 0,
      is_active: true,
    };
    bannerFile = null;
    bannerPreview = null;
    showModal = true;
  }

  function openEdit(partner) {
    editingPartner = partner;
    form = {
      name: partner.name,
      url: partner.url,
      description: partner.description || "",
      type: partner.type,
      sort_order: partner.sort_order,
      is_active: partner.is_active,
    };
    bannerFile = null;
    bannerPreview = partner.banner_url || null;
    showModal = true;
  }

  function handleFileChange(e) {
    const file = e.target.files[0];
    if (file) {
      bannerFile = file;
      const reader = new FileReader();
      reader.onload = (e) => (bannerPreview = e.target.result);
      reader.readAsDataURL(file);
    }
  }

  async function save() {
    saving = true;
    try {
      const formData = new FormData();
      formData.append("name", form.name);
      formData.append("url", form.url);
      formData.append("description", form.description);
      formData.append("type", form.type);
      formData.append("sort_order", form.sort_order);
      formData.append("is_active", form.is_active);

      if (bannerFile) {
        formData.append("banner_file", bannerFile);
      }

      if (editingPartner) {
        await updatePartner(editingPartner.uuid, formData);
      } else {
        await createPartner(formData);
      }
      showModal = false;
      await loadPartners();
    } catch (e) {
      alert("Error saving partner");
      console.error(e);
    } finally {
      saving = false;
    }
  }

  async function remove(uuid) {
    if (!confirm("Are you sure you want to delete this partner?")) return;
    try {
      await deletePartner(uuid);
      await loadPartners();
    } catch (e) {
      alert("Error deleting");
    }
  }

  function getSrcset(sources) {
    if (!sources || sources.length === 0) return "";
    return sources.map((s) => `${s.url} ${s.width}w`).join(", ");
  }
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold text-on-surface">Partners & Communities</h1>
      <p class="text-on-surface-variant/70">Manage banners for partner sites and data sources.</p>
    </div>
    <button
      onclick={openCreate}
      class="flex items-center gap-2 bg-primary text-on-primary px-4 py-2 rounded-sm font-medium hover:bg-primary/90 transition-colors"
    >
      <Plus size={20} />
      New Partner
    </button>
  </div>

  {#if loading}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each Array(3) as _}
        <div class="h-48 bg-surface-container rounded-md animate-pulse"></div>
      {/each}
    </div>
  {:else if (partners?.length || 0) === 0}
    <div class="bg-surface-container rounded-md p-12 text-center border border-outline-variant/10">
      <div class="bg-surface-highest w-16 h-16 rounded-md flex items-center justify-center mx-auto mb-4 text-on-surface-variant/40">
        <Globe size={32} />
      </div>
      <h2 class="text-xl font-semibold text-on-surface">No partners configured</h2>
      <p class="text-on-surface-variant/70 mt-2">Add a partner to show it in the communities section.</p>
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each partners as partner}
        <div class="bg-surface-container rounded-md overflow-hidden border border-outline-variant/10 flex flex-col hover:border-primary/30 transition-all group">
          <div class="aspect-video relative bg-surface-highest overflow-hidden">
            {#if partner.banner_url}
              <img
                src={partner.banner_url}
                srcset={getSrcset(partner.banner_sources)}
                sizes="(max-width: 768px) 100vw, 33vw"
                alt={partner.name}
                class="w-full h-full object-cover"
                loading="lazy"
                decoding="async"
              />
            {:else}
              <div class="w-full h-full flex items-center justify-center text-on-surface-variant/20">
                <Image size={48} />
              </div>
            {/if}
            <div class="absolute top-3 right-3 flex items-center gap-1">
              {#if partner.is_active}
                <span class="bg-primary text-on-primary text-[10px] uppercase font-bold px-2 py-1 rounded-sm flex items-center gap-1">
                  <CheckCircle size={10} /> Active
                </span>
              {:else}
                <span class="bg-on-surface-variant/20 text-on-surface-variant text-[10px] uppercase font-bold px-2 py-1 rounded-sm flex items-center gap-1">
                  <XCircle size={10} /> Inactive
                </span>
              {/if}
            </div>
          </div>

          <div class="p-5 flex-1 flex flex-col justify-between">
            <div>
              <div class="flex items-center justify-between mb-2">
                <h3 class="font-bold text-lg text-on-surface">{partner.name}</h3>
                <span class="text-[10px] text-on-surface-variant/50 font-mono">#{partner.sort_order}</span>
              </div>
              <p class="text-sm text-on-surface-variant/70 line-clamp-2 mb-4">
                {partner.description || 'No description.'}
              </p>
              <div class="flex items-center gap-2">
                <span class="bg-surface-highest text-on-surface-variant text-[10px] uppercase font-bold px-2 py-1 rounded-sm border border-outline-variant/10">
                  {partner.type}
                </span>
                <a href={partner.url} target="_blank" class="text-primary text-xs flex items-center gap-1 hover:underline">
                  <ExternalLink size={14} /> Visit site
                </a>
              </div>
            </div>

            <div class="flex items-center justify-end gap-2 mt-6 pt-4 border-t border-outline-variant/10">
              <button
                onclick={() => openEdit(partner)}
                class="p-2 hover:bg-primary/10 text-primary rounded-sm transition-colors"
              >
                <Edit2 size={18} />
              </button>
              <button
                onclick={() => remove(partner.uuid)}
                class="p-2 hover:bg-red-500/10 text-red-500 rounded-sm transition-colors"
              >
                <Trash2 size={18} />
              </button>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

{#if showModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-on-surface/40">
    <div class="bg-surface-container rounded-md w-full max-w-lg shadow-2xl overflow-hidden border border-outline-variant/10">
      <div class="p-6 border-b border-outline-variant/10 flex justify-between items-center">
        <h2 class="text-xl font-bold text-on-surface">
          {editingPartner ? "Edit Partner" : "New Partner"}
        </h2>
        <button onclick={() => (showModal = false)} class="text-on-surface-variant hover:text-on-surface">
          <Plus class="rotate-45" size={24} />
        </button>
      </div>

      <div class="p-6 space-y-4 max-h-[70vh] overflow-y-auto">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="col-span-2">
            <label class="block text-sm font-medium text-on-surface-variant mb-1">Name</label>
            <input
              bind:value={form.name}
              type="text"
              placeholder="e.g. AniList"
              class="w-full bg-surface-highest border border-outline-variant/10 rounded-sm px-4 py-3 text-on-surface focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
          </div>

          <div class="col-span-2">
            <label class="block text-sm font-medium text-on-surface-variant mb-1">URL</label>
            <input
              bind:value={form.url}
              type="text"
              placeholder="https://..."
              class="w-full bg-surface-highest border border-outline-variant/10 rounded-sm px-4 py-3 text-on-surface focus:outline-none focus:ring-2 focus:ring-primary/50 font-mono text-sm"
            />
          </div>

          <div class="col-span-2">
            <label class="block text-sm font-medium text-on-surface-variant mb-1">Description (Optional)</label>
            <textarea
              bind:value={form.description}
              rows="3"
              placeholder="Short description of the partner..."
              class="w-full bg-surface-highest border border-outline-variant/10 rounded-sm px-4 py-3 text-on-surface focus:outline-none focus:ring-2 focus:ring-primary/50 resize-none"
            ></textarea>
          </div>

          <div>
            <label class="block text-sm font-medium text-on-surface-variant mb-1">Type</label>
            <select
              bind:value={form.type}
              class="w-full bg-surface-highest border border-outline-variant/10 rounded-sm px-4 py-3 text-on-surface focus:outline-none focus:ring-2 focus:ring-primary/50"
            >
              <option value="partner">Partner</option>
              <option value="source">Data Source</option>
              <option value="alliance">Alliance</option>
              <option value="community">Community</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-on-surface-variant mb-1">Order</label>
            <input
              bind:value={form.sort_order}
              type="number"
              class="w-full bg-surface-highest border border-outline-variant/10 rounded-sm px-4 py-3 text-on-surface focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-on-surface-variant mb-2">Banner</label>
          <div class="relative group cursor-pointer aspect-video bg-surface-highest rounded-md overflow-hidden border-2 border-dashed border-outline-variant/20 hover:border-primary/50 transition-all">
            {#if bannerPreview}
              <img src={bannerPreview} alt="Preview" class="w-full h-full object-cover" />
              <div class="absolute inset-0 bg-surface-container/60 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity">
                <span class="text-on-surface text-sm font-bold">Change Image</span>
              </div>
            {:else}
              <div class="w-full h-full flex flex-col items-center justify-center text-on-surface-variant/40">
                <Image size={32} />
                <span class="text-xs mt-2">Upload Banner</span>
              </div>
            {/if}
            <input
              type="file"
              accept="image/*"
              onchange={handleFileChange}
              class="absolute inset-0 opacity-0 cursor-pointer"
            />
          </div>
        </div>

        <div>
          <label class="flex items-center gap-3 cursor-pointer group w-fit">
            <div class="relative">
              <input
                type="checkbox"
                bind:checked={form.is_active}
                class="sr-only peer"
              />
              <div class="w-11 h-6 bg-surface-highest peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
            </div>
            <span class="text-sm font-medium text-on-surface">
              {form.is_active ? 'Visible on site' : 'Hidden'}
            </span>
          </label>
        </div>
      </div>

      <div class="p-6 bg-surface-highest flex justify-end gap-3">
        <button
          onclick={() => (showModal = false)}
          class="px-6 py-2.5 rounded-sm font-medium text-on-surface-variant hover:bg-surface-highest transition-colors"
        >
          Cancel
        </button>
        <button
          onclick={save}
          disabled={saving || !form.name || !form.url}
          class="px-8 py-2.5 bg-primary text-on-primary rounded-sm font-bold hover:bg-primary/90 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
        >
          {#if saving}
            <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            Saving...
          {:else}
            {editingPartner ? "Update" : "Create Partner"}
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}
