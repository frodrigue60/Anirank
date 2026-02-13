<?php

namespace App\Http\Livewire;

use Livewire\Component;
use App\Models\Song;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\Log;

class GlobalBottomPlayer extends Component
{
    public $song;
    public $isVisible = false;
    public $isPlaying = false;
    public $progress = 0;
    public $currentTime = '0:00';
    public $duration = '0:00';
    public $hasVideoLoaded = false;

    protected $listeners = ['playSong' => 'handlePlaySong'];

    public function handlePlaySong($songId)
    {
        Log::info("GlobalBottomPlayer: Requested song ID $songId");
        $this->song = Song::with(['post', 'artists', 'firstSongVariant.video'])->find($songId);

        if (!$this->song) {
            Log::error("GlobalBottomPlayer: Song not found $songId");
            return;
        }

        $this->calculateScore();
        $this->isPlaying = true;
        $this->isVisible = true;
        $this->hasVideoLoaded = true;

        // Satisfy user requested logic: check first variant then fall back to direct videos
        $video = ($this->song->firstSongVariant && $this->song->firstSongVariant->video)
            ? $this->song->firstSongVariant->video
            : $this->song->videos()->first();

        if ($video) {
            $videoUrl = $video->video_src;
            if ($video->isLocal()) {
                if (Storage::disk('public')->exists($video->video_src)) {
                    $videoUrl = Storage::url($video->video_src);
                } else {
                    $videoUrl = Storage::url($video->video_src); // Fallback
                }
            } else if ($video->isEmbed()) {
                $videoUrl = $video->video_src;
            }

            Log::info("GlobalBottomPlayer: Loading video of type {$video->type} with URL: $videoUrl");

            $this->dispatchBrowserEvent('song-loaded', [
                'url' => $videoUrl,
                'type' => $video->type, // 'file' or 'embed'
                'title' => $this->song->name,
                'anime' => $this->song->post->title,
                'artists' => $this->song->artists->pluck('name')->join(', '),
                'thumbnail' => $this->song->thumbnail_src
            ]);
        } else {
            Log::warning("GlobalBottomPlayer: No video found for song ID $songId");
        }
    }

    private function calculateScore()
    {
        if (!$this->song) return;

        $user = Auth::check() ? Auth::user() : null;
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

        $this->song->formattedScore = $isDecimalFormat
            ? round($this->song->averageRating * $factor, 1)
            : (int) round($this->song->averageRating * $factor);
    }

    public function togglePlay()
    {
        $this->isPlaying = !$this->isPlaying;
        $this->dispatchBrowserEvent('toggle-playback', ['playing' => $this->isPlaying]);
    }

    public function render()
    {
        return view('livewire.global-bottom-player');
    }
}
