export default (baseUrl, type, feedUrl) => {
    let data = {
        type: type,
        feed_url: feedUrl
    };

    return fetch(baseUrl + "/api/v1/fetchPrice", {
        method: "POST",
        body: JSON.stringify(data),
    }).then(resp => {
        return resp.json();
    });
};