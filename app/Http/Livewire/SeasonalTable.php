<?php

namespace App\Http\Livewire;

use Livewire\Component;
use Livewire\WithPagination;
use App\Models\Song;
use App\Models\Season;
use App\Models\Year;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;

class SeasonalTable extends Component
{
    use WithPagination;

    public $currentSection = 'OP'; // OP or ED
    public $perPage = 15;
    public $page = 1;
    public $hasMorePages = true;

    protected $listeners = ['loadMore'];

    public function mount()
    {
        $this->currentSection = 'OP';
    }

    public function updatedCurrentSection()
    {
        $this->resetPage();
        $this->hasMorePages = true;
    }

    public function loadMore()
    {
        if ($this->hasMorePages) {
            $this->perPage += 15;
        }
    }

    public function render()
    {
        $currentSeason = Season::where('current', true)->first();
        $currentYear = Year::where('current', true)->first();
        $user = Auth::user();

        $query = Song::with(['post', 'artists'])
            ->where('type', $this->currentSection);

        // Filter by Current Season and Year
        if ($currentSeason && $currentYear) {
            $query->whereHas('post', function ($q) use ($currentSeason, $currentYear) {
                $q->where('status', true)
                    ->where('season_id', $currentSeason->id)
                    ->where('year_id', $currentYear->id);
            });
        } else {
            // Fallback if no current season is set, reasonably shouldn't happen but good for safety
            $query->where('id', 0);
        }

        $songsCount = $query->count();

        // Sorting: Default to Average Rating for the leaderboard look, 
        // or could be 'title' if we strictly follow the old controller. 
        // Given the "Ranking" aesthetic, score makes sense, but let's stick to the RankingTable logic which sorts by score.
        $songs = $query->get()
            ->sortByDesc('averageRating')
            ->values() // Reset keys for correct ranking numbers
            ->take($this->perPage);

        if ($songs->count() >= $songsCount) {
            $this->hasMorePages = false;
        }

        // Calculate User Scores
        $songs->each(function ($song) use ($user) {
            $song->userScore = null;
            if ($user) {
                $userRating = DB::table('ratings')
                    ->where('rateable_type', Song::class)
                    ->where('rateable_id', $song->id)
                    ->where('user_id', $user->id)
                    ->first(['rating']);

                if ($userRating) {
                    $song->userScore = round($userRating->rating);
                }
            }
        });

        return view('livewire.seasonal-table', [
            'songs' => $songs,
            'currentSeason' => $currentSeason,
            'currentYear' => $currentYear
        ]);
    }
}
