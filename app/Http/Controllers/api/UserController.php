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
    public function uploadAvatar(Request $request)
    {
        $validated = $request->validate([
            'image' => 'required|image|mimes:jpeg,png,jpg,webp|max:512',
        ]);

        $user = Auth::user();
        $old_user_image = $user->image;

        try {
            $extension = $request->image->extension();
            $file_name = $user->slug . '-' . time() . '.' . $extension;
            $path = 'profile';

            $storedPath = $request->file('image')->storeAs(
                $path,
                $file_name,
                'public'
            );

            if (!Storage::disk('public')->exists($storedPath)) {
                throw new \Exception('El archivo no se pudo guardar en el almacenamiento');
            }

            $user->image = $storedPath;
            $user->save();

            if (isset($old_user_image) && Storage::disk('public')->exists($old_user_image)) {
                Storage::disk('public')->delete($old_user_image);
            }

            return response()->json([
                'success' => true,
                'message' => 'Avatar actualizado correctamente',
                'avatar_url' => asset("storage/" . $storedPath),
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
        $validated = $request->validate([
            'banner' => 'required|image|mimes:jpeg,png,jpg,webp|max:512',
        ]);

        $user = Auth::user();
        $old_banner_image = $user->banner;

        try {
            $extension = $request->banner->extension();
            $file_name = $user->slug . '-' . time() . '.' . $extension;
            $path = 'banner';

            $storedPath = $request->file('banner')->storeAs(
                $path,
                $file_name,
                'public'
            );

            if (!Storage::disk('public')->exists($storedPath)) {
                throw new \Exception('El archivo no se pudo guardar en el almacenamiento');
            }

            $user->banner = $storedPath;
            $user->save();

            if (isset($old_banner_image) && Storage::disk('public')->exists($old_banner_image)) {
                Storage::disk('public')->delete($old_banner_image);
            }

            return response()->json([
                'success' => true,
                'message' => 'Banner actualizado correctamente',
                'banner_url' => asset("storage/" . $storedPath),
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
            ]);
        } catch (\Throwable $th) {
            return response()->json([
                'success' => false,
                'message' => $th->getMessage(),
            ]);
        }
    }

    public function userList(Request $request, $id)
    {
        $user = User::where('id', $id)->select('id', 'slug', 'score_format', 'image', 'banner', 'name')->first();

        $status = true;
        $perPage = 15;

        $season_id = $request->season_id;
        $year_id = $request->year_id;
        $type = $request->type;
        $sort = $request->sort;
        $name = $request->name;

        $songs = Song::with(['post'])
            ->when($type, function ($query, $type) {
                $query->where('type', $type);
            })
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
            ->favoritedBy($user->id)
            ->get();

        $songs = $this->setScoreSongs($songs, $user);
        $songs = $this->sortSongs($sort, $songs);
        $songs = $this->paginate($songs, $perPage, $request->page);

        return response()->json([
            'html' => view('partials.songs.cards-v2', compact('songs'))->render(),
            'songs' => $songs,
        ]);
    }

    public function favorites(Request $request)
    {
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
            ->when($type, function ($query, $type) {
                $query->where('type', $type);
            })
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
            ->favoritedBy($user->id)
            ->get();

        $songs = $this->setScoreSongs($songs, $user);
        $songs = $this->sortSongs($sort, $songs);
        $songs = $this->paginate($songs, $perPage, $request->page);

        return response()->json([
            'html' => view('partials.songs.cards-v2', compact('songs'))->render(),
            'songs' => $songs,
        ]);
    }

    public function setScoreOnlyVariants($variants, $user = null)
    {
        $variants->each(function ($variant) use ($user) {
            $variant->userScore = null;
            $factor = 1;
            $isDecimalFormat = false;
            $denominator = 100;

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

    public function setScoreSongs($songs, $user = null)
    {
        $songs->each(function ($song) use ($user) {
            $song->formattedScore = null;
            $song->rawScore = null;
            $song->scoreString = null;

            $factor = 1;
            $isDecimalFormat = false;
            $denominator = 100;

            if ($user) {
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
