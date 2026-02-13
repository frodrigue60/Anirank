<?php

namespace App\Http\Livewire;

use App\Models\Post;
use App\Models\Year;
use App\Models\Season;
use App\Models\Format;
use Livewire\Component;
use Livewire\WithPagination;

class ProducerAnimesTable extends Component
{
    use WithPagination;

    public $producerId;
    public $name = '';
    public $year_id = '';
    public $season_id = '';
    public $format_id = '';
    public $perPage = 18;

    protected $queryString = [
        'name' => ['except' => ''],
        'year_id' => ['except' => ''],
        'season_id' => ['except' => ''],
        'format_id' => ['except' => ''],
    ];

    public function mount($producerId)
    {
        $this->producerId = $producerId;
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
        $this->perPage += 12;
    }

    public function render()
    {
        $posts = Post::query()
            ->when(!auth()->check() || !auth()->user()->isStaff(), function ($query) {
                $query->where('status', true);
            })
            ->whereHas('producers', function ($query) {
                $query->where('producers.id', $this->producerId);
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
            ->orderBy('title', 'asc')
            ->paginate($this->perPage);

        return view('livewire.producer-animes-table', [
            'posts' => $posts,
            'years' => Year::orderBy('name', 'desc')->get(),
            'seasons' => Season::all(),
            'formats' => Format::all(),
        ]);
    }
}
