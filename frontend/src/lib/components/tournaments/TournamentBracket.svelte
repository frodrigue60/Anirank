<script lang="ts">
  import type { Tournament } from "$lib/types/tournament";
  import type { Song } from "$lib/types/song";
  import MatchupCard from "./MatchupCard.svelte";

  interface Props {
    tournament: Tournament;
    userVotedMatchupIds?: (number | string)[];
    canVote?: boolean;
    onvoteRequest?: (
      matchupId: number | string,
      songId: number | string,
      song: Song | undefined,
    ) => void;
    onpreview?: (song: Song | undefined) => void;
  }

  let {
    tournament,
    userVotedMatchupIds = [],
    canVote = false,
    onvoteRequest,
    onpreview,
  }: Props = $props();

  // Group matchups by round
  let rounds = $derived(
    tournament.matchups
      ? [
          ...new Set(tournament.matchups.map((m) => m.round)),
        ].sort((a, b) => b - a)
      : [],
  );

  function getMatchupsForRound(round: number) {
    return (
      tournament.matchups
        ?.filter((m) => m.round === round)
        .sort((a, b) => a.position - b.position) || []
    );
  }
</script>

<div class="w-full overflow-x-auto py-10 no-scrollbar">
  <div class="flex gap-20 min-w-max pb-8 px-4">
    {#each rounds as round}
      <div class="flex flex-col gap-8 min-w-[340px]">
        <div class="flex flex-col items-center gap-1 mb-4">
          <h3 class="text-xs font-black uppercase tracking-[0.3em] text-primary">
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
          <div class="w-8 h-1 bg-primary/20 rounded-full"></div>
        </div>

        <div class="flex flex-col justify-around h-full gap-12">
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
  /* Optional: hide scrollbars for cleaner look but keep scrolling functional */
  .no-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .no-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
</style>
