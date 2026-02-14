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
    use Traits\HasRankingScore;

    public $readyToLoad = false;
    public $currentSection = 'ALL'; // ALL, OP or ED
    public $perPage = 15;
    public $page = 1;
    public $hasMorePages = true;
    public $seasonId;
    public $yearId;
    public $seasonName;
    public $yearName;

    protected $listeners = ['loadMore'];

    public function loadData()
    {
        $this->readyToLoad = true;
    }

    public function mount($season, $year)
    {
        $this->seasonId = $season;
        $this->yearId = $year;
        $this->seasonName = Season::find($season)?->name;
        $this->yearName = Year::find($year)?->name;
        $this->currentSection = 'ALL';
    }

    public function updatedCurrentSection()
    {
        $this->page = 1;
        $this->hasMorePages = true;
    }

    public function loadMore()
    {
        if (!$this->hasMorePages || !$this->readyToLoad) return;
        $this->page++;
    }

    public function toggleFavorite($songId)
    {
        if (!Auth::check()) {
            return $this->dispatch('showLoginModal');
        }

        $song = Song::find($songId);
        if ($song) {
            $song->toggleFavorite();
        }
    }

    public function getSongsProperty()
    {
        if (!$this->readyToLoad) return collect();

        $status = true;
        $limit = 100;
        $perPage = $this->perPage * $this->page;

        $query = Song::query()
            ->with(['post:id,title,slug,banner,thumbnail', 'artists:id,name'])
            ->withAvg('ratings', 'rating')
            ->whereHas('post', function ($query) use ($status) {
                $query->where('status', $status)
                    ->where('season_id', $this->seasonId)
                    ->where('year_id', $this->yearId);
            });

        if ($this->currentSection !== 'ALL') {
            $query->where('type', $this->currentSection);
        }

        $query->orderByDesc('ratings_avg_rating');

        $songs = $query->take(min($perPage, $limit))->get();

        $this->hasMorePages = $songs->count() >= $perPage && $songs->count() < $limit;

        return $this->setScoreSongs($songs, Auth::user());
    }

    public function render()
    {
        return view('livewire.seasonal-table', [
            'songs' => $this->songs,
        ]);
    }
}
