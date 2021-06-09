import React, { useEffect } from "react";
import { useAuth0 } from "@auth0/auth0-react";

const LoginRedirect = () => {
  const { user,isLoading,isAuthenticated, loginWithRedirect, getAccessTokenSilently } = useAuth0();

  if (isLoading) {
    return <></>;
  }

  if (!isAuthenticated) {
    return loginWithRedirect();
  }

  if (isLoading) {
    return <></>;
  }

  useEffect(() => {
      const userIDKey = 'user_id';
      const Auth0ID = JSON.stringify(user.sub).slice(7).replace('"', '')
      localStorage.setItem(userIDKey, Auth0ID);

      try {
        getAccessTokenSilently().then(token => localStorage.setItem("accessToken", token)).then(window.location.reload());
      } catch (e) {
        console.log(e.message);
      }
  }, []);

  return (
    <></>
  );
};

export default LoginRedirect;