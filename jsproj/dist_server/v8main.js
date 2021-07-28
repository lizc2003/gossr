BASE_URL = "http://localhost:9090";
if(APP_ENV === 'prod') {
    BASE_URL = 'https://localhost:9090';
}
API_BASE_URL = BASE_URL;

v8worker.send(101, JSON.stringify({base: BASE_URL, api: API_BASE_URL}));

renderToString = require("vue-server-renderer/basic.js");
serverBundle = require("server.js").default;
