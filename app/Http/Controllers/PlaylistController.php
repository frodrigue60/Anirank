<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use App\Models\Playlist;
use App\Models\Post;
use Illuminate\Support\Facades\Auth;

class PlaylistController extends Controller
{
    public function index() {
        $playlists = Auth::user()->playlists()->withCount('songs')->get();
        return view('public.playlists.index', compact('playlists'));
    }

    public function create() {
        return view('public.playlists.create');
    }

    public function store(Request $request) {
        $request->validate([
            'name' => 'required|string|max:255',
            /* 'description' => 'nullable|string', */
        ]);

        $playlist = new Playlist();
        $playlist->name = $request->input('name');
        /* $playlist->description = $request->input('description'); */
        $playlist->user_id = Auth::id();
        $playlist->save();

        $message = 'Playlist created successfully.';

        return redirect()->route('playlists.index')->with('success', $message);
    }

    public function show(Playlist $playlist) {
        return view('public.playlists.show', compact('playlist'));
    }

    public function edit(Playlist $playlist) {
        return view('public.playlists.edit', compact('playlist'));
    }

    public function update(Request $request, Playlist $playlist) {
         $request->validate([
            'name' => 'required|string|max:255',
            /* 'description' => 'nullable|string', */
        ]);

        $playlist = new Playlist();
        $playlist->name = $request->input('name');
        /* $playlist->description = $request->input('description'); */
        $playlist->user_id = Auth::id();
        $playlist->save();

        $message = 'Playlist updated successfully.';

        return redirect()->route('playlists.index')->with('success', $message);
    }

    public function destroy(Playlist $playlist) {

        $playlist->delete();

        $message = 'Playlist deleted successfully.';

        return redirect()->route('playlists.index')->with('success', $message);
    }
}
