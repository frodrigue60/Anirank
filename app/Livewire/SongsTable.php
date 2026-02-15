<?php

namespace App\Livewire;

use Livewire\Component;
use Livewire\WithPagination;
use Livewire\Attributes\Url;
use App\Models\Song;
use App\Models\Year;
use App\Models\Season;

class SongsTable extends Component
{
    use WithPagination;

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

    public $perPage = 18;
    public $hasMorePages = true;
    public $readyToLoad = false;

    public function loadData()
    {
        $this->readyToLoad = true;
    }

    public function updatingName()
    {
        $this->resetPage();
        $this->perPage = 18;
    }

    public function updatingType()
    {
        $this->resetPage();
        $this->perPage = 18;
    }

    public function updatingYearId()
    {
        $this->resetPage();
        $this->perPage = 18;
    }

    public function updatingSeasonId()
    {
        $this->resetPage();
        $this->perPage = 18;
    }

    public function updatingSort()
    {
        $this->resetPage();
        $this->perPage = 18;
    }

    public function loadMore()
    {
        $this->perPage += 18;
    }

    public function render()
    {
        if (!$this->readyToLoad) {
            return view('livewire.songs-table', [
                'songs' => collect(),
                'years' => collect(),
                'seasons' => collect(),
                'types' => [],
                'sortMethods' => []
            ]);
        }

        $query = Song::query()
            ->with(['post:id,title,slug,banner,thumbnail', 'artists:id,name'])
            ->withAvg('ratings', 'rating')
            ->withCount('likes')
            ->whereHas('post', function ($q) {
                $q->where('status', true);
                if ($this->name) {
                    $q->where('title', 'LIKE', '%' . $this->name . '%');
                }
                if ($this->season_id) {
                    $q->where('season_id', $this->season_id);
                }
                if ($this->year_id) {
                    $q->where('year_id', $this->year_id);
                }
            });

        if ($this->type && $this->type !== 'all') {
            $query->where('type', $this->type);
        }

        switch ($this->sort) {
            case 'title':
                $query->join('posts', 'songs.post_id', '=', 'posts.id')
                    ->orderBy('posts.title')
                    ->select('songs.*');
                break;
            case 'averageRating':
                $query->orderByDesc('ratings_avg_rating');
                break;
            case 'view_count':
                $query->orderByDesc('views');
                break;
            case 'likeCount':
                $query->orderByDesc('likes_count');
                break;
            case 'recent':
            default:
                $query->orderByDesc('songs.created_at');
                break;
        }

        $totalSelection = $query->count();
        $songs = $query->take($this->perPage)->get();

        $this->hasMorePages = ($songs->count() < $totalSelection);

        return view('livewire.songs-table', [
            'songs' => $songs,
            'years' => Year::select('id', 'name')->orderBy('name', 'desc')->get(),
            'seasons' => Season::select('id', 'name')->get(),
            'types' => [
                ['name' => 'Opening', 'value' => 'OP'],
                ['name' => 'Ending', 'value' => 'ED'],
                ['name' => 'Insert', 'value' => 'INS'],
                ['name' => 'Other', 'value' => 'OTH'],
            ],
            'sortMethods' => [
                ['name' => 'Recent', 'value' => 'recent'],
                ['name' => 'Title', 'value' => 'title'],
                ['name' => 'Score', 'value' => 'averageRating'],
                ['name' => 'Views', 'value' => 'view_count'],
                ['name' => 'Popular', 'value' => 'likeCount'],
            ]
        ]);
    }
}
