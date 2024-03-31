

const pc = new RTCPeerConnection({
    iceServers: [{
        urls: 'stun:stun.l.google.com:19302'
    }]
});


const sendOffer = async () => {
    console.log("here")
    var peerId = document.getElementById('peerId').value;
    var offer = pc.localDescription.sdp;
    var requestData = {
        offer: offer
    };
    
    await fetch(`/offer?peerId=${peerId}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestData)
    })
    .then(response => response.json())
    .then(data => {
        console.log('Response:', JSON.parse(atob(data.offer)));
        pc.setRemoteDescription(JSON.parse(atob(data.offer)))
        pc.ontrack = function (event) {
            console.log(event)
            const el = document.createElement(event.track.kind)
            el.srcObject = event.streams[0]
            el.autoplay = true
            el.controls = true
        
            document.getElementById('remoteVideos').appendChild(el)
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
}

const log = msg => {
    document.getElementById('div').innerHTML += msg + '<br>'
}

pc.ontrack = function (event) {
    console.log(event)
    const el = document.createElement(event.track.kind)
    el.srcObject = event.streams[0]
    el.autoplay = true
    el.controls = true

    document.getElementById('remoteVideos').appendChild(el)
}

pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
pc.onicecandidate = event => {
    if (event.candidate === null) {
        document.getElementById('localSessionDescription').value = btoa(JSON.stringify(pc.localDescription))
    }
}

navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(stream => {
        console.log("here")
        document.getElementById('video1').srcObject = stream
        stream.getTracks().forEach(track => {
            console.log(stream)
            console.log(track)
            return pc.addTrack(track, stream)
        })

        pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)
    }).catch(log)

// window.startSession = () => {
//     const sd = document.getElementById('remoteSessionDescription').value
//     if (sd === '') {
//         return alert('Session Description must not be empty')
//     }

//     try {
//         pc.setRemoteDescription(JSON.parse(atob(sd)))
//     } catch (e) {
//         alert(e)
//     }
// }

window.copySessionDescription = () => {
    const browserSessionDescription = document.getElementById('localSessionDescription')

    browserSessionDescription.focus()
    browserSessionDescription.select()

    try {
        const successful = document.execCommand('copy')
        const msg = successful ? 'successful' : 'unsuccessful'
        log('Copying SessionDescription was ' + msg)
    } catch (err) {
        log('Oops, unable to copy SessionDescription ' + err)
    }
}
