export default (baseUrl, botName) => {
    return fetch(baseUrl + "/api/v1/getBotConfig", {
        method: "POST",
        body: botName,
    }).then(resp => {
        return resp.json();
    });
};