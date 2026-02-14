<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use Illuminate\Http\Request;
use App\Models\Post;
use App\Models\Artist;
use App\Models\User;
use Illuminate\Pagination\Paginator;
use Illuminate\Support\Collection;
use Illuminate\Pagination\LengthAwarePaginator;

class PostController extends Controller
{
    public function search($q)
    {
        $posts = Post::where('title', 'LIKE', '%' . $q . '%')->limit(5)->get(['title', 'slug']);

        $artists = Artist::where('name', 'LIKE', '%' . $q . '%')->limit(5)->get(['name', 'slug']);

        $users = User::where('name', 'LIKE', '%' . $q . '%')->limit(5)->get(['name', 'slug']);

        return response()->json([
            "posts" => $posts,
            "artists" => $artists,
            "users" => $users
        ]);
    }

    public function animes(Request $request)
    {
        $season_id = $request->season_id;
        $year_id = $request->year_id;
        $name = $request->name;
        $format_id = $request->format_id;

        $status = true;

        $posts = Post::where('status', $status)
            ->when($season_id, function ($query, $season_id) {
                $query->where('season_id', $season_id);
            })
            ->when($year_id, function ($query, $year_id) {
                $query->where('year_id', $year_id);
            })
            ->when($name, function ($query, $name) {
                $query->where('title', 'LIKE', '%' . $name . '%');
            })
            ->when($name, function ($query, $name) {
                $query->where('title', 'LIKE', '%' . $name . '%');
            })
            ->when($format_id, function ($query, $format_id) {
                $query->where('format_id', $format_id);
            })
            ->get();

        $posts = $posts->sortBy(function ($post) {
            return $post->title;
        });

        $posts = $this->paginate($posts, 15);
        $posts = $this->setShowUrl($posts);

        return response()->json([
            'html' => view('partials.posts.cards-v2', compact('posts'))->render(),
            'posts' => $posts,
            "lastPage" => $posts->lastPage()
        ]);
    }

    public function paginate($collection, $perPage = 18, $page = null, $options = [])
    {
        $page = Paginator::resolveCurrentPage();
        $options = ['path' => Paginator::resolveCurrentPath()];
        $items = $collection instanceof Collection ? $collection : Collection::make($collection);
        $collection = new LengthAwarePaginator($items->forPage($page, $perPage), $items->count(), $perPage, $page, $options);
        return $collection;
    }

    public function setShowUrl($posts)
    {
        $posts->each(function ($post) {
            $post->url = route('post.show', [$post->slug]);
        });
        return $posts;
    }
}
