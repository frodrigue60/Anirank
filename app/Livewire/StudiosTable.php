<?php

namespace App\Livewire;

use App\Models\Studio;
use Livewire\Component;
use Livewire\WithPagination;
use Livewire\Attributes\Url;
use Illuminate\Support\Facades\Auth;

class StudiosTable extends Component
{
    use WithPagination;

    #[Url(except: '')]
    public $search = '';
    
    #[Url(except: 'name_asc')]
    public $sort = 'name_asc';
    
    public $perPage = 18;
    public $hasMorePages = false;
    public $readyToLoad = false;

    public function loadData()
    {
        $this->readyToLoad = true;
    }

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
        if ($this->hasMorePages && $this->readyToLoad) {
            $this->perPage += 12;
        }
    }

    public function render()
    {
        if (!$this->readyToLoad) {
            return view('livewire.studios-table', [
                'studios' => collect(),
            ]);
        }

        $studiosQuery = Studio::query()
            ->withCount(['posts' => function ($query) {
                if (!Auth::check() || !Auth::user()->isStaff()) {
                    $query->where('status', true);
                }
            }])
            ->whereHas('posts', function ($query) {
                if (!Auth::check() || !Auth::user()->isStaff()) {
                    $query->where('status', true);
                }
            })
            ->with(['posts' => function ($query) {
                if (!Auth::check() || !Auth::user()->isStaff()) {
                    $query->where('status', true);
                }
                $query->select(['posts.id', 'posts.banner', 'posts.thumbnail']);
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
            });

        $results = $studiosQuery->take($this->perPage + 1)->get();
        $this->hasMorePages = $results->count() > $this->perPage;
        $studios = $results->take($this->perPage);

        return view('livewire.studios-table', [
            'studios' => $studios,
        ]);
    }
}
