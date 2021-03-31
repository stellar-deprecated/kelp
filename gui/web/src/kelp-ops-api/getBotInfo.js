import getUserData from "./getUserData";

export default (baseUrl, botName, signal) => {
    return fetch(baseUrl + "/api/v1/getBotInfo", {
        method: "POST",
        body: JSON.stringify({
            user_data: getUserData(),
            bot_name: botName,
        }),
        signal: signal,
    }).then(resp => {
        return resp.json();
    });
};