<?php

namespace App\Http\Livewire;

use App\Models\Post;
use App\Models\Year;
use App\Models\Season;
use App\Models\Format;
use Livewire\Component;
use Livewire\WithPagination;
use Illuminate\Support\Facades\Auth;
use App\Models\Studio;

class StudioAnimesTable extends Component
{
    public $studioId;
    public $studio;
    public $name = '';
    public $year_id = '';
    public $season_id = '';
    public $format_id = '';
    public $perPage = 18;
    public $viewMode = 'grid_small'; // grid_small, grid_large, list
    public $hasMorePages = false;
    public $readyToLoad = false;

    protected $queryString = [
        'name' => ['except' => ''],
        'year_id' => ['except' => ''],
        'season_id' => ['except' => ''],
        'format_id' => ['except' => ''],
        'viewMode' => ['except' => 'grid_small'],
    ];

    public function mount($studioId)
    {
        $this->studioId = $studioId;
        $this->studio = Studio::findOrFail($studioId);
    }

    public function loadData()
    {
        $this->readyToLoad = true;
    }

    public function updatingName()
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

    public function updatingFormatId()
    {
        $this->resetPage();
    }

    public function loadMore()
    {
        if ($this->readyToLoad) {
            $this->perPage += 12;
        }
    }

    public function setViewMode($mode)
    {
        $this->viewMode = $mode;
    }

    public function render()
    {
        if (!$this->readyToLoad) {
            return view('livewire.studio-animes-table', [
                'posts' => collect(),
                'years' => collect(),
                'seasons' => collect(),
                'formats' => collect(),
            ]);
        }

        $posts = Post::query()
            ->when(!Auth::check() || !Auth::user()->isStaff(), function ($query) {
                $query->where('status', true);
            })
            ->whereHas('studios', function ($query) {
                $query->where('studios.id', $this->studioId);
            })
            ->when($this->name, function ($query) {
                $query->where('title', 'like', '%' . $this->name . '%');
            })
            ->when($this->year_id, function ($query) {
                $query->where('year_id', $this->year_id);
            })
            ->when($this->season_id, function ($query) {
                $query->where('season_id', $this->season_id);
            })
            ->when($this->format_id, function ($query) {
                $query->where('format_id', $this->format_id);
            })
            ->with([
                'format:id,name',
                'season:id,name',
                'year:id,name',
                'studios:id,name',
                'producers:id,name'
            ])
            ->orderBy('title', 'asc')
            ->paginate($this->perPage);

        $this->hasMorePages = $posts->hasMorePages();

        return view('livewire.studio-animes-table', [
            'posts' => $posts,
            'years' => Year::orderBy('name', 'desc')->get(['id', 'name']),
            'seasons' => Season::all(['id', 'name']),
            'formats' => Format::all(['id', 'name']),
        ]);
    }
}
