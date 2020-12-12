export default (baseUrl, kelpError) => {
    return fetch(baseUrl + "/api/v1/removeKelpError", {
        method: "POST",
        body: { kelp_error: kelpError }
    }).then(resp => {
        return resp.json();
    });
};