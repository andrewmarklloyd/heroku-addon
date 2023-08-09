import { Outlet } from "react-router-dom";

const Account = (props) => {
  return (
    <>
    <h1>Account</h1>
    <h3>Email: {props.user.email}</h3>
    <h3>Login Method: {props.user.provenance}</h3>
    <Outlet />
  </>
  );
}

export default Account;
