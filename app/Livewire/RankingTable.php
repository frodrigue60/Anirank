<?php

namespace App\Livewire;

use Livewire\Component;
use Livewire\Attributes\On;
use App\Models\Song;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use Illuminate\Database\Eloquent\Collection;

class RankingTable extends Component
{
    use Traits\HasRankingScore;

    public $currentSection = 'ALL';
    public $perPage = 15;
    public $page = 1;
    public $hasMorePages = true;
    public $readyToLoad = false;

    public function loadData()
    {
        $this->readyToLoad = true;
    }

    public function mount()
    {
        $this->currentSection = 'ALL';
    }

    public function updatedCurrentSection()
    {
        $this->page = 1;
        $this->hasMorePages = true;
    }

    #[On('loadMore')]
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
                $query->where('status', $status);
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
        return view('livewire.ranking-table', [
            'songs' => $this->songs,
        ]);
    }
}
