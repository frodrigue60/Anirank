<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class Artist extends Model
{
    use HasFactory;

    protected $fillable = [
        'name',
        'name_jp',
        'slug',
        'thumbnail',
        'thumbnail_src'
    ];

    /* public function posts()
    {
        return $this->hasMany('App\Models\Post');
    } */

    public function songs()
    {
        return $this->belongsToMany(Song::class);
    }

    protected static function boot()
    {
        parent::boot();

        // Escucha el evento de eliminación del artista
        static::deleting(function ($artist) {
            // Desvincula todas las canciones asociadas
            $artist->songs()->detach();
        });
    }
}
