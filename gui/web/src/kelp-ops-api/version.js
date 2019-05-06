export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/version").then(resp => {
        return resp.text();
    });
};