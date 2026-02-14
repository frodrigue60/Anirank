@foreach ($posts as $post)
    @php
        $thumbnailUrl = $post->thumbnail_src;

        if ($post->thumbnail && Storage::disk('public')->exists($post->thumbnail)) {
            $thumbnailUrl = Storage::url($post->thumbnail);
        }
    @endphp
    {{-- <article class="tarjeta">
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
    </article> --}}
    <div class="d-flex flex-column p-2 color1 rounded-1" {{-- style="border: solid 1px red;" --}}>
        <a class="d-flex gap-3 no-deco " data-bs-toggle="collapse" href="#collapseExample{{ $post->id }}" role="button"
            aria-expanded="false" aria-controls="collapseExample{{ $post->id }}">
            <div class="d-flex">
                <img class="rounded-1" src="{{ $thumbnailUrl }}" alt="" style="max-width: 80px;height:auto;">
            </div>
            <div>
                <div>
                    <p class="d-inline-block text-truncate">{{ $post->title }}</p>
                    <p> {{ $post->season->name }} {{ $post->year->name }}</p>
                </div>
            </div>
        </a>
        <div class="collapse mt-2" id="collapseExample{{ $post->id }}">
            <div class="">
                <table class="table table-sm  mb-0">
                    <thead>
                        <tr>
                            <th>Type</th>
                            <th>Name</th>
                            <th>Flag</th>
                        </tr>
                    </thead>
                    <tbody>
                        @foreach ($post->songs as $song)
                            @foreach ($song->songVariants as $variant)
                                <tr>
                                    <td>{{ $variant->song->slug }} {{ $variant->slug }}</td>
                                    <td>
                                        <a href="{{ $variant->url }}">
                                            {{ $variant->song->name }} {{ $variant->slug }}
                                        </a>
                                    </td>
                                    <td></td>
                                </tr>
                            @endforeach
                        @endforeach
                    </tbody>
                </table>
            </div>
        </div>
    </div>
@endforeach
