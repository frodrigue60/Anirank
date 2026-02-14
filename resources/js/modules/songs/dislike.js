import { API, csrfToken, token } from "@/app.js";

let headersData = {};
let params = {};

if (token) {
    const likeBtn = document.querySelector("#like-button");
    const likesSpan = document.querySelector("#like-counter");

    const dislikeBtn = document.querySelector("#dislike-button");
    const dislikesSpan = document.querySelector("#dislike-counter");

    dislikeBtn.addEventListener("click", dislikeSong);

    async function dislikeSong() {
        likeBtn.setAttribute("disabled", "");
        dislikeBtn.setAttribute("disabled", "");
        try {
            headersData = {
                Accept: "application/json, text/html;q=0.9",
                "X-CSRF-TOKEN": csrfToken,
                Authorization: "Bearer " + token,
            };

            const response = await API.get(
                API.SONGS.DISLIKE(dislikeBtn.dataset.songId),
                headersData,
                params,
            );

            if (response.success) {
                likesSpan.textContent = response.likesCount;
                dislikesSpan.textContent = response.dislikesCount;
            }
        } catch (error) {
            throw new Error(error);
        } finally {
            likeBtn.removeAttribute("disabled");
            dislikeBtn.removeAttribute("disabled");
        }
    }
}

/* onDomReady(function() {
    // Tu código aquí
    console.log('DOM está listo');
    // puedes inicializar eventos, etc.
}); */
