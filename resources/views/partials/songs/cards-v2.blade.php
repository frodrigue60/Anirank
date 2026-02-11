@foreach ($songs as $song)
    @php
        /* $thumb_url = file_exists(asset('/storage/thumbnails/' . $song->post->thumbnail)) ? asset('/storage/thumbnails/' . $song->post->thumbnail) : $song->post->thumbnail_src; */
        $thumbnailUrl = '';

        if (Storage::disk('public')->exists($song->post->thumbnail)) {
            $thumbnailUrl = Storage::url($song->post->thumbnail);
        } else {
            $thumbnailUrl = $song->post->thumbnail_src;
        }
    @endphp
    {{-- <article class="tarjeta">
        <div class="textos">
            <div class="tarjeta-header ">
                <h3 class="text-shadow text-uppercase post-titles">{{ $song->post->title }}</h3>
            </div>
            @if ($song->theme_num > 1)
                <div class="{{ $song->type == 'OP' ? 'tag' : 'tag2' }}">
                    <span class="tag-content ">{{ $song->theme_num > 1 ? $song->slug : $song->type }}</span>
                </div>
            @endif

            <a class="no-deco" href="{{ $song->urlFirstVariant }}" rel="nofollow noopener noreferrer">
                <img class="thumb" loading="lazy" src="{{ $thumbnailUrl }}" alt="{{ $song->post->title }}"
                    title="{{ $song->post->title }}">
            </a>
            <div class="tarjeta-footer ">
                <span>{{ $song->likeCount }} <i class="fa fa-heart"></i></span>
                <span>{{ $song->view_count }} <i class="fa fa-eye"></i></span>
                @if (isset($song->rating))
                    <span style="color: rgb(162, 240, 181)">{{ $song->rating != null ? $song->rating : '0' }} <i
                            class="fa fa-star" aria-hidden="true"></i>
                    </span>
                @else
                    <span>{{ $song->score != null ? $song->score : 'n/a' }} <i class="fa fa-star"
                            aria-hidden="true"></i>
                    </span>
                @endif
            </div>
        </div>
        <div>
            <span>{{ $song->post->title }}</span>
        </div>
    </article> --}}

    <div class="media-card">
        <div class="position-relative overflow-hidden">
            <span class="absolute top-0 end-0 w-100 bg-black/50 rounded-sm text-sm m-1 px-1 z-20"
                style="">{{ $song->slug }}</span>
            @isset($song->score)
                <span class="absolute top-0 start-0 w-100 bg-black/50 rounded-sm text-sm m-1 px-1 z-20"
                    style="">{{ $song->scoreString }}</span>
            @endisset

            <a href="{{ $song->url }}" class="cover">
                <img class="image loaded z-0" loading="lazy" src="{{ $thumbnailUrl }}" alt="{{ $song->post->title }}">
            </a>
        </div>
        <div>
            <a href="{{ $song->url }}" class="title">
                {{ $song->post->title }}
            </a>
        </div>
    </div>
@endforeach
