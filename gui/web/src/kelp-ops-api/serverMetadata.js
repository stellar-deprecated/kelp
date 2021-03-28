export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/serverMetadata").then(resp => {
        return resp.json();
    });
};