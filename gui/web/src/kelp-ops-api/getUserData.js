import { v4 as uuidv4 } from "uuid";

export default () => {
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