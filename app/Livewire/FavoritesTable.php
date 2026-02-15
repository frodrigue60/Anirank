<?php

namespace App\Livewire;

use App\Models\Song;
use App\Models\Year;
use App\Models\Season;
use Livewire\Component;
use Livewire\WithPagination;
use Livewire\Attributes\Url;
use Illuminate\Support\Facades\Auth;
use App\Livewire\Traits\HasRankingScore;

class FavoritesTable extends Component
{
    use WithPagination;
    use HasRankingScore;

    #[Url(except: '')]
    public $name = '';
    
    #[Url(except: '')]
    public $type = '';
    
    #[Url(except: '')]
    public $year_id = '';
    
    #[Url(except: '')]
    public $season_id = '';
    
    #[Url(except: 'recent')]
    public $sort = 'recent';
    
    public $perPage = 15;
    public $hasMorePages = true;
    public $readyToLoad = false;

    public function loadData()
    {
        $this->readyToLoad = true;
    }

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
        if ($this->readyToLoad) {
            $this->perPage += 12;
        }
    }

    public function render()
    {
        if (!$this->readyToLoad) {
            return view('livewire.favorites-table', [
                'songs' => collect(),
                'years' => collect(),
                'seasons' => collect(),
                'sortMethods' => [],
                'types' => []
            ]);
        }

        $user = Auth::user();
        if (!$user) {
            return view('livewire.favorites-table', ['songs' => collect()]);
        }

        $query = Song::query()
            ->with(['post:id,title,slug,banner,thumbnail,thumbnail_src,season_id,year_id', 'post.season:id,name', 'post.year:id,name', 'artists:id,name'])
            ->withAvg('ratings', 'rating')
            ->favoritedBy($user->id)
            ->when($this->type, function ($query) {
                $query->where('type', $this->type);
            })
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
            case 'recent':
            default:
                $query->orderBy('songs.created_at', 'desc');
                break;
        }

        $songs = $query->paginate($this->perPage);
        $this->hasMorePages = $songs->hasMorePages();

        $this->setScoreSongs($songs, $user);

        return view('livewire.favorites-table', [
            'songs' => $songs,
            'years' => Year::orderBy('name', 'desc')->get(['id', 'name']),
            'seasons' => Season::all(['id', 'name']),
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
}
