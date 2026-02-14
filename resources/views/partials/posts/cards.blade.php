@foreach ($posts as $post)
    @php
        $thumbnail_url = $post->thumbnail_src;

        if ($post->thumbnail && Storage::disk('public')->exists($post->thumbnail)) {
            $thumbnail_url = Storage::url($post->thumbnail);
        }
    @endphp
    <article class="tarjeta">
        <a class="no-deco" href="{{ $post->url }}" rel="nofollow noopener noreferrer">
            <div class="textos">
                <div class="tarjeta-header ">
                    <h3 class="text-shadow text-uppercase post-titles">{{ $post->title }}</h3>
                </div>
                <img class="thumb" loading="lazy" src="{{ $thumbnail_url }}" alt="{{ $post->title }}"
                    title="{{ $post->title }}">
                <div class="tarjeta-footer justify-content-center">
                    <span class="">
                        {{ $post->songs->count() }} <i class="fa-solid fa-music"></i>
                    </span>
                </div>
            </div>
        </a>
    </article>
@endforeach
