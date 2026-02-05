<div class="modal fade" id="modal-search" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-dialog-scrollable">
        <div class="modal-content  ">
            {{-- <div class="modal-header">
                <span>Search</span>

            </div> --}}
            <div class="modal-body p-2">
                <div class="my-3 d-flex align-items-center">
                    {{--  <div class="p-2">
                        <i class="fa-solid fa-search"></i>
                    </div> --}}
                    <form class="d-flex w-100" role="search" id="form-search" data-url-base="{{ config('app.url') }}">
                        <input id="searchInputModal" class="form-control " type="search" placeholder="Search"
                            aria-label="Search" autocomplete="off">
                    </form>
                    {{-- <div class="">
                        <button type="button" class="btn " data-bs-dismiss="modal" aria-label="Close">
                            <i class="fa-solid fa-xmark"></i>
                        </button>
                    </div> --}}
                </div>

                <div id="modalBody" class="overflow-hidden ">
                    <div class="res text-nowrap text-truncate hidden">
                        <span class="catTitle fs-5 fw-bold">Anime</span>
                        <div id="posts">
                        </div>
                        <hr>
                        <span class="catTitle fs-5 fw-bold">Artist</span>
                        <div id="artists">
                        </div>
                        <hr>
                        <span class="catTitle fs-5 fw-bold">Users</span>
                        <div id="users">
                        </div>
                    </div>
                </div>
            </div>
            {{-- <div class="modal-footer justify-content-center">
                <div class="d-flex">
                    <a href="{{ route('themes') }}" class="btn btn-primary color3">More
                        options</a>
                </div>
            </div> --}}
        </div>
    </div>
</div>
