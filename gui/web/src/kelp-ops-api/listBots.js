export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/listBots").then(resp => {
        return resp.json();
    });
};