<?php

namespace App\Policies;

use App\Models\User;
use Illuminate\Auth\Access\HandlesAuthorization;
use App\Models\Playlist;

class PlaylistPolicy
{
    use HandlesAuthorization;

    /**
     * Create a new policy instance.
     *
     * @return void
     */
    public function __construct()
    {
        //
    }

    public function view(User $user, Playlist $playlist)
    {
        return $playlist->user_id === $user->id || $playlist->is_public;
    }

    public function update(User $user, Playlist $playlist)
    {
        return $playlist->user_id === $user->id;
    }

    public function delete(User $user, Playlist $playlist)
    {
        return $playlist->user_id === $user->id;
    }
}
