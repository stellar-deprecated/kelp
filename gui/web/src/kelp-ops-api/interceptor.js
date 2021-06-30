import fetchIntercept from 'fetch-intercept';

/* this file is not referenced anywhere but still being used because its registering interceptor on javascript fetch function globally */

let AccessToken = localStorage.getItem('accessToken');

export const interceptor = fetchIntercept.register({
    request: function (url, config) {
        // Modify the url or config here
        const withDefaults = Object.assign({}, config);
        if (!config || !config.headers) {
            withDefaults.headers = new Headers({})
        };
        withDefaults.headers.append('AUTHORIZATION', `Bearer ${AccessToken}`)
    return [url, withDefaults]
    },
 
    requestError: function (error) {
        // Called when an error occured during another 'request' interceptor call
        return Promise.reject(error);
    },
 
    response: function (response) {
        // Modify the reponse object
        return response;
    },
 
    responseError: function (error) {
        // Handle an fetch error
        return Promise.reject(error);
    }
});
