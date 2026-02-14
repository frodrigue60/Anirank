@foreach ($songs as $song)
    @php
        $thumbnailUrl = $song->post->thumbnail_src;

        if ($song->post->thumbnail && Storage::disk('public')->exists($song->post->thumbnail)) {
            $thumbnailUrl = Storage::url($song->post->thumbnail);
        }
    @endphp
    <article class="tarjeta">
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
    </article>
@endforeach
