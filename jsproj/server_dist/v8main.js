BASE_URL = "http://dev.olist.ng";
if(APP_ENV === 'prod') {
    BASE_URL = 'https://olist.ng';
} else if(APP_ENV === 'test') {
    BASE_URL = 'http://t1.dev.olist.ng';
}
AJAX_BASE_URL = BASE_URL;

v8worker.send(101, BASE_URL);
v8worker.send(102, AJAX_BASE_URL);

renderToString = require("vue-server-renderer/basic.js");
serverBundle = require("server.js").default;
