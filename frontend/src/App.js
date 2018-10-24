import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';
import { Button, Collapse } from '@blueprintjs/core'
import { Circle } from 'rc-progress'
import { LineChart, BarChart, Legend, XAxis, YAxis, Bar, CartesianGrid, Line, Tooltip} from 'recharts'

let Months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

class App extends Component {
    state = {
        Subs: Object,
        Status: Object
    };
    constructor(){
        super();
        this.getStatus();
    }

  render() {
    return (
      <div className="App">
        <header className="App-header" style={{display: 'flex', flexDirection: 'row'}}>
            <div style={{width: '33%', display: 'flex', flexDirection: 'column'}}>
                Extraction Queue:
                {this.getLabels(this.state.Status.ExtractQueue)}
            </div>
            <div style={{width: '33%', display: 'flex', flexDirection: 'row'}}>
                {this.state.Status.Extracting ?
                    (
                        <div style={{width: '20%', height: '20%'}}>
                            <Circle strokeWidth="10" percent={this.state.Status.ExtractProgress}/>
                            "{this.state.Status.ExtractQueue[0]}" {this.state.Status.ExtractProgress.toFixed(3)}% Extracted
                        </div>
                    ) : <div style={{width: '20%', height: '20%'}}>
                        <Circle strokeWidth="10" percent={this.state.Status.ExtractProgress}/>
                        Nothing to Extract!
                    </div>
                }
                <img style={{width: '60%'}} src={logo} className="App-logo" alt="logo" />
                {this.state.Status.Processing ?
                    (
                        <div style={{width: '20%', height: '20%'}}>
                            <Circle strokeWidth="10" percent={this.state.Status.ProcessProgress}/>
                            "{this.state.Status.ProcessQueue[0]}" {this.state.Status.ProcessProgress.toFixed(3)}% Processed
                        </div>
                    ) : <div style={{width: '20%', height: '20%'}}>
                        <Circle strokeWidth="10" percent={this.state.Status.ProcessProgress}/>
                        Nothing to Process!
                    </div>
                }
            </div>
            <div style={{width: '33%', display: 'flex', flexDirection: 'column'}}>
                Processing Queue:
                {this.getLabels(this.state.Status.ProcessQueue)}
            </div>
        </header>

          <button onClick={() => this.getStatus()}>
              Refresh Status
          </button>

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
        let totalComments = 0;
        let substatus = this.state.Subs[key];
        for (let mo in Months) {
            months.push({key: Months[mo], comments: substatus.ExtractedMonthCommentCounts[Months[mo]]});
            totalComments += substatus.ExtractedMonthCommentCounts[Months[mo]];
        }
        let words = [];
        for (let wo in substatus.ProcessedSummary.KeywordCommentTallies) {
            words.push({key: wo, 'Percent of Comments Containing': substatus.ProcessedSummary.KeywordCommentTallies[wo],
                'Karma per Comment Containing': substatus.ProcessedSummary.KeywordCommentKarmas[wo]})
        }
      arr.push(

          <div key={key} style={{display: 'flex', flexDirection: 'column', marginBottom: '5%'}}>
              <b>{key}</b>
              ({totalComments.toLocaleString()} total comments)
              <CollapseExample text={months}/>
              {this.state.Subs[key].Processed ? <div><CollapseExample text={words}/></div> : this.processButton(key)}
          </div>);
    }
      arr.sort(function(a,b){
          var x = a.key.toLowerCase();
          var y = b.key.toLowerCase();
          if (x < y) {return -1;}
          if (x > y) {return 1;}
          return 0;});
      return arr
  }

  getSubs() {
      fetch('http://localhost:5000/api/subs')
          .then(results => {
              return results.json();
          }).then(data => {
         
          if (JSON.stringify(this.state.Subs) !== JSON.stringify(data)) {
            // console.log('Updated');
            this.setState({ ...this.state, Subs: data });
          }
        
      });
  }

  getStatus(){
      fetch('http://localhost:5000/api/status')
          .then(results => {
              return results.json();
          }).then(data => {

          if (JSON.stringify(this.state.Status) !== JSON.stringify(data)) {
              console.log('Updated' + data.ExtractProgress);
              this.setState({ ...this.state, Status: data });
          }

      });
  }

  getLabels(queue){
      let arr = [];
      if (!queue || queue.length === 0) {
          return <div>Empty Queue</div>
      }
      for (let i in queue) {
          if (i === "0") {
              arr.push(<div key={i}>{i}. {queue[i]}</div>)
          } else {
              arr.push(<div key={i}>{i}. {queue[i]}</div>)
          }

      }
      return arr
  }

  processButton(subreddit){
        return <div><Button onClick={() => this.processSubreddit(subreddit)}>
            Process {subreddit}
        </Button></div>
  }

  processSubreddit(sub){
      fetch('http://localhost:5000/api/processSub/' + sub, {
          method: 'post'
      })
          .then(results => {
              return results;
          }).then(data => {
              console.log(data)
      });
  }
}

export class CollapseExample extends React.Component {
    constructor(text) {
        super(text);
        this.state={
            isOpen: false,
            text: text,
        }
    }
    state = {
        isOpen: false,
        text: null,
    };

    render() {
        let width = document.documentElement.clientWidth
        return (
            <div>
                <Button onClick={this.handleClick}>
                    {this.state.isOpen ? "Hide" : "Show"} Details
                </Button>
                <Collapse isOpen={this.state.isOpen}>

                    <BarChart width={width * 0.9} height={500} data={this.state.text.text} margin={{left: width*0.05}}>
                        <XAxis dataKey="key"/>
                        <YAxis />
                        <CartesianGrid stroke="#eee" strokeDasharray="5 5"/>
                        <Bar dataKey="Percent of Comments Containing" fill="blue" />\
                        <Bar dataKey="Karma per Comment Containing" fill="red" />\
                        <Tooltip/>
                        <Legend/>
                    </BarChart>
                </Collapse>
            </div>
        );
    }

    handleClick = () => {
        this.setState({ isOpen: !this.state.isOpen });
    }
}

export default App;
