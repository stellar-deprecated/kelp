import uuid from "uuid";

export default () => {
    const userIDKey = 'user_id';
    
    let userId = localStorage.getItem(userIDKey);
    if (userId == null) {
        userId = uuid.v4();
        localStorage.setItem(userIDKey, userId);
    }
    
    return {
        ID: userId,
    };
};