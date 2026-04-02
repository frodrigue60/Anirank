<script lang="ts">
  import { onMount } from "svelte";
  import { fade, slide } from "svelte/transition";

  interface Option {
    value: string;
    label: string;
  }

  let {
    label = "",
    value = $bindable(""),
    options = [],
    placeholder = "Select an option",
    icon: IconComponent = null,
    onchange = () => {},
  } = $props<{
    label?: string;
    value?: string;
    options: Option[];
    placeholder?: string;
    icon?: any;
    onchange?: (val: string) => void;
  }>();

  let isOpen = $state(false);
  let container: HTMLElement;

  const selectedLabel = $derived(
    options.find((opt: Option) => opt.value === value)?.label || placeholder,
  );

  function toggle() {
    isOpen = !isOpen;
  }

  function selectOption(val: string) {
    value = val;
    isOpen = false;
    onchange(val);
  }

  function handleClickOutside(event: MouseEvent) {
    if (container && !container.contains(event.target as Node)) {
      isOpen = false;
    }
  }

  onMount(() => {
    window.addEventListener("click", handleClickOutside);
    return () => window.removeEventListener("click", handleClickOutside);
  });
</script>

<div class="relative group {isOpen ? 'z-50' : 'z-20'}" bind:this={container}>
  {#if label}
    <span
      class="block text-[10px] uppercase font-black text-white/40 mb-2 ml-1 tracking-widest"
    >
      {label}
    </span>
  {/if}

  <button
    type="button"
    onclick={toggle}
    class="w-full h-12 bg-surface-darker/50 border border-white/10 rounded-xl px-4 flex items-center justify-between text-sm text-white focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer shadow-lg shadow-black/20"
  >
    <div class="flex items-center gap-3">
      {#if IconComponent}
        <span
          class="text-white/20 group-focus-within:text-primary transition-colors"
        >
          <IconComponent size={18} />
        </span>
      {/if}
      <span class={value ? "text-white" : "text-white/30"}>
        {selectedLabel}
      </span>
    </div>

    <span
      class="material-symbols-outlined text-white/20 transition-transform duration-300 {isOpen
        ? 'rotate-180'
        : ''}"
    >
      expand_more
    </span>
  </button>

  {#if isOpen}
    <div
      in:fade={{ duration: 150 }}
      class="absolute z-50 w-full mt-2 py-2 bg-surface-dark/95 border border-white/10 rounded-2xl shadow-[0_20px_50px_rgba(0,0,0,0.5)] overflow-hidden"
    >
      <div class="max-h-60 overflow-y-auto custom-scrollbar">
        {#each options as option}
          <button
            type="button"
            onclick={() => selectOption(option.value)}
            class="w-full px-4 py-2.5 text-left text-sm transition-all flex items-center justify-between group/opt {value ===
            option.value
              ? 'bg-primary/20 text-primary font-bold'
              : 'text-white/60 hover:bg-white/5 hover:text-white'}"
          >
            {option.label}
            {#if value === option.value}
              <span class="material-symbols-outlined text-lg">check_circle</span
              >
            {/if}
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
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.2);
  }
</style>
