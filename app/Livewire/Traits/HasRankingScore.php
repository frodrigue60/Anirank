<?php

namespace App\Livewire\Traits;

use App\Models\Song;
use Illuminate\Support\Facades\DB;

trait HasRankingScore
{
    private function setScoreSongs($songs, $user = null)
    {
        $songs->each(function ($song) use ($user) {
            $song->userScore = null;
            $factor = 1;
            $denominator = 100;
            $isDecimalFormat = false;

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

                $userRating = DB::table('ratings')
                    ->where('rateable_id', $song->id)
                    ->where('rateable_type', Song::class)
                    ->where('user_id', $user->id)
                    ->first(['rating']);

                if ($userRating) {
                    $song->userScore = $isDecimalFormat
                        ? round($userRating->rating * $factor, 1)
                        : (int) round($userRating->rating * $factor);
                }
            }

            $song->score = $isDecimalFormat
                ? round($song->averageRating * $factor, 1)
                : (int) round($song->averageRating * $factor);
        });

        return $songs;
    }
}
