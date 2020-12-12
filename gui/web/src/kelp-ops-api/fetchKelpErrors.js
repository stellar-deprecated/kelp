export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/fetchKelpErrors").then(resp => {
        return resp.json();
    });
};