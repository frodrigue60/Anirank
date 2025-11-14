<?php

use Illuminate\Support\Facades\Route;
use Illuminate\Support\Facades\Auth;

use App\Http\Controllers\PostController as PostController;
use App\Http\Controllers\Admin\PostController as AdminPostController;
use App\Http\Controllers\ArtistController as ArtistController;
use App\Http\Controllers\Admin\ArtistController as AdminArtistController;
use App\Http\Controllers\UserController as UserController;
use App\Http\Controllers\Admin\UserController as AdminUserController;
use App\Http\Controllers\ReportController as ReportController;
use App\Http\Controllers\Admin\ReportController as AdminReportController;
use App\Http\Controllers\Admin\UserRequestController as AdminUserRequestController;
use App\Http\Controllers\Admin\SongController as AdminSongController;
use App\Http\Controllers\Admin\VideoController as AdminVideoController;
use App\Http\Controllers\Admin\SongVariantController as AdminSongVariantController;
use App\Http\Controllers\SongVariantController as SongVariantController;
use App\Http\Controllers\Admin\YearController as AdminYearController;
use App\Http\Controllers\Admin\SeasonController as AdminSeasonController;
use App\Http\Controllers\CommentController as CommentController;
use App\Http\Controllers\SeasonController;
use App\Http\Controllers\YearController;
use App\Http\Controllers\Admin\CommentControlle as AdminCommentController;
use App\Http\Controllers\SongController as SongController;
use App\Http\Controllers\admin\StudioController as adminStudioController;
use App\Http\Controllers\StudioController as StudioController;
use App\Http\Controllers\ProducerController as ProducerController;

//POST PUBLIC
Route::get('/',       [PostController::class, 'index'])->name('/');
Route::get('/themes', [PostController::class, 'themes'])->name('themes');
Route::get('/welcome',       [UserController::class, 'welcome'])->name('welcome');
Route::get('/users/{slug}', [UserController::class, 'userList'])->name('user.list');

Route::get('/anime/{slug}',   [PostController::class, 'show'])->name('post.show');
Route::get('/animes',   [PostController::class, 'animes'])->name('animes');

Route::get('/offline', function () {
    return view('offline');
});

//VARIANTS PUBLIC
//Route::get('/anime/{anime_slug}/{song_slug}/{variant_slug}', [SongVariantController::class, 'show'])->name('variants.show');

//SONGS PUBLIC
Route::get('/anime/{anime_slug}/{song_slug}', [SongController::class, 'show'])->name('songs.show');
Route::get('/seasonal',   [SongController::class, 'seasonal'])->name('seasonal');
Route::get('/ranking',   [SongController::class, 'ranking'])->name('ranking');

//ARTISTS PUBLIC
//Route::get('/artists/{slug}',    [ArtistController::class, 'show'])->name('artists.show');
//Route::get('/artists',    [ArtistController::class, 'index'])->name('artists.index');
Route::resource('artists', ArtistController::class)->only(['index','show']);

//YEARS PUBLIC
Route::resource('years', YearController::class);

//SEASONS PUBLIC
Route::resource('seasons', SeasonController::class);

//STUDIO PUBLIC
Route::resource('studios', StudioController::class);

//ADMIN ROUTES
Route::group(['middleware' => 'staff'], function () {
    Route::prefix('admin')->group(function () {
        //DASHBOARD
        Route::get('/dashboard', [AdminPostController::class, 'dashboard'])->name('admin.dashboard');

        //SONGS
        Route::get('songs/{song}/variants/add', [AdminSongController::class, 'addVariant'])->name('admin.songs.variants.add');
        Route::get('songs/{song}/variants', [AdminSongController::class, 'variants'])->name('admin.songs.variants');
        Route::resource('songs', AdminSongController::class, ['as' => 'admin']);

        //VARIANTS

        Route::get('/variants/{variant}/videos', [AdminSongVariantController::class, 'videos'])->name('admin.variants.videos');
        Route::get('/variants/{variant}/videos/add', [AdminSongVariantController::class, 'addVideos'])->name('admin.variants.videos.add');
        Route::resource('variants', AdminSongVariantController::class, ['as' => 'admin']);

        //VIDEOS
        Route::resource('videos', AdminVideoController::class, ['as' => 'admin']);

        //REQUESTS
        Route::resource('requests', AdminUserRequestController::class, ['as' => 'admin']);

        //REPORTS

        Route::get('/reports/{report}/toggle', [AdminReportController::class, 'toggleStatus'])->name('admin.reports.toggle');
        Route::resource('reports', AdminReportController::class, ['as' => 'admin']);

        //POSTS
        Route::post('/posts/search', [AdminPostController::class, 'search'])->name('admin.posts.search');
        Route::post('/posts/{post}/toggle-status', [AdminPostController::class, 'toggleStatus'])->name('admin.posts.toggle.status');
        Route::get('/posts/{post}/songs/add', [AdminPostController::class, 'addSong'])->name('admin.posts.songs.add');
        Route::get('/posts/{post}/songs', [AdminPostController::class, 'songs'])->name('admin.posts.songs');
        Route::post('/posts/search-animes', [AdminPostController::class, 'searchInAnilist'])->name('admin.search.animes');
        Route::get('/posts/get-by-id/{id}', [AdminPostController::class, 'getById'])->name('get.by.id');
        Route::post('/posts/get-seasonal-animes', [AdminPostController::class, 'getSeasonalAnimes'])->name('get.seasonal.animes');
        Route::get('/posts/{id}/force-update', [AdminPostController::class, 'forceUpdate'])->name('admin.posts.force.update');
        Route::get('/posts/wipe', [AdminPostController::class, 'wipePosts'])->name('posts.wipe');
        Route::resource('posts', AdminPostController::class, ['as' => 'admin']);

        //ARTISTS
        Route::get('/artists/search', [AdminArtistController::class, 'searchArtist'])->name('admin.artists.search');
        Route::resource('artists', AdminArtistController::class, ['as' => 'admin']);


        //USERS
        Route::get('/users/search', [AdminUserController::class, 'searchUser'])->name('admin.users.search');
        Route::resource('users', AdminUserController::class, ['as' => 'admin']);

        //YEARS
        Route::get('years/{year}/toggle', [AdminYearController::class, 'toggle'])->name('admin.years.toggle');
        Route::resource('years', AdminYearController::class, ['as' => 'admin']);


        //SEASONS
        Route::get('seasons/{season}/toggle', [AdminSeasonController::class, 'toggle'])->name('admin.seasons.toggle');
        Route::resource('seasons', AdminSeasonController::class, ['as' => 'admin']);

        //COMMENTS
        Route::resource('comments', AdminCommentController::class, ['as' => 'admin']);

        //STUDIOS
        Route::resource('studios', adminStudioController::class, ['as' => 'admin']);
    });
});

//AUTH ROUTES
Auth::routes();

//COMMENTS
Route::post('/comments/{comment}/like', [CommentController::class, 'like'])->name('comments.like');
Route::post('/comments/{comment}/dislike', [CommentController::class, 'dislike'])->name('comments.dislike');
Route::post('/comments/{parentComment}/reply', [CommentController::class, 'reply'])->name('comments.reply');
Route::resource('comments', CommentController::class);

//REQUEST ROUTES
Route::get('/requests/create', [App\Http\Controllers\UserRequestController::class, 'create'])->name('request.create');
Route::post('/requests/store', [App\Http\Controllers\UserRequestController::class, 'store'])->name('request.store');

//USER ROUTES
Route::post('/change-score-format', [App\Http\Controllers\UserController::class, 'changeScoreFormat'])->name('change.score.format');
Route::get('/profile', [App\Http\Controllers\UserController::class, 'index'])->name('profile');
Route::post('/upload-profile-pic', [App\Http\Controllers\UserController::class, 'uploadProfilePic'])->name('upload.profile.pic');
Route::post('/upload-banner-pic', [App\Http\Controllers\UserController::class, 'uploadBannerPic'])->name('upload.banner.pic');

//Variants ROUTES
Route::get('/favorites', [UserController::class, 'favorites'])->name('favorites');

//REPORTS
Route::post('reports/store', [ReportController::class, 'store'])->name('reports.store');

//PLAYLIST ROUTES
Route::resource('playlists', App\Http\Controllers\PlaylistController::class);

//SONG VARIANT ROUTES
Route::post('/variant/{variant}/rate', [SongVariantController::class, 'rate'])->name('variant.rate');
Route::post('variants/{variant}/like', [SongVariantController::class, 'like'])->name('variants.like');
Route::post('variants/{variant}/dislike', [SongVariantController::class, 'dislike'])->name('variants.dislike');
Route::post('variants/{variant}/favorite', [SongVariantController::class, 'toggleFavorite'])->name('variants.toggle.favorite');
