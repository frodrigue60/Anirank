<script lang="ts">
  import { onMount } from "svelte";
  import {
    getWebhooks,
    createWebhook,
    updateWebhook,
    deleteWebhook,
    testWebhook,
  } from "$lib/api";
  import Plus from "lucide-svelte/icons/plus";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import Edit2 from "lucide-svelte/icons/edit-2";
  import Send from "lucide-svelte/icons/send";
  import CheckCircle from "lucide-svelte/icons/check-circle";
  import XCircle from "lucide-svelte/icons/x-circle";
  import ExternalLink from "lucide-svelte/icons/external-link";

  let webhooks = $state([]);
  let loading = $state(true);
  let error = $state(null);

  // Modal State
  let showModal = $state(false);
  let editingWebhook = $state(null);
  let form = $state({
    name: "",
    url: "",
    provider: "discord",
    is_active: true,
    content_types: ["animes", "songs"],
  });

  let saving = $state(false);

  async function loadWebhooks() {
    loading = true;
    try {
      webhooks = await getWebhooks();
    } catch (e) {
      error = "Error al cargar webhooks";
      console.error(e);
    } finally {
      loading = false;
    }
  }

  onMount(loadWebhooks);

  function openCreate() {
    editingWebhook = null;
    form = {
      name: "",
      url: "",
      provider: "discord",
      is_active: true,
      content_types: ["animes", "songs"],
    };
    showModal = true;
  }

  function openEdit(webhook) {
    editingWebhook = webhook;
    form = {
      name: webhook.name,
      url: webhook.url,
      provider: webhook.provider,
      is_active: webhook.is_active,
      content_types: [...webhook.content_types],
    };
    showModal = true;
  }

  async function save() {
    saving = true;
    try {
      if (editingWebhook) {
        await updateWebhook(editingWebhook.uuid, form);
      } else {
        await createWebhook(form);
      }
      showModal = false;
      await loadWebhooks();
    } catch (e) {
      alert("Error al guardar webhook");
    } finally {
      saving = false;
    }
  }

  async function remove(uuid) {
    if (!confirm("¿Estás seguro de eliminar este webhook?")) return;
    try {
      await deleteWebhook(uuid);
      await loadWebhooks();
    } catch (e) {
      alert("Error al eliminar");
    }
  }

  async function test(uuid) {
    try {
      await testWebhook(uuid);
      alert("¡Mensaje de prueba enviado!");
    } catch (e) {
      alert("Error al enviar prueba: " + (e.response?.data?.message || e.message));
    }
  }

  function toggleContentType(type) {
    if (form.content_types.includes(type)) {
      form.content_types = form.content_types.filter((t) => t !== type);
    } else {
      form.content_types = [...form.content_types, type];
    }
  }
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold text-on-surface">Webhooks Discord</h1>
      <p class="text-on-surface-variant/70">Gestiona las integraciones para anuncios automáticos.</p>
    </div>
    <button
      onclick={openCreate}
      class="flex items-center gap-2 bg-primary text-on-primary px-4 py-2 rounded-xl font-medium hover:bg-primary/90 transition-colors"
    >
      <Plus size={20} />
      Nuevo Webhook
    </button>
  </div>

  {#if loading}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each Array(3) as _}
        <div class="h-48 bg-surface-container rounded-3xl animate-pulse"></div>
      {/each}
    </div>
  {:else if webhooks.length === 0}
    <div class="bg-surface-container rounded-3xl p-12 text-center border border-gray-500/30">
      <div class="bg-surface-highest w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-4 text-on-surface-variant/40">
        <Send size={32} />
      </div>
      <h2 class="text-xl font-semibold text-on-surface">No hay webhooks configurados</h2>
      <p class="text-on-surface-variant/70 mt-2">Agrega un webhook de Discord para empezar a enviar notificaciones.</p>
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each webhooks as webhook}
        <div class="bg-surface-container rounded-3xl p-6 border border-gray-500/30 flex flex-col justify-between hover:border-primary/30 transition-all group">
          <div>
            <div class="flex items-start justify-between mb-4">
              <div class="flex items-center gap-3">
                <div class="bg-primary/10 text-primary p-3 rounded-2xl">
                  <ExternalLink size={24} />
                </div>
                <div>
                  <h3 class="font-bold text-lg text-on-surface">{webhook.name}</h3>
                  <span class="text-xs text-on-surface-variant/50 font-mono truncate max-w-[150px] block">
                    {webhook.uuid}
                  </span>
                </div>
              </div>
              <div class="flex items-center gap-1">
                {#if webhook.is_active}
                  <span class="bg-green-500/10 text-green-500 text-[10px] uppercase font-bold px-2 py-1 rounded-full flex items-center gap-1">
                    <CheckCircle size={10} /> Activo
                  </span>
                {:else}
                  <span class="bg-gray-500/10 text-gray-500 text-[10px] uppercase font-bold px-2 py-1 rounded-full flex items-center gap-1">
                    <XCircle size={10} /> Inactivo
                  </span>
                {/if}
              </div>
            </div>

            <div class="space-y-3 mt-4">
              <div class="flex flex-wrap gap-2">
                {#each webhook.content_types as type}
                  <span class="bg-surface-highest text-on-surface-variant text-xs px-3 py-1 rounded-lg border border-gray-500/20">
                    {type === 'animes' ? '📺 Animes' : '🎵 Temas'}
                  </span>
                {/each}
              </div>
            </div>
          </div>

          <div class="flex items-center justify-end gap-2 mt-6 pt-4 border-t border-gray-500/10">
             <button
              onclick={() => test(webhook.uuid)}
              class="p-2 hover:bg-blue-500/10 text-blue-500 rounded-xl transition-colors tooltip"
              title="Probar Webhook"
            >
              <Send size={18} />
            </button>
            <button
              onclick={() => openEdit(webhook)}
              class="p-2 hover:bg-primary/10 text-primary rounded-xl transition-colors"
            >
              <Edit2 size={18} />
            </button>
            <button
              onclick={() => remove(webhook.uuid)}
              class="p-2 hover:bg-red-500/10 text-red-500 rounded-xl transition-colors"
            >
              <Trash2 size={18} />
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

{#if showModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
    <div class="bg-surface-container rounded-3xl w-full max-w-lg shadow-2xl overflow-hidden border border-gray-500/20">
      <div class="p-6 border-b border-gray-500/10 flex justify-between items-center">
        <h2 class="text-xl font-bold text-on-surface">
          {editingWebhook ? "Editar Webhook" : "Nuevo Webhook"}
        </h2>
        <button onclick={() => (showModal = false)} class="text-on-surface-variant hover:text-on-surface">
          <Plus class="rotate-45" size={24} />
        </button>
      </div>

      <div class="p-6 space-y-4">
        <div>
          <label class="block text-sm font-medium text-on-surface-variant mb-1">Nombre</label>
          <input
            bind:value={form.name}
            type="text"
            placeholder="Ej: Canal #noticias-discord"
            class="w-full bg-surface-highest border border-gray-500/20 rounded-xl px-4 py-3 text-on-surface focus:outline-none focus:ring-2 focus:ring-primary/50"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-on-surface-variant mb-1">URL del Webhook</label>
          <input
            bind:value={form.url}
            type="text"
            placeholder="https://discord.com/api/webhooks/..."
            class="w-full bg-surface-highest border border-gray-500/20 rounded-xl px-4 py-3 text-on-surface focus:outline-none focus:ring-2 focus:ring-primary/50 font-mono text-xs"
          />
        </div>

        <div class="grid grid-cols-2 gap-4">
           <div>
            <label class="block text-sm font-medium text-on-surface-variant mb-2">Estado</label>
            <label class="flex items-center gap-3 cursor-pointer group">
              <div class="relative">
                <input
                  type="checkbox"
                  bind:checked={form.is_active}
                  class="sr-only peer"
                />
                <div class="w-11 h-6 bg-surface-highest peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
              </div>
              <span class="text-sm font-medium text-on-surface">
                {form.is_active ? 'Activo' : 'Inactivo'}
              </span>
            </label>
          </div>

          <div>
             <label class="block text-sm font-medium text-on-surface-variant mb-2">Contenido</label>
             <div class="flex gap-2">
                <button
                  type="button"
                  onclick={() => toggleContentType('animes')}
                  class="px-3 py-1.5 rounded-lg text-xs font-medium border transition-all {form.content_types.includes('animes') ? 'bg-primary text-on-primary border-primary' : 'bg-surface-highest text-on-surface-variant border-gray-500/20'}"
                >
                  📺 Animes
                </button>
                <button
                  type="button"
                  onclick={() => toggleContentType('songs')}
                  class="px-3 py-1.5 rounded-lg text-xs font-medium border transition-all {form.content_types.includes('songs') ? 'bg-primary text-on-primary border-primary' : 'bg-surface-highest text-on-surface-variant border-gray-500/20'}"
                >
                  🎵 Temas
                </button>
             </div>
          </div>
        </div>
      </div>

      <div class="p-6 bg-surface-highest/50 flex justify-end gap-3">
        <button
          onclick={() => (showModal = false)}
          class="px-6 py-2.5 rounded-xl font-medium text-on-surface-variant hover:bg-surface-highest transition-colors"
        >
          Cancelar
        </button>
        <button
          onclick={save}
          disabled={saving || !form.name || !form.url}
          class="px-8 py-2.5 bg-primary text-on-primary rounded-xl font-bold hover:bg-primary/90 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 shadow-lg shadow-primary/20"
        >
          {#if saving}
            <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            Guardando...
          {:else}
            {editingWebhook ? "Actualizar" : "Crear Webhook"}
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .tooltip {
    position: relative;
  }
</style>
