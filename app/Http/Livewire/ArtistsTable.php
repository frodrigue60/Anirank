<?php

namespace App\Http\Livewire;

use Livewire\Component;
use Livewire\WithPagination;
use App\Models\Artist;
use Illuminate\Support\Facades\Auth;

class ArtistsTable extends Component
{
    use WithPagination;

    public $name = '';
    public $perPage = 24;
    public $hasMorePages = false;
    public $sortBy = 'A-Z';
    public $sortByThemes = 'Most Themes';
    public $readyToLoad = false;

    protected $queryString = [
        'name' => ['except' => ''],
        'sortBy' => ['except' => 'A-Z'],
        'sortByThemes' => ['except' => 'Most Themes'],
    ];

    public function loadData()
    {
        $this->readyToLoad = true;
    }

    public function updatingName()
    {
        $this->resetPage();
    }

    public function updatingSortBy()
    {
        $this->resetPage();
    }

    public function updatingSortByThemes()
    {
        $this->resetPage();
    }

    public function loadMore()
    {
        if ($this->hasMorePages && $this->readyToLoad) {
            $this->perPage += 24;
        }
    }

    public function clearFilters()
    {
        $this->reset(['name', 'sortBy', 'sortByThemes', 'perPage']);
    }

    public function render()
    {
        if (!$this->readyToLoad) {
            return view('livewire.artists-table', [
                'artists' => collect(),
                'total' => 0,
            ]);
        }

        $query = Artist::query()
            ->select(['id', 'name', 'slug', 'thumbnail'])
            ->withCount('songs');

        if ($this->name) {
            $query->where('name', 'LIKE', '%' . $this->name . '%');
        }

        // Sorting logic
        if ($this->sortBy === 'A-Z') {
            $query->orderBy('name', 'asc');
        } elseif ($this->sortBy === 'Z-A') {
            $query->orderBy('name', 'desc');
        } elseif ($this->sortBy === 'most_themes') {
            $query->orderBy('songs_count', 'desc');
        } elseif ($this->sortBy === 'least_themes') {
            $query->orderBy('songs_count', 'asc');
        }

        $results = $query->take($this->perPage + 1)->get();
        $this->hasMorePages = $results->count() > $this->perPage;
        $artists = $results->take($this->perPage);

        return view('livewire.artists-table', [
            'artists' => $artists,
        ]);
    }
}
