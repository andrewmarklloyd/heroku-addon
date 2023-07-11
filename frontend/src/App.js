import logo from './logo.svg';
import ResponsiveAppBar from './components/appBar';
import './App.css';

function App() {
  return (
    <div className="App">
      <ResponsiveAppBar></ResponsiveAppBar>
      <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <h1>
          My Great Product
        </h1>
      </header>
    </div>
  );
}

export default App;
