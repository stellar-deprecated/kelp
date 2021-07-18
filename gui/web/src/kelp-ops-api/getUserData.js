import { v4 as uuidv4 } from "uuid";

export default () => {
    /* dont rename the user_id key because it is re-used in LoginRedirect.js and both keys need to match*/
    const userIDKey = 'user_id';
    
    let userId = localStorage.getItem(userIDKey);
    if (userId == null) {
        userId = uuidv4();
        localStorage.setItem(userIDKey, userId);
    }
    
    return {
        ID: userId,
    };
};