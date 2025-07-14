import { API, csrfToken, token, hideModal } from "@/app.js";
const scoreFormat = document.querySelector('meta[name="score-format"]').content;
const ratingForm = document.querySelector('#rating-form');
const ratingBtn = document.querySelector('#rating-button');
const scoreSpan = document.querySelector('#score-span');
let scoreInput = document.querySelector('#scoreInput');
let submitScoreBtn = document.querySelector('#submit-score-btn');
let headersData = {};
let bodyData = {};
let checkboxes = undefined;

if (scoreFormat == 'POINT_5') {
    let checkboxes = document.querySelectorAll('#rating-form input[name="score"]');
    let userScore = 0;

    function actualizarPuntuacionesSeleccionadas() {
        let checkedValue = Array.from(checkboxes)
            .filter(cb => cb.checked)
            .map(cb => cb.value);

        userScore = checkedValue.join();

        if ((checkedValue.join() > 0) && (checkedValue.join() <= 100)) {
            rate(userScore)
        }
    }

    checkboxes.forEach(checkbox => {
        checkbox.addEventListener('change', actualizarPuntuacionesSeleccionadas);
    });
}

ratingForm.addEventListener("submit", function (event) {
    event.preventDefault()
    let userScore = document.querySelector('#scoreInput').value;

    if ((userScore != '') && (userScore > 0) && (userScore <= 100)) {
        rate(userScore)
    }
});

async function rate(userScore) {
    if (scoreFormat == 'POINT_5') {
        checkboxes = document.querySelectorAll('#rating-form input[name="score"]');
        checkboxes.forEach(checkbox => {
            checkbox.setAttribute('disabled', '');
        });
    } else {
        scoreInput.setAttribute('disabled', '');
        submitScoreBtn.setAttribute('disabled', '');
    }

    try {
        headersData = {
            'Content-Type': 'application/json',
            'X-CSRF-TOKEN': csrfToken,
            'Authorization': 'Bearer ' + token,
        };

        bodyData = JSON.stringify({
            score: userScore,
        });

        const response = await API.post(API.SONGS.RATE(ratingForm.dataset.song), headersData, bodyData);

        if (response.success == true) {
            ratingBtn.classList.remove('btn-primary');
            ratingBtn.classList.add('btn-warning');
            scoreSpan.textContent = response.average;
            hideModal('modal-rating');


            swal('Success', 'Thanks for rate this song!', 'success', {
                buttons: false,
                timer: 1500,
            });
        }

    } catch (error) {
        console.log(error)
    } finally {
        if (scoreFormat == 'POINT_5') {
            setTimeout(function () {
                checkboxes.forEach(checkbox => {
                    checkbox.removeAttribute('disabled');
                });
            }, 500);

        } else {
            setTimeout(function () {
                scoreInput.removeAttribute('disabled');
                submitScoreBtn.removeAttribute('disabled');
            }, 500);
        }


    }
}
