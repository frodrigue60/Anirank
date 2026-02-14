<?php

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;
use App\Http\Controllers\Api\PostController as apiPostController;
use App\Http\Controllers\Api\SongVariantController as apiSongVariantController;
use App\Http\Controllers\Api\CommentController as apiCommentController;
use App\Http\Controllers\Api\UserController as apiUserController;
use App\Http\Controllers\Api\ArtistController as apiArtistController;
use App\Http\Controllers\Api\SongController as apiSongController;
use App\Http\Controllers\Api\UserRequestController as apiUserRequestController;
use App\Http\Controllers\Api\StudioController as apiStudioController;
use App\Http\Controllers\Api\PlaylistController as apiPlaylistController;

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
|
| Here is where you can register API routes for your application. These
| routes are loaded by the RouteServiceProvider within a group which
| is assigned the "api" middleware group. Enjoy building your API!
|
*/

Route::middleware('auth:sanctum')->get('/user', function (Request $request) {
    return $request->user();
});

#POSTS
//Route::resource('posts', apiPostController::class);
Route::get('search/{q}', [apiPostController::class, 'search'])->name('api.posts.search');
Route::get('animes', [apiPostController::class, 'animes'])->name('api.posts.animes');

#SONGS
//Route::resource('songs', apiSongController::class);
Route::get('songs/seasonal', [apiSongController::class, 'seasonal'])->name('api.songs.seasonal');
Route::get('songs/ranking', [apiSongController::class, 'ranking'])->name('api.songs.ranking');


#SONG VARIANTS
Route::get('variants/{variant}/get-videos', [apiSongVariantController::class, 'getVideos'])->name('api.variants.get-video');

#ARTISTS
//Route::resource('artists', apiArtistController::class);
Route::get('artists/{artist}/filter', [apiArtistController::class, 'songsFilter'])->name('api.artists.songs.filter');

#USERS
Route::get('users/{id}/list', [apiUserController::class, 'userList'])->name('api.users.list');

#COMMENTS
Route::get('songs/{song}/comments', [apiSongController::class, 'comments'])->name('api.songs.comments');

#STUDIOS
Route::get('studios/filter', [apiStudioController::class, 'filter'])->name('api.studios.filter');
Route::get('studios/{studio}/songs', [apiStudioController::class, 'songsFilter'])->name('api.studios.songs');
Route::get('studios/{studio}/animes', [apiStudioController::class, 'postsFilter'])->name('api.studios.posts');

#AUTH ROUTES
Route::middleware(['auth:sanctum'])->group(function () {
    #PLAYLISTS
    Route::resource('playlists', apiPlaylistController::class)->names('api.playlists')->only('index', 'store');
    Route::post('/playlists/{playlist}/toggle-song', [apiPlaylistController::class, 'toggleSong'])->name('api.playlists.toggle.song');

    #VARIANTS

    #SONGS
    Route::post('songs/{song}/like', [apiSongController::class, 'like'])->name('api.songs.like');
    Route::post('songs/{song}/dislike', [apiSongController::class, 'dislike'])->name('api.songs.dislike');
    Route::post('songs/{song}/favorite', [apiSongController::class, 'toggleFavorite'])->name('api.songs.toggle.favorite');
    Route::post('songs/{song}/rate', [apiSongController::class, 'rate'])->name('api.songs.rate');
    Route::post('songs/comments', [apiSongController::class, 'storeComment'])->name('api.songs.store.comment');
    Route::post('songs/reports', [apiSongController::class, 'storeReport'])->name('api.songs.reports');

    #COMMENTS
    Route::post('comments/{id}/like', [apiCommentController::class, 'like'])->name('api.comments.like');
    Route::post('comments/{id}/dislike', [apiCommentController::class, 'dislike'])->name('api.comments.dislike');
    Route::post('comments/{parentComment}/reply', [apiCommentController::class, 'reply'])->name('comments.reply');
    Route::resource('comments', apiCommentController::class, ['as' => 'api']);

    #USER REQUESTS
    Route::resource('requests', apiUserRequestController::class, ['as' => 'api'])->only('store');

    #USER
    //Route::resource('users', apiUserController::class, ['as' => 'api']);
    Route::post('users/avatar', [apiUserController::class, 'uploadAvatar'])->name('api.users.upload.avatar');
    Route::post('users/banner', [apiUserController::class, 'uploadBanner'])->name('api.users.upload.banner');
    Route::post('users/score-format', [apiUserController::class, 'setScoreFormat'])->name('api.users.score.format');
    Route::post('users/favorites', [apiUserController::class, 'favorites'])->name('api.users.favorites');
});
