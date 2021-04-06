import getUserData from "./getUserData";

export default (baseUrl, botName) => {
    return fetch(baseUrl + "/api/v1/getBotConfig", {
        method: "POST",
        body: JSON.stringify({
            user_data: getUserData(),
            bot_name: botName,
        }),
    }).then(resp => {
        return resp.json();
    });
};