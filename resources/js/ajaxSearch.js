import { API, csrfToken } from "@/app.js";

const formSearch = document.querySelector("#form-search");
const modalSearch = document.querySelector("#modal-search");
const postsDiv = document.querySelector("#posts");
const artistsDiv = document.querySelector("#artists");
const usersDiv = document.querySelector("#users");
const input = formSearch.querySelector("#searchInputModal");
const modalBody = document.querySelector("#modalBody");
const resDiv = document.querySelector(".res");
let headersData = {};
let params = {};
const baseUrl = formSearch.dataset.urlBase;

let typingTimer;
const delay = 250;

nullValueInput();

// Handle focus when modal is shown via Nav button
document.addEventListener("click", (e) => {
    if (
        e.target.closest('[data-modal-toggle="modal-search"]') ||
        (e.target.closest("button") &&
            e.target.closest("button").onclick &&
            e.target
                .closest("button")
                .onclick.toString()
                .includes("modal-search"))
    ) {
        setTimeout(() => input.focus(), 100);
    }
});

input.addEventListener("input", () => {
    resetDivs();
    insertLoader();

    clearTimeout(typingTimer);
    if (input.value.length >= 1) {
        typingTimer = setTimeout(function () {
            apiSearch();
        }, delay);
    } else {
        resetDivs();
        nullValueInput();
    }
});

async function apiSearch() {
    try {
        headersData = {
            Accept: "application/json, text/html;q=0.9",
            "X-CSRF-TOKEN": csrfToken,
        };

        let q = input.value;
        const response = await API.get(
            API.POSTS.SEARCH(q),
            headersData,
            params,
        );

        resetDivs();

        if (
            response.posts.length === 0 &&
            response.artists.length === 0 &&
            response.users.length === 0
        ) {
            nullValueInput();
        } else {
            response.posts.forEach((post) => {
                let url = baseUrl + "/anime/" + post.slug;
                let resultDiv = document.createElement("div");
                resultDiv.className =
                    "group p-3 hover:bg-white/5 rounded-xl transition-all truncate border border-transparent hover:border-white/5";

                let a = document.createElement("a");
                a.href = url;
                a.className =
                    "text-white/80 group-hover:text-primary font-bold transition-colors flex items-center gap-3";
                a.innerHTML = `<span class="material-symbols-outlined text-white/20 text-[18px]">movie</span> ${post.title}`;

                resultDiv.appendChild(a);
                postsDiv.appendChild(resultDiv);
            });

            response.artists.forEach((artist) => {
                let url = baseUrl + "/artists/" + artist.slug;
                let resultDiv = document.createElement("div");
                resultDiv.className =
                    "group p-3 hover:bg-white/5 rounded-xl transition-all truncate border border-transparent hover:border-white/5";

                let a = document.createElement("a");
                a.href = url;
                a.className =
                    "text-white/80 group-hover:text-primary font-bold transition-colors flex items-center gap-3";
                a.innerHTML = `<span class="material-symbols-outlined text-white/20 text-[18px]">person</span> ${artist.name}`;

                resultDiv.appendChild(a);
                artistsDiv.appendChild(resultDiv);
            });

            response.users.forEach((user) => {
                let url = baseUrl + "/users/" + user.slug;
                let resultDiv = document.createElement("div");
                resultDiv.className =
                    "group p-3 hover:bg-white/5 rounded-xl transition-all truncate border border-transparent hover:border-white/5";

                let a = document.createElement("a");
                a.href = url;
                a.className =
                    "text-white/80 group-hover:text-primary font-bold transition-colors flex items-center gap-3";
                a.innerHTML = `<span class="material-symbols-outlined text-white/20 text-[18px]">group</span> ${user.name}`;

                resultDiv.appendChild(a);
                usersDiv.appendChild(resultDiv);
            });
        }

        resDiv.classList.remove("hidden");
    } catch (error) {
        console.error("Search error:", error);
    }
}

function resetDivs() {
    postsDiv.innerHTML = "";
    artistsDiv.innerHTML = "";
    usersDiv.innerHTML = "";
}

function nullValueInput() {
    postsDiv.appendChild(createResultDiv("No results matched your search"));
    artistsDiv.appendChild(createResultDiv("No results matched your search"));
    usersDiv.appendChild(createResultDiv("No results matched your search"));
}

function createResultDiv(text) {
    let div = document.createElement("div");
    div.className = "p-3 text-white/20 italic text-sm font-medium";

    let span = document.createElement("span");
    span.textContent = text;

    div.appendChild(span);
    return div;
}

function createLoadingElement() {
    const div = document.createElement("div");
    div.className = "flex justify-center py-8";

    const spinner = document.createElement("div");
    spinner.className =
        "animate-spin rounded-full h-8 w-8 border-b-2 border-primary";

    div.appendChild(spinner);
    return div;
}

function insertLoader() {
    postsDiv.innerHTML = ""; // Clear before loading
    postsDiv.appendChild(createLoadingElement());
}
