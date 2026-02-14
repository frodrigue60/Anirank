<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\SongVariant;
use Illuminate\Http\Request;
use App\Models\Year;
use App\Models\Season;
use Illuminate\Support\Facades\Auth;
use App\Models\Reaction;
use App\Models\Favorite;
use Illuminate\Support\Facades\Validator;
use Illuminate\Support\Facades\DB;
use Illuminate\Pagination\Paginator;
use Illuminate\Support\Collection;
use Illuminate\Pagination\LengthAwarePaginator;
use Illuminate\Support\Facades\Storage;

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

    public function like(Request $request)
    {
        $songVariant = SongVariant::find($request->songVariant_id);
        $this->handleReaction($songVariant, 1); // 1 para like
        $songVariant->updateReactionCounters(); // Actualiza los contadores manualmente

        return response()->json([
            'success' => true,
            'message' => 'Liked',
            'likesCount' => $songVariant->likesCount,
            'dislikesCount' => $songVariant->dislikesCount,
        ], 200);
    }

    // Método para dislike
    public function dislike(Request $request)
    {
        $songVariant = SongVariant::find($request->songVariant_id);
        $this->handleReaction($songVariant, -1); // -1 para dislike
        $songVariant->updateReactionCounters(); // Actualiza los contadores manualmente

        return response()->json([
            'success' => true,
            'message' => 'disliked',
            'likesCount' => $songVariant->likesCount,
            'dislikesCount' => $songVariant->dislikesCount,
        ], 200);
    }

    // Método privado para manejar la reacción
    private function handleReaction($songVariant, $type)
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
        $status = true;
        $type = $request->type;
        $currentSeason = Season::where('current', true)->first();
        $currentYear = Year::where('current', true)->first();

        $user = Auth::check() ? Auth::User() : null;

        $themes = SongVariant::with(['song.post'])
            /* SONG QUERY */
            ->whereHas('song', function ($query) use ($type) {
                $query->when($type, function ($query, $type) {
                    $query->where('type', $type);
                });
            })
            ->whereHas('song.post', function ($query) use ($currentSeason, $currentYear, $status) {
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

        $song_variants = $this->setScoreOnlyVariants($themes, $user);
        $themes = view('partials.variants.cards', compact('song_variants'))->render();

        //$data = ['openings' => $openings, 'endings' => $endings];
        $data = ['themes' => $themes];
        return response()->json($data);
    }

    public function ranking(Request $request)
    {
        $validated = $request->validate([
            'ranking_type' => 'required|integer'
        ]);

        $rankingType = $request->ranking_type; //0 = GLOBAL, 1 = SEASONAL

        $user = Auth::check() ? Auth::User() : null;
        $currentSeason = null;
        $currentYear = null;
        $limit = 100;
        $status = true;

        switch ($rankingType) {
            #GLOBAL
            case '0':
                $openings = SongVariant::with(['song.post'])
                    /* SONG QUERY */
                    ->whereHas('song', function ($query) {
                        $query->where('type', 'OP');
                    })
                    ->whereHas('song.post', function ($query) use ($status) {
                        /* POST QUERY */
                        $query->where('status', $status);
                    })
                    /* SONG VARIANT QUERY */
                    ->get()
                    ->sortByDesc('averageRating')
                    ->take($limit);

                $endings = SongVariant::with(['song.post'])
                    /* SONG QUERY */
                    ->whereHas('song', function ($query) {
                        $query->where('type', 'ED');
                    })
                    ->whereHas('song.post', function ($query) use ($status) {
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

                $openings = SongVariant::with(['song.post'])
                    /* SONG QUERY */
                    ->whereHas('song', function ($query) {
                        $query->where('type', 'OP');
                    })
                    ->whereHas('song.post', function ($query) use ($currentSeason, $currentYear, $status) {
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

                $endings = SongVariant::with(['song.post'])
                    /* SONG QUERY */
                    ->whereHas('song', function ($query) {
                        $query->where('type', 'ED');
                    })
                    ->whereHas('song.post', function ($query) use ($currentSeason, $currentYear, $status) {
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
                $openings = SongVariant::with(['song.post'])
                    /* SONG QUERY */
                    ->whereHas('song', function ($query) {
                        $query->where('type', 'OP');
                    })
                    ->whereHas('song.post', function ($query) use ($status) {
                        /* POST QUERY */
                        $query->where('status', $status);
                    })
                    /* SONG VARIANT QUERY */
                    ->get()
                    ->sortByDesc('averageRating')
                    ->take($limit);

                $endings = SongVariant::with(['song.post'])
                    /* SONG QUERY */
                    ->whereHas('song', function ($query) {
                        $query->where('type', 'ED');
                    })
                    ->whereHas('song.post', function ($query) use ($status) {
                        /* POST QUERY */
                        $query->where('status', $status);
                    })
                    /* SONG VARIANT QUERY */
                    ->get()
                    ->sortByDesc('averageRating')
                    ->take($limit);
                break;
        }

        //$openings = $this->paginate($openings, 5)->withQueryString();
        //$endings = $this->paginate($endings, 5)->withQueryString();

        $openings = $this->setScoreOnlyVariants($openings, $user);
        //$openings = view('partials.top.positions', ['items' => $openings])->render();

        $endings = $this->setScoreOnlyVariants($endings, $user);
        //$endings = view('partials.top.positions', ['items' => $endings])->render();

        $data = [
            'openings' => view('partials.top.cards', ['items' => $openings])->render(),
            'endings' => view('partials.top.cards', ['items' => $endings])->render(),
            'currentSeason' => $currentSeason,
            'currentYear' => $currentYear
        ];
        //$data = ['themes' => $themes];
        return response()->json($data);
    }

    public function toggleFavorite(Request $request)
    {

        $songVariant = SongVariant::find($request->songVariant_id);
        $user = Auth::user();

        // Verificar si el post ya está en favoritos
        $favorite = Favorite::where('user_id', $user->id)
            ->where('favoritable_id', $songVariant->id)
            ->where('favoritable_type', SongVariant::class)
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
                'favoritable_id' => $songVariant->id,
                'favoritable_type' => SongVariant::class,
            ]);

            return response()->json([
                'success' => true,
                'message' => 'Added to favorites',
                'favorite' => true,
            ], 200);
        }
    }

    public function rate(Request $request, $variant_id)
    {

        $user = Auth::check() ? Auth::user() : null;
        $songVariant = SongVariant::find($variant_id);

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
        $songVariant->rateOnce($score, $user->id);
        $average = $songVariant->averageRating;
        $average = round($average * $factor, 1);

        return response()->json([
            'success' => true,
            'message' => 'Rated sucessfully',
            'score' => $score,
            'average' => $average,
        ], 200);
    }

    public function comments(SongVariant $variant)
    {
        $comments = $variant->comments;
        return response()->json([
            'success' => true,
            'message' => 'Comments',
            'comments' => $comments,
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

        $song_variants = null;
        $status = true;

        $song_variants = SongVariant::with(['song.post'])
            #SONG QUERY
            ->whereHas('song', function ($query) use ($type) {
                $query->when($type, function ($query, $type) {
                    $query->where('type', $type);
                });
            })
            #POST QUERY
            ->whereHas('song.post', function ($query) use ($name, $season_id, $year_id, $status) {
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

        $song_variants = $this->setScoreOnlyVariants($song_variants, $user);
        $song_variants = $this->sort_variants($sort, $song_variants);
        $song_variants = $this->paginate($song_variants);

        $view = view('partials.variants.cards', compact('song_variants'))->render();
        return response()->json(['html' => $view, "lastPage" => $song_variants->lastPage()]);
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

    public function sort_variants($sort, $song_variants)
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

    public function getUserRating(int $song_variant_id, int $user_id)
    {
        return DB::table('ratings')
            ->where('rateable_type', SongVariant::class)
            ->where('rateable_id', $song_variant_id)
            ->where('user_id', $user_id)
            ->first(['rating']);
    }

    public function getVideos(SongVariant $variant)
    {

        $video = $variant->video;
        $video->publicUrl = Storage::url($video->video_src);

        //$variants = $song->songVariants;

        return response()->json([
            'video' => $video
        ]);
    }
}
