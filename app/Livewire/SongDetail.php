<?php

namespace App\Livewire;

use Livewire\Component;
use Livewire\Attributes\Validate;
use App\Models\Song;
use App\Models\Post;
use App\Models\Comment;
use App\Models\Reaction;
use App\Models\Favorite;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Storage;

class SongDetail extends Component
{
    public $song;
    public $post;
    public $currentVariant;
    public $comments;
    public $relatedSongs;

    // Comment Form
    #[Validate('required|min:3|max:1000')]
    public $commentBody = '';

    #[Validate('required|min:3|max:1000')]
    public $replyBody = '';
    
    public $replyingTo = null;
    public $editingCommentId = null;

    #[Validate('required|min:3|max:1000')]
    public $editingBody = '';

    #[Validate('required|min:3|max:50')]
    public $newPlaylistName = '';

    // Playlist State
    public $showPlaylistModal = false;
    public $userPlaylists = [];

    // Rating State
    public $showRatingModal = false;
    public $ratingValue = 0;

    public function mount(Song $song, Post $post)
    {
        $this->song = $song;
        $this->post = $post;

        $this->loadVariant();
        $this->loadComments();
        $this->loadRelated();
        $this->calculateScore();
    }

    private function calculateScore()
    {
        $user = Auth::check() ? Auth::user() : null;
        $song = $this->song;

        $song->formattedScore = null;
        $song->rawScore = null;
        $song->scoreString = null;

        $factor = 1;
        $isDecimalFormat = false;
        $denominator = 100;

        if ($user) {
            switch ($user->score_format) {
                case 'POINT_100':
                    $factor = 1;
                    $denominator = 100;
                    break;
                case 'POINT_10_DECIMAL':
                    $factor = 0.1;
                    $denominator = 10;
                    $isDecimalFormat = true;
                    break;
                case 'POINT_10':
                    $factor = 0.1;
                    $denominator = 10;
                    break;
                case 'POINT_5':
                    $factor = 0.05;
                    $denominator = 5;
                    $isDecimalFormat = true;
                    break;
            }
        }

        $song->rawScore = round($song->averageRating, 1);

        $song->formattedScore = $isDecimalFormat
            ? round($song->averageRating * $factor, 1)
            : (int) round($song->averageRating * $factor);

        $song->scoreString = $song->formattedScore . '/' . $denominator;
    }

    public function loadVariant()
    {
        $this->currentVariant = $this->song->songVariants->sortBy('version_number')->first();
    }

    public function loadComments()
    {
        $this->comments = Comment::with(['user', 'replies' => function ($query) {
            $query->orderBy('created_at', 'asc');
        }, 'replies.user'])
            ->where('commentable_id', $this->song->id)
            ->where('commentable_type', Song::class)
            ->where('parent_id', null)
            ->orderByDesc('created_at')
            ->get();
    }

    public function loadRelated()
    {
        $this->relatedSongs = $this->post->songs()
            ->where('id', '!=', $this->song->id)
            ->with(['artists'])
            ->get();
    }

    public function switchVariant($variantId)
    {
        $this->currentVariant = $this->song->songVariants->find($variantId);
        $this->dispatch('video-changed', src: $this->getVideoUrl());
    }

    public function getVideoUrl()
    {
        if (!$this->currentVariant || !$this->currentVariant->video) {
            return '';
        }

        if (Storage::disk('public')->exists($this->currentVariant->video->video_src)) {
            return Storage::url($this->currentVariant->video->video_src);
        }

        return $this->currentVariant->video->video_src;
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

        $this->song->loadCount(['likes', 'dislikes']);
        $this->song->refresh();
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

        $this->song->refresh();
    }

    public function setReplyTo($commentId)
    {
        if (!Auth::check()) return redirect()->route('login');
        $this->replyingTo = $commentId;
        $this->replyBody = '';
    }

    public function cancelReply()
    {
        $this->replyingTo = null;
        $this->replyBody = '';
    }

    public function postComment()
    {
        if (!Auth::check()) return redirect()->route('login');

        if ($this->replyingTo) {
            $this->validateOnly('replyBody');

            $this->song->comments()->create([
                'user_id' => Auth::id(),
                'content' => $this->replyBody,
                'parent_id' => $this->replyingTo
            ]);

            $this->replyBody = '';
            $this->replyingTo = null;
        } else {
            $this->validateOnly('commentBody');

            $this->song->comments()->create([
                'user_id' => Auth::id(),
                'content' => $this->commentBody
            ]);

            $this->commentBody = '';
        }

        $this->loadComments();
        $this->dispatch('comment-posted');
        $this->dispatch('toast', type: 'success', message: 'Comment posted successfully!');
    }

    public function deleteComment($commentId)
    {
        if (!Auth::check()) return redirect()->route('login');

        $comment = Comment::find($commentId);

        if ($comment && ($comment->user_id === Auth::id() || Auth::user()->isAdmin())) {
            $comment->delete();
            $this->loadComments();
            $this->dispatch('toast', type: 'success', message: 'Comment deleted.');
        }
    }

    public function startEditing($commentId)
    {
        if (!Auth::check()) return redirect()->route('login');

        $comment = Comment::find($commentId);
        if ($comment && ($comment->user_id === Auth::id() || Auth::user()->isAdmin())) {
            $this->editingCommentId = $commentId;
            $this->editingBody = $comment->content;
            $this->cancelReply();
        }
    }

    public function cancelEditing()
    {
        $this->editingCommentId = null;
        $this->editingBody = '';
    }

    public function updateComment()
    {
        if (!Auth::check()) return redirect()->route('login');

        $comment = Comment::find($this->editingCommentId);
        if ($comment && ($comment->user_id === Auth::id() || Auth::user()->isAdmin())) {
            $this->validateOnly('editingBody');

            $comment->update(['content' => $this->editingBody]);

            $this->editingCommentId = null;
            $this->editingBody = '';
            $this->loadComments();
            $this->dispatch('toast', type: 'success', message: 'Comment updated!');
        }
    }

    public function openPlaylistModal()
    {
        if (!Auth::check()) return redirect()->route('login');

        $this->userPlaylists = Auth::user()->playlists()->withCount(['songs' => function ($query) {
            $query->where('song_id', $this->song->id);
        }])->get();

        $this->showPlaylistModal = true;
    }

    public function createPlaylist()
    {
        $this->validateOnly('newPlaylistName');

        $playlist = Auth::user()->playlists()->create([
            'name' => $this->newPlaylistName,
            'is_public' => true
        ]);

        $playlist->songs()->attach($this->song->id, ['position' => 1]);

        $this->newPlaylistName = '';
        $this->openPlaylistModal();
        $this->dispatch('toast', type: 'success', message: 'Playlist created successfully!');
    }

    public function togglePlaylist($playlistId)
    {
        $playlist = Auth::user()->playlists()->find($playlistId);

        if ($playlist->songs()->where('song_id', $this->song->id)->exists()) {
            $playlist->songs()->detach($this->song->id);
            $this->dispatch('toast', type: 'info', message: 'Removed from playlist');
        } else {
            $maxPos = $playlist->songs()->max('position') ?? 0;
            $playlist->songs()->attach($this->song->id, ['position' => $maxPos + 1]);
            $this->dispatch('toast', type: 'success', message: 'Added to playlist');
        }

        $this->openPlaylistModal();
    }

    public function openRatingModal()
    {
        if (!Auth::check()) return redirect()->route('login');

        $rating = $this->song->ratings()->where('user_id', Auth::id())->first();
        $this->ratingValue = $rating ? $rating->rating : 0;

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
            $this->calculateScore();
            $this->showRatingModal = false;

            $this->dispatch('toast',
                type: 'success',
                message: 'Rating Saved!',
                description: "You rated {$this->song->name} with {$value} points."
            );
        } catch (\Exception $e) {
            $this->dispatch('toast',
                type: 'error',
                message: 'Error saving rating',
                description: $e->getMessage()
            );
        }
    }

    public function render()
    {
        return view('livewire.song-detail');
    }
}
