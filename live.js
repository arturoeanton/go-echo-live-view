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

    if (json_data.type == 'style') {
        out.style.cssText = json_data.value
    }

    if (json_data.type == 'set') {
        out.value = json_data.value;
    }


    if (json_data.type == 'script') {
        eval(json_data.value);
    }

    if (json_data.type == 'get') {
        str = JSON.stringify({ "type": "get", "id_ret": json_data.id_ret, "data": null })
        if (json_data.sub_type == 'style') {
            str = JSON.stringify({ "type": "get", "id_ret": json_data.id_ret, "data": document.getElementById(json_data.id).style[json_data.value] })
        }
        if  (json_data.sub_type == 'value') {
            str = JSON.stringify({ "type": "get", "id_ret": json_data.id_ret, "data": document.getElementById(json_data.id).value })
        }
        if  (json_data.sub_type == 'html') {
            str = JSON.stringify({ "type": "get", "id_ret": json_data.id_ret, "data": document.getElementById(json_data.id).innerHTML })
        }
        if  (json_data.sub_type == 'text') {
            str = JSON.stringify({ "type": "get", "id_ret": json_data.id_ret, "data": document.getElementById(json_data.id).innerHTML })
        }
        ws.send(str)
    }
}

function send_event(id, event, data) {
    var str = JSON.stringify({ "type": "data", "id": id, "event": event, "data": data })
    ws.send(str)
}