window.addEventListener('beforeunload', function (event) {
    navigator.sendBeacon("/exit",null)
})