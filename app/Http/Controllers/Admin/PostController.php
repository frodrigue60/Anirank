<?php

namespace App\Http\Controllers\Admin;

use App\Models\Post;
use App\Models\Artist;
use App\Http\Controllers\Controller;
use App\Models\ExternalLink;
use App\Models\Format;
use App\Models\Producer;
use App\Models\Season;
use App\Models\Studio;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Str;
use Illuminate\Support\Facades\Validator;
use Illuminate\Support\Facades\Redirect;
use Intervention\Image\ImageManagerStatic as Image;
use Illuminate\Support\Facades\Storage;
use App\Models\Year;
use GuzzleHttp\Client;
use App\Services\Breadcrumb;

class PostController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
        $posts = Post::orderByDesc('id')->paginate(20);

        $breadcrumb = Breadcrumb::generate([
            [
                'name' => 'Index',
                'url' => route('admin.posts.index'),
            ],
        ]);
        //dd($posts);
        return view('admin.posts.index', compact('posts', 'breadcrumb'));
    }

    /**
     * Show the form for creating a new resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function create()
    {
        $artists = Artist::all();
        $seasons = Season::all();
        $years = Year::all();

        $types = [
            ['name' => 'Opening', 'value' => 'OP'],
            ['name' => 'Ending', 'value' => 'ED'],
            ['name' => 'Insert', 'value' => 'INS'],
            ['name' => 'Other', 'value' => 'OTH'],
        ];

        $postStatus = [
            ['name' => 'Stagged', 'value' => false],
            ['name' => 'Published', 'value' => true]
        ];

        $breadcrumb = Breadcrumb::generate([
            [
                'name' => 'Index',
                'url' => route('admin.posts.index'),
            ],
            [
                'name' => 'Create post',
                'url' => '',
            ],
        ]);

        return view('admin.posts.create', compact('years', 'seasons', 'types', 'artists', 'postStatus', 'breadcrumb'));
    }

    /**
     * Store a newly created resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @return \Illuminate\Http\Response
     */
    public function store(Request $request)
    {
        //dd($request->all());
        /* return redirect()->back()
                         ->with('error', 'Hubo un error al procesar el formulario.')
                         ->withInput(); */
        $user = Auth::User()->type;
        if ($user == 'admin' || $user == 'editor' || $user == 'creator') {

            $post = new Post;
            $post->title = $request->title;
            $post->slug = Str::slug($request->title);
            $post->description = $request->description;
            $post->year_id = $request->year;
            $post->season_id = $request->season;

            switch (Auth::user()->type) {
                case 'creator':
                    $post->status = false;
                    break;
                case 'admin' || 'editor':
                    if ($request->postStatus == null) {
                        $post->status = false;
                    } else {
                        $post->status = $request->postStatus;
                    }
                    break;
                default:
                    $post->status = true;
                    break;
            }

            $this->storePostImages($post, $request);

            if ($post->save()) {
                //$post->retag($request->tags);
                $post = Post::where('slug', $post->slug);
                $msg = 'Post created successfully';
                return redirect(route('admin.posts.songs', $post->id))->with('success', $msg);
            } else {
                $msg = 'Somethis was wrong!';
                return redirect(route('admin.posts.index'))->with('error', $msg);
            }
        } else {
            $error = 'User is not authorized!';
            return redirect(route('admin.posts.index'))->with('error', $error);
        }
    }

    /**
     * Display the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function show($id)
    {
        if (Auth::check() && Auth::user()->isStaff()) {

            $post = Post::findOrFail($id);
            $score_format = Auth::user()->score_format;

            $breadcrumb = Breadcrumb::generate([
                [
                    'name' => 'Index',
                    'url' => route('admin.posts.index'),
                ],
                [
                    'name' => $post->title,
                    'url' => route('admin.posts.show', $post->id),
                ],
            ]);

            $ops = $post->songs->filter(function ($song) {
                return $song->type === 'OP';
            });
            $eds = $post->songs->filter(function ($song) {
                return $song->type === 'ED';
            });

            //dd($ops, $eds);
            return view('admin.posts.show', compact('post', 'score_format', 'ops', 'eds', 'breadcrumb'));
        }
    }

    /**
     * Show the form for editing the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function edit($id)
    {
        $post = Post::find($id);
        $artists = Artist::all();
        $seasons = Season::all();
        $years = Year::all();

        $types = [
            ['name' => 'Opening', 'value' => 'OP'],
            ['name' => 'Ending', 'value' => 'ED'],
            ['name' => 'Insert', 'value' => 'INS'],
            ['name' => 'Other', 'value' => 'OTH'],
        ];

        $postStatus = [
            ['name' => 'Stagged', 'value' => false],
            ['name' => 'Published', 'value' => true]
        ];

        $breadcrumb = Breadcrumb::generate([
            [
                'name' => 'Index',
                'url' => route('admin.posts.index'),
            ],
            [
                'name' => $post->title,
                'url' => '',
            ],
        ]);

        return view('admin.posts.edit', compact('post', 'types', 'artists', 'postStatus', 'breadcrumb', 'years', 'seasons'));
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
        //dd($request->all());
        $user = Auth::User()->type;
        if ($user == 'admin' || $user == 'editor' || $user == 'creator') {
            $post = Post::find($id);
            $old_thumbnail = $post->thumbnail;
            $old_banner = $post->banner;
            $post->title = $request->title;
            $post->slug = Str::slug($request->title);
            $post->description = $request->description;


            switch (Auth::user()->type) {
                case 'creator':
                    $post->status = false;
                    break;
                case 'admin' || 'editor':
                    if ($request->postStatus == null) {
                        $post->status = false;
                    } else {
                        $post->status = $request->postStatus;
                    }
                    break;
                default:
                    $post->status = false;
                    break;
            }

            $this->storePostImages($post, $request);

            if ($post->update()) {
                if ($old_thumbnail && $old_thumbnail !== $post->thumbnail) {
                    Storage::disk('public')->delete($old_thumbnail);
                }
                if ($old_banner && $old_banner !== $post->banner) {
                    Storage::disk('public')->delete($old_banner);
                }
                return redirect(route('admin.posts.index'))->with('success', 'Post Updated Successfully');
            } else {
                return redirect(route('admin.posts.index'))->with('error', 'Something has wrong');
            }
        } else {
            $error = 'User is not authorized!';
            return redirect(route('admin.posts.index'))->with('error', $error);
        }
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function destroy($id)
    {
        $post = Post::find($id);

        if ($post->delete()) {
            return Redirect::route('admin.posts.index')->with('success', 'Post Deleted successfully!');
        } else {
            return Redirect::route('admin.posts.index')->with('error', 'Post has been not deleted!');
        }
    }

    //seach posts in admin pannel
    public function search(Request $request)
    {
        $posts = Post::query()
            ->where('title', 'LIKE', '%' . $request->input('q') . '%')
            ->paginate(10);
        //dd($posts);
        return view('admin.posts.index', compact('posts'));
    }

    public function toggleStatus(Post $post)
    {
        try {
            $post->update([
                'status' => !$post->status
            ]);

            return redirect()->back()->with('success', 'Post status updated: ' . $post->id);
        } catch (\Throwable $th) {
            return redirect(route('admin.posts.index'))->with('error', $th->getMessage());
        }
    }

    public function searchInAnilist(Request $request)
    {
        $q = $request->q;
        $type = $request->type;


        $variables = [
            'search' => $q,
            'format_in' => $type,
        ];

        $query = $this->buildGraphQLQuerySearch();

        $client = new Client();
        $response = $client->post('https://graphql.anilist.co', [
            'json' => [
                'query' => $query,
                'variables' => $variables,
            ]
        ]);

        $body = $response->getBody()->__toString();

        $json = json_decode($body);
        $data = $json->data->Page->media;
        $posts = [];
        foreach ($data as $item) {
            array_push($posts, $item);
        }

        return view('admin.posts.select', compact('posts'));
    }
    public function getById($anilist_id)
    {
        $variables = [
            "id" => $anilist_id
        ];

        $client = new \GuzzleHttp\Client;
        $response = $client->post('https://graphql.anilist.co', [
            'json' => [
                'query' => $this->buildGraphQLQueryId(),
                'variables' => $variables,
            ]
        ]);
        $body = $response->getBody()->__toString();
        $json = json_decode($body);

        $data[] = $json->data->Media;
        //dd($data);
        $this->generateMassive($data);
        $success = 'Single post created successfully';
        return redirect(route('admin.posts.index'))->with('success', $success);
    }
    public function syncAllFromAnilist()
    {
        $posts = Post::whereNotNull('anilist_id')->get();
        $client = new Client();

        foreach ($posts as $post) {
            try {
                $variables = ["id" => $post->anilist_id];
                $response = $client->post('https://graphql.anilist.co', [
                    'json' => [
                        'query' => $this->buildGraphQLQueryId(),
                        'variables' => $variables,
                    ]
                ]);
                $body = $response->getBody()->__toString();
                $json = json_decode($body);
                $item = $json->data->Media;

                $this->updatePostFromAnilistData($post, $item);
            } catch (\Throwable $th) {
                // Skip or log error
                continue;
            }
        }

        return redirect(route('admin.posts.index'))->with('success', 'All posts synchronized with AniList');
    }

    private function updatePostFromAnilistData($post, $item)
    {
        $post->description = $item->description;
        $post->anilist_id = $item->id;

        $externalLinks = $item->externalLinks;
        $studios = $item->studios->nodes;
        $idStudios = [];
        $idProducers = [];
        $idLinks = [];
        $format_name = $item->format;

        foreach ($studios as $key => $value) {
            if ($value->isAnimationStudio) {
                $studio = Studio::firstOrCreate(
                    ['slug' => Str::slug($value->name)],
                    ['name' =>  $value->name, 'slug' => Str::slug($value->name)]
                );
                array_push($idStudios, $studio->id);
            } else {
                $producer = Producer::firstOrCreate(
                    ['slug' => Str::slug($value->name)],
                    ['name' =>  $value->name, 'slug' => Str::slug($value->name)]
                );
                array_push($idProducers, $producer->id);
            }
        }

        foreach ($externalLinks as $key => $value) {
            $externalLink = ExternalLink::firstOrCreate(
                ['url' => $value->url],
                ['icon' => $value->icon, 'name' =>  $value->site, 'type' => $value->type, 'url' => $value->url]
            );
            array_push($idLinks, $externalLink->id);
        }

        $format = Format::firstOrCreate(
            ['slug' => Str::slug($format_name)],
            ['name' =>  $format_name, 'slug' => Str::slug($format_name)]
        );

        $post->format()->associate($format);

        $this->saveAnimeBanner($item, $post);
        $this->saveAnimeThumbnail($item, $post);

        if (!empty($item->season) && !empty($item->seasonYear)) {
            $season = Season::firstOrCreate(['name' =>  $item->season]);
            $post->season_id = $season->id;
            $year = Year::firstOrCreate(['name' =>  $item->seasonYear]);
            $post->year_id = $year->id;
        } else if (!$item->season and !$item->seasonYear && isset($item->startDate->month)) {
            $month_al = $item->startDate->month;
            $year_al = $item->startDate->year;

            $season = Season::firstOrCreate(['name' =>  $this->assignSeason($month_al)]);
            $post->season_id = $season->id;

            $year = Year::firstOrCreate(['name' =>  $year_al]);
            $post->year_id = $year->id;
        }

        if ($post->save()) {
            $post->studios()->sync($idStudios);
            $post->producers()->sync($idProducers);
            $post->externalLinks()->sync($idLinks);
        }
    }

    public function generateMassive($data)
    {
        foreach ($data as $item) {
            $post_exist = Post::where('title', $item->title->romaji)->first();
            if ($post_exist) {
                continue;
            }
            $post = new Post;
            $post->title = $item->title->romaji;
            $post->slug = Str::slug($post->title);
            $post->status = true;

            $this->updatePostFromAnilistData($post, $item);
        }
    }

    public function forceUpdate($id)
    {
        $post = Post::find($id);

        try {
            $variables = ["id" => $post->anilist_id];
            $client = new \GuzzleHttp\Client;
            $response = $client->post('https://graphql.anilist.co', [
                'json' => [
                    'query' => $this->buildGraphQLQueryId(),
                    'variables' => $variables,
                ]
            ]);
            $body = $response->getBody()->__toString();
            $json = json_decode($body);

            $this->updatePostFromAnilistData($post, $json->data->Media);

            return redirect(route('post.show', $post->slug))->with('success', 'Post updated');
        } catch (\Throwable $th) {
            return redirect(route('post.show', $post->slug))->with('error', $th->getMessage());
        }
    }
    public function wipePosts()
    {
        $posts = Post::all();

        foreach ($posts as $post) {
            $post->delete();
        }

        $thumbnail_files = Storage::disk('public')->files('thumbnails');
        Storage::disk('public')->delete($thumbnail_files);

        $banner_files = Storage::disk('public')->files('anime_banner');
        Storage::disk('public')->delete($banner_files);

        $success = 'All posts deleted';
        return redirect(route('admin.posts.index'))->with('success', $success);
    }
    function buildGraphQLQuerySearch()
    {
        $query = '
            query ($search: String, $format_in: [MediaFormat]) {
                Page {
                    media (search: $search, type: ANIME, format_in: $format_in) {
                        id
                        title {
                            romaji
                            english
                            native
                        }
                        coverImage {
                            extraLarge
                        }
                        bannerImage
                    }
                }
            }
        ';
        return $query;
    }
    function buildGraphQLQuerySeasonal()
    {
        $query = '
            query ($year: Int, $season: MediaSeason, $page: Int, $perPage: Int, $format_in: [MediaFormat]) {
                Page (page: $page, perPage: $perPage) {
                    pageInfo {
                        total
                        perPage
                        currentPage
                        lastPage
                        hasNextPage
                    }
                    media (
                            seasonYear: $year,
                            season: $season,
                            type: ANIME,
                            format_in: $format_in,
                            isAdult:false
                        ) {
                        id
                        title {
                            romaji
                            english
                            native
                        }
                        description
                        season
                        seasonYear
                        averageScore
                        format
                        genres
                        studios {
                            nodes {
                                name
                                isAnimationStudio
                            }
                        }
                        coverImage {
                            large
                            extraLarge
                        }
                        bannerImage
                        trailer{
                            id
                            site
                            thumbnail
                        }
                        format
                        externalLinks {
                            icon
                            type
                            site
                            url
                        }
                        startDate {
                            month
                            year
                        }
                        synonyms
                            }
                        }
            }';
        return $query;
    }
    function buildGraphQLQueryId()
    {
        $query = '
        query ($id: Int) { # Define which variables will be used in the query (id)
            Media (id: $id, type: ANIME) { # Insert our variables into the query arguments (id) (type: ANIME is hard-coded in the query)
                id
                title {
                    romaji
                    english
                    native
                }
                description
                season
                seasonYear
                format
                genres
                averageScore
                studios {
                    nodes {
                        name
                        isAnimationStudio
                    }
                }
                coverImage {
                    large
                    extraLarge
                }
                bannerImage
                episodes
                trailer{
                    id
                    site
                    thumbnail
                }
                format
                externalLinks {
                    icon
                    type
                    site
                    url
                }
                startDate {
                    month
                    year
                }
                synonyms
            }
        }
        ';
        return $query;
    }
    /* Used by generateMassive Method
     *
     */
    function saveAnimeThumbnail($item, $post)
    {
        if ($item->coverImage->extraLarge != null) {
            $client = new Client();
            $response = $client->get($item->coverImage->extraLarge);
            $imageContent = $response->getBody()->getContents();

            if (extension_loaded('gd')) {
                $imageContent = Image::make($imageContent)->encode('webp', 100); //->resize(150, 212)
                $file_name = Str::slug($post->slug) . '-' . time() . '.webp';
            }
            $path = 'thumbnails/' . $file_name;
            $this->storeSingleImage($path, $imageContent);
            $post->thumbnail = $path;
            $post->thumbnail_src = $item->coverImage->extraLarge;
        }
        return $post;
    }
    /* Used by generateMassive Method
     *
     */
    function saveAnimeBanner($item, $post)
    {
        if ($item->bannerImage != null) {

            $client = new Client();
            $response = $client->get($item->bannerImage);
            $imageContent = $response->getBody()->getContents();

            if (extension_loaded('gd')) {
                $file_name = Str::slug($post->slug) . '-' . time() . '.webp';
                $imageContent = Image::make($imageContent)->encode('webp', 100); //->resize(150, 212)
            }
            $path = 'anime_banner/' . $file_name;
            $this->storeSingleImage($path, $imageContent);
            $post->banner = $path;
            $post->banner_src = $item->bannerImage;
        }
        return $post;
    }

    /* Used by Store and Update Method
     *
     */
    public function storePostImages($post, $request)
    {
        /* Thumnail with file store */
        if ($request->hasFile('file')) {

            $validator = Validator::make($request->all(), [
                'file' => 'mimes:png,jpg,jpeg,webp|max:2048'
            ]);

            if ($validator->fails()) {
                $errors = $validator->getMessageBag();
                $request->flash();
                return Redirect::back()
                    ->with('error', $errors);
            }

            $imageContent = $request->file;

            if (extension_loaded('gd')) {
                $file_name = Str::slug($request->title) . '-' . time() . '.' . 'webp';
                $imageContent = Image::make($request->file)->encode('webp', 100);
            } else {
                $file_extension = $request->file->extension();
                $file_name = Str::slug($request->title) . '-' . time() . '.' . $file_extension;
            }
            $path = 'thumbnails/' . $file_name;
            $this->storeSingleImage($path, $imageContent);
            $post->thumbnail = $path;
        } else {
            /* Thumbnail witn url store */
            if ($request->thumbnail_src != null) {

                $post->thumbnail_src = $request->thumbnail_src;

                $client = new Client();
                $response = $client->get($request->thumbnail_src);
                $imageContent = $response->getBody()->getContents();

                if (extension_loaded('gd')) {
                    $file_name = Str::slug($request->title) . '-' . time() . '.' . 'webp';
                    $imageContent = Image::make($imageContent)->encode('webp', 100);
                } else {
                    $headers = $response->getHeaders();
                    $contentType = $headers['Content-Type'][0] ?? null;

                    $extension = match ($contentType) {
                        'image/jpeg' => 'jpg',
                        'image/png'  => 'png',
                        'image/gif'  => 'gif',
                        'image/webp' => 'webp',
                        default      => 'bin',
                    };

                    $file_name = Str::slug($request->title) . '-' . time() . '.' . $extension;
                }

                $path = 'thumbnails/' . $file_name;
                $this->storeSingleImage($path, $imageContent);
                $post->thumbnail = $path;
            } else {
                $request->flash();
                return Redirect::back()->with('error', "Post not created, thumbnail image not found");
            }
        }

        /* Banner with file store */
        if ($request->hasFile('banner')) {
            $validator = Validator::make($request->all(), [
                'banner' => 'mimes:png,jpg,jpeg,webp|max:2048'
            ]);

            if ($validator->fails()) {
                $errors = $validator->getMessageBag();
                $request->flash();
                return Redirect::back()
                    ->with('error', $errors);
            }

            $imageContent = $request->banner;

            if (extension_loaded('gd')) {
                $file_name = Str::slug($request->title) . '-' . time() . '.' . 'webp';
                $imageContent = Image::make($request->banner)->encode('webp', 100);
            } else {
                $extension = $request->file->extension();
                $file_name = Str::slug($request->title) . '-' . time() . '.' . $extension;
            }
            $path = 'anime_banner/' . $file_name;
            $this->storeSingleImage($path, $imageContent);
            $post->banner = $path;
        } else {
            /* Bannter with url store */
            if ($request->banner_src != null) {

                $post->banner_src = $request->banner_src;

                $client = new Client();
                $response = $client->get($request->banner_src);
                $imageContent = $response->getBody()->getContents();

                if (extension_loaded('gd')) {
                    $file_name = Str::slug($request->title) . '-' . time() . '.' . 'webp';
                    $imageContent = Image::make($imageContent)->encode('webp', 100);
                } else {
                    $headers = $response->getHeaders();
                    $contentType = $headers['Content-Type'][0] ?? null;
                    $extension = match ($contentType) {
                        'image/jpeg' => 'jpg',
                        'image/png'  => 'png',
                        'image/gif'  => 'gif',
                        'image/webp' => 'webp',
                        default      => 'bin', // Extensión por defecto si no se reconoce
                    };

                    $file_name = Str::slug($request->title) . '-' . time() . '.' . $extension;
                }
                $path = 'anime_banner/' . $file_name;
                $this->storeSingleImage($path, $imageContent);
                $post->banner = $path;
            }
        }
        return $post;
    }

    public function storeSingleImage($path, $imageContent)
    {
        Storage::disk('public')->put($path, $imageContent);
    }

    public function addSong($post_id)
    {
        $post = Post::find($post_id);
        $seasons = Season::all();
        $years = Year::all();
        $types = [
            ['name' => 'Opening', 'value' => 'OP'],
            ['name' => 'Ending', 'value' => 'ED'],
            ['name' => 'Insert', 'value' => 'INS'],
            ['name' => 'Other', 'value' => 'OTH']
        ];
        $artists = Artist::all();

        $post = Post::with('songs')->find($post_id);

        $breadcrumb = Breadcrumb::generate([
            [
                'name' => 'Index',
                'url' => route('admin.posts.index'),
            ],
            [
                'name' => $post->title,
                'url' => route('admin.posts.songs', $post->id),
            ],
            [
                'name' => 'Add song',
                'url' => '',
            ],
        ]);

        return view('admin.songs.create', compact('artists', 'types', 'post', 'seasons', 'years', 'breadcrumb'));
    }

    public function songs($post_id)
    {
        //dd(true);
        $post = Post::with('songs')->find($post_id);

        $breadcrumb = Breadcrumb::generate([
            [
                'name' => 'Index',
                'url' => route('admin.posts.index'),
            ],
            [
                'name' => $post->title,
                'url' => route('admin.posts.songs', $post->id),
            ],
        ]);


        return view('admin.songs.manage', compact('post', 'breadcrumb'));
    }

    public function dashboard()
    {
        return view('admin.dashboard');
    }

    private function assignSeason(int $month)
    {
        if ($month == 12 || $month == 1 || $month == 2) {
            return 'WINTER';
        } else {
            if ($month == 3 || $month == 4 || $month == 5) {
                return 'SPRING';
            } else {
                if ($month == 6 || $month == 7 || $month == 8) {
                    return 'SUMMER';
                } else {
                    return 'FALL';
                }
            }
        }
    }
}
