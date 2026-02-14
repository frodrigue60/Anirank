<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\SongVariant;
use Illuminate\Http\Request;
use App\Models\Year;
use App\Models\Season;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use Illuminate\Pagination\Paginator;
use Illuminate\Support\Collection;
use Illuminate\Pagination\LengthAwarePaginator;
use Illuminate\Support\Facades\Storage;

class SongVariantController extends Controller
{
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

        return response()->json([
            'video' => $video
        ]);
    }
}
