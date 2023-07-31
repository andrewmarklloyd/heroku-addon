import { useEffect, useState } from 'react';
import { Outlet, useNavigate } from "react-router-dom";
import { Button } from '@mui/material';
import InstanceTable from '../components/InstanceTable';

const Home = (props) => {
  const navigate = useNavigate();

  var [instances, setInstances] = useState([]);

  const handleCreateInstance = () => {
    navigate("/instance/create")
  }

  useEffect(() => {
    fetch("/api/instances", {
        method: 'GET',
        credentials: 'same-origin',
        headers: {
          'Content-Type': 'application/json'
        },
        referrerPolicy: 'no-referrer'
      })
      .then(r => r.json())
      .then(r => {
        setInstances(r);
      })
  }, [])

  return (
    <>
    <br></br>
    <br></br>
    <Button onClick={handleCreateInstance} size="small" variant="outlined">+ Create Instance</Button>
    <br></br>
    <h1>Instances</h1>
    <InstanceTable instances={instances}></InstanceTable>
    <Outlet />
  </>
  );
}

export default Home;