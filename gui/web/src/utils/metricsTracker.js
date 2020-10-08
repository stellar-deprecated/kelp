class MetricsTracker {
    constructor(apiKey) {
        this.apiKey = apiKey;
        this.apiUrl = "https://api2.amplitude.com/2/httpapi";

        this._asyncRequests ={};
        this.sendEvent = this.sendEvent.bind(this);
        // TODO: Implement more functions.
    }

    sendEvent(event) {
        var _this = this
        // TODO: Change namespace.
        // TODO: Replace `fetch` with a custom routing function.
        this._asyncRequests["fetch"] = fetch(this.apiUrl, {
            method: "POST",
            eventName: event,
        }).then(resp => {
            if (!_this._asyncRequests["fetch"]) {
                // if it has been deleted, we don't want to process the result
                return
            }
            delete _this._asyncRequests["fetch"];

            // TODO: log error or ignore successful response.
            return resp.json();
        });
    }
}

export default MetricsTracker;