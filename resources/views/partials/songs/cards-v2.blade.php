@foreach ($songs as $song)
    <div class="media-card">
        <div class="position-relative overflow-hidden">
            <span class="absolute top-0 end-0 w-100 bg-black/50 rounded-sm text-sm m-1 px-1 z-20"
                style="">{{ $song->slug }}</span>
            @isset($song->score)
                <span class="absolute top-0 start-0 w-100 bg-black/50 rounded-sm text-sm m-1 px-1 z-20"
                    style="">{{ $song->scoreString }}</span>
            @endisset
            <a href="{{ route('songs.show', $song->id) }}" class="cover">
                <img class="image loaded z-0" loading="lazy" src="{{ Storage::url($song->post->thumbnail) }}"
                    alt="{{ $song->post->title }}">
            </a>
        </div>
        <div>
            <a href="{{ route('songs.show', $song->id) }}" class="title">
                {{ $song->post->title }}
            </a>
        </div>
    </div>
@endforeach
