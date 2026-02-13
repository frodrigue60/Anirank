<?php

namespace App\Http\Livewire;

use Livewire\Component;
use App\Models\Song;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use Illuminate\Database\Eloquent\Collection;

class RankingTable extends Component
{
    public $currentSection = 'ALL'; // ALL, OP or ED
    public $perPage = 15;
    public $page = 1;
    public $hasMorePages = true;

    protected $listeners = ['loadMore'];

    public function mount()
    {
        $this->currentSection = 'ALL';
    }


    public function updatedCurrentSection()
    {
        $this->page = 1;
        $this->hasMorePages = true;
    }

    public function loadMore()
    {
        if (!$this->hasMorePages) return;
        $this->page++;
    }

    public function toggleFavorite($songId)
    {
        if (!Auth::check()) {
            return $this->emit('showLoginModal'); // O el nombre del evento para login si existe
        }

        $song = Song::find($songId);
        if ($song) {
            $song->toggleFavorite();
        }
    }

    public function getSongsProperty()
    {
        $status = true;
        $limit = 100;
        $perPage = $this->perPage * $this->page;

        $query = Song::with(['post', 'artists'])
            ->whereHas('post', function ($query) use ($status) {
                $query->where('status', $status);
            });

        if ($this->currentSection !== 'ALL') {
            $query->where('type', $this->currentSection);
        }


        $songsCount = $query->count();
        $songs = $query->get()
            ->sortByDesc('averageRating')
            ->values()
            ->take($perPage);

        if ($songs->count() >= $songsCount || $songs->count() >= $limit) {
            $this->hasMorePages = false;
        }

        return $this->setScoreSongs($songs, Auth::user());
    }

    private function setScoreSongs($songs, $user = null)
    {
        $songs->each(function ($song) use ($user) {
            $song->userScore = null;
            $factor = 1;
            $denominator = 100;
            $isDecimalFormat = false;

            if ($user) {
                switch ($user->score_format) {
                    case 'POINT_100':
                        $factor = 1;
                        $denominator = 100;
                        break;
                    case 'POINT_10_DECIMAL':
                        $factor = 0.1;
                        $denominator = 10;
                        $isDecimalFormat = true;
                        break;
                    case 'POINT_10':
                        $factor = 1 / 10;
                        $denominator = 10;
                        break;
                    case 'POINT_5':
                        $factor = 1 / 20;
                        $denominator = 5;
                        $isDecimalFormat = true;
                        break;
                }

                $userRating = DB::table('ratings')
                    ->where('rateable_id', $song->id)
                    ->where('rateable_type', Song::class)
                    ->where('user_id', $user->id)
                    ->first(['rating']);

                if ($userRating) {
                    $song->userScore = $isDecimalFormat
                        ? round($userRating->rating * $factor, 1)
                        : (int) round($userRating->rating * $factor);
                }
            }

            $song->score = $isDecimalFormat
                ? round($song->averageRating * $factor, 1)
                : (int) round($song->averageRating * $factor);
        });

        return $songs;
    }

    public function render()
    {
        return view('livewire.ranking-table', [
            'songs' => $this->songs,
        ]);
    }
}
