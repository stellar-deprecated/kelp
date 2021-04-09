import getUserData from "./getUserData";

export default (baseUrl) => {
    return fetch(baseUrl + "/api/v1/genBotName", {
        method: "POST",
        body: JSON.stringify({
            user_data: getUserData(),
        }),
    }).then(resp => {
        return resp.jtext();
    });
};