var loc = window.location;
var uri = 'ws:';

if (loc.protocol === 'https:') {
    uri = 'wss:';
}
uri += '//' + loc.host;
uri += loc.pathname + 'ws_goliveview';

ws = new WebSocket(uri)

ws.onopen = function () {
    console.log('Connected')
}

ws.onmessage = function (evt) {
    json_data = JSON.parse(evt.data)
    var out = document.getElementById(json_data.id);

    if (json_data.type == 'fill') {
        out.innerHTML = json_data.value;
    }

    if (json_data.type == 'set') {
        out.value = json_data.value;
    }


    if (json_data.type == 'script') {
        eval(json_data.value);
    }

    if (json_data.type == 'get') {
        var str = JSON.stringify({ "type": "get", "id_ret": json_data.id_ret, "data": document.getElementById(json_data.id).value })
        ws.send(str)
    }
}

function send_event(id, event, data) {
    var str = JSON.stringify({ "type": "data", "id": id, "event": event, "data": data })
    ws.send(str)
}