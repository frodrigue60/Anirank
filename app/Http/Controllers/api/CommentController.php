<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use App\Models\Comment;
use App\Models\User;
use App\Models\Reaction;

class CommentController extends Controller
{
    public function destroy($id)
    {
        $comment = Comment::findOrFail($id);
        try {
            $user = Auth::check() ? Auth::user() : null;

            if (($comment->user_id == $user->id) || $user->isAdmin()) {
                $comment->delete();
            }

            return response()->json([
                'message' => 'Comment deleted successfully',
                'success' => true,
            ]);
        } catch (\Throwable $th) {
            return response()->json([
                'message' => 'Ah error has been occurred',
                'success' => false,
                'th' => $th
            ]);
        }
    }

    public function like($comment_id)
    {
        try {
            $comment = Comment::findOrFail($comment_id);
            $this->handleReaction($comment, 1);
            $comment->updateReactionCounters();

            return response()->json([
                'success' => true,
                'comment' => $comment,
                'likesCount' => $comment->likesCount,
                'dislikesCount' => $comment->dislikesCount
            ]);
        } catch (\Throwable $th) {
            return response()->json(['success' => false, 'error' => $th]);
        }
    }

    public function dislike($comment_id)
    {
        try {
            $comment = Comment::findOrFail($comment_id);
            $this->handleReaction($comment, -1);
            $comment->updateReactionCounters();

            return response()->json([
                'success' => true,
                'comment' => $comment,
                'likesCount' => $comment->likesCount,
                'dislikesCount' => $comment->dislikesCount
            ]);
        } catch (\Throwable $th) {
            return response()->json(['success' => false, 'error' => $th]);
        }
    }

    private function handleReaction($comment, $type)
    {
        $user = Auth::user();

        $reaction = Reaction::where('user_id', $user->id)
            ->where('reactable_id', $comment->id)
            ->where('reactable_type', Comment::class)
            ->first();

        if ($reaction) {
            if ($reaction->type === $type) {
                $reaction->delete();
            } else {
                $reaction->update(['type' => $type]);
            }
        } else {
            Reaction::create([
                'user_id' => $user->id,
                'reactable_id' => $comment->id,
                'reactable_type' => Comment::class,
                'type' => $type,
            ]);
        }
    }

    public function reply(Request $request, Comment $parentComment)
    {
        try {
            $request->validate(['content' => 'required|string']);

            $reply = new Comment();
            $reply->content = $request->content;
            $reply->user_id = Auth::User()->id;
            $reply->parent_id = $parentComment->id;
            $reply->commentable_type = $parentComment->commentable_type;
            $reply->commentable_id = $parentComment->commentable_id;

            $reply->save();

            return response()->json([
                'success' => true,
                'html' => view('partials.songs.show.comments.comment', ['comment' => $reply])->render(),
                'parentComment' => $reply->parent_id
            ]);
        } catch (\Throwable $th) {
            return response()->json([
                'error' => $th->getMessage(),
                'request' => $request->all()
            ]);
        }
    }
}
