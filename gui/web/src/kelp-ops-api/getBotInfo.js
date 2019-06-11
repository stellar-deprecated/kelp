export default (baseUrl, botName) => {
    return fetch(baseUrl + "/api/v1/getBotInfo", {
        method: "POST",
        body: botName,
    }).then(resp => {
        return resp.json();
    });
};