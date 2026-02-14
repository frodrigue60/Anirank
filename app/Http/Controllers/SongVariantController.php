<?php

namespace App\Http\Controllers;

use App\Models\SongVariant;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Validator;
use Illuminate\Support\Facades\DB;
use App\Models\Post;
use App\Models\Song;
use App\Models\Reaction;
use App\Models\Season;
use App\Models\Year;
use App\Models\Favorite;
use App\Models\User;
use Illuminate\Pagination\Paginator;
use Illuminate\Support\Collection;
use Illuminate\Pagination\LengthAwarePaginator;

class SongVariantController extends Controller
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
        //
    }

    /**
     * Display the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function show(SongVariant $songVariant)
    {
        $user = Auth::check() ? Auth::User() : null;

        if (!$songVariant) {
            return redirect(route('home'))->with('warning', 'Item no exist!');
        }

        if ($songVariant->song->post->status == 'stagged') {
            return redirect(route('home'))->with('warning', 'Paused post!');
        }

        $comments = $songVariant->comments;
        //dd($comments[0]->user);
        $factor = 1;

        $songVariant->score = round($songVariant->averageRating * $factor, 1);

        #Is used by rating form
        $userRating = null;

        if ($user) {

            $userRating = $this->getUserRating($songVariant, $user);

            if ($userRating) {

                switch ($user->score_format) {
                    case 'POINT_100':
                        $factor = 1;
                        $userRating->formatRating = round($userRating->rating);
                        break;

                    case 'POINT_10_DECIMAL':
                        $factor = 0.1;
                        $userRating->formatRating = round($userRating->rating / 10, 1);
                        break;

                    case 'POINT_10':
                        $factor = 1 / 10;
                        $userRating->formatRating = round($userRating->rating / 10);
                        break;

                    case 'POINT_5':
                        $factor = 1 / 20;
                        //Divide the score in segments of [20, 40, 60, 80, 100]
                        $userRating->formatRating = (int) max(20, min(100, ceil($userRating->rating / 20) * 20));
                        break;

                    default:
                        $factor = 1;
                        $userRating->formatRating = round($userRating->rating);
                        break;
                }
            }
        }

        $songVariant = $this->setScoreOnlyOneVariant($songVariant, $user);

        //dd($songVariant);

        $songVariant->incrementViews();

        return view('public.variants.show', compact('songVariant', 'comments', 'userRating'));
    }

    /**
     * Show the form for editing the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function edit($id)
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
    public function update(Request $request, $id)
    {
        //
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function destroy($id)
    {
        //
    }

    public function showVariant($animeSlug, $songSlug, $variantSlug)
    {
        $user = Auth::check() ? Auth::User() : null;
        $post = Post::where('slug', $animeSlug)->first();

        if (!$post) {
            return redirect(route('/'))->with('warning', 'Post not exist!');
        }
        if (!$post->status) {
            if ($user) {
                if (!$user->isAdmin()) {
                    return redirect('/')->with('danger', 'User not autorized!');
                }
            } else {
                return redirect('/')->with('danger', 'Post status: Private');
            }
        }

        $song = Song::where('slug', $songSlug)
            ->where('post_id', $post->id)
            ->firstOrFail();

        $songVariant = SongVariant::where('slug', $variantSlug)
            ->where('song_id', $song->id)
            ->with('reactionsCounter')
            ->firstOrFail();

        //dd($songVariant);

        if (!$songVariant) {
            return redirect(route('/'))->with('warning', 'Item no exist!');
        }

        if ($songVariant->song->post->status == 'stagged') {
            return redirect(route('/'))->with('warning', 'Paused post!');
        }

        $comments = $songVariant->comments;
        //dd($comments[0]->user);
        $factor = 1;

        $songVariant->score = round($songVariant->averageRating * $factor, 1);

        #Is used by rating form
        $userRating = null;

        if ($user) {

            $userRating = $this->getUserRating($songVariant->id, $user->id);

            if ($userRating) {

                switch ($user->score_format) {
                    case 'POINT_100':
                        $factor = 1;
                        $userRating->formatRating = round($userRating->rating);
                        break;

                    case 'POINT_10_DECIMAL':
                        $factor = 0.1;
                        $userRating->formatRating = round($userRating->rating / 10, 1);
                        break;

                    case 'POINT_10':
                        $factor = 1 / 10;
                        $userRating->formatRating = round($userRating->rating / 10);
                        break;

                    case 'POINT_5':
                        $factor = 1 / 20;
                        //Divide the score in segments of [20, 40, 60, 80, 100]
                        $userRating->formatRating = (int) max(20, min(100, ceil($userRating->rating / 20) * 20));
                        break;

                    default:
                        $factor = 1;
                        $userRating->formatRating = round($userRating->rating);
                        break;
                }
            }
        }

        $songVariant = $this->setScoreOnlyOneVariant($songVariant, $user);

        //dd($songVariant);

        $songVariant->incrementViews();

        return view('public.variants.show', compact('songVariant', 'comments', 'userRating'));
    }

    public function rate(Request $request, SongVariant $songVariant)
    {
        //dd($request->all());
        if (Auth::check()) {

            $score_format = Auth::user()->score_format;

            $validator = Validator::make($request->all(), [
                'score' => 'required|numeric'
            ]);

            if ($validator->fails()) {
                $messageBag = $validator->getMessageBag();
                return redirect()
                    ->back()
                    ->with('error', $messageBag);
            } else {
                $score = $request->score;
            }

            if ($score_format === 'POINT_5') {
                // Ajustar el score según las reglas específicas para POINT_5
                $score = max(20, min(100, ceil($score / 20) * 20));
            } else {
                // Ajustar el score según las reglas comunes para POINT_100, POINT_10_DECIMAL y POINT_10
                $score = max(1, min(100, ($score_format === 'POINT_10_DECIMAL') ? round($score * 10) : round($score)));
            }

            // Validar el rango del score
            if ($score >= 1 && $score <= 100) {
                // Utilizar el score ajustado
                $songVariant->rateOnce($score, Auth::User()->id);
                return redirect()->back()->with('success', 'Post rated Successfully');
            } else {
                return redirect()->back()->with('warning', 'Only values between 1 and 100');
            }
        } else {
            return redirect()->route('login');
        }
    }

    public function getUserRating(SongVariant $songVariant, User $user)
    {
        $userRating = DB::table('ratings')
            ->where('rateable_type', SongVariant::class)
            ->where('rateable_id', $songVariant->id)
            ->where('user_id', $user->id)
            ->first(['rating']);

        return $userRating;
    }

    public function like(SongVariant $songVariant)
    {
        $this->handleReaction($songVariant, 1); // 1 para like
        $songVariant->updateReactionCounters(); // Actualiza los contadores manualmente
        return redirect()->back(); // Redirige de vuelta a la página anterior
    }

    // Método para dislike
    public function dislike(SongVariant $songVariant)
    {
        $this->handleReaction($songVariant, -1); // -1 para dislike
        $songVariant->updateReactionCounters(); // Actualiza los contadores manualmente
        return redirect()->back(); // Redirige de vuelta a la página anterior
    }

    // Método privado para manejar la reacción
    private function handleReaction(SongVariant $songVariant, int $type)
    {
        $user = Auth::user();

        // Buscar si ya existe una reacción del usuario para este post
        $reaction = Reaction::where('user_id', $user->id)
            ->where('reactable_id', $songVariant->id)
            ->where('reactable_type', SongVariant::class)
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
                'reactable_id' => $songVariant->id,
                'reactable_type' => SongVariant::class,
                'type' => $type,
            ]);
        }
    }

    public function seasonal(Request $request)
    {
        $currentSeason = Season::where('current', true)->first();
        $currentYear = Year::where('current', true)->first();

        return view('public.seasonal', compact('currentSeason', 'currentYear'));
    }

    public function setScoreOnlyVariants($variants, $user = null)
    {
        $variants->each(function ($variant) use ($user) {
            $variant->userScore = null;
            $factor = 1;
            $isDecimalFormat = false;
            $denominator = 100; // Por defecto para POINT_100

            if ($user) {
                switch ($user->score_format) {
                    case 'POINT_100':
                        $factor = 1;
                        $denominator = 100;
                        break;
                    case 'POINT_10_DECIMAL':
                        $factor = 0.1;
                        $denominator = 10;
                        $isDecimalFormat = true;
                        break;
                    case 'POINT_10':
                        $factor = 1 / 10;
                        $denominator = 10;
                        break;
                    case 'POINT_5':
                        $factor = 1 / 20;
                        $denominator = 5;
                        $isDecimalFormat = true;
                        break;
                }

                if ($userRating = $this->getUserRating($variant->id, $user->id)) {
                    $variant->userScore = $isDecimalFormat
                        ? round($userRating->rating * $factor, 1)
                        : (int) round($userRating->rating * $factor);
                }
            }

            $variant->score = $isDecimalFormat
                ? round($variant->averageRating * $factor, 1)
                : (int) round($variant->averageRating * $factor);

            // Agregar la propiedad scoreString formateada
            $variant->scoreString = $this->formatScoreString(
                $variant->score,
                $user->score_format ?? 'POINT_100',
                $denominator
            );
        });

        return $variants;
    }

    #New
    public function setScoreOnlyOneVariant(SongVariant $variant, User $user)
    {
        $variant->userScore = null;
        $factor = 1;
        $isDecimalFormat = false;
        $denominator = 100; // Por defecto para POINT_100

        if ($user) {
            switch ($user->score_format) {
                case 'POINT_100':
                    $factor = 1;
                    $denominator = 100;
                    break;
                case 'POINT_10_DECIMAL':
                    $factor = 0.1;
                    $denominator = 10;
                    $isDecimalFormat = true;
                    break;
                case 'POINT_10':
                    $factor = 1 / 10;
                    $denominator = 10;
                    break;
                case 'POINT_5':
                    $factor = 1 / 20;
                    $denominator = 5;
                    $isDecimalFormat = true;
                    break;
            }

            if ($userRating = $this->getUserRating($variant, $user)) {
                $variant->userScore = $isDecimalFormat
                    ? round($userRating->rating * $factor, 1)
                    : (int) round($userRating->rating * $factor);
            }
        }

        $variant->score = $isDecimalFormat
            ? round($variant->averageRating * $factor, 1)
            : (int) round($variant->averageRating * $factor);

        // Agregar la propiedad scoreString formateada
        $variant->scoreString = $this->formatScoreString(
            $variant->score,
            $user->score_format ?? 'POINT_100',
            $denominator
        );


        return $variant;
    }

    public function toggleFavorite(SongVariant $songVariant)
    {

        if (!Auth::check()) {
            return redirect()->back()->with('warning', 'Please login');
        }

        $user = Auth::user();

        // Verificar si el post ya está en favoritos
        $favorite = Favorite::where('user_id', $user->id)
            ->where('favoritable_id', $songVariant->id)
            ->where('favoritable_type', SongVariant::class)
            ->first();

        if ($favorite) {
            $favorite->delete();
            return redirect()->back()->with('success', 'Theme removed to favorites');
        } else {
            Favorite::create([
                'user_id' => $user->id,
                'favoritable_id' => $songVariant->id,
                'favoritable_type' => SongVariant::class,
            ]);
            return redirect()->back()->with('success', 'Theme added to favorites');
        }
    }

    public function ranking()
    {
        return view('public.ranking');
    }

    public function paginate($collection, $perPage = 18, $page = null, $options = [])
    {
        $page = Paginator::resolveCurrentPage();
        $options = ['path' => Paginator::resolveCurrentPath()];
        $items = $collection instanceof Collection ? $collection : Collection::make($collection);
        $collection = new LengthAwarePaginator($items->forPage($page, $perPage), $items->count(), $perPage, $page, $options);
        return $collection;
    }

    protected function formatScoreString($score, $format, $denominator)
    {
        switch ($format) {
            case 'POINT_100':
                return $score . '/' . $denominator;
            case 'POINT_10_DECIMAL':
                return number_format($score, 1) . '/' . $denominator;
            case 'POINT_10':
                return $score . '/' . $denominator;
            case 'POINT_5':
                return number_format($score, 1) . '/' . $denominator;
            default:
                return $score . '/' . $denominator;
        }
    }
}
