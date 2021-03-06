function getHostAndRoom(){
    const currURL = new URL(window.location.href);
    const host = currURL.host
    const room = currURL.pathname.split('/').slice(-1).pop()
    return {host, room};
}

function createSocket({host, room}, onMessageCallback, onCloseCallback){
    let socket = null;
    if(!window["WebSocket"]){
        alert("Error: Your browser does not support web sockets.");
    } else {
        socket = new WebSocket(`ws://${host}/joinroom/${room}`);
        socket.onclose = onCloseCallback;
        socket.onmessage = onMessageCallback;
    }
    return socket;
}

const app = new Vue({
    el: "#app",
    delimiters: ["${", "}"],
    data:{
        announcements: [],
        room: null,
        socket: null,
        newMessage: null
    },
    methods:{
        onMessage: function(m){
            this.announcements.unshift(m.data);
        },
        sendMessage: function(){
            if(!this.newMessage || this.newMessage.length == 0) return;
            this.socket.send(this.newMessage);
            this.newMessage = null;
        }
    },
    created: function(){
        const hostRoom = getHostAndRoom();
        this.room = hostRoom.room;
        this.socket = createSocket(hostRoom, this.onMessage.bind(this));
    }
});
