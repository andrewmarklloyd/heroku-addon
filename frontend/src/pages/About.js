import { Outlet } from "react-router-dom";

const About = (props) => {
  return (
    <>
    <h1>About</h1>
    <h3>This service allows you to create, manage, and interact with nothing. Each instance of nothing is fully managed, secured, and monitored by nothing.</h3>
    <Outlet />
  </>
  );
}

export default About;
