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

    public function filter(Request $request)
    {
        $user = Auth::check() ? Auth::user() : null;
        $type = $request->type;
        $sort = $request->sort;
        $name = $request->name;
        $season_id = $request->season_id;
        $year_id = $request->year_id;

        $songVariants = null;
        $status = true;

        $songVariants = SongVariant::with(['song.post'])
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

        $songVariants = $this->setScoreOnlyVariants($songVariants, $user);
        $songVariants = $this->sortVariants($sort, $songVariants);
        $songVariants = $this->paginate($songVariants);

        $view = view('partials.variants.cards', compact('songVariants'))->render();
        return response()->json(['html' => $view, "lastPage" => $songVariants->lastPage()]);
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

    public function sortVariants($sort, $songVariants)
    {
        //dd($song_variants);
        switch ($sort) {
            case 'title':
                $songVariants = $songVariants->sortBy(function ($song_variant) {
                    return $song_variant->song->post->title;
                });
                return $songVariants;
                break;

            case 'averageRating':
                $songVariants = $songVariants->sortByDesc('averageRating');
                return $songVariants;
                break;

            case 'view_count':
                $songVariants = $songVariants->sortByDesc('views');
                return $songVariants;
                break;

            case 'likeCount':
                $songVariants = $songVariants->sortByDesc('likeCount');
                return $songVariants;
                break;

            case 'recent':
                $songVariants = $songVariants->sortByDesc('created_at');
                return $songVariants;
                break;

            default:
                $songVariants = $songVariants->sortBy(function ($song_variant) {
                    return $song_variant->song->post->title;
                });
                return $songVariants;
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

    public function getUserRating(int $songVariantId, int $userId)
    {
        return DB::table('ratings')
            ->where('rateable_type', SongVariant::class)
            ->where('rateable_id', $songVariantId)
            ->where('user_id', $userId)
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
