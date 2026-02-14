<?php

namespace App\Http\Livewire;

use Livewire\Component;
use Livewire\WithPagination;
use App\Models\Song;
use Illuminate\Support\Facades\Auth;
use App\Models\Year;
use App\Models\Season;
use App\Models\Artist;
use Illuminate\Support\Facades\Storage;

class ArtistThemesTable extends Component
{
    use WithPagination;

    public $artist;
    public $name = '';
    public $type = '';
    public $year_id = '';
    public $season_id = '';
    public $sort = 'recent';

    public $perPage = 18;
    public $hasMorePages = false;

    protected $queryString = [
        'name' => ['except' => ''],
        'type' => ['except' => ''],
        'year_id' => ['except' => ''],
        'season_id' => ['except' => ''],
        'sort' => ['except' => 'recent'],
    ];

    public function mount(Artist $artist)
    {
        $this->artist = $artist;
    }

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
            $this->perPage += 18;
        }
    }

    public function clearFilters()
    {
        $this->reset(['name', 'type', 'year_id', 'season_id', 'sort', 'perPage']);
    }

    public function render()
    {
        if (!$this->readyToLoad) {
            return view('livewire.artist-themes-table', [
                'songs' => collect(),
                'years' => collect(),
                'seasons' => collect(),
                'types' => [],
                'sortMethods' => []
            ]);
        }

        $query = Song::query()
            ->with(['post:id,title,slug,banner,banner_src,thumbnail,thumbnail_src', 'artists:id,name'])
            ->withAvg('ratings', 'rating')
            ->whereHas('artists', function ($q) {
                $q->where('artists.id', $this->artist->id);
            })
            ->whereHas('post', function ($q) {
                if (!Auth::check() || !Auth::user()->isStaff()) {
                    $q->where('status', true);
                }

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

        // Sorting Logic
        switch ($this->sort) {
            case 'title':
                $query->join('posts', 'songs.post_id', '=', 'posts.id')
                    ->orderBy('posts.title');
                break;
            case 'averageRating':
                $query->orderByDesc('ratings_avg_rating');
                break;
            case 'view_count':
                $query->orderByDesc('view_count');
                break;
            case 'likeCount':
                $query->orderByDesc('likeCount');
                break;
            case 'recent':
            default:
                $query->orderByDesc('songs.created_at');
                break;
        }

        if ($this->sort === 'title') {
            $query->select('songs.*');
        }

        $songs = $query->paginate($this->perPage);
        $this->hasMorePages = $songs->hasMorePages();

        return view('livewire.artist-themes-table', [
            'songs' => $songs,
            'years' => Year::orderBy('name', 'desc')->get(['id', 'name']),
            'seasons' => Season::all(['id', 'name']),
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
