import getUserData from "./getUserData";

export default (baseUrl, kelpErrorIDs) => {
    return fetch(baseUrl + "/api/v1/removeKelpErrors", {
        method: "POST",
        body: JSON.stringify({
            user_data: getUserData(),
            kelp_error_ids: kelpErrorIDs,
        }),
    }).then(resp => {
        return resp.json();
    });
};