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
