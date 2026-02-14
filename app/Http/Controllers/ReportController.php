<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use App\Models\Report;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Validator;

class ReportController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
        //
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
        if (Auth::check()) {
            //dd($request->all());

            $validator = Validator::make($request->all(), [
                'song_variant_id' => 'required|integer|exists:song_variants,id',
                'title' => 'required|max:255|string',
                'content' => 'string|nullable',
                'user_id' => 'required|integer|exists:users,id',
            ]);

            if ($validator->fails()) {
                return redirect($request->header('Referer'))
                    ->withErrors($validator)
                    ->withInput();
            }

            $report = new Report();
            $report->song_variant_id = $request->song_variant_id;
            $report->title = $request->title;
            $report->content = $request->content;
            $report->user_id = $request->user_id;
            $report->source = $request->header('Referer');
            $report->save();

            return redirect()->back()->with('success', 'Thanks for report this problem');
        }

        return redirect()->back()->with('warning', 'Please login to send a report');
    }

    /**
     * Display the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function show(Report $report)
    {
        //
    }

    /**
     * Show the form for editing the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function edit(Report $report)
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
    public function update(Request $request, Report $report)
    {
        //
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function destroy(Report $report)
    {
        //
    }
}
