export default (baseUrl, kelpErrorIDs) => {
    return fetch(baseUrl + "/api/v1/removeKelpErrors", {
        method: "POST",
        body: { kelp_error_ids: kelpErrorIDs }
    }).then(resp => {
        return resp.json();
    });
};