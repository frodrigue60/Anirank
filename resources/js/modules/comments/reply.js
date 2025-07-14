import { API, csrfToken, token } from '@/app.js';
let headersData = {};
let bodyData = {};
let formReply = undefined;
let parentId = undefined;

document.body.addEventListener('click', (event) => {
    const button = event.target.closest('button.btn-reply-comment');

    if (button) {
        toggleReplyForm(button.dataset.commentId);
    }
});

/* document.body.addEventListener('submit', (event) => {
    event.preventDefault();
    formReply = event.target.closest('form.form-reply');

    let replyContent = formReply.querySelector('textarea[name=content]').value;
    parentId = formReply.dataset.parentCommentId;

    if (replyContent.trim() != '') {
        replyComment(parentId, replyContent);
    } else {
        swal('Invalid reply', 'Reply content can not be null', 'error', {
            buttons: false,
            timer: 1500,
        });
        formReply.querySelector('textarea').value = '';
    }

}); */
document.body.addEventListener('submit', (event) => {
    event.preventDefault();

    const formReply = event.target.closest('form.form-reply');

    if (!formReply) return; // Evita errores si no es un formulario de respuesta

    const replyContent = formReply.querySelector('textarea[name=content]').value;
    const parentId = formReply.dataset.parentCommentId;

    if (replyContent.trim() !== '') {
        replyComment(parentId, replyContent);
    } else {
        swal('Invalid reply', 'Reply content can not be null', 'error', {
            buttons: false,
            timer: 1500,
        });
        formReply.querySelector('textarea').value = '';
    }
});


async function replyComment(parentId, replyContent) {

    try {
        headersData = {
            'Content-Type': 'application/json',
            'Accept': 'application/json, text/html;q=0.9',
            'X-CSRF-TOKEN': csrfToken,
            'Authorization': 'Bearer ' + token,
        }
        bodyData = JSON.stringify({
            content: replyContent,
            comment_parent_id: parentId,
        });

        const response = await API.post(API.COMMENTS.REPLY(parentId), headersData, bodyData);

        if (response.success == true) {

            let repliesContainer = document.getElementById('replies-' + parentId);
            repliesContainer.insertAdjacentHTML('afterbegin', response.html);
        }

    } catch (error) {
        error.message = `UserService: ${error.message}`;
        throw error;
    } finally {
        toggleReplyForm(parentId);
    }
}


function toggleReplyForm(commentId) {
    let form = document.getElementById(`reply-form-${commentId}`);
    form.style.display = form.style.display === 'none' ? 'block' : 'none';
}
