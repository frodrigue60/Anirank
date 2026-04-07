<script lang="ts">
  import api from "$lib/api";

  let {
    value = $bindable(""),
    endpoint,
    placeholder = "Add tags...",
    entityName = "Item",
  } = $props<{
    value: string;
    endpoint: string;
    placeholder?: string;
    entityName?: string;
  }>();

  interface Tag {
    name: string;
    isKnown: boolean;
  }

  const initialNames = value
    ? value
        .split(",")
        .map((s: string) => s.trim())
        .filter(Boolean)
    : [];

  let tags = $state<Tag[]>(
    initialNames.map((name: string) => ({ name, isKnown: true })),
  );
  let inputValue = $state("");
  let suggestions = $state<any[]>([]);
  let showSuggestions = $state(false);
  let isSearching = $state(false);
  let searchTimeout: any;
  let focusedIndex = $state(-1);
  let inputRef: HTMLInputElement;

  // Keep parent value in sync with tags
  $effect(() => {
    value = tags.map((t) => t.name).join(", ");
  });

  // Search API
  async function handleSearch() {
    if (inputValue.length < 2) {
      suggestions = [];
      showSuggestions = false;
      return;
    }

    isSearching = true;
    try {
      const res = await api.get(
        `${endpoint}?search=${encodeURIComponent(inputValue)}&limit=5`,
      );
      suggestions = res.data.data || [];
      showSuggestions = suggestions.length > 0;
      focusedIndex = -1;
    } catch (err) {
      console.error("Search error:", err);
      suggestions = [];
    } finally {
      isSearching = false;
    }
  }

  function debounceSearch() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(handleSearch, 300);
  }

  // Quick check if a specific exact name exists
  async function checkExists(name: string): Promise<boolean> {
    try {
      const res = await api.get(
        `${endpoint}?search=${encodeURIComponent(name)}&limit=1`,
      );
      const results = res.data.data || [];
      return results.some(
        (a: any) => a.name.toLowerCase() === name.toLowerCase(),
      );
    } catch {
      return false;
    }
  }

  // Add a tag to the list
  async function addTag(name: string, isKnownOverride?: boolean) {
    name = name.trim();
    if (!name) return;

    // Prevent duplicates
    if (tags.some((t) => t.name.toLowerCase() === name.toLowerCase())) {
      inputValue = "";
      suggestions = [];
      showSuggestions = false;
      return;
    }

    let isKnown = isKnownOverride;
    if (isKnown === undefined) {
      isKnown = await checkExists(name);
    }

    tags = [...tags, { name, isKnown }];
    inputValue = "";
    suggestions = [];
    showSuggestions = false;
    focusedIndex = -1;
    inputRef?.focus();
  }

  function removeTag(index: number) {
    tags = tags.filter((_, i) => i !== index);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      if (
        showSuggestions &&
        focusedIndex >= 0 &&
        focusedIndex < suggestions.length
      ) {
        addTag(suggestions[focusedIndex].name, true);
      } else {
        addTag(inputValue);
      }
    } else if (e.key === "Backspace" && inputValue === "" && tags.length > 0) {
      removeTag(tags.length - 1);
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      if (showSuggestions && suggestions.length > 0) {
        focusedIndex = (focusedIndex + 1) % suggestions.length;
      }
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      if (showSuggestions && suggestions.length > 0) {
        focusedIndex =
          focusedIndex <= 0 ? suggestions.length - 1 : focusedIndex - 1;
      }
    } else if (e.key === "Escape") {
      showSuggestions = false;
    }
  }

  async function handlePaste(e: ClipboardEvent) {
    e.preventDefault();
    const pasted = e.clipboardData?.getData("text") || "";
    if (!pasted) return;

    const newNames = pasted
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    const checks = await Promise.all(
      newNames.map(async (name) => {
        const isKnown = await checkExists(name);
        return { name, isKnown };
      }),
    );

    let updatedTags = [...tags];
    for (const check of checks) {
      if (
        !updatedTags.some(
          (t) => t.name.toLowerCase() === check.name.toLowerCase(),
        )
      ) {
        updatedTags.push(check);
      }
    }
    tags = updatedTags;
  }
</script>

<div class="relative w-full">
  <div
    class="w-full bg-surface-highest border border-outline-variant rounded-xl p-2 text-on-surface flex flex-wrap gap-2 items-center focus-within:border-primary transition-all min-h-[44px]"
  >
    {#each tags as tag, i}
      <div
        class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-sm font-medium border
          {tag.isKnown
          ? 'bg-primary/10 text-primary border-primary/20'
          : 'bg-surface-highest text-on-surface-variant border-outline-variant'}"
        title={tag.isKnown
          ? `Existing ${entityName}`
          : `New ${entityName} will be created`}
      >
        <span>{tag.name}</span>
        <button
          type="button"
          aria-label="Remove tag"
          onclick={() => removeTag(i)}
          class="hover:text-on-surface transition-colors flex items-center justify-center opacity-70 hover:opacity-100"
        >
          <svg
            class="w-3.5 h-3.5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>
    {/each}

    <input
      bind:this={inputRef}
      type="text"
      bind:value={inputValue}
      oninput={debounceSearch}
      onkeydown={handleKeydown}
      onpaste={handlePaste}
      onfocus={() => {
        if (suggestions.length > 0) showSuggestions = true;
      }}
      placeholder={tags.length === 0 ? placeholder : "Add more..."}
      class="flex-1 min-w-[120px] bg-transparent outline-none placeholder-on-surface-variant/40 text-sm py-1"
    />
  </div>

  {#if isSearching}
    <div class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none">
      <div
        class="animate-spin h-4 w-4 border-2 border-primary border-t-transparent rounded-full"
      ></div>
    </div>
  {/if}

  {#if showSuggestions && suggestions.length > 0}
    <ul
      class="absolute z-50 w-full mt-2 bg-surface-highest border border-outline-variant rounded-xl shadow-2xl overflow-hidden"
      onblur={() => setTimeout(() => (showSuggestions = false), 200)}
    >
      {#each suggestions as suggest, i}
        <li>
          <button
            type="button"
            onclick={() => addTag(suggest.name, true)}
            class="w-full text-left px-4 py-3
                   {i === focusedIndex
              ? 'bg-primary/20 text-primary'
              : 'hover:bg-primary-container/20 hover:text-primary'} 
                   transition-all flex items-center gap-3 border-b border-outline-variant last:border-0"
          >
            <span class="text-sm font-medium">{suggest.name}</span>
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</div>
