export default (baseUrl, eventData) => {
    return fetch(baseUrl + "/api/v1/sendMetricEvent", {
        method: "POST",
        body: eventData,
    }).then(resp => {
       return resp.json();
    });
}