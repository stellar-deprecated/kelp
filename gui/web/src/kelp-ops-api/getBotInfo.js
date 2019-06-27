export default (baseUrl, botName, signal) => {
    return fetch(baseUrl + "/api/v1/getBotInfo", {
        method: "POST",
        body: botName,
        signal: signal,
    }).then(resp => {
        return resp.json();
    });
};