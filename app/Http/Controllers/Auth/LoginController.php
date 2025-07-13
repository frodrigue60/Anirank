<?php

namespace App\Http\Controllers\Auth;

use App\Http\Controllers\Controller;
use App\Providers\RouteServiceProvider;
use Illuminate\Foundation\Auth\AuthenticatesUsers;
use Illuminate\Http\Request;

class LoginController extends Controller
{
    /*
    |--------------------------------------------------------------------------
    | Login Controller
    |--------------------------------------------------------------------------
    |
    | This controller handles authenticating users for the application and
    | redirecting them to your home screen. The controller uses a trait
    | to conveniently provide its functionality to your applications.
    |
    */

    use AuthenticatesUsers;

    /**
     * Where to redirect users after login.
     *
     * @var string
     */
    protected $redirectTo = RouteServiceProvider::HOME;

    /**
     * Create a new controller instance.
     *
     * @return void
     */
    public function __construct()
    {
        $this->middleware('guest')->except('logout');
    }

    protected function authenticated(Request $request, $user)
    {
        // Elimina tokens anteriores si quieres que solo haya uno activo
        $user->tokens()->delete();

        // Crea un nuevo token
        $token = $user->createToken('api-token')->plainTextToken;

        // Si es una petición AJAX o JSON, devuelve el token
        if ($request->expectsJson()) {
            return response()->json([
                'success' => true,
                'token' => $token,
                'user' => $user,
            ]);
        }

        // Si es una petición normal, redirige como siempre
        return redirect()->intended($this->redirectPath());
    }
}
