export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/getNewBotConfig").then(resp => {
        return resp.json();
    });
};