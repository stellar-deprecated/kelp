import getUserData from "./getUserData";

export default (baseUrl, eventType, eventData) => {
    return fetch(baseUrl + "/api/v1/sendMetricEvent", {
        method: "POST",
        body: JSON.stringify({
            user_data: getUserData(),
            event_type: eventType,
            event_data: eventData,
        }),
    }).then(resp => {
       return resp.json();
    });
}