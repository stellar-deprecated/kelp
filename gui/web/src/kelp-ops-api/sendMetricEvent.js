export default (baseUrl, name, eventData) => {
    return fetch(baseUrl + "/api/v1/sendMetricEvent", {
        method: "POST",
        body: {
            name: name,
            data: eventData,
        },
    }).then(resp => {
       return resp.json();
    });
}