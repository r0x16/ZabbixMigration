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

runStatus.addEventListener('logs', function (event) {
    if (event.data == "null" || event.data == "undefined") {
        return
    }

    let data = JSON.parse(event.data)
    let textarea = document.getElementById('runLogArea')

    for (let i = 0; i < data.length; i++) {
        textarea.value += data[i]
    }
    textarea.scrollTop = textarea.scrollHeight
})

runStatus.addEventListener('log', function (event) {
    if (event.data == "null" || event.data == "undefined") {
        return
    }

    let data = JSON.parse(event.data)
    let textarea = document.getElementById('runLogArea')

    textarea.value += data
    textarea.scrollTop = textarea.scrollHeight
})