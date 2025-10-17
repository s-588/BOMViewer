window.addEventListener('beforeunload', function (event) {
    navigator.sendBeacon("/debug/exit",null)
})