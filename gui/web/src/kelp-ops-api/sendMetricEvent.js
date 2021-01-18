export default (baseUrl, eventType, eventData) => {
    return fetch(baseUrl + "/api/v1/sendMetricEvent", {
        method: "POST",
        body: JSON.stringify({
            event_type: eventType,
            event_data: eventData,
        }),
    }).then(resp => {
       return resp.json();
    });
}