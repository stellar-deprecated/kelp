export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/version", {method: "GET"}).then(resp => {
        return resp.text();
    });
};