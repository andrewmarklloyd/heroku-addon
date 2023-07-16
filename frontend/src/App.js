import { useEffect, useState } from 'react';
import ResponsiveAppBar from './components/AppBar';
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import Home from "./pages/Home";

const darkTheme = createTheme({
  palette: {
    mode: 'dark',
  },
});

const App = () => {
  var [user, setUser] = useState({user: {}});

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
        console.log("user api response:",r)
        setUser(state => ({ ...state, user: {provenance: r.provenance} }));
      })
  }, [])

  return (
    <ThemeProvider theme={darkTheme}>
      <ResponsiveAppBar user={user}></ResponsiveAppBar>
      <CssBaseline />
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />}/>
      </Routes>
    </BrowserRouter>
    </ThemeProvider>
  );
}

export default App;
