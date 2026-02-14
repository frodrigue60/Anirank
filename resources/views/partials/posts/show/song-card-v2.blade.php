<div class="card d-flex flex-row gap-3 w-100 rounded-1 py-3 px-2 mb-3">
    <div class="d-flex align-items-center">
        <h3 class="fs-5">{{ $song->slug }}</h3>
    </div>
    <div class="d-flex flex-column me-auto overflow-hidden">
        <div>
            <a href="{{ route('songs.show', $song->id) }}" class="text-decoration-none ">{{ $song->name }}</a>
        </div>
        <div class="overflow-hidden text-nowrap text-truncate">
            @isset($song->artists)

                @foreach ($song->artists as $index => $item)
                    <a class=" text-decoration-none d-inline-block text" href="{{ $item->url }}">{{ $item->name }}</a>
                    @if ($index < count($song->artists) - 1)
                        ,
                    @endif
                @endforeach

            @endisset
        </div>
        {{-- <div class="d-flex flex-row gap-3">
            <div>
                <span><i class="fa-solid fa-film"></i> {{ '1-4' }}</span>
            </div>
            <div>
                <span><i class="fa-solid fa-triangle-exclamation"></i>
                    {{ 'Spoiler' }}</span>
            </div>
        </div> --}}
    </div>
    <div class="d-flex align-items-center">
        <a href="{{ route('songs.show', $song->id) }}" class="btn btn-sm btn-primary rounded-pill text-nowrap">
            <i class="fa-solid fa-play"></i> {{ 'Play' }}
        </a>
    </div>
</div>
