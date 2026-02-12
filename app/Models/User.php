<?php

namespace App\Models;

// use Illuminate\Contracts\Auth\MustVerifyEmail;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Foundation\Auth\User as Authenticatable;
use Illuminate\Notifications\Notifiable;
use Laravel\Sanctum\HasApiTokens;
use Illuminate\Support\Facades\Auth;
use App\Models\UserRequest;
use App\Models\Comment;
use App\Models\Favorite;
use Illuminate\Support\Str;

class User extends Authenticatable
{
    use HasApiTokens, HasFactory, Notifiable;

    /**
     * The attributes that are mass assignable.
     *
     * @var array<int, string>
     */
    protected $fillable = [
        'name',
        'email',
        'password',
        'image',
        'banner',
        'score_format',
        'slug'
    ];

    /**
     * The attributes that should be hidden for serialization.
     *
     * @var array<int, string>
     */
    protected $hidden = [
        'password',
        'remember_token',
    ];

    /**
     * The attributes that should be cast.
     *
     * @var array<string, string>
     */
    protected $casts = [
        'email_verified_at' => 'datetime',
    ];

    public function userRequests()
    {
        return $this->hasMany(UserRequest::class);
    }

    public function isStaff()
    {
        $user = Auth::user()->type;
        if ($user == 'admin' || $user == 'editor' || $user == 'creator') {
            return true;
        } else {
            return false;
        }
    }

    public function isAdmin()
    {
        $user = Auth::user()->type;
        if ($user == 'admin') {
            return true;
        } else {
            return false;
        }
    }
    public function isEditor()
    {
        $user = Auth::user()->type;
        if ($user == 'editor') {
            return true;
        } else {
            return false;
        }
    }
    public function isCreator()
    {
        $user = Auth::user()->type;
        if ($user == 'creator') {
            return true;
        } else {
            return false;
        }
    }

    public function generateSlug()
    {
        $slug = Str::slug($this->name); // Genera el slug a partir del nombre
        $originalSlug = $slug;
        $count = 1;

        // Verifica si el slug ya existe y agrega un sufijo numérico si es necesario
        while (User::where('slug', $slug)->exists()) {
            $slug = $originalSlug . '-' . $count++;
        }

        $this->slug = $slug;
    }

    public function reactions()
    {
        return $this->hasMany(Reaction::class);
    }

    public function favorites()
    {
        return $this->hasMany(Favorite::class);
    }

    public function comments()
    {
        return $this->hasMany(Comment::class);
    }

    /* PLAYLISTS */
    public function playlists()
    {
        return $this->hasMany(Playlist::class);
    }

    /**
     * Obtener playlists con conteo de posts (para eficiencia)
     */
    public function playlistsWithCount()
    {
        return $this->playlists()->withCount('posts');
    }

    /**
     * Obtener playlists públicas de otros usuarios
     */
    public function publicPlaylists()
    {
        return Playlist::where('is_public', true)
            ->where('user_id', '!=', $this->id)
            ->with('user')
            ->withCount('posts');
    }

    /**
     * Verificar si el usuario puede ver una playlist específica
     */
    public function canViewPlaylist(Playlist $playlist)
    {
        return $playlist->user_id === $this->id || $playlist->is_public;
    }

    /**
     * Obtener el número total de playlists del usuario
     */
    public function getPlaylistsCountAttribute()
    {
        return $this->playlists()->count();
    }

    /**
     * Obtener el número total de posts en todas las playlists del usuario
     */
    public function getTotalPlaylistPostsAttribute()
    {
        return $this->playlists()->withCount('posts')->get()->sum('posts_count');
    }
}
