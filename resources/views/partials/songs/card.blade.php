@php
    $url = route('song.show', [$song->id, $song->post->slug, $song->slug != null ? $song->slug : $song->type]);
@endphp

<article class="tarjeta">
    <div class="textos">
        <div class="tarjeta-header ">
            <h3 class="text-shadow text-uppercase post-titles">{{ $song->post->title }}</h3>
        </div>
        <div class="{{ $song->type == 'OP' ? 'tag' : 'tag2' }}">
            <span class="tag-content ">{{ $song->slug != null ? $song->slug : $song->type }}</span>
        </div>
        <a class="no-deco" href="{{ $url }}" rel="nofollow noopener noreferrer">
            <img class="thumb" loading="lazy" src="{{ Storage::url($song->post->thumbnail) }}"
                alt="{{ $song->post->title }}" title="{{ $song->post->title }}">
        </a>
        <div class="tarjeta-footer ">
            <span>{{ $song->likeCount }} <i class="fa fa-heart"></i></span>
            <span>{{ $song->view_count }} <i class="fa fa-eye"></i></span>
            @if (isset($song->rating))
                <span style="color: rgb(162, 240, 181)">{{ $song->rating != null ? $song->rating : '0' }} <i
                        class="fa fa-star" aria-hidden="true"></i>
                </span>
            @else
                <span>{{ $song->score != null ? $song->score : 'n/a' }} <i class="fa fa-star" aria-hidden="true"></i>
                </span>
            @endif
        </div>
    </div>
</article>
