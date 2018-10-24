import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';

class App extends Component {
    state = {
        Subs: Object
    };

  render() {
    return (
      <div className="App">
        <header className="App-header">
          <img src={logo} className="App-logo" alt="logo" />
          <button onClick={() => this.getBlocks()}>
            Fetch Subs
            </button>
            {this.displaySubs()}
              {/* {this.state.Subs.EarthPorn ? (this.state.Subs.EarthPorn.Processing ? <p>'True'</p> : <p>'False'</p>) : ''} */}
        </header>
      </div>
    );
  }
  displaySubs() {
    let arr = []
    for (var key in this.state.Subs) {
      console.log(key)
      arr.push(<div key={key}>
        {key}:
        <p>
          Extracted: {this.state.Subs[key].Extracting ? 'True' : 'False'}
          </p>
          <p>
          Processing: {this.state.Subs[key].Processing ? 'True' : 'False'}
          </p>
        </div>)
    }
    return arr
  }

  getBlocks() {
      fetch('http://dquane.tplinkdns.com:5000/api/subs')
          .then(results => {
              return results.json();
          }).then(data => {
         
          if (JSON.stringify(this.state.Subs) !== JSON.stringify(data)) {
            console.log('Updated');
            this.setState({ ...this.state, Subs: data });
          }
        
      });
  }
}

export default App;
