<?php

use Illuminate\Support\Facades\Route;
use Illuminate\Support\Facades\Auth;

use App\Http\Controllers\{
    PostController,
    ArtistController,
    UserController,
    ReportController,
    CommentController,
    SongController,
    SongVariantController,
    SeasonController,
    YearController,
    StudioController,
    ProducerController,
    UserRequestController,
    PlaylistController
};

use App\Http\Controllers\Admin\{
    PostController as AdminPostController,
    ArtistController as AdminArtistController,
    UserController as AdminUserController,
    ReportController as AdminReportController,
    UserRequestController as AdminUserRequestController,
    SongController as AdminSongController,
    VideoController as AdminVideoController,
    SongVariantController as AdminSongVariantController,
    YearController as AdminYearController,
    SeasonController as AdminSeasonController,
    CommentController as AdminCommentController,
    StudioController as AdminStudioController,
    ProducerController as AdminProducerController
};

/*
|--------------------------------------------------------------------------
| Public Routes
|--------------------------------------------------------------------------
*/

Route::controller(PostController::class)->group(function () {
    Route::get('/', 'index')->name('/');
    Route::get('/themes', 'themes')->name('themes');
    Route::get('/anime/{slug}', 'show')->name('post.show');
    Route::get('/animes', 'animes')->name('animes');
});

Route::controller(UserController::class)->group(function () {
    Route::get('/welcome', 'welcome')->name('welcome');
    Route::get('/users/{slug}', 'userList')->name('user.list');
    Route::get('/profile', 'index')->name('profile');
    Route::post('/change-score-format', 'changeScoreFormat')->name('change.score.format');
    Route::post('/upload-profile-pic', 'uploadProfilePic')->name('upload.profile.pic');
    Route::post('/upload-banner-pic', 'uploadBannerPic')->name('upload.banner.pic');
    Route::get('/favorites', 'favorites')->name('favorites');
});

Route::controller(SongController::class)->group(function () {
    Route::get('/anime/{anime-slug}/{song-slug}', 'show')->name('songs.show');
    Route::get('/seasonal', 'seasonal')->name('seasonal');
    Route::get('/ranking', 'ranking')->name('ranking');
});

Route::get('/offline', fn() => view('offline'));

// Resources
Route::resource('artists', ArtistController::class)->only(['index', 'show']);
Route::resource('years', YearController::class);
Route::resource('seasons', SeasonController::class);
Route::resource('studios', StudioController::class);
Route::resource('producers', ProducerController::class);
Route::resource('playlists', PlaylistController::class);
Route::resource('variants', SongVariantController::class);

/*
|--------------------------------------------------------------------------
| Admin Routes (Staff Middleware)
|--------------------------------------------------------------------------
*/

Route::middleware('staff')->prefix('admin')->as('admin.')->group(function () {
    Route::get('/dashboard', [AdminPostController::class, 'dashboard'])->name('dashboard');

    // Songs & Variants
    Route::get('songs/{song}/variants/add', [AdminSongController::class, 'addVariant'])->name('songs.variants.add');
    Route::get('songs/{song}/variants', [AdminSongController::class, 'variants'])->name('songs.variants');
    Route::resource('songs', AdminSongController::class);

    Route::get('/variants/{variant}/videos', [AdminSongVariantController::class, 'videos'])->name('variants.videos');
    Route::get('/variants/{variant}/videos/add', [AdminSongVariantController::class, 'addVideos'])->name('variants.videos.add');
    Route::resource('variants', AdminSongVariantController::class);

    // Common Resources
    Route::resource('videos', AdminVideoController::class);
    Route::resource('requests', AdminUserRequestController::class);
    Route::resource('comments', AdminCommentController::class);
    Route::resource('studios', AdminStudioController::class);
    Route::resource('producers', AdminProducerController::class);

    // Reports
    Route::get('/reports/{report}/toggle', [AdminReportController::class, 'toggleStatus'])->name('reports.toggle');
    Route::resource('reports', AdminReportController::class);

    // Posts
    Route::controller(AdminPostController::class)->prefix('posts')->as('posts.')->group(function () {
        Route::post('/search', 'search')->name('search');
        Route::post('/{post}/toggle-status', 'toggleStatus')->name('toggle.status');
        Route::get('/{post}/songs/add', 'addSong')->name('songs.add');
        Route::get('/{post}/songs', 'songs')->name('songs');
        Route::post('/search-animes', 'searchInAnilist')->name('search.animes');
        Route::get('/get-by-id/{id}', 'getById')->name('get.by.id');
        Route::post('/get-seasonal-animes', 'getSeasonalAnimes')->name('get.seasonal.animes');
        Route::get('/{post}/force-update', 'forceUpdate')->name('force.update');
        Route::get('/sync-all', 'syncAllFromAnilist')->name('sync.all');
        Route::get('/wipe', 'wipePosts')->name('wipe');
    });
    Route::resource('posts', AdminPostController::class);

    // Artists & Users search
    Route::get('/artists/search', [AdminArtistController::class, 'searchArtist'])->name('artists.search');
    Route::resource('artists', AdminArtistController::class);

    Route::get('/users/search', [AdminUserController::class, 'searchUser'])->name('users.search');
    Route::resource('users', AdminUserController::class);

    // Years & Seasons toggle
    Route::get('years/{year}/toggle', [AdminYearController::class, 'toggle'])->name('years.toggle');
    Route::resource('years', AdminYearController::class);

    Route::get('seasons/{season}/toggle', [AdminSeasonController::class, 'toggle'])->name('seasons.toggle');
    Route::resource('seasons', AdminSeasonController::class);
});

/*
|--------------------------------------------------------------------------
| Auth & Interaction Routes
|--------------------------------------------------------------------------
*/

Auth::routes();

// Comments Interactions
Route::controller(CommentController::class)->prefix('comments')->as('comments.')->group(function () {
    Route::post('/{comment}/like', 'like')->name('like');
    Route::post('/{comment}/dislike', 'dislike')->name('dislike');
    Route::post('/{parentComment}/reply', 'reply')->name('reply');
});
Route::resource('comments', CommentController::class);

// Requests
Route::resource('requests', UserRequestController::class)->only(['create', 'store'])->names([
    'create' => 'request.create',
    'store' => 'request.store'
]);

// Song Variant Interactions
Route::controller(SongVariantController::class)->group(function () {
    Route::post('/variant/{variant}/rate', 'rate')->name('variant.rate');
    Route::post('variants/{variant}/like', 'like')->name('variants.like');
    Route::post('variants/{variant}/dislike', 'dislike')->name('variants.dislike');
    Route::post('variants/{variant}/favorite', 'toggleFavorite')->name('variants.toggle.favorite');
});

// Reports Store
Route::post('reports/store', [ReportController::class, 'store'])->name('reports.store');
