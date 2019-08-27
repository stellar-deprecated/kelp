export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/optionsMetadata", {
        method: "GET",
    }).then(resp => {
        return resp.json();
    });
};