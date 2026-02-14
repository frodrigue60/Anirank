<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\User;
use Illuminate\Http\Request;
use Illuminate\Pagination\LengthAwarePaginator;
use Illuminate\Pagination\Paginator;
use Illuminate\Support\Collection;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Validator;
use App\Models\SongVariant;
use App\Models\Song;

class UserController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
        $users = User::all(['id', 'name']);
        $users = $this->paginate($users, 5);
        return response()->json([
            'users' => $users,
        ]);
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
        $user = User::where('id', $id)->first(['id', 'name']);

        return response()->json([
            'users' => $user,
        ]);
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

    public function uploadAvatar(Request $request)
    {
        //return response()->json($request->all());
        $validated = $request->validate([
            'image' => 'required|image|mimes:jpeg,png,jpg,webp|max:512',
        ]);

        $user = auth()->user(); // O tu método para obtener el usuario

        $old_user_image = $user->image;

        try {
            // Generar nombre del archivo
            $extension = $request->image->extension();
            $file_name = $user->slug . '-' . time() . '.' . $extension; // Añadimos timestamp para evitar caché
            $path = 'profile';

            // Almacenar el archivo
            $storedPath = $request->file('image')->storeAs(
                $path,
                $file_name,
                'public'
            );

            // Verificación física del archivo
            if (!Storage::disk('public')->exists($storedPath)) {
                throw new \Exception('El archivo no se pudo guardar en el almacenamiento');
            }

            // Actualizar modelo de usuario si es necesario
            $user->image = $storedPath;
            $user->save();

            if (isset($old_user_image) && Storage::disk('public')->exists($old_user_image)) {
                Storage::disk('public')->delete($old_user_image);
            }

            return response()->json([
                'success' => true,
                'message' => 'Avatar actualizado correctamente',
                'avatar_url' => asset("storage/" . $storedPath),
                /* 'file_path' => $storedPath // Para depuración */
            ]);
        } catch (\Exception $e) {
            return response()->json([
                'success' => false,
                'message' => 'Error al subir la imagen',
                'error' => $e->getMessage()
            ], 500);
        }
    }

    public function uploadBanner(Request $request)
    {
        //return response()->json($request->all());
        $validated = $request->validate([
            'banner' => 'required|image|mimes:jpeg,png,jpg,webp|max:512',
        ]);

        $user = auth()->user(); // O tu método para obtener el usuario

        $old_banner_image = $user->banner;

        try {
            // Generar nombre del archivo
            $extension = $request->banner->extension();
            $file_name = $user->slug . '-' . time() . '.' . $extension;
            $path = 'banner';

            // Almacenar el archivo
            $storedPath = $request->file('banner')->storeAs(
                $path,
                $file_name,
                'public'
            );

            // Verificación física del archivo
            if (!Storage::disk('public')->exists($storedPath)) {
                throw new \Exception('El archivo no se pudo guardar en el almacenamiento');
            }

            // Actualizar modelo de usuario si es necesario
            $user->banner = $storedPath;
            $user->save();

            if (isset($old_banner_image) && Storage::disk('public')->exists($old_banner_image)) {
                Storage::disk('public')->delete($old_banner_image);
            }

            return response()->json([
                'success' => true,
                'message' => 'Avatar actualizado correctamente',
                'banner_url' => asset("storage/" . $storedPath),
                /* 'file_path' => $storedPath // Para depuración */
            ]);
        } catch (\Exception $e) {
            return response()->json([
                'success' => false,
                'message' => 'Error al subir la imagen',
                'error' => $e->getMessage()
            ], 500);
        }
    }

    public function setScoreFormat(Request $request)
    {

        //return response()->json([$request->json()->all()]);
        try {
            $validator = Validator::make($request->all(), [
                'score_format' => 'required|in:POINT_100,POINT_10_DECIMAL,POINT_10,POINT_5'
            ]);

            if ($validator->fails()) {
                return response()->json([
                    'success' => false,
                    'message' => $validator->getMessageBag()
                ]);
            }

            $user = Auth::check() ? Auth::User() : null;

            $user = User::find($user->id);
            $user->score_format = $request->score_format;
            $user->update();

            return response()->json([
                'success' => true,
                'message' => 'User score format updated successfully',
                /* 'user' => $user, */
                /* 'request' => $request->all() */
            ]);
        } catch (\Throwable $th) {
            return response()->json([
                'success' => false,
                'message' => $th->getMessage(),
                /* 'user' => $user, */
                /* 'request' => $request->all() */
            ]);
        }
    }

    /* public function userList(Request $request, $slug)
    {
        $user = User::where('slug', $slug)->select('id', 'score_format', 'image', 'banner', 'name')->first();

        if (!$user) {
            return response()->json([
                'success' => false,
                'message' => 'Invalid user',
            ]);
        }

        $status = true;

        $filterBy = $request->filterBy;
        $type = $request->type;
        $sort = $request->sort;
        $name = $request->name;

        $season_id = $request->season_id;
        $year_id = $request->year_id;

        $types = $this->filterTypesSortChar()['types'];
        $sortMethods = $this->filterTypesSortChar()['sortMethods'];

        $song_variants = null;

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
                    ->when($season_id, function ($query, $season_id) {
                        $query->where('season_id', $season_id);
                    })
                    ->when($year_id, function ($query, $year_id) {
                        $query->where('year_id', $year_id);
                    })
                    ->when($name, function ($query, $name) {
                        $query->where('title', 'LIKE', '%' . $name . '%');
                    });
            })
            #SONG VARIANT QUERY
            ->whereLikedBy($user->id)
            ->get();

        $song_variants = $this->setScoreOnlyVariants($song_variants, $user);
        $song_variants = $this->sort_variants($sort, $song_variants);
        $song_variants = $this->paginate($song_variants);

        return response()->json([
            'html' => view('partials.variants.cards', compact('song_variants'))->render(),
            "lastPage" => $song_variants->lastPage(),
            'request' => $request->all()
        ]);
    } */

    public function userList(Request $request, $id)
    {
        //return response()->json(['request' => $request->all()]);

        $user = User::where('id', $id)->select('id', 'slug', 'score_format', 'image', 'banner', 'name')->first();

        //return response()->json(['user' => $user]);

        $status = true;
        $perPage = 15;

        $season_id = $request->season_id;
        $year_id = $request->year_id;
        $type = $request->type;
        $sort = $request->sort;
        $name = $request->name;

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
                        $query->where('season_id', $season_id->id);
                    })
                    ->when($year_id, function ($query, $year_id) {
                        $query->where('year_id', $year_id);
                    });
            })
            #SONG VARIANT QUERY
            ->favoritedBy($user->id)
            ->get();

        $songs = $this->setScoreSongs($songs, $user);
        $songs = $this->sortSongs($sort, $songs);
        $songs = $this->paginate($songs, $perPage, $request->page);

        return response()->json([
            'html' => view('partials.songs.cards-v2', compact('songs'))->render(),
            'songs' => $songs,
            /* 'request' => $request->all(), */
        ]);
    }

    public function favorites(Request $request)
    {
        //return response()->json(['request' => $request->all()]);

        $user = Auth::check() ? Auth::user() : null;

        if (!$user) {
            return response()->json([
                'message' => 'Please login or Re-login'
            ]);
        }

        $status = true;
        $perPage = 15;

        $season_id = $request->season_id;
        $year_id = $request->year_id;
        $type = $request->type;
        $sort = $request->sort;
        $name = $request->name;

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
                        $query->where('season_id', $season_id->id);
                    })
                    ->when($year_id, function ($query, $year_id) {
                        $query->where('year_id', $year_id);
                    });
            })
            #SONG VARIANT QUERY
            ->favoritedBy($user->id)
            ->get();

        $songs = $this->setScoreSongs($songs, $user);
        $songs = $this->sortSongs($sort, $songs);
        $songs = $this->paginate($songs, $perPage, $request->page);

        return response()->json([
            'html' => view('partials.songs.cards-v2', compact('songs'))->render(),
            'songs' => $songs,
            /* 'request' => $request->all(), */
        ]);
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
            ['name' => 'Other', 'value' => 'OTH']
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
    public function getUserRating($song_id, $user_id)
    {
        return DB::table('ratings')
            ->where('rateable_type', Song::class)
            ->where('rateable_id', $song_id)
            ->where('user_id', $user_id)
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
