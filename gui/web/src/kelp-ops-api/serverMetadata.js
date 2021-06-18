export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/serverMetadata", {method: "GET"}).then(resp => {
        return resp.json();
    });
};