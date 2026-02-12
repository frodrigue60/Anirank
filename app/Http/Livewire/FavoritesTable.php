<?php

namespace App\Http\Livewire;

use App\Models\Song;
use App\Models\Year;
use App\Models\Season;
use Livewire\Component;
use Livewire\WithPagination;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use Illuminate\Pagination\Paginator;
use Illuminate\Support\Collection;
use Illuminate\Pagination\LengthAwarePaginator;

class FavoritesTable extends Component
{
    use WithPagination;

    public $name = '';
    public $type = '';
    public $year_id = '';
    public $season_id = '';
    public $sort = 'recent';
    public $perPage = 15;

    protected $queryString = [
        'name' => ['except' => ''],
        'type' => ['except' => ''],
        'year_id' => ['except' => ''],
        'season_id' => ['except' => ''],
        'sort' => ['except' => 'recent'],
    ];

    public function updatingName()
    {
        $this->resetPage();
    }

    public function updatingType()
    {
        $this->resetPage();
    }

    public function updatingYearId()
    {
        $this->resetPage();
    }

    public function updatingSeasonId()
    {
        $this->resetPage();
    }

    public function updatingSort()
    {
        $this->resetPage();
    }

    public function loadMore()
    {
        $this->perPage += 12;
    }

    public function render()
    {
        $user = Auth::user();
        if (!$user) {
            return view('livewire.favorites-table', ['songs' => collect()]);
        }

        $query = Song::query()
            ->with(['post.season', 'post.year', 'artists', 'favorites'])
            ->withAvg('ratings', 'rating')
            ->whereHas('favorites', function ($q) use ($user) {
                $q->where('user_id', $user->id);
            })
            #SONG QUERY
            ->when($this->type, function ($query) {
                $query->where('type', $this->type);
            })
            #POST QUERY
            ->whereHas('post', function ($query) {
                $query->where('status', true)
                    ->when($this->name, function ($query) {
                        $query->where('title', 'LIKE', '%' . $this->name . '%');
                    })
                    ->when($this->season_id, function ($query) {
                        $query->where('season_id', $this->season_id);
                    })
                    ->when($this->year_id, function ($query) {
                        $query->where('year_id', $this->year_id);
                    });
            });

        // Apply Sorting
        switch ($this->sort) {
            case 'title':
                $query->join('posts', 'songs.post_id', '=', 'posts.id')
                    ->orderBy('posts.title', 'asc')
                    ->select('songs.*');
                break;
            case 'averageRating':
                $query->orderBy('ratings_avg_rating', 'desc');
                break;
            case 'view_count':
                $query->orderBy('views', 'desc');
                break;
            case 'likeCount':
                // For simplicity, using average rating if like count complex
                $query->orderBy('ratings_avg_rating', 'desc');
                break;
            case 'recent':
            default:
                $query->orderBy('created_at', 'desc');
                break;
        }

        $songs = $query->paginate($this->perPage);

        // Prep data for view (scoring etc)
        $this->prepareSongs($songs, $user);

        return view('livewire.favorites-table', [
            'songs' => $songs,
            'years' => Year::orderBy('name', 'desc')->get(),
            'seasons' => Season::all(),
            'sortMethods' => [
                ['name' => 'Recent', 'value' => 'recent'],
                ['name' => 'Title', 'value' => 'title'],
                ['name' => 'Score', 'value' => 'averageRating'],
                ['name' => 'Views', 'value' => 'view_count'],
            ],
            'types' => [
                ['name' => 'Opening', 'value' => 'OP'],
                ['name' => 'Ending', 'value' => 'ED'],
                ['name' => 'Insert', 'value' => 'INS'],
                ['name' => 'Other', 'value' => 'OTH']
            ]
        ]);
    }

    protected function prepareSongs($songs, $user)
    {
        $songs->each(function ($song) use ($user) {
            $song->formattedScore = null;
            $song->scoreString = null;
            $factor = 1;
            $isDecimalFormat = false;
            $denominator = 100;

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
            }

            $avgRating = $song->ratings_avg_rating ?? 0;
            $song->formattedScore = $isDecimalFormat
                ? round($avgRating * $factor, 1)
                : (int) round($avgRating * $factor);

            $song->scoreString = $song->formattedScore . '/' . $denominator;

            // Thumbnail logic
            $song->thumbnailUrl = $song->post->thumbnail_src;
            if ($song->post->thumbnail && \Storage::disk('public')->exists($song->post->thumbnail)) {
                $song->thumbnailUrl = \Storage::url($song->post->thumbnail);
            }
        });
    }
}
