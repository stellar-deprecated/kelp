const constants = {
    BotState: {
        initializing: "initializing",
        stopped: "stopped",
        running: "running",
        stopping: "stopping",
    },

    ErrorType: {
        bot: "object_type_bot",
    },

    ErrorLevel: {
        info: "info",
        error: "error",
        warning: "warning",
    },

    BaseURL: "",

    setGlobalBaseURL: (baseUrl) => {
        constants.BaseURL = baseUrl;
    },
};

export default constants;