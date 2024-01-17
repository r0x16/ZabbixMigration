let runStatus = new EventSource(runEventsUrl)

runStatus.addEventListener('open', function (event) {
    console.log('Connection opened')
})

runStatus.addEventListener('ready', function (event) {
    let data = JSON.parse(event.data)
    runStatus.close()
    window.location.reload()
})

runStatus.addEventListener('error', function (event) {
    let data = JSON.parse(event.data)
    runStatus.close()
    alert(data.message)
})

let pendingData = [];

function updateTextArea() {
    let textarea = document.getElementById('runLogArea');
    let newData = pendingData.join('');
    let fullData = textarea.value + newData;
    let lines = fullData.split('\n');

    if (lines.length > maxLines) {
        textarea.value = lines.slice(-maxLines).join('\n');
    } else {
        textarea.value = fullData;
    }

    textarea.scrollTop = textarea.scrollHeight;
    pendingData = [];
}

runStatus.addEventListener('logs', function (event) {
    if (!event.data) {
        return;
    }

    let data;
    try {
        data = JSON.parse(event.data);
    } catch (e) {
        console.error('Error parsing JSON:', e);
        return;
    }

    if (data.length === 0) {
        return;
    }

    for (let i = 0; i < data.length; i++) {
        pendingData.push(data[i]);
    }

    updateTextArea();
});

runStatus.addEventListener('log', function (event) {
    if (event.data == "null" || event.data == "undefined") {
        return
    }

    let data = JSON.parse(event.data)
    let textarea = document.getElementById('runLogArea')

    textarea.value += data
    textarea.scrollTop = textarea.scrollHeight
})