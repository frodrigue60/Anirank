<?php

namespace App\Http\Controllers;

use App\Models\Producer;
use Illuminate\Http\Request;
use App\Models\Year;
use App\Models\Season;
use App\Models\Format;

class ProducerController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
        return view('public.producers.index');
    }

    /**
     * Display the specified resource.
     *
     * @param  string  $slug
     * @return \Illuminate\Http\Response
     */
    public function show(Producer $producer)
    {
        $years = Year::all()->sortBy('name', null, true);
        $seasons = Season::all();
        $sortMethods = $this->filterTypesSortChar()['sortMethods'];
        $formats = Format::all();

        return view('public.producers.show', compact('producer', 'seasons', 'years', 'sortMethods', 'formats'));
    }

    public function filterTypesSortChar()
    {
        $sortMethods = [
            ['name' => 'Recent', 'value' => 'recent'],
            ['name' => 'Title', 'value' => 'title'],
            ['name' => 'Score', 'value' => 'averageRating'],
            ['name' => 'Views', 'value' => 'view_count'],
            ['name' => 'Popular', 'value' => 'likeCount']
        ];

        return [
            'sortMethods' => $sortMethods,
        ];
    }
}
