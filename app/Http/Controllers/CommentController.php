<?php

namespace App\Http\Controllers;

use App\Models\Comment;
use Illuminate\Http\Request;
use App\Models\SongVariant;
use Illuminate\Support\Facades\Auth;
use App\Models\Reaction;

class CommentController extends Controller
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
        //dd($request->all());
        $user = Auth::User();

        $validatedData = $request->validate([
            'content' => 'required',
        ]);

        $songVariant = SongVariant::findOrFail($request->song_variant_id);

        $comment = new Comment($validatedData);
        $comment->user_id = $user->id;
        $songVariant->comments()->save($comment);

        return redirect()->back()->with('status', '¡Comentario añadido con éxito!');
    }

    /**
     * Display the specified resource.
     *
     * @param  \App\Models\Comment  $comment
     * @return \Illuminate\Http\Response
     */
    public function show(Comment $comment)
    {
        //
    }

    /**
     * Show the form for editing the specified resource.
     *
     * @param  \App\Models\Comment  $comment
     * @return \Illuminate\Http\Response
     */
    public function edit(Comment $comment)
    {
        //return view('comments.edit', compact('comment'));
    }

    /**
     * Update the specified resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @param  \App\Models\Comment  $comment
     * @return \Illuminate\Http\Response
     */
    public function update(Request $request, Comment $comment)
    {
        $user = Auth::User();
        dd(
            $request->all(),
            $comment,
            $user,
        );

        $validatedData = $request->validate([
            'content' => 'required',
            'user_id' => $user->id(),
        ]);

        $comment->update($validatedData);

        return redirect()->back()->with('status', '¡Comentario actualizado con éxito!');
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  \App\Models\Comment  $comment
     * @return \Illuminate\Http\Response
     */
    public function destroy(Comment $comment)
    {
        $comment->delete();

        return redirect()->back()->with('status', '¡Comentario eliminado con éxito!');
    }

    public function like(Comment $comment)
    {
        if (Auth::check()) {
            $this->handleReaction($comment, 1); // 1 para like
            $comment->updateReactionCounters(); // Actualiza los contadores manualmente
            return redirect()->back(); // Redirige de vuelta a la página anterior
        }
        return redirect()->route('/')->with('warning', 'Please login');
    }

    public function dislike(Comment $comment)
    {
        if (Auth::check()) {
            $this->handleReaction($comment, -1); // 1 para like
            $comment->updateReactionCounters(); // Actualiza los contadores manualmente
            return redirect()->back(); // Redirige de vuelta a la página anterior
        }
        return redirect()->route('/')->with('warning', 'Please login');
    }

    private function handleReaction(Comment $comment, int $type)
    {
        $user = Auth::user();

        // Buscar si ya existe una reacción del usuario para este post
        $reaction = Reaction::where('user_id', $user->id)
            ->where('reactable_id', $comment->id)
            ->where('reactable_type', Comment::class)
            ->first();

        if ($reaction) {
            if ($reaction->type === $type) {
                // Si la reacción es la misma, eliminarla (toggle)
                $reaction->delete();
            } else {
                // Si la reacción es diferente, actualizarla
                $reaction->update(['type' => $type]);
            }
        } else {
            // Si no existe una reacción, crear una nueva
            Reaction::create([
                'user_id' => $user->id,
                'reactable_id' => $comment->id,
                'reactable_type' => Comment::class,
                'type' => $type,
            ]);
        }
    }

    public function reply(Request $request, Comment $comment)
    {
        try {
            $request->validate(['content' => 'required|string']);

            $reply = new Comment();
            $reply->content = $request->content;
            $reply->user_id = Auth::User()->id;
            $reply->parent_id = $comment->id; // Asignar el padre
            $reply->commentable_type = $comment->commentable_type; // Heredar el tipo polimórfico
            $reply->commentable_id = $comment->commentable_id; // Heredar el ID polimórfico

            //dd($parentComment, $reply);
            $reply->save();

            return back()->with('success', 'Respuesta enviada.');
        } catch (\Throwable $th) {
            return back()->with('error', $th->getMessage());
        }
    }
}
