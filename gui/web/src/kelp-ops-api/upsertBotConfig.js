export default (baseUrl, configData) => {
    return fetch(baseUrl + "/api/v1/upsertBotConfig", {
        method: "POST",
        body: configData,
    }).then(resp => {
        return resp.json();
    });
};