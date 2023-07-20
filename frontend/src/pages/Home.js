import { useEffect, useState } from 'react';
import { Outlet } from "react-router-dom";
import { Button } from '@mui/material';
import InstanceTable from '../components/InstanceTable';

const Home = () => {
  var [instances, setInstances] = useState([]);

  const handleCreateInstance = (event) => {
    console.log("todo")
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
    <h1>Instances</h1>
    <Button onClick={handleCreateInstance} size="small" variant="outlined">+ Create Instance</Button>
    <InstanceTable instances={instances}></InstanceTable>
    <Outlet />
  </>
  );
}

export default Home;