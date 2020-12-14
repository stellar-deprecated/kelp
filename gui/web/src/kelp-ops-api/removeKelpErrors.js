export default (baseUrl, kelpErrorIDs) => {
    const data = {
        kelp_error_ids: kelpErrorIDs
    };
    
    return fetch(baseUrl + "/api/v1/removeKelpErrors", {
        method: "POST",
        body: JSON.stringify(data)
    }).then(resp => {
        return resp.json();
    });
};