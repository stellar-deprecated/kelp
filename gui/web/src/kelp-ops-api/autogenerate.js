export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/autogenerate").then(resp => {
        return resp.json();
    });
};