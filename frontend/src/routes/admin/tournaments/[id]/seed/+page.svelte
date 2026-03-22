<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import api from '$lib/api';
  import { toastState } from '$lib/state/toast.svelte';
  import type { Tournament } from '$lib/types/tournament';
  import type { Song } from '$lib/types/song';

  const tournamentId = $page.params.id;
  let tournament: Tournament | null = null;
  let years: any[] = [];
  let seasons: any[] = [];
  let genres: any[] = [];
  
  let step = 1;
  let loading = true;
  let method: 'random' | 'filtered' | 'manual' = 'random';
  
  // Filters for Step 1
  let selectedYear: number | null = null;
  let selectedSeason: number | null = null;
  let selectedGenre: number | null = null;
  let selectedType: string | null = null;
  let sort: 'rating' | 'random' = 'rating';

  // Manual Selection (Step 2)
  let searchQuery = '';
  let searchResults: Song[] = [];
  let selectedSongs: Song[] = [];
  let searching = false;
  let previewSong: Song | null = null;

  onMount(async () => {
    try {
      const [tRes, yRes, sRes, gRes] = await Promise.all([
        api.get(`/admin/tournaments/${tournamentId}`),
        api.get('/admin/taxonomies/years'),
        api.get('/admin/taxonomies/seasons'),
        api.get('/admin/genres')
      ]);
      
      tournament = tRes.data.data;
      if (!tournament) throw new Error('Tournament not found');

      // Default type from tournament filter if available
      if (tournament.type_filter) {
        selectedType = tournament.type_filter.toLowerCase();
      }
      
      years = yRes.data.data;
      seasons = sRes.data.data;
      genres = gRes.data.data;

      // Load saved progress from localStorage
      const saved = localStorage.getItem(`tournament_seed_${tournamentId}`);
      if (saved) {
        const data = JSON.parse(saved);
        selectedSongs = data.selectedSongs || [];
        method = data.method || 'random';
        // maybe add a "Resume" button if they want to
      }
    } catch (error) {
      console.error('Error initialization:', error);
    } finally {
      loading = false;
    }
  });

  async function handleSearch() {
    if (searchQuery.length < 2) return;
    searching = true;
    try {
      const res = await api.get('/admin/songs', { params: { search: searchQuery, limit: 10, type: selectedType } });
      searchResults = res.data.data;
    } catch (error) {
      console.error('Search failed:', error);
    } finally {
      searching = false;
    }
  }

  function addSong(song: Song) {
    if (!tournament) return;
    if (selectedSongs.find(s => s.id === song.id)) return;
    if (selectedSongs.length >= tournament.size) {
      alert(`Limit reached (${tournament.size} songs)`);
      return;
    }
    selectedSongs = [...selectedSongs, song];
    saveProgress();
  }

  function removeSong(id: number) {
    selectedSongs = selectedSongs.filter(s => s.id !== id);
    saveProgress();
  }

  function saveProgress() {
    localStorage.setItem(`tournament_seed_${tournamentId}`, JSON.stringify({
      selectedSongs,
      method
    }));
    toastState.addToast('Progress saved successfully', 'success', 3000);
  }

  async function finalizeSeeding() {
    if (!tournament) return;
    if (method === 'manual' && selectedSongs.length < tournament.size) {
      alert(`Please select ${tournament.size} songs (current: ${selectedSongs.length})`);
      return;
    }

    try {
      loading = true;
      const payload: any = {
        method: method === 'manual' ? 'manual' : (method === 'filtered' ? 'filtered' : 'random'),
        manual_songs: method === 'manual' ? selectedSongs.map(s => s.id) : [],
        year_id: selectedYear,
        season_id: selectedSeason,
        genre_id: selectedGenre,
        song_type: selectedType,
        sort: sort
      };

      await api.post(`/admin/tournaments/${tournamentId}/seed`, payload);
      alert('Tournament seeded successfully!');
      localStorage.removeItem(`tournament_seed_${tournamentId}`);
      window.location.href = '/admin/tournaments';
    } catch (error: any) {
      console.error('Seeding failed:', error);
      alert(error.response?.data?.message || 'Failed to seed tournament');
    } finally {
      loading = false;
    }
  }

  function nextStep() {
    if (step === 1 && method === 'random') {
       // Skip to confirmation? Or show preview?
       // Let's just go to step 2 for all
    }
    step++;
  }

  function prevStep() {
    step--;
  }

  $: canProceed = tournament && (
    (method === 'random') ||
    (method === 'filtered') ||
    (method === 'manual' && selectedSongs.length === tournament.size)
  );

</script>

<div class="seeding-wizard p-6 max-w-6xl mx-auto">
  {#if loading && !tournament}
    <div class="text-center py-20 opacity-50">Loading tournament details...</div>
  {:else if tournament}
    <div class="header mb-8">
      <div class="flex items-center gap-4 text-sm opacity-50 mb-2">
        <a href="/admin/tournaments" class="hover:underline">Tournaments</a>
        <span>/</span>
        <span>Seed Tournament #{tournamentId}</span>
      </div>
      <h1 class="text-3xl font-bold">{tournament.name}</h1>
      <p class="opacity-60 text-lg">Seeding {tournament.size} songs</p>
    </div>

    <!-- Stepper -->
    <div class="flex items-center gap-4 mb-10 overflow-x-auto">
      {#each [1, 2, 3] as s}
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-full flex items-center justify-center font-bold {step === s ? 'bg-primary text-white' : 'bg-white/10 opacity-50'}">
            {s}
          </div>
          <span class="whitespace-nowrap {step === s ? 'font-bold' : 'opacity-50'}">
            {s === 1 ? 'Configuration' : (s === 2 ? 'Selection' : 'Confirmation')}
          </span>
          {#if s < 3}
            <div class="w-12 h-px bg-white/10"></div>
          {/if}
        </div>
      {/each}
    </div>

    <!-- Step 1: Configuration -->
    {#if step === 1}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-8 animate-in fade-in duration-300">
        <div class="space-y-6">
          <h2 class="text-xl font-bold mb-4">Select Seeding Method</h2>
          
          <label class="method-card {method === 'random' ? 'active' : ''}">
            <input type="radio" bind:group={method} value="random" class="hidden" />
            <div class="flex items-start gap-4">
              <div class="icon text-2xl">🎲</div>
              <div>
                <div class="font-bold">Automatic (Random)</div>
                <div class="text-sm opacity-60">Picks random songs based on base tournament filters.</div>
              </div>
            </div>
          </label>

          <label class="method-card {method === 'filtered' ? 'active' : ''}">
            <input type="radio" bind:group={method} value="filtered" class="hidden" />
            <div class="flex items-start gap-4">
              <div class="icon text-2xl">🪄</div>
              <div>
                <div class="font-bold">Semi-Automatic</div>
                <div class="text-sm opacity-60">Apply specific year, season or genre filters first.</div>
              </div>
            </div>
          </label>

          <label class="method-card {method === 'manual' ? 'active' : ''}">
            <input type="radio" bind:group={method} value="manual" class="hidden" />
            <div class="flex items-start gap-4">
              <div class="icon text-2xl">✍️</div>
              <div>
                <div class="font-bold">Manual Selection</div>
                <div class="text-sm opacity-60">Search and pick songs one by one. Supports saving progress.</div>
              </div>
            </div>
          </label>
        </div>

        <div class="bg-white/5 rounded-2xl p-6 h-fit {method === 'manual' ? 'opacity-30 pointer-events-none' : ''}">
          <h2 class="text-xl font-bold mb-6">Discovery Filters</h2>
          <div class="space-y-4">
            <div>
              <label for="year-select" class="block text-sm opacity-50 mb-1">Year</label>
              <select id="year-select" bind:value={selectedYear} class="w-full bg-black/40 border border-white/10 rounded-lg p-3">
                <option value={null}>Any Year</option>
                {#each years as y}
                  <option value={y.id}>{y.name}</option>
                {/each}
              </select>
            </div>
            <div>
              <label for="season-select" class="block text-sm opacity-50 mb-1">Season</label>
              <select id="season-select" bind:value={selectedSeason} class="w-full bg-black/40 border border-white/10 rounded-lg p-3">
                <option value={null}>Any Season</option>
                {#each seasons as s}
                  <option value={s.id}>{s.name}</option>
                {/each}
              </select>
            </div>
            <div>
              <label for="genre-select" class="block text-sm opacity-50 mb-1">Genre</label>
              <select id="genre-select" bind:value={selectedGenre} class="w-full bg-black/40 border border-white/10 rounded-lg p-3">
                <option value={null}>Any Genre</option>
                {#each genres as g}
                  <option value={g.id}>{g.name}</option>
                {/each}
              </select>
            </div>
            <div>
              <label for="type-select" class="block text-sm opacity-50 mb-1">Theme Type</label>
              <select id="type-select" bind:value={selectedType} class="w-full bg-black/40 border border-white/10 rounded-lg p-3">
                <option value={null}>Any Type</option>
                <option value="op">Openings (OP)</option>
                <option value="ed">Endings (ED)</option>
                <option value="ins">Insert Songs (INS)</option>
                <option value="oth">Others (OTH)</option>
              </select>
            </div>
            <div>
              <label for="priority-select" class="block text-sm opacity-50 mb-1">Priority</label>
              <div class="flex gap-2">
                <button 
                  class="flex-1 p-2 rounded-lg border {sort === 'rating' ? 'border-primary bg-primary/20' : 'border-white/10 bg-white/5'}"
                  on:click={() => sort = 'rating'}
                >
                  Top Rated
                </button>
                <button 
                  class="flex-1 p-2 rounded-lg border {sort === 'random' ? 'border-primary bg-primary/20' : 'border-white/10 bg-white/5'}"
                  on:click={() => sort = 'random'}
                >
                  Random
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="mt-12 flex justify-end">
        <button class="bg-primary hover:scale-105 transition-transform text-white px-8 py-3 rounded-xl font-bold text-lg" on:click={nextStep}>
          Continue to Selection →
        </button>
      </div>

    <!-- Step 2: Selection / Preview -->
    {:else if step === 2}
      {#if method === 'manual'}
        <div class="manual-panel grid grid-cols-1 lg:grid-cols-3 gap-8 ">
          <!-- Left: Search & Results -->
          <div class="lg:col-span-2 space-y-6">
            <div class="relative">
              <input 
                type="text" 
                bind:value={searchQuery}
                on:input={handleSearch}
                placeholder="Search by song name, anime or artist..."
                class="w-full bg-white/5 border border-white/10 rounded-2xl p-4 pl-12 text-lg focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all"
              />
              <span class="absolute left-4 top-1/2 -translate-y-1/2 opacity-40 text-xl">🔍</span>
            </div>

            <div class="results-list bg-white/5 rounded-2xl overflow-hidden min-h-[400px]">
              {#if searching}
                <div class="p-8 text-center opacity-50">Searching songs...</div>
              {:else if searchResults.length > 0}
                {#each searchResults as song}
                  <div class="song-row p-4 border-b border-white/5 flex items-center gap-4 hover:bg-white/10 transition-colors">
                    <img src={song.anime?.cover_url || '/placeholder.jpg'} alt="" class="w-12 h-16 object-cover rounded shadow" />
                    <div class="flex-1">
                      <div class="text-[10px] text-primary font-bold uppercase tracking-wider mb-0.5">{song.anime?.title}</div>
                      <div class="font-bold flex items-center gap-2">
                        {song.song_romaji}
                        <span class="bg-white/10 px-1.5 py-0.5 rounded text-[10px] font-bold uppercase">{song.type}{song.theme_num}</span>
                      </div>
                      {#if song.song_variants && song.song_variants.length > 0}
                        <div class="flex gap-1 mt-1">
                          {#each song.song_variants as variant}
                            <span class="text-[9px] opacity-40 bg-white/5 px-1 rounded">v{variant.version_number}</span>
                          {/each}
                        </div>
                      {/if}
                    </div>
                    <button 
                      type="button"
                      class="bg-white/10 hover:bg-white/20 p-2 rounded flex items-center justify-center w-10 h-10" 
                      on:click={() => previewSong = song}
                      title="Preview song"
                      aria-label="Preview song"
                    >
                      ▶️
                    </button>
                    <button 
                      type="button"
                      class="bg-primary hover:bg-primary/90 text-white px-4 py-2 rounded-lg font-bold disabled:opacity-30 h-10 min-w-[80px]"
                      on:click={() => addSong(song)}
                      disabled={selectedSongs.find(s => s.id === song.id) !== undefined}
                    >
                      {selectedSongs.find(s => s.id === song.id) ? 'Added' : 'Add'}
                    </button>
                  </div>
                {/each}
              {:else if searchQuery.length > 1}
                <div class="p-8 text-center opacity-50">No results found for "{searchQuery}"</div>
              {:else}
                <div class="p-8 text-center opacity-50">Start typing to find anime songs...</div>
              {/if}
            </div>
          </div>

          <!-- Right: Selection Bucket -->
          <div class="bg-white/5 border border-white/10 rounded-2xl p-6 flex flex-col h-[700px]">
            <div class="flex justify-between items-center mb-6">
              <h2 class="text-xl font-bold">Selected Themes</h2>
              <span class="bg-primary/20 text-primary px-3 py-1 rounded-full font-bold text-sm">
                {selectedSongs.length} / {tournament.size}
              </span>
            </div>

            <div class="flex-1 overflow-y-auto space-y-2 mb-6 scrollbar-thin">
              {#each selectedSongs as song, i}
                <div class="flex items-center gap-3 p-3 bg-white/5 rounded-xl group">
                  <span class="text-xs opacity-30 w-4">{i + 1}</span>
                  <div class="flex-1 truncate">
                    <div class="font-bold text-sm truncate">{song.song_romaji}</div>
                    <div class="text-[10px] opacity-40 truncate">{song.anime?.title}</div>
                  </div>
                  <button 
                    type="button"
                    class="opacity-0 group-hover:opacity-100 text-red-500 hover:scale-125 transition-all text-lg"
                    on:click={() => removeSong(song.id)}
                    title="Remove song"
                    aria-label="Remove song"
                  >
                    ×
                  </button>
                </div>
              {:else}
                <div class="h-full flex flex-col items-center justify-center opacity-30 text-center">
                  <div class="text-4xl mb-4">🧺</div>
                  <p>Your selection is empty</p>
                </div>
              {/each}
            </div>

            <div class="space-y-3">
              <button 
                type="button"
                class="w-full bg-white/10 hover:bg-white/20 py-3 rounded-xl font-bold transition-all flex items-center justify-center gap-2"
                on:click={saveProgress}
              >
                💾 Save Progress
              </button>
              <button 
                type="button"
                class="w-full bg-primary py-4 rounded-xl font-bold text-lg disabled:opacity-40 disabled:cursor-not-allowed hover:scale-[1.02] transition-all"
                disabled={selectedSongs.length < tournament.size}
                on:click={nextStep}
              >
                Next Step →
              </button>
            </div>
          </div>
        </div>
      {:else}
        <!-- Preview Case for Auto/Semi -->
        <div class="text-center py-20 bg-white/5 rounded-3xl border border-dashed border-white/20">
          <div class="text-5xl mb-6">⚡</div>
          <h2 class="text-2xl font-bold mb-2">Ready to Generate</h2>
          <p class="opacity-60 mb-8 max-w-md mx-auto">
            The system will automatically find {tournament.size} songs matching your criteria.
            Method: <span class="text-primary font-bold">{method === 'random' ? 'Pure Random' : 'Filtered Discovery'}</span>
          </p>
          <div class="flex justify-center gap-4">
            <button type="button" class="px-8 py-3 rounded-xl border border-white/10 hover:bg-white/5" on:click={prevStep}>Back</button>
            <button type="button" class="px-8 py-3 rounded-xl bg-primary font-bold" on:click={nextStep}>Final Confirmation →</button>
          </div>
        </div>
      {/if}

    <!-- Step 3: Confirmation -->
    {:else if step === 3}
      <div class="max-w-2xl mx-auto space-y-8 animate-in zoom-in-95 duration-500">
        <div class="bg-white/5 border border-white/10 rounded-3xl p-8">
          <h2 class="text-2xl font-bold mb-8 flex items-center gap-3">
            <span class="text-green-500">✅</span> Final Review
          </h2>
          
          <div class="space-y-6">
            <div class="flex justify-between border-b border-white/5 pb-4">
              <span class="opacity-50">Tournament</span>
              <span class="font-bold">{tournament.name}</span>
            </div>
            <div class="flex justify-between border-b border-white/5 pb-4">
              <span class="opacity-50">Size</span>
              <span class="font-bold">{tournament.size} Themes</span>
            </div>
            <div class="flex justify-between border-b border-white/5 pb-4">
              <span class="opacity-50">Method</span>
              <span class="font-bold uppercase text-primary">{method}</span>
            </div>
            
            {#if method !== 'manual'}
              <div class="space-y-3">
                <span class="opacity-50 block text-sm">Active Filters</span>
                <div class="flex flex-wrap gap-2">
                  {#if selectedYear} <span class="chip">Year: {years.find(y => y.id === selectedYear)?.name}</span> {/if}
                  {#if selectedSeason} <span class="chip">Season: {seasons.find(s => s.id === selectedSeason)?.name}</span> {/if}
                  {#if selectedGenre} <span class="chip">Genre: {genres.find(g => g.id === selectedGenre)?.name}</span> {/if}
                  <span class="chip italic">Sorting: {sort}</span>
                </div>
              </div>
            {/if}
          </div>

          <p class="text-sm opacity-40 mt-8 leading-relaxed">
            By clicking start, the tournament status will change to <span class="text-white">Active</span>. 
            The first round matchups will be generated immediately and users can start voting.
          </p>
        </div>

        <div class="flex items-center gap-4">
          <button 
             type="button"
             class="flex-1 py-4 rounded-2xl border border-white/10 hover:bg-white/5 font-bold transition-all"
             on:click={prevStep}
             disabled={loading}
          >
            ← Back
          </button>
          <button 
             type="button"
             class="flex-2 py-4 rounded-2xl bg-primary text-white font-bold text-xl shadow-lg shadow-primary/20 hover:scale-[1.02] active:scale-95 transition-all disabled:opacity-50"
             on:click={finalizeSeeding}
             disabled={loading}
          >
            {loading ? 'Starting Tournament...' : 'START TOURNAMENT 🚀'}
          </button>
        </div>
      </div>
    {/if}

    <!-- Preview Modal -->
    {#if previewSong}
      <div 
        class="fixed inset-0 bg-black/80 backdrop-blur-md z-50 flex items-center justify-center p-4" 
        on:click|self={() => previewSong = null}
        on:keydown={(e) => e.key === 'Escape' && (previewSong = null)}
        role="button"
        tabindex="0"
        aria-label="Close preview"
      >
        <div class="bg-[#111] border border-white/10 rounded-3xl overflow-hidden max-w-2xl w-full shadow-2xl relative animate-in zoom-in-95 duration-300">
          <button 
            type="button"
            class="absolute right-6 top-6 text-2xl opacity-50 hover:opacity-100 z-10 w-10 h-10 flex items-center justify-center bg-black/40 rounded-full" 
            on:click={() => previewSong = null}
            title="Close preview"
            aria-label="Close preview"
          >×</button>
          
          <div class="aspect-video w-full bg-black relative">
            {#if previewSong.song_variants && previewSong.song_variants.length > 0}
               {@const v = previewSong.song_variants[0].video}
               {#if v}
                 {#if v.type === 'file'}
                   <!-- svelte-ignore a11y-media-has-caption -->
                   <video src={v.local_url} controls autoplay class="w-full h-full object-contain"></video>
                 {:else if v.type === 'embed'}
                   <iframe src={v.embed_url} title="Video Preview" frameborder="0" allow="autoplay; encrypted-media" allowfullscreen class="w-full h-full"></iframe>
                 {/if}
               {:else}
                 <div class="w-full h-full flex items-center justify-center opacity-40">No preview video available</div>
               {/if}
            {:else}
              <div class="w-full h-full flex items-center justify-center opacity-40">No variants or videos available</div>
            {/if}
          </div>

          <div class="p-8">
            <div class="flex gap-6 mb-6">
               <img src={previewSong.anime?.cover_url} alt="" class="w-24 h-32 object-cover rounded-xl shadow-2xl" />
               <div class="flex-1 pt-2">
                 <h3 class="text-2xl font-bold mb-1">{previewSong.song_romaji}</h3>
                 <p class="text-primary font-bold mb-4">{previewSong.anime?.title}</p>
                 <div class="flex flex-wrap gap-2">
                   <span class="text-xs bg-white/10 px-2 py-1 rounded uppercase tracking-wider font-bold">{previewSong.type}{previewSong.theme_num}</span>
                   {#if previewSong.song_variants && previewSong.song_variants.length > 0}
                     <span class="text-xs bg-white/10 px-2 py-1 rounded uppercase tracking-wider font-bold">Version {previewSong.song_variants[0].version_number}</span>
                   {/if}
                 </div>
               </div>
            </div>
            
            <div class="pt-4 border-t border-white/5 flex justify-center gap-4">
              <button type="button" class="bg-white/10 hover:bg-white/20 text-white px-8 py-3 rounded-xl font-bold transition-all" on:click={() => previewSong = null}>Close</button>
              <button type="button" class="bg-primary text-white px-8 py-3 rounded-xl font-bold shadow-lg shadow-primary/20 hover:scale-[1.05] transition-all" on:click={() => { if(previewSong) addSong(previewSong); previewSong = null; }}>
                Add Theme to Selection
              </button>
            </div>
          </div>
        </div>
      </div>
    {/if}

  {/if}
</div>

<style>
  :global(:root) {
    --primary: #ff4e50;
  }
  
  .method-card {
    display: block;
    background: rgba(255, 255, 255, 0.03);
    border: 1px border-white/5;
    padding: 1.5rem;
    border-radius: 1.25rem;
    cursor: pointer;
    border-style: solid;
    border-width: 1px;
    border-color: rgba(255,255,255,0.05);
    transition: all 0.2s ease;
  }
  
  .method-card:hover {
    background: rgba(255, 255, 255, 0.06);
    transform: translateY(-2px);
  }
  
  .method-card.active {
    background: rgba(255, 78, 80, 0.1);
    border-color: var(--primary);
  }

  .chip {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    padding: 0.25rem 0.75rem;
    border-radius: 999px;
    font-size: 0.8rem;
  }

  .scrollbar-thin::-webkit-scrollbar {
    width: 6px;
  }
  .scrollbar-thin::-webkit-scrollbar-track {
    background: transparent;
  }
  .scrollbar-thin::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 10px;
  }
  
  @keyframes fade-in {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }
  
  .animate-in {
    animation: fade-in 0.4s ease-out forwards;
  }
</style>
