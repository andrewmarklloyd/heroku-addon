import { useEffect, useState } from 'react';
import ResponsiveAppBar from './components/AppBar';
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import Home from "./pages/Home";
import Account from "./pages/Account";
import About from "./pages/About";
import { GetPricing } from './helpers/Pricing'
import { CreateInstance, ConfirmInstance, EditInstance } from './pages/Instance';
import { OrderComplete } from './pages/OrderComplete';

const darkTheme = createTheme({
  palette: {
    mode: 'dark',
  },
});

const App = () => {
  var [user, setUser] = useState({});
  var [pricingState, setPricingState] = useState([]);

  useEffect(() => {
    const pricing = GetPricing()
    pricing.then(p=>{
      setPricingState(p)
    })

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
          name: r.name,
          userID: r.userID
        }));
      })
  }, [])

  return (
    <ThemeProvider theme={darkTheme}>
      <ResponsiveAppBar user={user} pricing={pricingState}></ResponsiveAppBar>
      <CssBaseline />
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home user={user}/>}/>
        <Route path="/about" element={<About/>}/>
        <Route path="/account" element={<Account user={user} />}/>
        <Route path="/instance/create" element={<CreateInstance pricing={pricingState} />}/>
        <Route path="/instance/confirm" element={<ConfirmInstance pricing={pricingState} />}/>
        <Route path="/instance/edit" element={<EditInstance pricing={pricingState} />}/>
        <Route path="/order/complete" element={<OrderComplete/>}/>
      </Routes>
    </BrowserRouter>
    </ThemeProvider>
  );
}

export default App;
