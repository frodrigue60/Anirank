import { API, csrfToken, token } from "@/app.js";
const formRequest = document.querySelector("#form-request");
const textareaRequest = formRequest.querySelector("#textarea-request");
let headersData = {};
let bodyData = {};

formRequest.addEventListener("submit", function (event) {
    event.preventDefault();
    sendRequest();
});

async function sendRequest() {
    try {
        headersData = {
            "Content-Type": "application/json",
            Accept: "application/json, text/html;q=0.9",
            "X-CSRF-TOKEN": csrfToken,
            Authorization: "Bearer " + token,
        };
        bodyData = JSON.stringify({
            content: textareaRequest.value,
        });

        const response = await API.post(
            API.REQUESTS.STORE,
            headersData,
            bodyData,
        );

        if (response.success == true) {
            swal("Thanks!", response.message, "success", {
                buttons: false,
                timer: 2000,
            });
            hideModal("requestModal");

            formRequest.reset();
        }
    } catch (error) {
        throw new Error(error);
    }
}
