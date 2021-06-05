/* eslint-disable no-use-before-define */
/* eslint-disable no-plusplus */
/* eslint-disable no-console */
/* eslint-disable prefer-const */
const tempfcmtoken =
    'BLJ4_Ju_C-wNAZwOwtSzQc0_OJGwzO8qst5AbpnDFBUzToY6bIUJMaEcTAhpk0iqYnPlxOJakvBNnxvXVkixb_E'
const messaging = firebase.messaging()
messaging
    .getToken({
        vapidKey: tempfcmtoken,
    })
    .then((currentToken) => {
        if (currentToken) {
            document.cookie = `fcm=${currentToken}`
            // Send the token to your server and update the UI if necessary
            // ...
            TokenElem.innerHTML = 'Device token is : <br>' + currentToken
        } else {
            // Show permission request UI
            console.log(
                'No registration token available. Request permission to generate one.'
            )
            // ...
        }
    })
    .catch((err) => {
        console.log('An error occurred while retrieving token. ', err)
        // ...
    })

// ws
let conn
let id
let log
let MsgElem
let TokenElem
let NotisElem
let ErrElem

window.onload = () => {
    MsgElem = document.getElementById('msg')
    TokenElem = document.getElementById('token')
    NotisElem = document.getElementById('notis')
    ErrElem = document.getElementById('err')
    log = document.getElementById('log')
    id = document.getElementById('uid')
}
const connectWS = () => {
    console.log(conn)
    if (!conn || conn.readyState === 3) {
        document.cookie = `UID=${id.value}`
        document.cookie = `lastID=${getLastMessageID(log)}`
        messaging.getToken({ vapidKey: tempfcmtoken }).then((tok) => {
            document.cookie = `fcm=${currentToken}`
        })
        conn = new WebSocket(`ws://${document.location.host}/ws`)
    }
    conn.onclose = function (evt) {
        let item = document.createElement('div')
        item.innerHTML = '<b>Connection closed.</b>'
        appendLog(item)
    }
    conn.onmessage = function (evt) {
        let messages = evt.data
        let decodedMessages = JSON.parse(messages)
        for (let i = 0; i < decodedMessages.length; i++) {
            let item = document.createElement('div')
            item.id = decodedMessages[i].Id
            item.innerText = decodedMessages[i].Message
            appendLog(item)
        }
    }
}
function appendLog(item) {
    log.appendChild(item)
}
const disconnectWS = () => {
    conn.close()
}
const getLastMessageID = (messageNode) => {
    let messageArray = Array.prototype.slice.call(
        messageNode.getElementsByTagName('*')
    )
    let lastID = 0
    for (const message in messageArray) {
        if (messageArray[message].hasAttribute('id')) {
            if (messageArray[message].id != '-1')
                lastID = messageArray[message].id
        }
    }
    return lastID
}
