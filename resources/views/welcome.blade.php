@extends('layouts.app')
@section('meta')
    @if (Request::is('welcome'))
        <title>Welcome to AniRank</title>
        <meta title="Welcome to AniRank">
    @endif
@endsection
@section('content')
    <div class="container ">
        <section>
            <article>
                <div class="text-center">
                    <h1>Welcome to AniRank</h1>
                </div>
                <div>
                    <h2><strong>What is it?</strong></h2>
                    <h4>It is a personal project that arises from the idea of ​​qualifying the openings and endings of the
                        animes that are released in each season, and thus be able to place them in a ranking, to determine
                        which
                        are the best according to the opinion of users.
                    </h4>
                </div>
            </article>
        </section>
        <br>
        <hr>
        <section>
            <article>
                <div>
                    <h2>
                        Search <strong>openings</strong> and <strong>endings</strong>, by anime titles, artists, or season
                    </h2>
                </div>
                <div>
                    <h5>
                        On the site a search engine has been implemented that, although it is simple, will help you to
                        search
                        for the opening/ending, artist, or specific season in which you are interested. In this way it will
                        be
                        easier for you to find the musical themes, so you can add them to your favorites and rate them.
                    </h5>
                </div>
            </article>
        </section>
        <br>
        <hr>
        <section>
            <article>
                <div>
                    <h2>
                        Ranking OP & ED animes
                    </h2>
                </div>
                <div>
                    <h5>
                        Based on the opinions of the users, an average rating of the musical themes is generated, and they
                        are
                        placed in a ranking that can be divided according to the current anime season, or in a global
                        ranking of
                        all time.
                    </h5>
                </div>
            </article>
        </section>
        <div class="d-flex justify-content-center">
            <a href="{{ route('home') }}" class="btn btn-ms color4">Let's Do it!</a>
        </div>
    </div>
@endsection
