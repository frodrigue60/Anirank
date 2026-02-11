<?php

namespace App\Http\Controllers;

use App\Models\Comment;
use Illuminate\Http\Request;
use App\Models\Song;
use Illuminate\Support\Facades\Auth;
use App\Models\Year;
use App\Models\Season;
use App\Models\Post;
use Illuminate\Support\Facades\DB;

class SongController extends Controller
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
    public function show($anime_slug, $song_slug)
    {
        //dd($anime_slug);

        $post = Post::with(['songs'])->where('slug', $anime_slug)->first();
        $user = Auth::check() ? Auth::User() : null;

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

        $song = Song::with(['songVariants.video', 'comments.user'])
            ->where('slug', $song_slug)
            ->where('post_id', $post->id)
            ->first();

        $factor = 1;
        $song->rawScore = (int) round($song->averageRating * $factor, 1);
        $userRating = null;

        if ($user) {

            $userRating = $this->getUserRating($song->id, $user->id);

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

        #Add to song ["formattedUserScore" => 8.0,"formattedScore" => 8.0,"rawUserScore" => 80.0,"rawScore" => 80.0,"scoreString" => "8.0/10"] attributes
        $song = $this->setScoreSong($song, $user);

        $song->incrementViews();

        $firstVariant = $post->songs->flatMap(function ($song) {
            return $song->songVariants;
        })->sortBy('version_number')->first();
        //dd($song);
        //$comments = $song->comments->sortByDesc('created_at');

        $comments = Comment::with('replies', 'user')
            ->where('commentable_id', $song->id)
            ->where('commentable_type', Song::class)
            ->where('parent_id', null)
            ->get()
            ->sortByDesc('created_at');



        //dd($comments);



        return view('public.songs.show', compact('song', 'post', 'userRating', 'firstVariant', 'comments'));
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

    public function seasonal()
    {
        $currentSeason = Season::where('current', true)->first();
        $currentYear = Year::where('current', true)->first();

        //dd($currentSeason, $currentYear);

        return view('public.seasonal', compact('currentSeason', 'currentYear'/* , 'openings', 'endings' */));
    }

    public function ranking()
    {
        return view('public.ranking');
    }

    public function getUserRating($song_id, $user_id)
    {
        $userRating = DB::table('ratings')
            ->where('rateable_type', Song::class)
            ->where('rateable_id', $song_id)
            ->where('user_id', $user_id)
            ->first(['rating']);

        return $userRating;
    }

    public function setScoreSong($song, $user = null)
    {
        #Inizialided attributes
        $song->formattedScore = null;
        $song->rawScore = null;
        $song->scoreString = null;

        $factor = 1;
        $isDecimalFormat = false;
        $denominator = 100; // Por defecto para POINT_100

        if ($user) {
            #Inizialided attributes
            $song->formattedUserScore = null;
            $song->rawUserScore = null;

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

            if ($userRating = $this->getUserRating($song->id, $user->id)) {
                $song->formattedUserScore = $isDecimalFormat
                    ? round($userRating->rating * $factor, 1)
                    : (int) round($userRating->rating * $factor);

                $song->rawUserScore = round($userRating->rating);
            }
        }

        $song->rawScore = round($song->averageRating, 1);

        $song->formattedScore = $isDecimalFormat
            ? round($song->averageRating * $factor, 1)
            : (int) round($song->averageRating * $factor);

        // Agregar la propiedad scoreString formateada
        $song->scoreString = $this->formatScoreString(
            $song->formattedScore,
            $user->score_format ?? 'POINT_100',
            $denominator
        );


        return $song;
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
