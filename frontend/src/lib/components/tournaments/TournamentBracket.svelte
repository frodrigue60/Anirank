<script lang="ts">
  import type { Tournament } from '$lib/types/tournament';
  import type { Song } from '$lib/types/song';
  import MatchupCard from './MatchupCard.svelte';

  interface Props {
    tournament: Tournament;
    userVotedMatchupIds?: number[];
    canVote?: boolean;
    onvoteRequest?: (matchupId: number, songId: number, song: Song | undefined) => void;
    onpreview?: (song: Song | undefined) => void;
  }

  let { 
    tournament, 
    userVotedMatchupIds = [], 
    canVote = false,
    onvoteRequest,
    onpreview
  }: Props = $props();

  // Group matchups by round
  let rounds = $derived(tournament.matchups ? [...new Set(tournament.matchups.map(m => m.round))].sort((a, b) => b - a) : []);

  function getMatchupsForRound(round: number) {
    return tournament.matchups?.filter(m => m.round === round).sort((a, b) => a.position - b.position) || [];
  }
</script>

<div class="tournament-bracket">
  <div class="rounds-container">
    {#each rounds as round}
      <div class="round-column">
        <h3 class="round-title">
          {#if round === 2}
            Final
          {:else if round === 4}
            Semifinals
          {:else if round === 8}
            Quarterfinals
          {:else}
            Round of {round}
          {/if}
        </h3>
        
        <div class="matchups-list">
          {#each getMatchupsForRound(round) as matchup}
            <MatchupCard 
              {matchup} 
              canVote={canVote && !userVotedMatchupIds.includes(matchup.id)}
              {onvoteRequest}
              {onpreview}
            />
          {/each}
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
  .tournament-bracket {
    width: 100%;
    overflow-x: auto;
    padding: 20px 0;
  }

  .rounds-container {
    display: flex;
    gap: 40px;
    min-width: max-content;
    padding-bottom: 20px;
  }

  .round-column {
    display: flex;
    flex-direction: column;
    gap: 20px;
    min-width: 300px;
  }

  .round-title {
    text-align: center;
    font-size: 1.1rem;
    font-weight: 700;
    margin-bottom: 10px;
    color: var(--primary-color, #ff4e50);
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .matchups-list {
    display: flex;
    flex-direction: column;
    justify-content: space-around;
    height: 100%;
    gap: 30px;
  }
</style>
