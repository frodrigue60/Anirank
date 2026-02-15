@foreach ($songs as $song)
    <div class="media-card">
        <a class="cover">
            <div class="absolute top-0 z-20">
                <div class="flex gap-1 p-1">
                    <span class="bg-primary rounded-sm text-sm px-1">{{ $song->slug }}</span>
                    <span
                        class="bg-primary rounded-sm text-sm px-1">{{ number_format($song->average_rating ?? 0, 1) }}</span>
                </div>
            </div>
            <img class="image loaded" loading="lazy" src="{{ Storage::url($song->post->thumbnail) }}"
                alt="{{ $song->post->title }}">
        </a>
        <a href="{{ $song->url }}" class="title">
            {{ $song->post->title }}
        </a>
    </div>
@endforeach
