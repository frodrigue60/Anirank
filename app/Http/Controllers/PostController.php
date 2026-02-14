<?php

namespace App\Http\Controllers;

use App\Models\Artist;
use App\Models\Format;
use Illuminate\Http\Request;
use App\Models\Post;
use App\Models\Season;
use Illuminate\Pagination\Paginator;
use Illuminate\Support\Collection;
use Illuminate\Pagination\LengthAwarePaginator;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use App\Models\Year;
use App\Models\Song;
use App\Models\User;

class PostController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
        $user = Auth::check() ? Auth::User() : null;
        $status = true;

        $recently = Song::with(['post' => function ($q) {
            $q->select('id', 'title', 'slug', 'banner', 'thumbnail');
        }])
            ->whereHas('post', function ($query) use ($status) {
                $query->where('status', $status);
            })
            ->latest()
            ->take(25)
            ->get();

        $popular = Song::with(['post' => function ($q) {
            $q->select('id', 'title', 'slug', 'banner', 'thumbnail');
        }])
            ->withCount('likes')
            ->whereHas('post', function ($query) use ($status) {
                $query->where('status', $status);
            })
            ->orderByDesc('likes_count')
            ->take(25)
            ->get();

        $viewed = Song::with(['post' => function ($q) {
            $q->select('id', 'title', 'slug', 'banner', 'thumbnail');
        }])
            ->whereHas('post', function ($query) use ($status) {
                $query->where('status', $status);
            })
            ->orderByDesc('views')
            ->take(25)
            ->get();

        $openings = Song::with(['post' => function ($q) {
            $q->select('id', 'title', 'slug', 'banner', 'thumbnail');
        }, 'artists:id,name'])
            ->withAvg('ratings', 'rating')
            ->where('type', 'OP')
            ->whereHas('post', function ($q) use ($status) {
                $q->where('status', $status);
            })
            ->orderByDesc('ratings_avg_rating')
            ->take(3)
            ->get();

        $endings = Song::with(['post' => function ($q) {
            $q->select('id', 'title', 'slug', 'banner', 'thumbnail');
        }, 'artists:id,name'])
            ->withAvg('ratings', 'rating')
            ->where('type', 'ED')
            ->whereHas('post', function ($q) use ($status) {
                $q->where('status', $status);
            })
            ->orderByDesc('ratings_avg_rating')
            ->take(3)
            ->get();

        $weaklyRanking = $openings->concat($endings);
        $weaklyRanking = $this->setScoreSongs($weaklyRanking, $user);

        $artists = Artist::select('id', 'name', 'slug', 'thumbnail')->latest()->take(20)->get();

        $featuredSong = Song::with(['post' => function ($q) {
            $q->select('id', 'title', 'slug', 'banner', 'thumbnail');
        }, 'artists:id,name'])
            ->whereHas('post', function ($q) use ($status) {
                $q->where('status', $status);
            })
            ->inRandomOrder()
            ->first();

        return view('index', compact('weaklyRanking', 'recently', 'popular', 'viewed', 'artists', 'featuredSong'));
    }

    public function animes(Request $request)
    {
        $seasons = Season::all();
        $years = Year::all()->sortByDesc('name');
        $formats = Format::all();

        return view('public.posts.index', compact('seasons', 'years', 'formats'));
    }


    /**
     * Display the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function show(Post $post)
    {
        $post = Post::with(['songs' => function ($q) {
            $q->with(['songVariants.video', 'artists:id,name', 'favorites', 'ratings']);
            $q->withAvg('ratings', 'rating');
        }])->where('id', $post->id)->first();

        $user = Auth::user();

        if (!$post) {
            return redirect(route('home'))->with('warning', 'Post not exist!');
        }

        if (!$post->status) {
            if ($user && $user->isAdmin()) {
                // Admin can view
            } else {
                return redirect('/')->with('danger', $user ? 'User not autorized!' : 'Post status: Private');
            }
        }

        $openings = $post->songs->where('type', 'OP')->sortBy('theme_num');
        $endings = $post->songs->where('type', 'ED')->sortBy('theme_num');

        return view('public.posts.show', compact('post', 'openings', 'endings'));
    }

    public function showAnime($animeSlug)
    {
        $post = Post::with(['songs' => function ($q) {
            $q->with(['songVariants.video', 'artists:id,name', 'favorites', 'ratings']);
            $q->withAvg('ratings', 'rating');
        }])->where('slug', $animeSlug)->first();

        $user = Auth::user();

        if (!$post) {
            return redirect(route('home'))->with('warning', 'Post not exist!');
        }

        if (!$post->status) {
            if ($user && $user->isAdmin()) {
                // Admin can view
            } else {
                return redirect('/')->with('danger', $user ? 'User not autorized!' : 'Post status: Private');
            }
        }

        $openings = $post->songs->where('type', 'OP')->sortBy('theme_num');
        $endings = $post->songs->where('type', 'ED')->sortBy('theme_num');

        return view('public.posts.show', compact('post', 'openings', 'endings'));
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function destroy(Post $post) {}

    public function themes(Request $request)
    {
        $years = Year::select('id', 'name')->orderBy('name', 'desc')->get();
        $seasons = Season::select('id', 'name')->get();
        $types = $this->filterTypesSortChar()['types'];
        $sortMethods = $this->filterTypesSortChar()['sortMethods'];

        return view('public.themes.index', compact('seasons', 'years', 'sortMethods', 'types'));
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

    public function setScoreSongs($songs, $user = null)
    {
        $songs->each(function ($song) use ($user) {

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
        });

        return $songs;
    }

    public function paginate($songs, $perPage = 18, $page = null, $options = [])
    {
        $page = Paginator::resolveCurrentPage();
        $options = ['path' => Paginator::resolveCurrentPath()];
        $items = $songs instanceof Collection ? $songs : Collection::make($songs);
        $songs = new LengthAwarePaginator($items->forPage($page, $perPage), $items->count(), $perPage, $page, $options);
        return $songs;
    }

    public function sortSongs($sort, $songs)
    {
        switch ($sort) {
            case 'title':

                $songs = $songs->sortBy(function ($song) {
                    return $song->post->title;
                });
                return $songs;
                break;
            case 'averageRating':
                $songs = $songs->sortByDesc('averageRating');
                return $songs;
            case 'view_count':
                $songs = $songs->sortByDesc('view_count');
                return $songs;

            case 'likeCount':
                $songs = $songs->sortByDesc('likeCount');
                return $songs;
                break;
            case 'recent':
                $songs = $songs->sortByDesc('created_at');
                return $songs;
                break;

            default:
                $songs = $songs->sortBy(function ($song) {
                    return $song->post->title;
                });
                return $songs;
                break;
        }
    }
    public function sortVariants($sort, $song_variants)
    {
        //dd($song_variants);
        switch ($sort) {
            case 'title':
                $song_variants = $song_variants->sortBy(function ($song_variant) {
                    return $song_variant->song->post->title;
                });
                return $song_variants;
                break;

            case 'averageRating':
                $song_variants = $song_variants->sortByDesc('averageRating');
                return $song_variants;
                break;

            case 'view_count':
                $song_variants = $song_variants->sortByDesc('views');
                return $song_variants;
                break;

            case 'likeCount':
                $song_variants = $song_variants->sortByDesc('likeCount');
                return $song_variants;
                break;

            case 'recent':
                $song_variants = $song_variants->sortByDesc('created_at');
                return $song_variants;
                break;

            default:
                $song_variants = $song_variants->sortBy(function ($song_variant) {
                    return $song_variant->song->post->title;
                });
                return $song_variants;
                break;
        }
    }

    public function filterTypesSortChar()
    {
        $filters = [
            ['name' => 'All', 'value' => 'all'],
            ['name' => 'Only Rated', 'value' => 'rated']
        ];

        $types = [
            ['name' => 'Opening', 'value' => 'OP'],
            ['name' => 'Ending', 'value' => 'ED'],
            ['name' => 'Insert', 'value' => 'INS'],
            ['name' => 'Other', 'value' => 'OTH'],
        ];

        $sortMethods = [
            ['name' => 'Recent', 'value' => 'recent'],
            ['name' => 'Title', 'value' => 'title'],
            ['name' => 'Score', 'value' => 'averageRating'],
            ['name' => 'Views', 'value' => 'view_count'],
            ['name' => 'Popular', 'value' => 'likeCount']
        ];

        $characters = range('A', 'Z');

        $data = [
            'filters' => $filters,
            'types' => $types,
            'sortMethods' => $sortMethods,
            'characters' => $characters
        ];
        return $data;
    }

    public function getUserRating($songId, $userId)
    {
        return DB::table('ratings')
            ->where('rateable_type', Song::class)
            ->where('rateable_id', $songId)
            ->where('user_id', $userId)
            ->first(['rating']);
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
