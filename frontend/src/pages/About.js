import { Outlet } from "react-router-dom";

const About = (props) => {
  return (
    <>
    <h1>About</h1>
    <p>This service allows you to create, manage, and interact with nothing. Each instance of nothing is fully managed, secured, and monitored by nothing.</p>
    <Outlet />
  </>
  );
}

export default About;
