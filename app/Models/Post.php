<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Support\Facades\Storage;

class Post extends Model
{
    use HasFactory;

    protected $fillable = [
        'title',
        'slug',
        'type',
        'thumbnail',
        'thumbnail_src',
        'status',
        'format',
        'banner',
        'banner_src'
    ];

    protected static function boot()
    {
        parent::boot();

        static::deleting(function ($post) {
            if ($post->thumbnail_src != null && Storage::disk('public')->exists($post->thumbnail)) {
                Storage::disk('public')->delete($post->thumbnail);
            }

            if ($post->banner_src != null && Storage::disk('public')->exists($post->banner)) {
                Storage::disk('public')->delete($post->banner);
            }
        });
    }

    public function songs()
    {
        return $this->hasMany(Song::class);
    }
    public function reports()
    {
        return $this->hasMany(Report::class);
    }

    public function getUrlAttribute()
    {
        return route('post.show', [
            'slug' => $this->slug,
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

    public function openings()
    {
        return $this->hasMany(Song::class)->where('type', Song::TYPE_OPENING);
    }

    public function endings()
    {
        return $this->hasMany(Song::class)->where('type', Song::TYPE_ENDING);
    }

    public function studios()
    {
        return $this->belongsToMany(Studio::class);
    }

    public function producers()
    {
        return $this->belongsToMany(Studio::class, 'post_producer', 'post_id', 'studio_id');
    }

    public function format()
    {
        return $this->belongsTo(Format::class);
    }

    public function externalLinks()
    {
        return $this->belongsToMany(ExternalLink::class);
    }
}
