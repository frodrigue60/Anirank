<?php

namespace App\Http\Livewire;

use App\Models\Studio;
use Livewire\Component;
use Livewire\WithPagination;

class StudiosTable extends Component
{
    use WithPagination;

    public $search = '';
    public $sort = 'name_asc';
    public $perPage = 18;

    protected $queryString = [
        'search' => ['except' => ''],
        'sort' => ['except' => 'name_asc'],
    ];

    public function updatingSearch()
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
        $studios = Studio::query()
            ->withCount('post')
            ->when($this->search, function ($query) {
                $query->where('name', 'like', '%' . $this->search . '%');
            })
            ->when($this->sort === 'name_asc', function ($query) {
                $query->orderBy('name', 'asc');
            })
            ->when($this->sort === 'name_desc', function ($query) {
                $query->orderBy('name', 'desc');
            })
            ->when($this->sort === 'most_series', function ($query) {
                $query->orderBy('post_count', 'desc');
            })
            ->when($this->sort === 'least_series', function ($query) {
                $query->orderBy('post_count', 'asc');
            })
            ->paginate($this->perPage);

        return view('livewire.studios-table', [
            'studios' => $studios,
        ]);
    }
}
