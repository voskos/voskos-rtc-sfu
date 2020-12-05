const { RTCSessionDescription, RTCPeerConnection} = require('wrtc');
const WebSocket = require('ws');

const { createCanvas, createImageData } = require('canvas');
const { hsv } = require('color-space');
const { performance } = require('perf_hooks');

var ws = null
var pc = null
var wsStatus = "connecting"

room_id = process.argv[2].split(" ")[1]
participant_id = process.argv[2].split(" ")[2]

const width = 640;
const height = 480;

const { RTCVideoSink, RTCVideoSource, i420ToRgba, rgbaToI420 } = require('wrtc').nonstandard;
 
// const ws = new WebSocket('ws://localhost:8080/ws');
// pc = null;

function createPeerConnection() {
    var config = {
        sdpSemantics: 'unified-plan'
    };

    //config.iceServers = [{urls: ['stun:stun.l.google.com:19302']}];

    pc = new RTCPeerConnection(config);

    pc.addEventListener('icegatheringstatechange', function() {
    }, false);

    pc.addEventListener('iceconnectionstatechange', function() {    }, false);

    pc.addEventListener('signalingstatechange', function() {
    }, false);

    // pc.addEventListener('icecandidate', async function({candidate}) {
    //     console.log("candidate ", candidate)
    //     await connection.send(JSON.stringify({'action' : 'NEW_ICE_CANDIDATE_CLIENT','user_id' : $("#uname").val(), 'ice_candidate' : candidate, 'room_id' : $("#room_id").val()}))
    // }, false);

    // connect audio / video
    // pc.ontrack = ({transceiver}) => {
    //     console.log("TRACK RECIEVED")    
    // };
    pc.addEventListener('track', function(event) {
        console.log("***************************TRACK RECIEVED for "+ participant_id+"***************************")
    }, false);

    return pc;
}

function negotiate() {
    return pc.createOffer().then(function(offer) {
        return pc.setLocalDescription(offer);
    }).then(function() {
        // wait for ICE gathering to complete
        return new Promise(function(resolve) {
            if (pc.iceGatheringState === 'complete') {
                resolve();
            } else {
                function checkState() {
                    if (pc.iceGatheringState === 'complete') {
                        pc.removeEventListener('icegatheringstatechange', checkState);
                        resolve();
                    }
                }
                pc.addEventListener('icegatheringstatechange', checkState);
            }
        });
    }).then(async function() {
        
        ws.send(JSON.stringify({'action' : 'INIT','user_id' : participant_id, 'sdp' : pc.localDescription, 'room_id' : room_id}))
        //console.log("SDP ANSWER FROM SERVER---> ", answer)
    })
}

function sleep(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
} 

function getWsConn(){
	let t = new WebSocket('ws://localhost:8080/ws');
	return t
}
async function startCall(){
	ws = await getWsConn()
	pc = await createPeerConnection();

	const source = new RTCVideoSource();
    const track = source.createTrack();
    const transceiver = pc.addTransceiver(track);


 	const sink = new RTCVideoSink(transceiver.receiver.track);

	  let lastFrame = null;

	  function onFrame({ frame }) {
	    lastFrame = frame;
	  }

	  sink.addEventListener('frame', onFrame);

	  // TODO(mroberts): Is pixelFormat really necessary?
	  const canvas = createCanvas(width, height);
	  const context = canvas.getContext('2d');
	  context.fillStyle = 'white';
	  context.fillRect(0, 0, width, height);

	  let hue = 0;

	  const interval = setInterval(() => {
	    if (lastFrame) {
	      const lastFrameCanvas = createCanvas(lastFrame.width,  lastFrame.height);
	      const lastFrameContext = lastFrameCanvas.getContext('2d');

	      const rgba = new Uint8ClampedArray(lastFrame.width *  lastFrame.height * 4);
	      const rgbaFrame = createImageData(rgba, lastFrame.width, lastFrame.height);
	      i420ToRgba(lastFrame, rgbaFrame);

	      lastFrameContext.putImageData(rgbaFrame, 0, 0);
	      context.drawImage(lastFrameCanvas, 0, 0);
	    } else {
	      context.fillStyle = 'rgba(255, 255, 255, 0.025)';
	      context.fillRect(0, 0, width, height);
	    }

	    hue = ++hue % 360;
	    const [r, g, b] = hsv.rgb([hue, 100, 100]);
	    const thisTime = performance.now();

	    context.font = '60px Sans-serif';
	    context.strokeStyle = 'black';
	    context.lineWidth = 1;
	    context.fillStyle = `rgba(${Math.round(r)}, ${Math.round(g)}, ${Math.round(b)}, 1)`;
	    context.textAlign = 'center';
	    context.save();
	    context.translate(width / 2, height / 2);
	    context.rotate(thisTime / 1000);
	    context.strokeText('node-webrtc', 0, 0);
	    context.fillText('node-webrtc', 0, 0);
	    context.restore();

	    const rgbaFrame = context.getImageData(0, 0, width, height);
	    const i420Frame = {
	      width,
	      height,
	      data: new Uint8ClampedArray(1.5 * width * height)
	    };
	    rgbaToI420(rgbaFrame, i420Frame);
	    source.onFrame(i420Frame);
	  });


	ws.on('open', function open() {
		wsStatus = "connected"
		return negotiate();
	  	//ws.send('something');
	});
	
	//sleep(3000)
	ws.on('message', async function incoming(data) {
	  //console.log(data);
	  let resp = JSON.parse(data)
	    if(resp.action === "SERVER_ANSWER"){
	      await pc.setRemoteDescription(resp.sdp)
	    }
	    else if(resp.action === "SERVER_OFFER"){
	    	try {
		      await pc.setRemoteDescription(resp.sdp)
		      let ans = await pc.createAnswer()
		      await pc.setLocalDescription(ans)
		      //sleep(3000)
		      ws.send(JSON.stringify({'action' : 'CLIENT_ANSWER','user_id' : participant_id, 'sdp' : pc.localDescription, 'room_id' : room_id}))
			} catch (error) {
			  // expected output: ReferenceError: nonExistentFunction is not defined
			  // Note - error messages will vary depending on browser
			}
	      
	    }
	});

	// while(true){
	// 	console.log(wsStatus)
	// 	if(wsStatus === "connected"){
			
	// 		break;
	// 	}
	// }

	
}


startCall()
