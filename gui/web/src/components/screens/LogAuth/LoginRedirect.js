import React, { useEffect } from "react";
import { useAuth0 } from "@auth0/auth0-react";

const LoginRedirect = () => {
  const { user,isLoading,isAuthenticated, loginWithRedirect, getAccessTokenSilently } = useAuth0();

  if (isLoading) {
    return "";
  }

  if (!isAuthenticated) {
    return loginWithRedirect();
  }

  if (isLoading) {
    return "";
  }

  useEffect(() => {

    /* dont rename the user_id key because it is re-used in getUserData.js and both keys need to match*/

      const userIDKey = 'user_id';
      const Auth0ID = JSON.stringify(user.sub).replace('"','').replace('|', '-').replace('"','')
      localStorage.setItem(userIDKey, Auth0ID);

      try {
        if(localStorage.getItem('accessToken') == null){
          getAccessTokenSilently().then(token => localStorage.setItem("accessToken", token)).then(window.location.reload());
        }
      } catch (e) {
        console.log(e.message);
      }
  }, []);

  return (
    <></>
  );
};

export default LoginRedirect;