package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sfu/src/signal"

	"github.com/pion/webrtc/v4"
)

type SFU struct {
	peers map[string]*webrtc.PeerConnection
}

func NewSFU() *SFU {
	return &SFU{
		peers: make(map[string]*webrtc.PeerConnection),
	}
}

func (sfu *SFU) AddPeer(peerID string, peerConnection *webrtc.PeerConnection) {
	log.Println("Peer added: ", peerID)
	sfu.peers[peerID] = peerConnection
}

func (sfu *SFU) RemovePeer(peerID string) {
	delete(sfu.peers, peerID)
}

func main() {
	sfu := NewSFU()

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/offer", func(w http.ResponseWriter, r *http.Request) {
		peerId := r.URL.Query().Get("peerId")

		var offerRequest struct {
			Offer string `json:"offer"`
		}
		err := json.NewDecoder(r.Body).Decode(&offerRequest)
		if err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}

		config := webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		}

		peerConnection, err := webrtc.NewPeerConnection(config)
		if err != nil {
			panic(err)
		}

		sfu.AddPeer(peerId, peerConnection)

		peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			fmt.Printf("Track has started streamId(%s) id(%s) rid(%s) \n", track.StreamID(), track.ID(), track.RID())

			for id, pc := range sfu.peers {
				log.Printf("Track forwarded to peer: %s", id)

				newTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, track.ID(), track.StreamID())
				if err != nil {
					log.Printf("Error creating new track: %v", err)
					return
				}
				log.Printf("Track created: %s", newTrack.ID())
				if _, err = pc.AddTrack(newTrack); err != nil {
					log.Printf("Error forwarding track to peer: %v", err)
				}

				for {
					rtpPackets, _, rtcpErr := track.ReadRTP()
					if rtcpErr != nil {
						panic(rtcpErr)
					}
					// log.Printf("Received RTP Packet from %s: %v", peerId, rtpPackets)

					if err = newTrack.WriteRTP(rtpPackets); err != nil {
						log.Printf("Error writing to track: %v", err)
						return
					}

				}
			}
			// for {
			// 	rtcpPackets, _, rtcpErr := receiver.ReadRTCP()
			// 	if rtcpErr != nil {
			// 		panic(rtcpErr)
			// 	}

			// 	for _, r := range rtcpPackets {
			// 		if stringer, canString := r.(fmt.Stringer); canString {
			// 			fmt.Printf("Received RTCP Packet from %s: %v", peerId, stringer.String())
			// 		}
			// 	}
			// }
		})

		peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			fmt.Printf("Connection State has changed %s \n", connectionState.String())
		})

		offer := webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  offerRequest.Offer,
		}
		// signal.Decode(signal.MustReadStdin(), &offer)

		err = peerConnection.SetRemoteDescription(offer)
		if err != nil {
			panic(err)
		}

		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}

		gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

		// Sets the LocalDescription, and starts our UDP listeners
		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			panic(err)
		}

		<-gatherComplete

		// Write the SDP offer to the response
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"offer":"%s"}`, signal.Encode(*peerConnection.LocalDescription()))
		// fmt.Println(signal.Encode(*peerConnection.LocalDescription()))
	})
	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
