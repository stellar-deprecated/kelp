export default (baseUrl, botName) => {
    return fetch(baseUrl + "/api/v1/stop", {
        method: "POST",
        body: botName,
    }).then(resp => {
        return resp.text();
    });
};