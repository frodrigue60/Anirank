<?php

namespace App\Http\Controllers\admin;

use App\Http\Controllers\Controller;
use App\Models\Producer;
use Illuminate\Http\Request;
use Illuminate\Support\Str;

class ProducerController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
        $producers = Producer::paginate(15);
        return view('admin.producers.index', compact('producers'));
    }

    /**
     * Show the form for creating a new resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function create()
    {
        return view('admin.producers.create');
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
            'name' => 'required|unique:producers,name',
        ]);

        Producer::create([
            'name' => $request->name,
            'slug' => Str::slug($request->name),
        ]);

        return redirect()->route('admin.producers.index')->with('success', 'Producer created successfully.');
    }

    /**
     * Show the form for editing the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function edit($id)
    {
        $producer = Producer::findOrFail($id);
        return view('admin.producers.edit', compact('producer'));
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
        $producer = Producer::findOrFail($id);

        $request->validate([
            'name' => 'required|unique:producers,name,' . $producer->id,
        ]);

        $producer->update([
            'name' => $request->name,
            'slug' => Str::slug($request->name),
        ]);

        return redirect()->route('admin.producers.index')->with('success', 'Producer updated successfully.');
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function destroy($id)
    {
        $producer = Producer::findOrFail($id);
        $producer->delete();

        return redirect()->route('admin.producers.index')->with('success', 'Producer deleted successfully.');
    }
}
