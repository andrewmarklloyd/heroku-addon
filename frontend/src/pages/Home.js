import { useEffect, useState } from 'react';
import { Outlet } from "react-router-dom";
import { Button, TextField, FormControl, InputLabel, Select, MenuItem } from '@mui/material';
import InstanceTable from '../components/InstanceTable';

const Home = (props) => {
  var email = props.user.email
  var userID = props.user.userID
  var [instances, setInstances] = useState([]);
  var [newInstanceName, setNewInstanceName] = useState('');
  var [newInstancePlan, setNewInstancePlan] = useState('');

  const handleCreateInstance = () => {
    if (!newInstanceName) {
      alert("Instance name cannot be empty")
      return
    }
    if (!newInstancePlan) {
      alert("Instance plan cannot be empty")
      return
    }

    fetch("/api/new-instance", {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json'
      },
      referrerPolicy: 'no-referrer',
      body: JSON.stringify({"name": newInstanceName, "plan": newInstancePlan})
    })
    .then(r => r.json())
    .then(r => {
      console.log(r)
    })
  }

  const handleUpdateInstanceName = (event) => {
    setNewInstanceName(event.target.value)
  }

  const handleUpdateInstancePlan = (event) => {
    setNewInstancePlan(event.target.value)
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
    <InstanceTable instances={instances}></InstanceTable>
    <Outlet />
    <br></br>
    <br></br>
    <TextField onChange={handleUpdateInstanceName} id="outlined-basic" label="Name" variant="outlined" />
    <FormControl fullWidth>
      <InputLabel id="plan">Plan</InputLabel>
      <Select
        labelId="plan"
        id="demo-simple-select"
        value={newInstancePlan}
        label="Plan"
        onChange={handleUpdateInstancePlan}
      >
      <MenuItem value={"free"}>Free</MenuItem>
      <MenuItem value={"staging"}>Staging</MenuItem>
      <MenuItem value={"production"}>Production</MenuItem>
    </Select>
  </FormControl>
  <Button onClick={handleCreateInstance} size="small" variant="outlined">+ Create Instance</Button>
    <br></br>
  </>
  );
}

export default Home;