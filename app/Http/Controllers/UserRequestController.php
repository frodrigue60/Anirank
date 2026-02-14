<?php

namespace App\Http\Controllers;

use App\Models\UserRequest;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Validator;

class UserRequestController extends Controller
{
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
        $userRequest = new UserRequest();
        $userRequest->content = $request->content;
        $userRequest->user_id = Auth::user()->id;

        $validator = Validator::make($request->all(), [
            'content' => 'required|string|max:255',
        ]);

        if ($validator->fails()) {
            $messageBag = $validator->getMessageBag();
            return redirect(route('requests.create'))
                ->back()
                ->with('error', $messageBag);
        }

        if ($userRequest->save()) {
            return redirect(route('home'))->with('success', 'Thank for your request');
        } else {
            return redirect(route('home'))->with('error', 'Something has been wrong');
        }
    }
}
