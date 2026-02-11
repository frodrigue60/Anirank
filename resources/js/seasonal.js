import API from "@api/index.js";
const token = localStorage.getItem("api_token");
const contentContainer = document.getElementById("data");
const sectionHeader = document.getElementById("section-header");
const toggleBtn = document.getElementById("toggle-type-btn");
let currentType = "OP";
let params = {};
let headersData = {};

const toggleOp = document.getElementById("toggle-op");
const toggleEd = document.getElementById("toggle-ed");
const toggleBg = document.getElementById("toggle-bg");

fetchData(currentType);

function updateVisualState(type) {
    if (type === "OP") {
        toggleBg.classList.replace("left-[calc(50%-4px)]", "left-1");
        toggleOp.classList.replace("text-white/40", "text-white");
        toggleEd.classList.replace("text-white", "text-white/40");
        toggleOp
            .querySelector(".material-symbols-outlined")
            .classList.add("filled");
        toggleEd
            .querySelector(".material-symbols-outlined")
            .classList.remove("filled");
    } else {
        toggleBg.classList.replace("left-1", "left-[calc(50%-4px)]");
        toggleOp.classList.replace("text-white", "text-white/40");
        toggleEd.classList.replace("text-white/40", "text-white");
        toggleOp
            .querySelector(".material-symbols-outlined")
            .classList.remove("filled");
        toggleEd
            .querySelector(".material-symbols-outlined")
            .classList.add("filled");
    }
}

async function fetchData(type) {
    try {
        if (toggleBtn) toggleBtn.disabled = true;
        updateVisualState(type);

        params = {
            type: currentType,
        };

        headersData = {
            Authorization: "Bearer " + token,
            Accept: "application/json, text/html;q=0.9",
        };

        const response = await API.get(API.SONGS.SEASONAL, headersData, params);

        try {
            validateResponse(response);
            //console.log(response.songs);
            renderData(response.html);
            updateHeader(type);
        } catch (error) {
            console.error(error);
        }
    } catch (error) {
        error.message = `UserService: ${error.message}`;
        throw error;
    } finally {
        toggleBtn.disabled = false;
    }
}

function updateHeader(type) {
    sectionHeader.textContent = type === "OP" ? "OPENINGS" : "ENDINGS";
    document.querySelector("#btn-toggle-text").textContent =
        type === "OP" ? "Endings" : "Openings";
}

function renderData(html) {
    contentContainer.innerHTML = html;
}

if (toggleOp) {
    toggleOp.addEventListener("click", () => {
        if (currentType !== "OP") {
            currentType = "OP";
            fetchData(currentType);
        }
    });
}

if (toggleEd) {
    toggleEd.addEventListener("click", () => {
        if (currentType !== "ED") {
            currentType = "ED";
            fetchData(currentType);
        }
    });
}

function validateResponse({ songs, html }) {
    if (!songs) {
        showError("Songs data is null or undefined!", "songs");
    } else if (songs.length === 0) {
        showError("No songs avaible!", "songs");
    }
    /* if (!html) {
        showError('Invalid HTML data structure!', 'html');
    } */
}

function showError(message, context) {
    swal("Error!", message, "error", {
        timer: 2000,
        buttons: false,
    });
    throw new Error(`${context}: ${message}`);
}
