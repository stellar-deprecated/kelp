export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/newSecretKey").then(resp => {
        return resp.text();
    });
};