<?php

namespace App\Http\Livewire;

use Livewire\Component;
use Livewire\WithPagination;
use App\Models\Artist;

class ArtistsTable extends Component
{
    use WithPagination;

    public $name = '';
    public $perPage = 24;
    public $hasMorePages = true;
    public $sortBy = 'A-Z';
    public $sortByThemes = 'Most Themes';

    protected $queryString = [
        'name' => ['except' => ''],
        'sortBy' => ['except' => 'A-Z'],
        'sortByThemes' => ['except' => 'Most Themes'],
    ];

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
        $this->perPage += 24;
    }

    public function clearFilters()
    {
        $this->reset(['name', 'sortBy', 'sortByThemes', 'perPage']);
    }

    public function render()
    {
        $query = Artist::query()->withCount('songs');

        if ($this->name) {
            $query->where('name', 'LIKE', '%' . $this->name . '%');
        }

        // Sorting logic
        if ($this->sortBy === 'A-Z') {
            $query->orderBy('name', 'asc');
        }
        if ($this->sortBy === 'Z-A') {
            $query->orderBy('name', 'desc');
        }
        if ($this->sortBy === 'most_themes') {
            $query->orderBy('songs_count', 'desc');
        }
        if ($this->sortBy === 'least_themes') {
            $query->orderBy('songs_count', 'asc');
        }

        $total = $query->count();

        $artists = $query->take($this->perPage)
            ->get();

        if ($artists->count() >= $total) {
            $this->hasMorePages = false;
        } else {
            $this->hasMorePages = true;
        }

        return view('livewire.artists-table', [
            'artists' => $artists,
            'total' => $total,
        ]);
    }
}
