import { useEffect, useState } from 'react';
import ResponsiveAppBar from './components/AppBar';
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import Home from "./pages/Home";
import Account from "./pages/Account";
import { CreateInstance, ConfirmInstance } from './pages/Instance';

const darkTheme = createTheme({
  palette: {
    mode: 'dark',
  },
});

const App = () => {
  var [user, setUser] = useState({});

  useEffect(() => {
    fetch("/api/user", {
        method: 'GET',
        credentials: 'same-origin',
        headers: {
          'Content-Type': 'application/json'
        },
        referrerPolicy: 'no-referrer'
      })
      .then(r => r.json())
      .then(r => {
        setUser(state => ({ ...state, 
          provenance: r.provenance,
          email: r.email,
          userID: r.userID
        }));
      })
  }, [])

  return (
    <ThemeProvider theme={darkTheme}>
      <ResponsiveAppBar user={user}></ResponsiveAppBar>
      <CssBaseline />
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home user={user}/>}/>
        <Route path="/account" element={<Account />}/>
        <Route path="/instance/create" element={<CreateInstance />}/>
        <Route path="/instance/confirm" element={<ConfirmInstance />}/>
      </Routes>
    </BrowserRouter>
    </ThemeProvider>
  );
}

export default App;
