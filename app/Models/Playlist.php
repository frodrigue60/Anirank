<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use App\Models\Song;

class Playlist extends Model
{
    use HasFactory;

    protected $fillable = ['name', 'description', 'user_id', 'is_public'];

    public function user()
    {
        return $this->belongsTo(User::class);
    }

    public function songs()
    {
        return $this->belongsToMany(Song::class)
            ->withPivot('position')
            ->orderBy('position');
    }

    public function addPost(Song $song)
    {
        $currentMaxPosition = $this->songs()->max('position') ?? 0;
        return $this->songs()->attach($song->id, ['position' => $currentMaxPosition + 1]);
    }

    public function removePost(Song $song)
    {
        return $this->songs()->detach($song->id);
    }
}
