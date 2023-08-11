import { useEffect, useState } from 'react';
import { Outlet } from "react-router-dom";
import InstanceTable from '../components/InstanceTable';

const Home = (props) => {
  var [instances, setInstances] = useState([]);

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
    <h1>Instances of Nothing</h1>
    <InstanceTable user={props.user} instances={instances}></InstanceTable>
    <Outlet />
  </>
  );
}

export default Home;