import { useState, useEffect, forwardRef } from 'react';
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { Button, TextField, FormControl, InputLabel, Select, MenuItem } from '@mui/material';
import MuiAlert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import { GetPricing } from '../helpers/Pricing'

const Alert = forwardRef(function Alert(props, ref) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

const CreateInstance = () => {
  const navigate = useNavigate();

  var [newInstanceName, setNewInstanceName] = useState('');
  var [newInstancePlan, setNewInstancePlan] = useState('');

  const handleCreateInstance = () => {
    navigate("/instance/confirm", {state:{name:newInstanceName,plan:newInstancePlan}})
  }

  const handleCancel = () => {
    navigate("/")
  }

  const handleUpdateInstanceName = (event) => {
    setNewInstanceName(event.target.value)
  }

  const handleUpdateInstancePlan = (event) => {
    setNewInstancePlan(event.target.value)
  }

  return (
    <>
    <h1>Create New Instance</h1>
    <Outlet />
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
      <MenuItem value={"free"}>Free ($0/month)</MenuItem>
      <MenuItem value={"staging"}>Staging ($10/month)</MenuItem>
      <MenuItem value={"production"}>Production ($35/month)</MenuItem>
    </Select>
  </FormControl>
  <Button onClick={handleCreateInstance} size="small" variant="outlined">Review</Button>
  <Button onClick={handleCancel} color="secondary" size="small" variant="outlined">Cancel</Button>
  </>
  );
}

const ConfirmInstance = (props) => {
  const [open, setOpen] = useState(false);
  const [createDisabled, setCreateDisabled] = useState(false);

  const navigate = useNavigate();
  const location = useLocation();
  const pricing = GetPricing()
  
  const handleCreateInstance = () => {
    setCreateDisabled(true)
    fetch("/api/new-instance", {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json'
      },
      referrerPolicy: 'no-referrer',
      body: JSON.stringify({"name": location.state.name, "plan": location.state.plan})
    })
    .then(r => r.json())
    .then(r => {
      if (r.status === 'success') {
        setOpen(true)
        setTimeout(() => {
          navigate("/")  
        }, 3000);
      } else {
        alert("failed to create instance: " + r)
      }
    })
  }

  const handleCancel = () => {
    navigate("/")
  }

  return (
    <>
    <h1>Confirm New Instance</h1>
    <h3>Name: {location.state.name}</h3>
    <h3>Plan: {location.state.plan}</h3>
    <h3>Total: ${pricing[location.state.plan]}/month</h3>
    <Button disabled={createDisabled} onClick={handleCreateInstance} size="small" variant="outlined">Create Instance</Button>
    <Button onClick={handleCancel} color="secondary" size="small" variant="outlined">Cancel</Button>
    <Snackbar open={open} autoHideDuration={3000}>
      <Alert severity="success" sx={{ width: '100%' }}>
        Successfully created instance, redirecting home in 3 seconds.
      </Alert>
    </Snackbar>
    <Outlet />
  </>
  );
}

const EditInstance = (props) => {
  const navigate = useNavigate();
  const pricing = GetPricing()
  const location = useLocation();
  var [instanceName, setInstanceName] = useState('');
  const [deleteDisabled, setDeleteDisabled] = useState(true);

  const handleInstanceName = (event) => {
    setInstanceName(event.target.value)
  }

  const handleDeleteInstance = () => {
    if (instanceName !== location.state.name) {
      alert("Entered text did not equal " + location.state.name)
      return
    }

    setDeleteDisabled(true)
    fetch("/api/delete-instance", {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json'
      },
      referrerPolicy: 'no-referrer',
      body: JSON.stringify({"id": location.state.id})
    })
    .then(r => r.json())
    .then(r => {
      if (r.status === 'success') {
        navigate("/")
      } else {
        alert("failed to delete instance: " + r)
        setDeleteDisabled(false)
      }
    })
  }

  const handleBack = () => {
    navigate("/")
  }

  useEffect(() => {
    if (instanceName === location.state.name) {
      setDeleteDisabled(false)
    } else {
      setDeleteDisabled(true)
    }
  }, [instanceName, location.state.name])

  return (
    <>
    <h1>Edit Instance</h1>
    <h3>Name: {location.state.name}</h3>
    <h3>Plan: {location.state.plan}</h3>
    <h3>Total: ${pricing[location.state.plan]}/month</h3>
    <br></br>
    <br></br>
    <h1>Delete Instance</h1>
    <div>Are you sure you want to delete {location.state.name}? Type <i>{location.state.name}</i> to confirm deletion.</div>
    <TextField onChange={handleInstanceName} id="outlined-basic" label="Name" variant="outlined" />
    <Button disabled={deleteDisabled} onClick={handleDeleteInstance} size="small" variant="outlined">Delete</Button>
    <Button onClick={handleBack} color="secondary" size="small" variant="outlined">Back</Button>
    <Outlet />
  </>
  );
}

export {
  CreateInstance,
  ConfirmInstance,
  EditInstance
};
