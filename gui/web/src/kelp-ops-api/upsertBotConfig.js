import getUserData from "./getUserData";

export default (baseUrl, configData) => {
    return fetch(baseUrl + "/api/v1/upsertBotConfig", {
        method: "POST",
        body: JSON.stringify({
            user_data: getUserData(),
            config_data: configData,
        }),
    }).then(resp => {
        return resp.json();
    });
};