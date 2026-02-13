<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use App\Models\SongVariant;
use Illuminate\Support\Facades\Storage;

class Video extends Model
{
    use HasFactory;

    protected static function boot()
    {
        parent::boot();

        static::deleting(function ($video) {
            if ($video->video_src && Storage::exists('public/' . $video->video_src)) {
                Storage::delete('public/' . $video->video_src);
            }
        });
    }

    public function song()
    {
        return $this->belongsTo(Song::class);
    }

    public function songVariant()
    {
        return $this->belongsTo(SongVariant::class);
    }

    public function isEmbed()
    {
        return $this->type === 'embed';
    }

    public function isLocal()
    {
        return $this->type === 'file';
    }

    public function getEmbedUrlAttribute()
    {
        if (!$this->isEmbed()) return null;

        return $this->embed_code;
    }

    public function getLocalUrlAttribute()
    {
        return $this->isLocal() ? env('APP_URL') . Storage::url($this->video_src) : null;
    }
}
