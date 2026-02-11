import { API, token } from "@/app.js";

let rankingType = "0"; // 0 = GLOBAL, 1 = SEASONAL
let currentSection = "OP"; // 'OP' or 'ED'

const sectionHeader = document.getElementById("section-header");
const toggleGlobal = document.getElementById("toggle-global");
const toggleSeasonal = document.getElementById("toggle-seasonal");
const rankingTypeBg = document.getElementById("ranking-type-bg");

const showOps = document.getElementById("show-ops");
const showEds = document.getElementById("show-eds");

const containerOps = document.getElementById("container-ops");
const containerEds = document.getElementById("container-eds");

const loaderOps = document.getElementById("loader-ops");
const loaderEds = document.getElementById("loader-eds");

const loadMoreOp = document.getElementById("load-more-op");
const loadMoreEd = document.getElementById("load-more-ed");

const paginatorBtnsOps = document.querySelectorAll(".load-more-op");
const paginatorBtnsEds = document.querySelectorAll(".load-more-ed");

let params = {};
let headersData = {};
let page_openings = 1;
let page_endings = 1;
let last_page_openings = undefined;
let last_page_endings = undefined;

// Initial fetch
getOpenings();
getEndings();

// Global/Seasonal Toggle
if (toggleGlobal && toggleSeasonal) {
    toggleGlobal.addEventListener("click", () => switchRankingType("0"));
    toggleSeasonal.addEventListener("click", () => switchRankingType("1"));
}

function switchRankingType(type) {
    if (rankingType === type) return;
    rankingType = type;

    // Visual animation
    if (type === "0") {
        rankingTypeBg.style.left = "6px";
        toggleGlobal.classList.replace("text-white/40", "text-white");
        toggleSeasonal.classList.replace("text-white", "text-white/40");
    } else {
        rankingTypeBg.style.left = "calc(50% + 1.5px)";
        toggleGlobal.classList.replace("text-white", "text-white/40");
        toggleSeasonal.classList.replace("text-white/40", "text-white");
    }

    // Reset and refetch
    containerOps.innerHTML = "";
    containerEds.innerHTML = "";
    page_openings = 1;
    page_endings = 1;
    getOpenings();
    getEndings();
}

// Openings/Endings Switcher
if (showOps && showEds) {
    showOps.addEventListener("click", () => switchSection("OP"));
    showEds.addEventListener("click", () => switchSection("ED"));
}

function switchSection(section) {
    currentSection = section;
    if (section === "OP") {
        showOps.classList.add("bg-primary", "text-white");
        showOps.classList.remove("text-white/40");
        showEds.classList.remove("bg-primary", "text-white");
        showEds.classList.add("text-white/40");

        containerOps.classList.remove("hidden");
        containerEds.classList.add("hidden");
        loadMoreOp.classList.remove("hidden");
        loadMoreEd.classList.add("hidden");
    } else {
        showEds.classList.add("bg-primary", "text-white");
        showEds.classList.remove("text-white/40");
        showOps.classList.remove("bg-primary", "text-white");
        showOps.classList.add("text-white/40");

        containerEds.classList.remove("hidden");
        containerOps.classList.add("hidden");
        loadMoreEd.classList.remove("hidden");
        loadMoreOp.classList.add("hidden");
    }
}

// Pagination
paginatorBtnsOps.forEach((btn) => {
    btn.addEventListener("click", () => {
        if (
            last_page_openings === undefined ||
            page_openings < last_page_openings
        ) {
            page_openings++;
            getOpenings();
        }
    });
});

paginatorBtnsEds.forEach((btn) => {
    btn.addEventListener("click", () => {
        if (
            last_page_endings === undefined ||
            page_endings < last_page_endings
        ) {
            page_endings++;
            getEndings();
        }
    });
});

async function getOpenings() {
    loaderOps.classList.remove("hidden");
    paginatorBtnsOps.forEach((btn) => btn.setAttribute("disabled", ""));

    try {
        params = {
            ranking_type: rankingType,
            page_op: page_openings,
            page_ed: page_endings,
        };

        headersData = {
            Authorization: "Bearer " + token,
            Accept: "application/json, text/html;q=0.9",
        };

        const response = await API.get(API.SONGS.RANKING, headersData, params);

        if (!response.html) throw new Error("html: Invalid data structure");

        page_openings = response.openings.current_page;
        last_page_openings = response.openings.last_page;

        renderData(containerOps, response.html.openings);
        handlePaginationVisibility("OP", page_openings, last_page_openings);

        // Update Header
        if (rankingType === "0") {
            sectionHeader.textContent = "Global Ranking";
        } else {
            let season = response.currentSeason;
            let year = response.currentYear;
            sectionHeader.textContent = `Ranking ${season.name} ${year.name}`;
        }
    } catch (error) {
        console.error(`RankingService: ${error.message}`);
    } finally {
        loaderOps.classList.add("hidden");
        paginatorBtnsOps.forEach((btn) => btn.removeAttribute("disabled"));
    }
}

async function getEndings() {
    loaderEds.classList.remove("hidden");
    paginatorBtnsEds.forEach((btn) => btn.setAttribute("disabled", ""));

    try {
        params = {
            ranking_type: rankingType,
            page_op: page_openings,
            page_ed: page_endings,
        };

        headersData = {
            Authorization: "Bearer " + token,
            Accept: "application/json, text/html;q=0.9",
        };

        const response = await API.get(API.SONGS.RANKING, headersData, params);

        if (!response.html) throw new Error("html: Invalid data structure");

        page_endings = response.endings.current_page;
        last_page_endings = response.endings.last_page;

        renderData(containerEds, response.html.endings);
        handlePaginationVisibility("ED", page_endings, last_page_endings);
    } catch (error) {
        console.error(`RankingService: ${error.message}`);
    } finally {
        loaderEds.classList.add("hidden");
        paginatorBtnsEds.forEach((btn) => btn.removeAttribute("disabled"));
    }
}

function handlePaginationVisibility(type, current, last) {
    const btn = type === "OP" ? loadMoreOp : loadMoreEd;
    if (last !== undefined && current >= last) {
        btn.classList.add("invisible");
    } else {
        btn.classList.remove("invisible");
    }
}

function renderData(container, html) {
    container.innerHTML += html;
}
