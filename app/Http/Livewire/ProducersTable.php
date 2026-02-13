<?php

namespace App\Http\Livewire;

use App\Models\Producer;
use Livewire\Component;
use Livewire\WithPagination;

class ProducersTable extends Component
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
        $producers = Producer::query()
            ->withCount(['posts' => function ($query) {
                if (!auth()->check() || !auth()->user()->isStaff()) {
                    $query->where('status', true);
                }
            }])
            ->whereHas('posts', function ($query) {
                if (!auth()->check() || !auth()->user()->isStaff()) {
                    $query->where('status', true);
                }
            })
            ->with(['posts' => function ($query) {
                if (!auth()->check() || !auth()->user()->isStaff()) {
                    $query->where('status', true);
                }
            }])
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
                $query->orderBy('posts_count', 'desc');
            })
            ->when($this->sort === 'least_series', function ($query) {
                $query->orderBy('posts_count', 'asc');
            })
            ->paginate($this->perPage);

        return view('livewire.producers-table', [
            'producers' => $producers,
        ]);
    }
}
