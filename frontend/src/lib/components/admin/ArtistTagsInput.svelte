<script lang="ts">
  import api from "$lib/api";

  let { value = $bindable("") } = $props<{ value: string }>();

  // Represents an individual tag (artist name) and whether it's known in the DB
  interface ArtistTag {
    name: string;
    isKnown: boolean;
  }

  const initialNames = value
    .split(",")
    .map((s: string) => s.trim())
    .filter(Boolean);

  let tags = $state<ArtistTag[]>(
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
        `/admin/artists?search=${encodeURIComponent(inputValue)}&limit=5`,
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

  // Quick check if a specific exact name exists (used for pasted tags or arbitrary adds)
  async function checkArtistExists(name: string): Promise<boolean> {
    try {
      const res = await api.get(
        `/admin/artists?search=${encodeURIComponent(name)}&limit=1`,
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

    // Prevent perfect duplicates
    if (tags.some((t) => t.name.toLowerCase() === name.toLowerCase())) {
      inputValue = "";
      suggestions = [];
      showSuggestions = false;
      return;
    }

    // Determine correctness
    let isKnown = isKnownOverride;
    if (isKnown === undefined) {
      isKnown = await checkArtistExists(name);
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

  // Handle mass pasting of comma-separated names
  async function handlePaste(e: ClipboardEvent) {
    e.preventDefault();
    const pasted = e.clipboardData?.getData("text") || "";
    if (!pasted) return;

    const newNames = pasted
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);

    // Concurrently check validation for the pasted names
    const checks = await Promise.all(
      newNames.map(async (name) => {
        const isKnown = await checkArtistExists(name);
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
    <!-- Render Tags -->
    {#each tags as tag, i}
      <div
        class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-sm font-medium border
          {tag.isKnown
          ? 'bg-green-500/10 text-green-400 border-green-500/20'
          : 'bg-zinc-800 text-zinc-300 border-zinc-700'}"
        title={tag.isKnown ? "Existing Artist" : "New Artist will be created"}
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

    <!-- Input Field -->
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
      placeholder={tags.length === 0
        ? "e.g. LiSA, YOASOBI (press Enter)"
        : "Add more..."}
      class="flex-1 min-w-[120px] bg-transparent outline-none placeholder-gray-500 text-sm py-1 font-mono"
    />
  </div>

  <!-- Loading indicator inside the input area -->
  {#if isSearching}
    <div class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none">
      <div
        class="animate-spin h-4 w-4 border-2 border-primary border-t-transparent rounded-full"
      ></div>
    </div>
  {/if}

  <!-- Suggestions Dropdown -->
  {#if showSuggestions && suggestions.length > 0}
    <ul
      class="absolute z-50 w-full mt-2 bg-zinc-900 border border-zinc-800 rounded-xl shadow-2xl overflow-hidden"
      onblur={() => setTimeout(() => (showSuggestions = false), 200)}
    >
      {#each suggestions as suggest, i}
        <li>
          <button
            type="button"
            onclick={() => addTag(suggest.name, true)}
            class="w-full text-left px-4 py-3
                   {i === focusedIndex
              ? 'bg-blue-600/20 text-blue-400'
              : 'hover:bg-primary-container/20 hover:text-blue-400'} 
                   transition-all flex items-center gap-3 border-b border-zinc-800/50 last:border-0"
          >
            <div class="flex flex-col">
              <span class="text-sm font-medium">{suggest.name}</span>
            </div>
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</div>
