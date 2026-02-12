<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use App\Models\Playlist;
use App\Models\Post;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Storage;

class PlaylistController extends Controller
{
    public function index()
    {
        $playlists = Auth::user()->playlists()
            ->withCount('songs')
            ->with(['songs' => function ($query) {
                $query->with('post')->limit(1);
            }])
            ->get();
        return view('public.playlists.index', compact('playlists'));
    }

    public function create()
    {
        return view('public.playlists.create');
    }

    public function store(Request $request)
    {
        $request->validate([
            'name' => 'required|string|max:255',
            'description' => 'nullable|string|max:255',
        ]);

        $playlist = new Playlist();
        $playlist->name = $request->input('name');
        $playlist->description = $request->input('description');
        $playlist->user_id = Auth::id();
        $playlist->save();

        $message = 'Playlist created successfully.';

        return redirect()->route('playlists.index')->with('success', $message);
    }

    public function show(Playlist $playlist)
    {
        $playlist->load(['songs.post', 'songs.songVariants.video']);

        $queue = $playlist->songs->map(function ($song) {
            // 1. Tomar la primera variante
            $firstVariant = $song->songVariants->first();

            if (!$firstVariant) {
                return null;
            }

            // 2. Obtener el video asociado
            $video = $firstVariant->video;

            if (!$video) {
                return null;
            }

            // 3. Determinar el thumbnail (preferencia al post)
            $thumbnail = asset('images/default.jpg');
            if ($song->post && $song->post->thumbnail && Storage::disk('public')->exists($song->post->thumbnail)) {
                $thumbnail = Storage::url($song->post->thumbnail);
            } elseif ($song->thumbnail) {
                $thumbnail = $song->thumbnail;
            }

            // 4. Construir item
            return [
                'song_id'        => $song->id,
                'song_title'     => $song->name,
                'variant_id'     => $firstVariant->id,
                'variant_quality' => $firstVariant->quality ?? 'unknown',
                'video_id'       => $video->id,
                'video_type'     => $video->type,
                'video_url'      => $video->type === 'embed'
                    ? $video->embed_url
                    : $video->local_url,
                'duration'       => $video->duration ?? 0,
                'thumbnail'      => $thumbnail,
            ];
        })
            ->filter()
            ->values();

        return view('public.playlists.show', compact('playlist', 'queue'));
    }

    public function edit(Playlist $playlist)
    {
        return view('public.playlists.edit', compact('playlist'));
    }

    public function update(Request $request, Playlist $playlist)
    {
        $request->validate([
            'name' => 'required|string|max:255',
            'description' => 'nullable|string|max:255',
        ]);

        $playlist->name = $request->input('name');
        $playlist->description = $request->input('description');
        $playlist->save();

        $message = 'Playlist updated successfully.';

        return redirect()->route('playlists.index')->with('success', $message);
    }

    public function destroy(Playlist $playlist)
    {

        $playlist->delete();

        $message = 'Playlist deleted successfully.';

        return redirect()->route('playlists.index')->with('success', $message);
    }
}
