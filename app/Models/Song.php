<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use App\Models\Post;
use App\Models\SongVariant;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use App\Models\Rateable;
use Illuminate\Support\Facades\Session;

class Song extends Model
{
    use HasFactory;
    use Rateable;

    protected $fillable = [
        'id',
        'song_romaji',
        'song_jp',
        'song_en',
        'theme_num',
        'type',
        'slug',
        'views'
    ];

    public const TYPE_OPENING = 'OP';
    public const TYPE_ENDING = 'ED';

    protected static function boot()
    {
        parent::boot();

        static::deleting(function ($song) {

            $song->load('songVariants.video');

            foreach ($song->songVariants as $variant) {
                if ($variant->video) {

                    $video = $variant->video;

                    if ($video->video_src != null && Storage::disk('public')->exists($video->video_src)) {
                        Storage::disk('public')->delete($video->video_src);
                    }
                }
            }
        });
    }

    public function year()
    {
        return $this->belongsTo(Year::class);
    }

    public function season()
    {
        return $this->belongsTo(Season::class);
    }

    public function post()
    {
        return $this->belongsTo(Post::class);
    }

    public function artists()
    {
        return $this->belongsToMany(Artist::class);
    }

    public function videos()
    {
        return $this->hasMany(Video::class)->whereNull('song_variant_id');
    }

    public function songVariants()
    {
        return $this->hasMany(SongVariant::class);
    }

    public function comments()
    {
        return $this->morphMany(Comment::class, 'commentable');
    }

    public function incrementViews()
    {
        $key = 'song_' . $this->id;

        if (!Session::has($key)) {
            DB::table('songs')
                ->where('id', $this->id)
                ->increment('views');

            Session::put($key, true);
        }
    }

    public function getNameAttribute()
    {
        if ($this->song_romaji != null) {
            return $this->song_romaji;
        } elseif ($this->song_en != null) {
            return $this->song_en;
        } elseif ($this->song_jp != null) {
            return $this->song_jp;
        } else {
            return 'n/a';
        }
    }

    public function getUrlAttribute()
    {
        if (!$this->relationLoaded('post')) {
            $this->load(['post']);
        }
        return route('songs.show', [
            'anime_slug' => $this->post->slug,
            'song_slug' => $this->slug
        ]);
    }

    public function getUrlFirstVariantAttribute()
    {
        // Cargar relaciones necesarias si no están ya cargadas
        if (!$this->relationLoaded('post') || !$this->post->relationLoaded('songs')) {
            $this->load(['post', 'songVariants']);
        }

        $smallestVariant = $this->post->songs->flatMap(function ($song) {
            return $song->songVariants;
        })->sortBy('version_number')->first();

        return route('variants.show', [
            'anime_slug' => $this->post->slug,
            'song_slug' => $this->slug,
            'variant_slug' => $smallestVariant->slug
        ]);
    }

    public function reactions()
    {
        return $this->morphMany(Reaction::class, 'reactable');
    }

    public function likes()
    {
        return $this->reactions()->where('type', 1);
    }

    public function dislikes()
    {
        return $this->reactions()->where('type', -1);
    }

    public function getLikesCountAttribute()
    {
        return $this->likes()->count();
    }

    public function getDislikesCountAttribute()
    {
        return $this->dislikes()->count();
    }

    // Método para verificar si el usuario actual ha dado like
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

    public function getViewsStringAttribute()
    {
        if ($this->views >= 1000000) {
            $views = number_format(intval($this->views / 1000000), 0) . 'M';
        } elseif ($this->views >= 1000) {
            $views = number_format(intval($this->views / 1000), 0) . 'K';
        } else {
            $views = $this->views;
        }

        return $views;
    }

    // Método para obtener los posts que un usuario ha dado like
    public function scopeWhereLikedBy($query, $userId)
    {
        return $query->whereHas('favorites', function ($q) use ($userId) {
            $q->where('user_id', $userId);
        });
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

    // Relación polimórfica para favoritos, retorna array con la relacion
    public function favorites()
    {
        return $this->morphMany(Favorite::class, 'favoritable');
    }

    // retorna el la cantidad de veces que ha sido marcado como favorito
    public function getFavoritesCountAttribute()
    {
        return $this->favorites()->count();
    }

    // Método para verificar si el usuario actual ha marcado este post como favorito
    public function isFavorited()
    {
        if (Auth::check()) {
            return $this->favorites()->where('user_id', Auth::id())->exists();
        }
        return false;
    }

    public function playlists()
    {
        return $this->belongsToMany(Playlist::class);
    }

    public function isInPlaylist($playlistId)
    {
        return $this->playlists()->where('playlist_id', $playlistId)->exists();
    }

    /* public function getUserRatingAttribute($song_id)
    {
        $user = Auth::user();
        $userRating = DB::table('ratings')
            ->where('rateable_type', $this::class)
            ->where('rateable_id', $song_id)
            ->where('user_id', $user->id)
            ->first(['rating']);

        return $userRating;
    } */
}
