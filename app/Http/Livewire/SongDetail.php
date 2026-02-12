<?php

namespace App\Http\Livewire;

use Livewire\Component;
use App\Models\Song;
use App\Models\Post;
use App\Models\SongVariant;
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
    public $commentBody = '';
    public $replyBody = '';
    public $replyingTo = null;
    public $editingCommentId = null;
    public $editingBody = '';

    protected $rules = [
        'commentBody' => 'required|min:3|max:1000',
        'replyBody' => 'required|min:3|max:1000',
        'newPlaylistName' => 'required|min:3|max:50',
    ];

    // Playlist State
    public $showPlaylistModal = false;
    public $userPlaylists = [];
    public $newPlaylistName = '';

    // Rating State
    public $showRatingModal = false;
    public $ratingValue = 0;

    public function mount(Song $song, Post $post)
    {
        $this->song = $song;
        $this->post = $post;

        // Load initial data
        $this->loadVariant();
        $this->loadComments();
        $this->loadRelated();
        $this->calculateScore();
    }

    public function hydrate()
    {
        $this->calculateScore();
    }

    private function calculateScore()
    {
        $user = Auth::check() ? Auth::user() : null;
        $song = $this->song;

        // Default values
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

        // Simple string format for now, or replicate formatScoreString if complex
        $song->scoreString = $song->formattedScore . '/' . $denominator;
    }

    public function loadVariant()
    {
        // Get the first variant (lowest version number) or the one specified via query/logic if we were doing that
        // For now, default to the first one.
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
        // Get other songs from the same post
        $this->relatedSongs = $this->post->songs()
            ->where('id', '!=', $this->song->id)
            ->with(['artists'])
            ->get();
    }

    public function switchVariant($variantId)
    {
        $this->currentVariant = $this->song->songVariants->find($variantId);
        $this->dispatchBrowserEvent('video-changed', ['src' => $this->getVideoUrl()]);
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
                $this->dispatchBrowserEvent('toast', ['type' => 'info', 'message' => "Removed $typeName"]);
            } else {
                $existingReaction->update(['type' => $type]);
                $this->dispatchBrowserEvent('toast', ['type' => 'success', 'message' => ucfirst($typeName) . "d the song"]);
            }
        } else {
            $this->song->reactions()->create([
                'user_id' => $userId,
                'type' => $type
            ]);
            $this->dispatchBrowserEvent('toast', ['type' => 'success', 'message' => ucfirst($typeName) . "d the song"]);
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
            $this->dispatchBrowserEvent('toast', ['type' => 'info', 'message' => 'Removed from favorites']);
        } else {
            $this->song->favorites()->create([
                'user_id' => $userId
            ]);
            $this->dispatchBrowserEvent('toast', ['type' => 'success', 'message' => 'Added to favorites!']);
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
            $this->validate(['replyBody' => 'required|min:3|max:1000']);

            $this->song->comments()->create([
                'user_id' => Auth::id(),
                'content' => $this->replyBody,
                'parent_id' => $this->replyingTo
            ]);

            $this->replyBody = '';
            $this->replyingTo = null;
        } else {
            $this->validate(['commentBody' => 'required|min:3|max:1000']);

            $this->song->comments()->create([
                'user_id' => Auth::id(),
                'content' => $this->commentBody
            ]);

            $this->commentBody = '';
        }

        $this->loadComments();
        $this->dispatchBrowserEvent('comment-posted');
        $this->dispatchBrowserEvent('toast', ['type' => 'success', 'message' => 'Comment posted successfully!']);
    }

    public function deleteComment($commentId)
    {
        if (!Auth::check()) return redirect()->route('login');

        $comment = Comment::find($commentId);

        if ($comment && ($comment->user_id === Auth::id() || Auth::user()->isAdmin())) {
            // Delete the comment (Recursive deletion is usually handled by cascade in DB or manually if needed)
            // If DB cascade is not set, we might need to delete replies first, but most systems use cascade.
            $comment->delete();
            $this->loadComments();
            $this->dispatchBrowserEvent('toast', ['type' => 'success', 'message' => 'Comment deleted.']);
        }
    }

    public function startEditing($commentId)
    {
        if (!Auth::check()) return redirect()->route('login');

        $comment = Comment::find($commentId);
        if ($comment && ($comment->user_id === Auth::id() || Auth::user()->isAdmin())) {
            $this->editingCommentId = $commentId;
            $this->editingBody = $comment->content;
            $this->cancelReply(); // Close reply if open
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
            $this->validate(['editingBody' => 'required|min:3|max:1000']);

            $comment->update(['content' => $this->editingBody]);

            $this->editingCommentId = null;
            $this->editingBody = '';
            $this->loadComments();
            $this->dispatchBrowserEvent('toast', ['type' => 'success', 'message' => 'Comment updated!']);
        }
    }

    // Playlist Logic
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
        $this->validate(['newPlaylistName' => 'required|min:3|max:50']);

        $playlist = Auth::user()->playlists()->create([
            'name' => $this->newPlaylistName,
            'is_public' => true
        ]);

        $playlist->songs()->attach($this->song->id, ['position' => 1]);

        $this->newPlaylistName = '';
        $this->openPlaylistModal(); // Reload lists
        $this->dispatchBrowserEvent('toast', ['type' => 'success', 'message' => 'Playlist created successfully!']);
    }

    public function togglePlaylist($playlistId)
    {
        $playlist = Auth::user()->playlists()->find($playlistId);

        if ($playlist->songs()->where('song_id', $this->song->id)->exists()) {
            $playlist->songs()->detach($this->song->id);
            $this->dispatchBrowserEvent('toast', ['type' => 'info', 'message' => 'Removed from playlist']);
        } else {
            $maxPos = $playlist->songs()->max('position') ?? 0;
            $playlist->songs()->attach($this->song->id, ['position' => $maxPos + 1]);
            $this->dispatchBrowserEvent('toast', ['type' => 'success', 'message' => 'Added to playlist']);
        }

        $this->openPlaylistModal(); // Reload status
    }

    // Rating Logic
    public function openRatingModal()
    {
        if (!Auth::check()) return redirect()->route('login');

        // Load current rating
        $rating = $this->song->ratings()->where('user_id', Auth::id())->first();
        $this->ratingValue = $rating ? $rating->rating : 0;

        $this->showRatingModal = true;
    }

    public function rate($value)
    {
        if (!Auth::check()) return redirect()->route('login');

        $this->song->rateOnce($value, Auth::id());
        $this->calculateScore(); // Update displayed score
        $this->showRatingModal = false;

        $this->dispatchBrowserEvent('toast', ['type' => 'success', 'message' => 'Song rated successfully!']);
    }

    public function render()
    {
        return view('livewire.song-detail');
    }
}
