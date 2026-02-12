<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Facades\Auth;
use App\Models\Reaction;

class Comment extends Model
{
    use HasFactory;

    protected $fillable = ['content', 'user_id', 'parent_id', 'commentable_id', 'commentable_type'];

    public function commentable()
    {
        return $this->morphTo();
    }

    // Relación para respuestas (hijos)
    public function replies()
    {
        return $this->hasMany(Comment::class, 'parent_id')->with('replies', 'user'); // Eager loading para anidar respuestas
    }

    // Relación para el comentario padre (opcional)
    public function parent()
    {
        return $this->belongsTo(Comment::class, 'parent_id');
    }

    public function user()
    {
        return $this->belongsTo(User::class);
    }

    public function reactions()
    {
        return $this->morphMany(Reaction::class, 'reactable');
    }

    // Relación con el contador de reacciones (nombre corregido)
    public function reactionsCounter()
    {
        return $this->morphOne(ReactionCounter::class, 'reactable');
    }

    // Método para actualizar los contadores
    public function updateReactionCounters()
    {
        $likesCount = $this->reactions()->where('type', 1)->count();
        $dislikesCount = $this->reactions()->where('type', -1)->count();

        $this->reactionsCounter()->updateOrCreate(
            ['reactable_id' => $this->id, 'reactable_type' => self::class],
            ['likes_count' => $likesCount, 'dislikes_count' => $dislikesCount]
        );
    }

    public function liked()
    {
        if (Auth::check()) { // Verifica si el usuario está autenticado
            return $this->reactions()
                ->where('user_id', Auth::id())
                ->where('type', 1)
                ->exists();
        }
        return false;
    }

    // Método para verificar si el usuario actual ha dado dislike
    public function disliked()
    {
        if (Auth::check()) { // Verifica si el usuario está autenticado
            return $this->reactions()
                ->where('user_id', Auth::id())
                ->where('type', -1)
                ->exists();
        }
        return false;
    }

    public function getLikesCountAttribute()
    {
        return $this->likes()->count();
    }

    public function getDislikesCountAttribute()
    {
        return $this->dislikes()->count();
    }

    public function likes()
    {
        return $this->reactions()->where('type', 1);
    }

    public function dislikes()
    {
        return $this->reactions()->where('type', -1);
    }
}
