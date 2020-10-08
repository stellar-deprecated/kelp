class MetricsTracker {
    constructor(apiKey) {
        this.apiKey = apiKey;
        this.apiUrl = "https://api2.amplitude.com/2/httpapi";

        // TODO: Implement more functions.
        this.sendEvent = this.sendEvent.bind(this);
    }

    sendEvent(event) {
        var _this = this
        this._asyncRequests["fetch"] = fetch(this.apiUrl, {
            method: "POST",
            eventName: event,
        }).then(resp => {
            if (!_this._asyncRequests["fetch"]) {
                // if it has been deleted, we don't want to process the result
                return
            }
            delete _this._asyncRequests["fetch"];

            return resp.json();
        });
    }
}

export default MetricsTracker;