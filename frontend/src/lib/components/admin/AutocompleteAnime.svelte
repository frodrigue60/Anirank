<script lang="ts">
    import { onMount } from "svelte";
    import axios from "axios";
    import api from "$lib/api";

    let { 
        value = $bindable(""), 
        id = "anime-search",
        showLabel = true,
        onselect = (id: string) => {},
        placeholder = "Search anime..." 
    } = $props<{
        value?: string;
        id?: string;
        showLabel?: boolean;
        onselect?: (id: string) => void;
        placeholder?: string;
    }>();

    let search = $state("");
    let suggestions = $state<any[]>([]);
    let isOpen = $state(false);
    let loading = $state(false);
    let container: HTMLElement;

    // To prevent immediate refetch when setting search from value
    let ignoreNext = false;

    $effect(() => {
        if (value && (search === "" || ignoreNext)) {
            // Initial load or value change
            fetchAnimeName(value);
        } else if (!value) {
            search = "";
        }
    });

    async function fetchAnimeName(id: string) {
        try {
            const res = await api.get(`/admin/animes/${id}`);
            if (res.data?.data) {
                search = res.data.data.title;
                ignoreNext = true;
            }
        } catch (err) {
            console.error("Failed to fetch anime name", err);
        }
    }

    async function handleInput() {
        if (ignoreNext) {
            ignoreNext = false;
            return;
        }

        if (search.length < 2) {
            suggestions = [];
            isOpen = false;
            return;
        }

        loading = true;
        try {
            const res = await api.get(`/admin/animes?search=${encodeURIComponent(search)}&limit=10`);
            suggestions = res.data.data || [];
            isOpen = suggestions.length > 0;
        } catch (err) {
            console.error("Autocomplete failed", err);
            suggestions = [];
        } finally {
            loading = false;
        }
    }

    function select(anime: any) {
        search = anime.title;
        value = anime.id.toString();
        isOpen = false;
        ignoreNext = true;
        onselect(value);
    }

    function handleClickOutside(e: MouseEvent) {
        if (container && !container.contains(e.target as Node)) {
            isOpen = false;
        }
    }

    onMount(() => {
        window.addEventListener("click", handleClickOutside);
        return () => window.removeEventListener("click", handleClickOutside);
    });
</script>

<div class="relative w-full z-30" bind:this={container}>
    {#if showLabel}
    <label for={id} class="block text-[10px] uppercase font-black text-white/40 mb-2 ml-1 tracking-widest">Anime</label>
    {/if}
    <div class="relative">
        <input
            id={id}
            type="text"
            bind:value={search}
            oninput={handleInput}
            onfocus={() => search.length >= 2 && (isOpen = true)}
            placeholder={placeholder}
            class="w-full h-12 bg-surface-darker/50 border border-white/10 rounded-xl px-4 text-sm text-white placeholder-white/30 focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer backdrop-blur-sm shadow-lg shadow-black/20"
            autocomplete="off"
        />
        {#if loading}
            <div class="absolute right-3 top-1/2 -translate-y-1/2">
                <div class="w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
            </div>
        {/if}
    </div>

    {#if isOpen && suggestions.length > 0}
        <div class="absolute z-50 w-full mt-2 py-2 bg-surface-dark/95 border border-white/10 rounded-2xl backdrop-blur-xl shadow-[0_20px_50px_rgba(0,0,0,0.5)] overflow-hidden">
            <div class="max-h-60 overflow-y-auto custom-scrollbar">
                {#each suggestions as anime}
                    <button
                        type="button"
                        onclick={() => select(anime)}
                        class="w-full px-4 py-3 text-left text-sm transition-all flex items-center justify-between group/opt text-white/60 hover:bg-white/5 hover:text-white"
                    >
                        <div class="flex flex-col">
                            <span class="font-bold">{anime.title}</span>
                            {#if anime.year || anime.season}
                            <span class="text-[10px] text-white/40 uppercase tracking-widest">{anime.season?.name || ''} {anime.year?.name || ''}</span>
                            {/if}
                        </div>
                        <span class="text-[10px] text-white/20 font-mono">ID: {anime.id}</span>
                    </button>
                {/each}
            </div>
        </div>
    {/if}
</div>

<style>
    .custom-scrollbar::-webkit-scrollbar {
        width: 4px;
    }
    .custom-scrollbar::-webkit-scrollbar-track {
        background: transparent;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb {
        background: rgba(255, 255, 255, 0.1);
        border-radius: 10px;
    }
</style>
