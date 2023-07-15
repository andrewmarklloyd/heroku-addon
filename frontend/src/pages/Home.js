import { Outlet } from "react-router-dom";
import InstanceTable from '../components/InstanceTable.js';

const Home = () => {
  return (
    <>
    <h1>Instances</h1>
    <InstanceTable></InstanceTable>
    <Outlet />
  </>
  );
}

export default Home;