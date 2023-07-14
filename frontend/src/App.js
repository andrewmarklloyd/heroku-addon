import ResponsiveAppBar from './components/appBar';
import InstanceTable from './components/instanceTable';


import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';

const darkTheme = createTheme({
  palette: {
    mode: 'dark',
  },
});

function App() {
  fetch("/api/user", {
    method: 'GET',
    credentials: 'same-origin',
    headers: {
      'Content-Type': 'application/json'
    },
    referrerPolicy: 'no-referrer'
  })
  .then(r => r.text())
  .then(r => {
    console.log(r)
  })

  return (
    <ThemeProvider theme={darkTheme}>
      <ResponsiveAppBar></ResponsiveAppBar>
      <CssBaseline />
      <main>
        <h1>Instances</h1>
        <InstanceTable></InstanceTable>
      </main>
    </ThemeProvider>
  );
}

export default App;
