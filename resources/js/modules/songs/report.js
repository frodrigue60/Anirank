import { API, csrfToken, token } from "@/app.js";

const formReport = document.querySelector("#form-report");
let reportTitle = formReport.querySelector("input[id=input-title]");
let reportTextarea = formReport.querySelector("textarea[id=textarea-content]");
let reportBtnSubmit = formReport.querySelector("button[type=submit]");
let songId = formReport.dataset.songid;
let headersData = {};
let bodyData = {};

formReport.addEventListener("submit", function (event) {
    event.preventDefault();

    if (reportTitle.value != null && reportTextarea.value != null) {
        sendReport();
    }
});

async function sendReport() {
    reportTitle.toggleAttribute("disabled");
    reportTextarea.toggleAttribute("disabled");
    reportBtnSubmit.toggleAttribute("disabled");
    try {
        headersData = {
            "Content-Type": "application/json",
            "X-CSRF-TOKEN": csrfToken,
            Authorization: "Bearer " + token,
        };

        bodyData = JSON.stringify({
            song_id: songId,
            title: reportTitle.value,
            content: reportTextarea.value,
        });

        const response = await API.post(
            API.SONGS.REPORTS,
            headersData,
            bodyData,
        );

        if (response.success == true) {
            reportTitle.value = "";
            reportTextarea.value = "";

            hideModal("modal-report");

            setTimeout(function () {
                swal(
                    "Thanks for reporting!",
                    "We are working on it!",
                    "success",
                    {
                        buttons: false,
                        timer: 2000,
                    },
                );
            }, 500);
        } else {
            if (response.message.content) {
                response.message.content.forEach((message) => {
                    console.log(message);
                });
            }
        }
    } catch (error) {
        throw new Error(error);
    } finally {
        setTimeout(function () {
            reportTitle.toggleAttribute("disabled");
            reportTextarea.toggleAttribute("disabled");
            reportBtnSubmit.toggleAttribute("disabled");
        }, 1000);
    }
}
