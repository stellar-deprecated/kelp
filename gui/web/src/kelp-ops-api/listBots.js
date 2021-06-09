import getUserData from "./getUserData";

export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/listBots", {
        method: "POST",
        body: JSON.stringify({
            user_data: getUserData(),
        }),
    }).then(resp => {
        return resp.json();
    });
};