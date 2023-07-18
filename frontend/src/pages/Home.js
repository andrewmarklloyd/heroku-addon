import { Outlet } from "react-router-dom";
import { Button } from '@mui/material';
import InstanceTable from '../components/InstanceTable';

const Home = () => {
  return (
    <>
    <h1>Instances</h1>
    <Button size="small" variant="outlined">+ Create Instance</Button>
    <InstanceTable></InstanceTable>
    <Outlet />
  </>
  );
}

export default Home;