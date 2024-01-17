const maxLines = 500;

document.addEventListener("DOMContentLoaded", function() {
    trimTextArea();
});

function trimTextArea() {
    let textarea = document.getElementById('runLogArea');
    let lines = textarea.value.split('\n');

    if (lines.length > maxLines) {
        textarea.value = lines.slice(-maxLines).join('\n');
    }

    textarea.scrollTop = textarea.scrollHeight;
}