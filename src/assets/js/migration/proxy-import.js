let importStatus = new EventSource(importEventsUrl)

importStatus.addEventListener('open', function (event) {
    console.log('Connection opened')
})

importStatus.addEventListener('ready', function (event) {
    let data = JSON.parse(event.data)
    importStatus.close()
    window.location.reload()
})

importStatus.addEventListener('error', function (event) {
    let data = JSON.parse(event.data)
    importStatus.close()
    alert(data.message)
})