<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use App\Models\Playlist;
use Illuminate\Support\Facades\DB;

class PlaylistController extends Controller
{
    public function index(Request $request)
    {
        $songId = $request->query('song_id') ?? $request->input('song_id');

        if (!$songId) {
            return response()->json(['error' => 'song_id is required'], 400);
        }

        $user = Auth::user();

        $playlists = $user->playlists()
            ->withCount('songs')
            ->get();

        $playlistsWithSong = DB::table('playlist_song')
            ->where('song_id', $songId)
            ->whereIn('playlist_id', $playlists->pluck('id'))
            ->pluck('playlist_id')
            ->toArray();

        $playlists = $playlists->map(function ($playlist) use ($playlistsWithSong) {
            $playlist->is_in_playlist = in_array($playlist->id, $playlistsWithSong);
            return $playlist;
        });

        return response()->json([
            'playlists' => $playlists,
            'song_id'   => $songId,
            'message'   => 'Playlists retrieved successfully',
            'status'    => 200
        ], 200);
    }

    public function store(Request $request)
    {
        $request->validate([
            'name' => 'required|string|max:255',
            'description' => 'nullable|string|max:255',
        ]);

        $playlist = Playlist::create([
            'name' => $request->name,
            'description' => $request->description,
            'user_id' => Auth::id(),
        ]);

        return response()->json([
            'playlist' => $playlist,
            'message' => 'Playlist created successfully',
            'status' => 201
        ], 201);
    }

    public function show(Playlist $playlist)
    {
        $this->authorize('view', $playlist);

        $playlist->load('posts');

        return response()->json([
            'playlist' => $playlist,
            'message' => 'Playlist retrieved successfully',
        ], 200);
    }

    public function destroy(Playlist $playlist)
    {
        $this->authorize('delete', $playlist);
        $playlist->delete();

        if (request()->wantsJson()) {
            return response()->json(['message' => 'Playlist eliminada']);
        }

        return back()->with('success', 'Playlist eliminada');
    }

    public function toggleSong(Request $request, Playlist $playlist)
    {
        try {
            $request->validate([
                'song_id' => 'required|exists:songs,id'
            ]);

            $songId = $request->song_id;
            $exists = $playlist->songs()->where('song_id', $songId)->exists();

            if ($exists) {
                $playlist->songs()->detach($songId);
                $action = 'removed';
                $message = 'Post removido de la playlist correctamente';
            } else {
                $playlist->songs()->attach($songId);
                $action = 'added';
                $message = 'Post agregado a la playlist correctamente';
            }

            return response()->json([
                'success' => true,
                'message' => $message,
                'data' => [
                    'playlist_id' => $playlist->id,
                    'song_id' => $songId,
                    'action' => $action,
                    'in_playlist' => !$exists
                ]
            ], 200);
        } catch (\Exception $e) {
            return response()->json([
                'success' => false,
                'message' => 'Error interno del servidor',
                'error' => $e->getMessage()
            ], 500);
        }
    }
}
