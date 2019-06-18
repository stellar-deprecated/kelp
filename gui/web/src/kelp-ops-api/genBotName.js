export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/genBotName").then(resp => {
        return resp.text();
    });
};