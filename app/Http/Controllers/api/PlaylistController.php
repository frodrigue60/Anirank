<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use App\Models\Playlist;
use Illuminate\Support\Facades\DB;

class PlaylistController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    /* public function index(Request $request)
    {
        //return response()->json(['song_id' => $request->all()], 200);
        // Obligatorio: el ID de la canción que el usuario quiere añadir/quitar
        $songId = $request->query('song_id') ?? $request->input('song_id');

        if (!$songId) {
            return response()->json(['error' => 'song_id is required'], 400);
        }

        $user = auth()->user();

        // Cargar playlists del usuario con el conteo de canciones
        $playlists = $user->playlists()
            ->withCount('songs')  // → genera $playlist->songs_count
            ->get();

        // Obtener solo los IDs de playlists que YA tienen esta canción
        // (asumiendo que la tabla pivot se llama playlist_song)
        $playlistsWithSong = DB::table('playlist_song')
            ->where('song_id', $songId)
            ->whereIn('playlist_id', $playlists->pluck('id'))
            ->pluck('playlist_id')
            ->toArray();

        return response()->json([
            'playlists'  => $playlists,
            'message'    => 'Playlists retrieved successfully',
            'status'     => 200
        ], 200);
    } */

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

        // IDs de playlists que contienen la canción
        $playlistsWithSong = DB::table('playlist_song')
            ->where('song_id', $songId)
            ->whereIn('playlist_id', $playlists->pluck('id'))
            ->pluck('playlist_id')
            ->toArray();

        // Añadir campo booleano a cada playlist
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

    /**
     * Show the form for creating a new resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function create()
    {
        //
    }

    /**
     * Store a newly created resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @return \Illuminate\Http\Response
     */
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

    /**
     * Display the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function show(Playlist $playlist)
    {
        $this->authorize('view', $playlist);

        $playlist->load('posts');

        return response()->json([
            'playlist' => $playlist,
            'message' => 'Playlist retrieved successfully',
        ], 200);
        //return view('playlists.show', compact('playlist'));
    }

    /**
     * Show the form for editing the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function edit($id)
    {
        //
    }

    /**
     * Update the specified resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function update(Request $request, $id)
    {
        //
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
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
                    'in_playlist' => !$exists // Estado actual después del toggle
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
