<?php

namespace App\Livewire;

use Livewire\Component;
use App\Models\Song;
use Illuminate\Support\Facades\Auth;
use Livewire\Attributes\On;

class SongInteractions extends Component
{
    public $songId;
    public $song;
    public $mode = 'bar';
    public $showRatingModal = false;
    public $ratingValue = 0;

    public function mount($songId, $mode = 'bar')
    {
        $this->songId = $songId;
        $this->mode = $mode;
        $this->loadSong();
    }

    #[On('songChanged')]
    public function handleSongChanged($songId)
    {
        $this->songId = $songId;
        $this->loadSong();
    }

    private function loadSong()
    {
        $this->song = Song::withCount(['likes', 'dislikes'])->find($this->songId);
        $this->calculateScore();

        if (Auth::check()) {
            $rating = $this->song->ratings()->where('user_id', Auth::id())->first();
            $this->ratingValue = $rating ? $rating->rating : 0;
        }
    }

    public function calculateScore()
    {
        $user = Auth::check() ? Auth::user() : null;
        $song = $this->song;

        if (!$song) return;

        $song->formattedScore = null;
        $factor = 1;
        $isDecimalFormat = false;

        if ($user) {
            switch ($user->score_format) {
                case 'POINT_100':
                    $factor = 1;
                    break;
                case 'POINT_10_DECIMAL':
                    $factor = 0.1;
                    $isDecimalFormat = true;
                    break;
                case 'POINT_10':
                    $factor = 0.1;
                    break;
                case 'POINT_5':
                    $factor = 0.05;
                    $isDecimalFormat = true;
                    break;
            }
        }

        $song->formattedScore = $isDecimalFormat
            ? round($song->averageRating * $factor, 1)
            : (int) round($song->averageRating * $factor);
    }

    public function toggleLike()
    {
        if (!Auth::check()) return redirect()->route('login');
        $this->toggleReaction(1);
    }

    public function toggleDislike()
    {
        if (!Auth::check()) return redirect()->route('login');
        $this->toggleReaction(-1);
    }

    private function toggleReaction($type)
    {
        $userId = Auth::id();
        $existingReaction = $this->song->reactions()
            ->where('user_id', $userId)
            ->first();

        $typeName = $type === 1 ? 'like' : 'dislike';

        if ($existingReaction) {
            if ($existingReaction->type == $type) {
                $existingReaction->delete();
                $this->dispatch('toast', type: 'info', message: "Removed $typeName");
            } else {
                $existingReaction->update(['type' => $type]);
                $this->dispatch('toast', type: 'success', message: ucfirst($typeName) . "d the song");
            }
        } else {
            $this->song->reactions()->create([
                'user_id' => $userId,
                'type' => $type
            ]);
            $this->dispatch('toast', type: 'success', message: ucfirst($typeName) . "d the song");
        }

        $this->loadSong();
    }

    public function toggleFavorite()
    {
        if (!Auth::check()) return redirect()->route('login');

        $userId = Auth::id();
        $existingFavorite = $this->song->favorites()
            ->where('user_id', $userId)
            ->first();

        if ($existingFavorite) {
            $existingFavorite->delete();
            $this->dispatch('toast', type: 'info', message: 'Removed from favorites');
        } else {
            $this->song->favorites()->create([
                'user_id' => $userId
            ]);
            $this->dispatch('toast', type: 'success', message: 'Added to favorites!');
        }

        $this->loadSong();
    }

    public function openRatingModal()
    {
        if (!Auth::check()) return redirect()->route('login');
        $this->showRatingModal = true;
    }

    public function rate($value = null)
    {
        if (!Auth::check()) return redirect()->route('login');

        try {
            $value = $value ?? $this->ratingValue;

            if ($value < 0 || $value > 100) {
                throw new \Exception('Invalid rating value.');
            }

            $this->song->rateOnce($value, Auth::id());
            $this->loadSong();
            $this->showRatingModal = false;

            $this->dispatch(
                'toast',
                type: 'success',
                message: 'Rating Saved!',
                description: "You rated {$this->song->name} with {$value} points."
            );
        } catch (\Exception $e) {
            $this->dispatch(
                'toast',
                type: 'error',
                message: 'Error saving rating',
                description: $e->getMessage()
            );
        }
    }

    public function render()
    {
        return view('livewire.song-interactions');
    }
}
