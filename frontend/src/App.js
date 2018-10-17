import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';
import { Card, Button, Collapse } from '@blueprintjs/core'
import { LineChart, XAxis, YAxis, CartesianGrid, Line, Tooltip} from 'recharts'

let Months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

class App extends Component {
    state = {
        Subs: Object
    };

  render() {
    return (
      <div className="App">
        <header className="App-header"> 
          <img src={logo} className="App-logo" alt="logo" />
          
        </header>
        <button onClick={() => this.getSubs()}>
            Fetch Subs
            </button>
            {this.displaySubs()}
              {/* {this.state.Subs.EarthPorn ? (this.state.Subs.EarthPorn.Processing ? <p>'True'</p> : <p>'False'</p>) : ''} */}
      </div>
    );
  }
  displaySubs() {
    let arr = [];
    for (let key in this.state.Subs) {
        let months = [];
        for (let mo in Months) {
            months.push({key: Months[mo], comments: this.state.Subs[key].ExtractedMonthCommentCounts[Months[mo]]})
        }
      arr.push(

          <div key={key} style={{display: 'flex', flexDirection: 'column'}}>
              {key}
              <CollapseExample text={months}/>
          </div>);
        arr.sort(function(a,b){
            var x = a.key.toLowerCase();
            var y = b.key.toLowerCase();
            if (x < y) {return -1;}
            if (x > y) {return 1;}
            return 0;})
    }
      return arr
  }

  getSubs() {
      fetch('http://dquane.tplinkdns.com:5000/api/subs')
          .then(results => {
              return results.json();
          }).then(data => {
         
          if (JSON.stringify(this.state.Subs) !== JSON.stringify(data)) {
            // console.log('Updated');
            this.setState({ ...this.state, Subs: data });
          }
        
      });
  }
}

export class CollapseExample extends React.Component {

    constructor(text) {
        super();
        this.state={
            isOpen: false,
            text: text,
        }
    }
    state = {
        isOpen: false,
        text: null
    };

    render() {
        return (
            <div>
                <Button onClick={this.handleClick}>
                    {this.state.isOpen ? "Hide" : "Show"} Details
                </Button>
                <Collapse isOpen={this.state.isOpen}>
                    <LineChart width={800} height={300} data={this.state.text.text} margin={{left: 30}}>
                        <XAxis dataKey="key"/>
                        <YAxis dataKey="comments"/>
                        <CartesianGrid stroke="#eee" strokeDasharray="5 5"/>
                        <Line type="monotone" dataKey="comments" stroke="black" />\
                        <Tooltip/>
                    </LineChart>
                </Collapse>
            </div>
        );
    }

    handleClick = () => {
        this.setState({ isOpen: !this.state.isOpen });
    }
}

export default App;
