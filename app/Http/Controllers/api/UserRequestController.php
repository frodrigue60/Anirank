<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use Illuminate\Http\Request;
use App\Models\UserRequest;
use Illuminate\Support\Facades\Validator;
use Illuminate\Support\Facades\Auth;

class UserRequestController extends Controller
{
    public function store(Request $request)
    {
        try {
            $userRequest = new UserRequest();
            $userRequest->content = $request->content;
            $userRequest->user_id = Auth::user()->id;

            $validator = Validator::make($request->all(), [
                'content' => 'required|string|max:255',
            ]);

            if ($validator->fails()) {
                $messageBag = $validator->getMessageBag();
                return response()->json([
                    'success' => false,
                    'message' => $messageBag
                ]);
            }

            $userRequest->save();

            return response()->json([
                'success' => true,
                'message' => 'Request send successfully!'
            ]);
        } catch (\Throwable $th) {
            return response()->json([
                'success' => false,
                'message' => $th
            ]);
        }
    }
}
