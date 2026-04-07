<script lang="ts">
    import { onMount } from "svelte";
    import api from "$lib/api";
    let { data } = $props();

    let log = $state<any>(null);
    let loading = $state(true);
    let error = $state<string | null>(null);

    const eventColors: Record<string, string> = {
        created: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
        updated: "bg-blue-500/10 text-blue-400 border-blue-500/20",
        deleted: "bg-rose-500/10 text-rose-400 border-rose-500/20",
        status_toggled: "bg-amber-500/10 text-amber-400 border-amber-500/20",
    };

    function formatJSON(json: any) {
        if (!json) return "N/A";
        return JSON.stringify(json, null, 2);
    }

    onMount(async () => {
        try {
            const res = await api.get(`/admin/audit-logs/${data.id}`);
            log = res.data.data;
        } catch (e: any) {
            console.error("Failed to load audit log:", e);
            error = e.response?.data?.message || "Failed to load audit log detail";
        } finally {
            loading = false;
        }
    });
</script>

<svelte:head>
    <title>Audit Log Detail {data.id ? `#${data.id}` : ''} | Admin</title>
</svelte:head>

<div class="max-w-6xl mx-auto px-4 py-8">
    <!-- Breadcrumbs -->
    <nav class="flex items-center gap-2 text-sm text-on-surface-variant/40 mb-6">
        <a href="/admin/audit-logs" class="hover:text-on-surface transition-colors">Audit Logs</a>
        <span class="material-symbols-outlined text-xs">chevron_right</span>
        <span class="text-on-surface-variant">Log Detail</span>
    </nav>

    {#if loading}
        <div class="flex flex-col items-center justify-center py-20 gap-4">
            <div class="w-12 h-12 border-4 border-primary/20 border-t-anirank-primary rounded-full animate-spin"></div>
            <p class="text-on-surface-variant/70 font-medium animate-pulse">Cargando detalles del registro...</p>
        </div>
    {:else if error}
        <div class="bg-rose-500/10 border border-rose-500/20 rounded-2xl p-8 text-center max-w-2xl mx-auto">
            <div class="w-16 h-16 rounded-full bg-rose-500/10 flex items-center justify-center text-rose-500 mx-auto mb-4">
                <span class="material-symbols-outlined text-3xl">error</span>
            </div>
            <h2 class="text-xl font-bold text-on-surface mb-2">Error al cargar</h2>
            <p class="text-rose-400/80 mb-6">{error}</p>
            <button 
                onclick={() => location.reload()}
                class="px-6 py-2 bg-rose-500 text-on-surface rounded-xl font-bold hover:bg-rose-600 transition-colors"
            >
                Intentar de nuevo
            </button>
        </div>
    {:else if log}
        <div class="flex flex-col md:flex-row md:items-end justify-between gap-4 mb-8">
            <div>
                <div class="flex items-center gap-3 mb-2">
                    <h1 class="text-3xl font-bold tracking-tight text-on-surface">Log #{log.id}</h1>
                    <span class="px-3 py-1 rounded-full text-xs font-bold border uppercase tracking-wider {eventColors[log.event] || 'bg-gray-500/10 text-on-surface-variant/70 border-gray-500/20'}">
                        {log.event}
                    </span>
                </div>
                <p class="text-on-surface-variant/70">
                    Action performed on <span class="text-on-surface font-medium capitalize">{log.auditable_type}</span> (ID: {log.auditable_id})
                </p>
            </div>

            <button
                onclick={() => history.back()}
                class="px-4 py-2 bg-surface-highest hover:bg-surface-highest text-on-surface rounded-xl transition-colors border border-outline-variant flex items-center gap-2 self-start"
            >
                <span class="material-symbols-outlined text-sm">arrow_back</span>
                Back
            </button>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
            <!-- Metadata Sidebar -->
            <div class="lg:col-span-1 space-y-6">
                <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-xl">
                    <h3 class="text-xs font-bold text-on-surface-variant/40 uppercase tracking-widest mb-6">Execution Context</h3>
                    
                    <div class="space-y-6">
                        <div class="flex items-start gap-3">
                            <div class="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center border border-blue-500/20 text-blue-400">
                                <span class="material-symbols-outlined text-sm">person</span>
                            </div>
                            <div>
                                <span class="block text-xs text-on-surface-variant/40 font-medium mb-0.5">Executor</span>
                                <span class="text-sm text-on-surface font-bold">{log.user_name || `User ID: ${log.user_id}`}</span>
                            </div>
                        </div>

                        <div class="flex items-start gap-3">
                            <div class="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20 text-emerald-400">
                                <span class="material-symbols-outlined text-sm">calendar_today</span>
                            </div>
                            <div>
                                <span class="block text-xs text-on-surface-variant/40 font-medium mb-0.5">Timestamp</span>
                                <span class="text-sm text-on-surface">{new Date(log.created_at).toLocaleString()}</span>
                            </div>
                        </div>

                        <div class="flex items-start gap-3">
                            <div class="w-8 h-8 rounded-lg bg-purple-500/10 flex items-center justify-center border border-purple-500/20 text-purple-400">
                                <span class="material-symbols-outlined text-sm">lan</span>
                            </div>
                            <div class="overflow-hidden">
                                <span class="block text-xs text-on-surface-variant/40 font-medium mb-0.5">IP Address</span>
                                <span class="text-sm text-on-surface font-mono">{log.ip_address || "None"}</span>
                            </div>
                        </div>

                        <div class="flex items-start gap-3 pt-4 border-t border-outline-variant">
                            <div class="w-full">
                                <span class="block text-xs text-on-surface-variant/40 font-medium mb-2">Target URL</span>
                                <div class="bg-black/20 rounded-lg p-3 text-xs font-mono text-on-surface-variant/70 border border-outline-variant break-all">
                                    {log.url || "/"}
                                </div>
                            </div>
                        </div>

                        <div class="pt-4">
                            <span class="block text-xs text-on-surface-variant/40 font-medium mb-2">User Agent</span>
                            <div class="bg-black/20 rounded-lg p-3 text-xs font-mono text-on-surface-variant/70 border border-outline-variant leading-relaxed">
                                {log.user_agent || "No agent data"}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Data Content -->
            <div class="lg:col-span-2 space-y-6">
                <div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden shadow-xl">
                    <div class="px-6 py-4 bg-surface-highest border-b border-outline-variant flex items-center justify-between">
                        <h3 class="text-sm font-bold text-on-surface flex items-center gap-2">
                            <span class="material-symbols-outlined text-sm opacity-50">data_object</span>
                            State Changes
                        </h3>
                    </div>

                    <div class="p-6">
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
                            <div>
                                <div class="flex items-center justify-between mb-3">
                                    <h4 class="text-xs font-bold text-on-surface-variant/40 uppercase tracking-wider">Before Action</h4>
                                    <span class="w-2 h-2 rounded-full bg-rose-500/50"></span>
                                </div>
                                <div class="bg-black/40 rounded-xl p-4 text-xs font-mono text-on-surface-variant/70 overflow-x-auto max-h-[500px] border border-outline-variant whitespace-pre shadow-inner">
                                    {formatJSON(log.old_values)}
                                </div>
                            </div>

                            <div>
                                <div class="flex items-center justify-between mb-3">
                                    <h4 class="text-xs font-bold text-on-surface-variant/40 uppercase tracking-wider">After Action</h4>
                                    <span class="w-2 h-2 rounded-full bg-emerald-500/50 pulse"></span>
                                </div>
                                <div class="bg-black/40 rounded-xl p-4 text-xs font-mono text-emerald-400/80 overflow-x-auto max-h-[500px] border border-outline-variant whitespace-pre shadow-inner">
                                    {formatJSON(log.new_values)}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    {/if}
</div>

<style>
    .pulse {
        animation: pulse-animation 2s infinite;
    }

    @keyframes pulse-animation {
        0% { transform: scale(0.95); opacity: 0.5; }
        50% { transform: scale(1.1); opacity: 1; }
        100% { transform: scale(0.95); opacity: 0.5; }
    }

    .whitespace-pre::-webkit-scrollbar {
        width: 4px;
        height: 4px;
    }
    .whitespace-pre::-webkit-scrollbar-track {
        background: transparent;
    }
    .whitespace-pre::-webkit-scrollbar-thumb {
        background: rgba(255, 255, 255, 0.1);
        border-radius: 10px;
    }
</style>
