<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use App\Models\Song;
use App\Models\Video;

class SongVariant extends Model
{
    use HasFactory;

    protected $fillable = [
        'id',
        'version',
        'song_id',
        'views',
        'spoiler',
    ];

    protected static function boot()
    {
        parent::boot();

        static::deleting(function ($songVariant) {
            if ($songVariant->video) {
                $songVariant->video->delete();
            }
        });
    }

    public function song()
    {
        return $this->belongsTo(Song::class);
    }

    public function video()
    {
        return $this->hasOne(Video::class);
    }

    public function getUrlAttribute()
    {
        return route('variants.show', [
            'anime_slug' => $this->song->post->slug,
            'song_slug' => $this->song->slug,
            'variant_slug' => $this->slug,
        ]);
    }

    public function year()
    {
        return $this->belongsTo(Year::class);
    }

    public function season()
    {
        return $this->belongsTo(Season::class);
    }

    public function favorites()
    {
        return $this->morphMany(Favorite::class, 'favoritable');
    }

    public function scopeFavoritedBy($query, $userId)
    {
        return $query->whereHas('favorites', function ($q) use ($userId) {
            $q->where('user_id', $userId);
        });
    }
}
