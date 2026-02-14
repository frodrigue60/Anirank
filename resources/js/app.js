import API from "@api/index.js";
import swal from "sweetalert";

function hideModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
        modal.classList.add("hidden");
    }
}

function showModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
        modal.classList.remove("hidden");
    }
}

window.hideModal = hideModal;
window.showModal = showModal;

let headers = document.querySelectorAll(".section-header");

headers.forEach((header) => {
    let words = header.textContent.split(" ");
    header.innerHTML = `<span class="first-word">${words[0]}</span> ${words.slice(1).join(" ")}`;
});

const token = localStorage.getItem("api_token");
const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

/* function onDomReady(callback) {
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', callback);
    } else {
        callback();
    }
} */

export { API, token, csrfToken, swal, hideModal, showModal };
