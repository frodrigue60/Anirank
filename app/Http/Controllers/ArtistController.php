<?php

namespace App\Http\Controllers;

use App\Models\Artist;
use App\Models\Season;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use App\Models\Song;
use App\Models\Year;
use stdClass;
use Illuminate\Pagination\Paginator;
use Illuminate\Support\Collection;
use Illuminate\Pagination\LengthAwarePaginator;
use App\Models\SongVariant;


class ArtistController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
        return view('public.artists.index');
    }

    /**
     * Display the specified resource.
     *
     * @param  int  $id
     * @param  mixed  $slug
     * @return \Illuminate\Http\Response
     */
    public function show($slug)
    {
        $years = Year::all()->sortByDesc('name');
        $seasons = Season::all()->sortByDesc('name');

        $types = [
            ['name' => 'Opening', 'value' => 'OP'],
            ['name' => 'Ending', 'value' => 'ED'],
            ['name' => 'Insert', 'value' => 'INS'],
            ['name' => 'Other', 'value' => 'OTH']
        ];

        $sortMethods = [
            ['name' => 'Recent', 'value' => 'recent'],
            ['name' => 'Title', 'value' => 'title'],
            ['name' => 'Score', 'value' => 'averageRating'],
            ['name' => 'Views', 'value' => 'view_count'],
            ['name' => 'Popular', 'value' => 'likeCount']
        ];

        $artist = Artist::with('songs')->where('slug', $slug)->first();
        //$songs = $artist->songs;

        //return view('public.artists.show', compact('artist', 'seasons', 'years', 'sortMethods', 'types', 'songs'));
        return view('public.artists.show', compact('artist', 'seasons', 'years', 'sortMethods', 'types'));
    }
}
