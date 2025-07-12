<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use Illuminate\Http\Request;
use App\Models\Song;
use Illuminate\Pagination\Paginator;
use Illuminate\Support\Collection;
use Illuminate\Pagination\LengthAwarePaginator;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use App\Models\Season;
use App\Models\Year;
use App\Models\Reaction;
use App\Models\Favorite;
use Illuminate\Support\Facades\Validator;
use App\Models\Comment;
use App\Models\Report;

class SongController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
        return response()->json(['data' => 'mathod: index']);
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
    public function show($id)
    {
        //
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

    public function like($song_id)
    {
        //return response()->json(['request' => $request->all()]);
        try {
            $song = Song::findOrFail($song_id);
            $this->handleReaction($song, 1); // 1 para like
            $song->updateReactionCounters(); // Actualiza los contadores manualmente

            return response()->json([
                'success' => true,
                'message' => 'Song liked successfully',
                'likesCount' => $song->likesCount,
                'dislikesCount' => $song->dislikesCount,
            ]);
        } catch (\Exception $e) {
            // Otro error general
            //Log::error('Error al crear usuario: ' . $e->getMessage());

            return response()->json([
                'message' => 'Error interno del servidor',
                'exception' => $e->getMessage()
            ], 500); // 500 Internal Server Error
        }
    }

    // Método para dislike
    public function dislike($song_id)
    {
        //return response()->json(['request' => $request->all()]);
        try {
            $song = Song::findOrFail($song_id);
            $this->handleReaction($song, -1); // 1 para like
            $song->updateReactionCounters(); // Actualiza los contadores manualmente

            return response()->json([
                'success' => true,
                'message' => 'Song disliked successfully',
                'likesCount' => $song->likesCount,
                'dislikesCount' => $song->dislikesCount,
            ]);
        } catch (\Exception $e) {
            // Otro error general
            //Log::error('Error al crear usuario: ' . $e->getMessage());

            return response()->json([
                'message' => 'Error interno del servidor',
                'exception' => $e->getMessage()
            ], 500); // 500 Internal Server Error
        }
    }
    // Método privado para manejar la reacción
    private function handleReaction($song, $type)
    {
        $user = Auth::check() ? Auth::user() : null;

        // Buscar si ya existe una reacción del usuario para este post
        $reaction = Reaction::where('user_id', $user->id)
            ->where('reactable_id', $song->id)
            ->where('reactable_type', Song::class)
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
                'reactable_id' => $song->id,
                'reactable_type' => Song::class,
                'type' => $type,
            ]);
        }
    }
    public function toggleFavorite($song_id)
    {

        $song = Song::find($song_id);
        $user = Auth::check() ? Auth::user() : null;

        // Verificar si el post ya está en favoritos
        $favorite = Favorite::where('user_id', $user->id)
            ->where('favoritable_id', $song->id)
            ->where('favoritable_type', Song::class)
            ->first();

        if ($favorite) {
            $favorite->delete();

            return response()->json([
                'success' => true,
                'message' => 'Removed from favorites',
                'favorite' => false,
            ], 200);
        } else {
            Favorite::create([
                'user_id' => $user->id,
                'favoritable_id' => $song->id,
                'favoritable_type' => Song::class,
            ]);

            return response()->json([
                'success' => true,
                'message' => 'Added to favorites',
                'favorite' => true,
            ], 200);
        }
    }
    public function rate(Request $request, $song_id)
    {

        $user = Auth::check() ? Auth::user() : null;
        $song = Song::find($song_id);

        $factor = 1;
        $isDecimalFormat = false;

        $score_format = $user->score_format;

        $validator = Validator::make($request->all(), [
            'score' => 'required|numeric|max:100'
        ]);

        if ($validator->fails()) {
            $messageBag = $validator->getMessageBag();
            return response()->json([
                'success' => false,
                'message' => $messageBag,
                'score' => $request->score,
            ], 200);
        } else {
            $score = $request->score;
        }

        switch ($score_format) {
            case 'POINT_5':
                $score = max(20, min(100, ceil($score / 20) * 20));
                $factor = 1 / 20;
                break;

            case 'POINT_10':
                $score = round($score * 10);
                $factor = 1 / 10;
                break;

            case 'POINT_10_DECIMAL':
                $score = round($score * 10, 1);
                $factor = 0.1;
                break;

            case 'POINT_100':
                $score = round($score);
                $factor = 1;
                break;
            default:
                $score = round($score);
                $factor = 1;
                break;
        }

        // Utilizar el score ajustado
        $song->rateOnce($score, $user->id);
        $average = $song->averageRating;
        $average = round($average * $factor, 1);

        return response()->json([
            'success' => true,
            'message' => 'Rated sucessfully',
            'score' => $score,
            'average' => $average,
        ], 200);
    }

    public function filter(Request $request)
    {
        $user = Auth::check() ? Auth::user() : null;
        $type = $request->type;
        $sort = $request->sort;
        $name = $request->name;
        $season_id = $request->season_id;
        $year_id = $request->year_id;

        $songs = null;
        $status = true;
        $perPage = 15;

        $songs = Song::with(['post'])
            #SONG QUERY
            ->when($type, function ($query, $type) {
                $query->where('type', $type);
            })
            #POST QUERY
            ->whereHas('post', function ($query) use ($name, $season_id, $year_id, $status) {
                $query->where('status', $status)
                    ->when($name, function ($query, $name) {
                        $query->where('title', 'LIKE', '%' . $name . '%');
                    })
                    ->when($season_id, function ($query, $season_id) {
                        $query->where('season_id', $season_id);
                    })
                    ->when($year_id, function ($query, $year_id) {
                        $query->where('year_id', $year_id);
                    });
            })
            #SONG VARIANT QUERY
            ->get();

        //$songs = $this->setScoreOnlyVariants($songs, $user);
        $songs = $this->sortSongs($sort, $songs);
        $songs = $this->paginate($songs, $perPage);

        return response()->json([
            'html' => view('partials.songs.cards-v2', compact('songs'))->render(),
            'songs' => $songs,
        ]);
    }

    public function seasonal(Request $request)
    {
        $status = true;
        $type = $request->type != null ? $request->type : 'OP';
        $currentSeason = Season::where('current', true)->first();
        $currentYear = Year::where('current', true)->first();
        $sort = 'title';
        //$user = Auth::check() ? Auth::User() : null;
        if ($currentSeason && $currentYear) {
            //$songs = $this->setScoreOnlyVariants($themes, $user);

            $songs = Song::with(['post'])
                /* SONG QUERY */
                ->where('type', $type)
                ->whereHas('post', function ($query) use ($currentSeason, $currentYear, $status) {
                    /* POST QUERY */
                    $query->where('status', $status)
                        ->when($currentSeason, function ($query, $currentSeason) {
                            $query->where('season_id', $currentSeason->id);
                        })
                        ->when($currentYear, function ($query, $currentYear) {
                            $query->where('year_id', $currentYear->id);
                        });
                })
                /* SONG VARIANT QUERY */
                ->get();


            $songs = $this->sortSongs($sort, $songs);

            return response()->json([
                'songs' => $songs,
                'html' => view('partials.songs.cards-v2', compact('songs'))->render()
            ]);
        } else {
            $songs = null;

            return response()->json([
                'songs' => $songs,
                'html' => null
            ]);
        }
    }

    public function ranking(Request $request)
    {

        $validator = Validator::make($request->all(['ranking_type']), [
            'ranking_type' => 'nullable|integer|min:0|max:1',
            'page_op' => 'nullable|integer|min:1',
            'page_ed' => 'nullable|integer|min:1'
        ]);

        if ($validator->fails()) {
            return response()->json([
                'message' => 'Validator fail',
                $validator->errors()
            ]);
        }

        //0 = GLOBAL, 1 = SEASONAL
        $rankingType = $request->ranking_type ? $request->ranking_type : 0;

        $user = Auth::check() ? Auth::User() : null;
        $currentSeason = null;
        $currentYear = null;
        $limit = 100;
        $status = true;

        $currentSeason = Season::where('current', true)->first();
        $currentYear = Year::where('current', true)->first();

        $perPage = 10;

        switch ($rankingType) {
            #GLOBAL
            case '0':
                $openings = Song::with(['post'])
                    /* SONG QUERY */
                    ->where('type', 'OP')
                    ->whereHas('post', function ($query) use ($status) {
                        /* POST QUERY */
                        $query->where('status', $status);
                    })
                    /* SONG VARIANT QUERY */
                    ->get()
                    ->sortByDesc('averageRating')
                    ->take($limit);

                $endings = Song::with(['post'])
                    /* SONG QUERY */
                    ->where('type', 'ED')
                    ->whereHas('post', function ($query) use ($status) {
                        /* POST QUERY */
                        $query->where('status', $status);
                    })
                    /* SONG VARIANT QUERY */
                    ->get()
                    ->sortByDesc('averageRating')
                    ->take($limit);
                break;
            #SEASONAL
            case '1':
                $currentSeason = Season::where('current', true)->first();
                $currentYear = Year::where('current', true)->first();

                $openings = Song::with(['post'])
                    /* SONG QUERY */
                    ->where('type', 'OP')
                    ->whereHas('post', function ($query) use ($currentSeason, $currentYear, $status) {
                        /* POST QUERY */
                        $query->where('status', $status)
                            ->when($currentSeason, function ($query, $currentSeason) {
                                $query->where('season_id', $currentSeason->id);
                            })
                            ->when($currentYear, function ($query, $currentYear) {
                                $query->where('year_id', $currentYear->id);
                            });
                    })
                    /* SONG VARIANT QUERY */
                    ->get()
                    ->sortByDesc('averageRating')
                    ->take($limit);

                $endings = Song::with(['post'])
                    /* SONG QUERY */
                    ->where('type', 'ED')
                    ->whereHas('post', function ($query) use ($currentSeason, $currentYear, $status) {
                        /* POST QUERY */
                        $query->where('status', $status)
                            ->when($currentSeason, function ($query, $currentSeason) {
                                $query->where('season_id', $currentSeason->id);
                            })
                            ->when($currentYear, function ($query, $currentYear) {
                                $query->where('year_id', $currentYear->id);
                            });
                    })
                    /* SONG VARIANT QUERY */
                    ->get()
                    ->sortByDesc('averageRating')
                    ->take($limit);


                break;

            default:
                $openings = Song::with(['post'])
                    /* SONG QUERY */
                    ->where('type', 'OP')
                    ->whereHas('post', function ($query) use ($status) {
                        /* POST QUERY */
                        $query->where('status', $status);
                    })
                    /* SONG VARIANT QUERY */
                    ->get()
                    ->sortByDesc('averageRating')
                    ->take($limit);

                $endings = Song::with(['post'])
                    /* SONG QUERY */
                    ->where('type', 'ED')
                    ->whereHas('post', function ($query) use ($status) {
                        /* POST QUERY */
                        $query->where('status', $status);
                    })
                    /* SONG VARIANT QUERY */
                    ->get()
                    ->sortByDesc('averageRating')
                    ->take($limit);
                break;
        }

        $openings = $this->setScoreSongs($openings, $user);
        $openings = $this->paginate($openings, $perPage, 'page_op')->withQueryString();
        $openings = $this->setShowUrl($openings);

        $endings = $this->setScoreSongs($endings, $user);
        $endings = $this->paginate($endings, $perPage, 'page_ed')->withQueryString();
        $endings = $this->setShowUrl($endings);

        return response()->json([
            'success' => true,
            'html' => [
                'openings' => view('partials.top.cards-v2', ['items' => $openings])->render(),
                'endings' => view('partials.top.cards-v2', ['items' => $endings])->render(),
            ],
            'openings' => $openings,
            'endings' => $endings,
            'currentSeason' => $currentSeason,
            'currentYear' => $currentYear
        ]);
    }

    public function setScoreSongs(Collection|array $songs, $user = null): Collection
    {
        $songs->each(function ($song) use ($user) {
            $song->userScore = null;
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

                if ($userRating = $this->getUserRating($song->id, $user->id)) {
                    $song->userScore = $isDecimalFormat
                        ? round($userRating->rating * $factor, 1)
                        : (int) round($userRating->rating * $factor);
                }
            }

            $song->score = $isDecimalFormat
                ? round($song->averageRating * $factor, 1)
                : (int) round($song->averageRating * $factor);

            // Agregar la propiedad scoreString formateada
            $song->scoreString = $this->formatScoreString(
                $song->score,
                $user->score_format ?? 'POINT_100',
                $denominator
            );
        });

        return $songs;
    }

    public function setShowUrl($songs)
    {
        $songs->each(function ($song) {
            $song->url = route('songs.show', [$song->post->slug, $song->slug]);
        });
        return $songs;
    }

    public function getUserRating(int $song_id, int $user_id)
    {
        return DB::table('ratings')
            ->where('rateable_type', Song::class)
            ->where('rateable_id', $song_id)
            ->where('user_id', $user_id)
            ->first(['rating']);
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

    /* ORIGINAL */
    /* public function paginate($collection, $perPage = 18, $page = null, $options = [])
    {
        $page = Paginator::resolveCurrentPage();
        $options = ['path' => Paginator::resolveCurrentPath()];
        $items = $collection instanceof Collection ? $collection : Collection::make($collection);
        $collection = new LengthAwarePaginator($items->forPage($page, $perPage), $items->count(), $perPage, $page, $options);
        return $collection;
    } */

    /* NEW */
    public function paginate($collection, $perPage = 18, $pageParam = 'page', $options = [])
    {
        // Obtiene el número de página desde el request usando el nombre personalizado
        $page = request()->input($pageParam, 1);

        // Define la ruta base sin parámetros de paginación
        $path = url()->current();

        // Opciones personalizadas para el paginador
        $options = array_merge([
            'path' => $path,
            'pageName' => $pageParam, // Nombre del parámetro de página
        ], $options);

        $items = $collection instanceof Collection ? $collection : Collection::make($collection);

        return new LengthAwarePaginator(
            $items->forPage($page, $perPage),
            $items->count(),
            $perPage,
            $page,
            $options
        );
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

    public function storeComment(Request $request)
    {
        $user = Auth::User();

        $validatedData = $request->validate([
            'content' => 'required',
        ]);

        $song = Song::findOrFail($request->song_id);

        $comment = new Comment($validatedData);
        $comment->user_id = $user->id;
        $song->comments()->save($comment);

        $comment::with('user');

        return response()->json([
            'success' => true,
            'message' => 'Commented sucessfully',
            /* 'comments' => $comment, */
            'html' => view('partials.songs.show.comments.comment', ['comment' => $comment])->render(),
        ], 200);
    }

    public function comments($song_id)
    {
        $comments = Comment::with('replies', 'user')
            ->where('commentable_id', $song_id)
            ->where('commentable_type', Song::class)
            ->where('parent_id', null)
            ->get()
            ->sortByDesc('created_at');

        $comments = $this->paginate($comments, 3)->withQueryString();

        return response()->json([
            'comments' => $comments,
            'html' => view('partials.songs.show.comments.comments', ['comments' => $comments])->render()
        ]);
    }

    public function storeReport(Request $request)
    {
        $user = Auth::check() ? Auth::user() : null;

        $validator = Validator::make($request->all(), [
            'song_id' => 'required|integer|exists:songs,id',
            'title' => 'required|max:255|string',
            'content' => 'string|required|max:500',
        ]);

        if ($validator->fails()) {
            return response()->json([
                'success' => false,
                'message' => $validator->getMessageBag(),
            ]);
        }

        $report = new Report();
        $report->song_id = $request->song_id;
        $report->title = $request->title;
        $report->content = $request->content;
        $report->user_id = $user->id;
        $report->source = $request->header('Referer');
        $report->save();

        return response()->json([
            'success' => true,
            'message' => 'Report stored successfully',
        ]);
    }
}
