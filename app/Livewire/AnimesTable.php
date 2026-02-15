<?php

namespace App\Livewire;

use Livewire\Component;
use Livewire\WithPagination;
use Livewire\Attributes\Url;
use Livewire\Attributes\On;
use App\Models\Post;
use App\Models\Season;
use App\Models\Year;
use App\Models\Format;

class AnimesTable extends Component
{
    use WithPagination;

    #[Url(except: '')]
    public $name = '';
    
    #[Url(except: '')]
    public $year_id = '';
    
    #[Url(except: '')]
    public $season_id = '';
    
    #[Url(except: '')]
    public $format_id = '';

    #[Url(except: 'grid_small')]
    public $viewMode = 'grid_small';
    
    public $perPage = 15;
    public $page = 1;
    public $hasMorePages = false;
    public $readyToLoad = false;

    public function loadData()
    {
        $this->readyToLoad = true;
    }

    public function mount()
    {
        $this->viewMode = 'grid_small';
    }

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
    }

    #[On('loadMore')]
    public function loadMore()
    {
        if ($this->hasMorePages && $this->readyToLoad) {
            $this->perPage += 15;
        }
    }

    public function render()
    {
        if (!$this->readyToLoad) {
            return view('livewire.animes-table', [
                'posts' => collect(),
                'years' => collect(),
                'seasons' => collect(),
                'formats' => collect(),
            ]);
        }

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

        $results = $query->with(['format:id,name', 'season:id,name', 'year:id,name', 'studios:id,name'])
            ->withCount('songs')
            ->orderBy('title')
            ->take($this->perPage + 1)
            ->get();

        $this->hasMorePages = $results->count() > $this->perPage;
        $posts = $results->take($this->perPage);

        $years = Year::orderBy('name', 'desc')->get(['id', 'name']);
        $seasons = Season::all(['id', 'name']);
        $formats = Format::all(['id', 'name']);

        return view('livewire.animes-table', [
            'posts' => $posts,
            'years' => $years,
            'seasons' => $seasons,
            'formats' => $formats,
        ]);
    }
}
