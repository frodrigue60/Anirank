<?php

namespace App\Http\Livewire;

use Livewire\Component;
use Livewire\WithPagination;
use App\Models\Post;
use App\Models\Season;
use App\Models\Year;
use App\Models\Format;

class AnimesTable extends Component
{
    use WithPagination;

    public $name = '';
    public $year_id = '';
    public $season_id = '';
    public $format_id = '';

    public $viewMode = 'grid_small'; // grid_small, grid_large, list
    public $perPage = 15;
    public $page = 1;
    public $hasMorePages = true;

    protected $queryString = [
        'name' => ['except' => ''],
        'year_id' => ['except' => ''],
        'season_id' => ['except' => ''],
        'format_id' => ['except' => ''],
        'viewMode' => ['except' => 'grid_small'],
    ];

    protected $listeners = ['loadMore'];

    public function updatedName()
    {
        $this->resetPage();
    }
    public function updatedYearId()
    {
        $this->resetPage();
    }
    public function updatedSeasonId()
    {
        $this->resetPage();
    }
    public function updatedFormatId()
    {
        $this->resetPage();
    }

    public function setViewMode($mode)
    {
        $this->viewMode = $mode;
        // Optionally adjust perPage based on mode if needed
    }

    public function loadMore()
    {
        if ($this->hasMorePages) {
            $this->perPage += 15;
        }
    }

    public function render()
    {
        $query = Post::where('status', true);

        if ($this->name) {
            $query->where('title', 'LIKE', '%' . $this->name . '%');
        }
        if ($this->year_id) {
            $query->where('year_id', $this->year_id);
        }
        if ($this->season_id) {
            $query->where('season_id', $this->season_id);
        }
        if ($this->format_id) {
            $query->where('format_id', $this->format_id);
        }

        $total = $query->count();

        $posts = $query->with(['format', 'season', 'year', 'studios'])
            ->orderBy('title')
            ->take($this->perPage)
            ->get();

        if ($posts->count() >= $total) {
            $this->hasMorePages = false;
        } else {
            $this->hasMorePages = true;
        }

        // Fetch filter options
        $years = Year::orderBy('name', 'desc')->get();
        $seasons = Season::all();
        $formats = Format::all();

        return view('livewire.animes-table', [
            'posts' => $posts,
            'years' => $years,
            'seasons' => $seasons,
            'formats' => $formats,
        ]);
    }
}
