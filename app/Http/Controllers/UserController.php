<?php

namespace App\Http\Controllers;

use App\Models\Season;
use Illuminate\Http\Request;
use App\Models\User;
use Illuminate\Pagination\Paginator;
use Illuminate\Support\Collection;
use Illuminate\Pagination\LengthAwarePaginator;
use Illuminate\Support\Facades\Redirect;
use Illuminate\Support\Facades\Validator;
use Illuminate\Support\Facades\DB;
use stdClass;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Storage;
use App\Models\SongVariant;
use App\Models\Year;

class UserController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
        $score_formats = [
            ['name' => ' 100 Point (55/100)', 'value' => 'POINT_100'],
            ['name' => '10 Point Decimal (5.5/10)', 'value' => 'POINT_10_DECIMAL'],
            ['name' => '10 Point (5/10)', 'value' => 'POINT_10'],
            ['name' => '5 Star (3/5)', 'value' => 'POINT_5'],
        ];

        if (Auth::check()) {
            $user = Auth::user();
            return view('public.users.profile', compact('score_formats', 'user'));
        } else {
            return redirect()->route('/')->with('warning', 'Please login');
        }
    }

    /**
     * Show the form for creating a new resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function create()
    {
        return abort(404);
    }

    /**
     * Store a newly created resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @return \Illuminate\Http\Response
     */
    public function store(Request $request)
    {
        return abort(404);
    }

    /**
     * Display the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function show($id)
    {
        return abort(404);
    }

    /**
     * Show the form for editing the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function edit($id)
    {
        return abort(404);
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
        return abort(404);
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function destroy($id)
    {
        return abort(404);
    }

    public function userList($slug)
    {
        //dd($userId);
        $user = User::where('slug', $slug)->select('id', 'slug', 'score_format', 'image', 'banner', 'name')->first();

        if (!$user) {
            return redirect('/')->with('warning', 'Invalid user');
        }

        $years = Year::all();
        $seasons = Season::all();
        $types = $this->filterTypesSortChar()['types'];
        $sortMethods = $this->filterTypesSortChar()['sortMethods'];
        //dd($songs);
        return view('public.users.list', compact('seasons', 'years', 'sortMethods', 'types', 'user'));
    }


    public function favorites(Request $request)
    {
        if (!Auth::check()) {
            return redirect()->route('login');
        }

        $status = true;

        $user = Auth::user();
        //$score_format = $user->score_format;

        $season = Season::where('name', $request->season)->first();
        $year = Year::where('name', $request->year)->first();
        $type = $request->type;
        $sort = $request->sort;
        $name = $request->name;

        $years = Year::all();
        $seasons = Season::all();

        $types = $this->filterTypesSortChar()['types'];
        $sortMethods = $this->filterTypesSortChar()['sortMethods'];

        $types = $this->filterTypesSortChar()['types'];
        $sortMethods = $this->filterTypesSortChar()['sortMethods'];

        return view('public.users.favorites', compact('seasons', 'years', 'sortMethods', 'types', 'user'));
    }

    public function paginate($songs, $perPage = 18, $page = null, $options = [])
    {
        $page = Paginator::resolveCurrentPage();
        $options = ['path' => Paginator::resolveCurrentPath()];
        $items = $songs instanceof Collection ? $songs : Collection::make($songs);
        $songs = new LengthAwarePaginator($items->forPage($page, $perPage), $items->count(), $perPage, $page, $options);
        return $songs;
    }

    public function sortPosts($sort, $songs)
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

    public function setScore($songs, $score_format)
    {
        $songs->each(function ($song) use ($score_format) {
            $song->score = null;
            $song->user_score = null;
            switch ($score_format) {
                case 'POINT_100':
                    $song->score = round($song->averageRating);
                    if ($song->rating != null) {
                        $song->user_score = round($song->rating);
                    }

                    break;
                case 'POINT_10_DECIMAL':
                    $song->score = round($song->averageRating / 10, 1);
                    if ($song->rating != null) {
                        $song->user_score = round($song->rating / 10, 1);
                    }

                    break;
                case 'POINT_10':
                    $song->score = round($song->averageRating / 10);
                    if ($song->rating != null) {
                        $song->user_score = round($song->rating / 10);
                    }

                    break;
                case 'POINT_5':
                    $song->score = round($song->averageRating / 20);
                    if ($song->rating != null) {
                        $song->user_score = round($song->rating / 20);
                    }

                    break;
                default:
                    $song->score = round($song->averageRating / 10);
                    if ($song->rating != null) {
                        $song->user_score = round($song->rating / 10);
                    }

                    break;
            }
        });
        return $songs;
    }

    public function uploadProfilePic(Request $request)
    {
        if ($request->hasFile('image')) {
            $validator = Validator::make($request->all(), [
                'image' => 'mimes:png,jpg,jpeg,webp|max:2048'
            ]);

            if ($validator->fails()) {
                return redirect(route('profile'))->with('error', 'Invalid file format or size.');
            }

            $user = Auth::user();
            $old_image = $user->image;

            $extension = $request->image->extension();
            $file_name = $user->slug . '-' . time() . '.' . $extension;
            $path = 'profile';

            $storedPath = $request->image->storeAs($path, $file_name, 'public');

            if ($storedPath) {
                $user->update(['image' => $storedPath]);

                if ($old_image && Storage::disk('public')->exists($old_image)) {
                    Storage::disk('public')->delete($old_image);
                }

                return redirect(route('profile'))->with('success', 'Profile picture updated successfully!');
            }
        }

        return redirect(route('profile'))->with('warning', 'File not found');
    }
    public function uploadBannerPic(Request $request)
    {
        if ($request->hasFile('banner')) {
            $validator = Validator::make($request->all(), [
                'banner' => 'mimes:png,jpg,jpeg,webp|max:2048'
            ]);

            if ($validator->fails()) {
                return redirect(route('profile'))->with('error', 'Invalid file format or size.');
            }

            $user = Auth::user();
            $old_banner = $user->banner;

            $extension = $request->banner->extension();
            $file_name = $user->slug . '-' . time() . '.' . $extension;
            $path = 'banner';

            $storedPath = $request->banner->storeAs($path, $file_name, 'public');

            if ($storedPath) {
                $user->update(['banner' => $storedPath]);

                if ($old_banner && Storage::disk('public')->exists($old_banner)) {
                    Storage::disk('public')->delete($old_banner);
                }

                return redirect(route('profile'))->with('success', 'Banner updated successfully!');
            }
        }
        return redirect(route('profile'))->with('warning', 'File not found');
    }
    public function changeScoreFormat(Request $request)
    {
        if ($request->score_format == 'null') {
            return redirect()->back()->with('warning', 'score method not changed');
        }

        $validator = Validator::make($request->all(), [
            'score_format' => 'required|in:POINT_100,POINT_10_DECIMAL,POINT_10,POINT_5'
        ]);

        if ($validator->fails()) {
            return Redirect::back()->with('error', '¡Ooops!');
        }


        if (Auth::check()) {
            $user = Auth::user();
            $user = User::find($user->id);
            $user->score_format = $request->score_format;
            $user->update();

            return redirect()->back()->with('success', 'score method changed successfully');
        } else {
            return redirect(route('login'));
        }
    }

    public function welcome()
    {
        return view('welcome');
    }
    public function SeasonsYears($tags)
    {
        $tagNames = [];
        $tagYears = [];

        for ($i = 0; $i < count($tags); $i++) {
            [$name, $year] = explode(' ', $tags[$i]->name);

            if (!in_array($year, $tagNames)) {
                $years[] = ['name' => $year, 'value' => $year];
                $tagNames[] = $year; // Agregamos el año al array de nombres para evitar duplicados
            }

            if (!in_array($name, $tagYears)) {
                $seasons[] = ['name' => $name, 'value' => $name];
                $tagYears[] = $name; // Agregamos el año al array de nombres para evitar duplicados
            }
        }

        $data = [
            'years' => $years,
            'seasons' => $seasons
        ];
        return $data;
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
            $isDecimalFormat = false; // Determina si el formato permite decimales

            if ($user) {
                switch ($user->score_format) {
                    case 'POINT_100':
                        $factor = 1;
                        break;
                    case 'POINT_10_DECIMAL':
                        $factor = 0.1;
                        $isDecimalFormat = true;
                        break;
                    case 'POINT_10':
                        $factor = 1 / 10;
                        break;
                    case 'POINT_5':
                        $factor = 1 / 20;
                        $isDecimalFormat = true;
                        break;
                    default:
                        $factor = 1;
                        break;
                }

                if ($userRating = $this->getUserRating($variant->id, $user->id)) {
                    $variant->userScore = $isDecimalFormat
                        ? round($userRating->rating * $factor, 1) // Conserva 1 decimal
                        : (int) round($userRating->rating * $factor); // Fuerza entero
                }
            }

            $variant->score = $isDecimalFormat
                ? round($variant->averageRating * $factor, 1) // Conserva 1 decimal
                : (int) round($variant->averageRating * $factor); // Fuerza entero
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
                $song_variants = $song_variants->sortByDesc('created_at');
                return $song_variants;
                break;
        }
    }
    public function getUserRating($song_variant_id, $user_id)
    {
        $userRating = DB::table('ratings')
            ->where('rateable_type', SongVariant::class)
            ->where('rateable_id', $song_variant_id)
            ->where('user_id', $user_id)
            ->first(['rating']);

        return $userRating;
    }
}
