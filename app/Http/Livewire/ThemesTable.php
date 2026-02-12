<?php

namespace App\Http\Livewire;

use Livewire\Component;
use Livewire\WithPagination;
use App\Models\Song;
use App\Models\Year;
use App\Models\Season;
use App\Models\Format;
use Illuminate\Support\Facades\Storage;

class ThemesTable extends Component
{
    use WithPagination;

    public $name = '';
    public $type = '';
    public $year_id = '';
    public $season_id = '';
    public $sort = 'recent';

    public $perPage = 18;
    public $hasMorePages = true;

    protected $queryString = [
        'name' => ['except' => ''],
        'type' => ['except' => ''],
        'year_id' => ['except' => ''],
        'season_id' => ['except' => ''],
        'sort' => ['except' => 'recent'],
    ];

    public function updatingName()
    {
        $this->resetPage();
        $this->perPage = 18; // Reset limit on filter change
    }

    public function updatingType()
    {
        $this->resetPage();
        $this->perPage = 18;
    }

    public function updatingYearId()
    {
        $this->resetPage();
        $this->perPage = 18;
    }

    public function updatingSeasonId()
    {
        $this->resetPage();
        $this->perPage = 18;
    }

    public function updatingSort()
    {
        $this->resetPage();
        $this->perPage = 18;
    }

    public function loadMore()
    {
        $this->perPage += 18;
    }

    public function render()
    {
        $query = Song::with(['post', 'artists'])
            ->withAvg('ratings', 'rating')
            ->whereHas('post', function ($q) {
                $q->where('status', true);

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
                // Sorting by related model column requires join or custom logic
                // Using simpler approach for now: sort by collection (less efficient) or join
                // For efficiency/standard paginator, doing join is better
                $query->join('posts', 'songs.post_id', '=', 'posts.id')
                    ->orderBy('posts.title');
                break;
            case 'averageRating':
                $query->orderByDesc('averageRating');
                break;
            case 'view_count':
                $query->orderByDesc('view_count');
                break;
            case 'likeCount':
                $query->orderByDesc('likeCount');
                break;
            case 'recent':
            default:
                $query->orderByDesc('created_at');
                break;
        }

        // Select explicitly to avoid column conflicts with join
        if ($this->sort === 'title') {
            $query->select('songs.*');
        }

        // Get total count for infinite scroll check
        // Cloning query before take() is not needed since we haven't applied take() yet
        $total = $query->count();

        $songs = $query->take($this->perPage)->get();

        if ($songs->count() >= $total) {
            $this->hasMorePages = false;
        } else {
            $this->hasMorePages = true;
        }

        // Processing thumbnails manually if needed or handling in blade
        $songs->each(function ($song) {
            $song->thumbnailUrl = $song->post->thumbnail_src;
            if ($song->post->thumbnail && Storage::disk('public')->exists($song->post->thumbnail)) {
                $song->thumbnailUrl = Storage::url($song->post->thumbnail);
            }

            // Generate URL
            $song->url = route('songs.show', [$song->post->slug, $song->slug]);

            // Score String (simplified logic from controller)
            $song->scoreString = $song->averageRating ? round($song->averageRating, 1) . '/100' : null;
        });

        return view('livewire.themes-table', [
            'songs' => $songs,
            'years' => Year::orderBy('name', 'desc')->get(),
            'seasons' => Season::all(),
            'types' => [
                ['name' => 'All', 'value' => 'all'],
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
