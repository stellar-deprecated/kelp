export default (baseUrl, name, eventData) => {
    return fetch(baseUrl + "/api/v1/sendMetricEvent", {
        method: "POST",
        body: {
            event_name: name,
            event_props: eventData,
        },
    }).then(resp => {
       return resp.json();
    });
}