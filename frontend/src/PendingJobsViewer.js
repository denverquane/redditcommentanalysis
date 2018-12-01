import React from "react";
import { connect } from "react-redux";
import { Card, Intent, Button } from "@blueprintjs/core";

import {postExtractSubreddits } from './reducer'

class PendingJobsViewer extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
        organized: []
    };
  }

  componentWillReceiveProps(props) {
    let organized = [];
    for (let joI in props.extractionJobs) {
        let jo = props.extractionJobs[joI];
        if (organized[jo.year]) {
          if (organized[jo.year][jo.month]) {
            organized[jo.year][jo.month].push(jo.subreddit);
          } else {
            organized[jo.year][jo.month] = [jo.subreddit];
          }
        } else {
          organized[jo.year] = [];
          organized[jo.year][jo.month] = [jo.subreddit];
        }
      }
      this.setState({organized: organized})
  }

  render() {
    let jobs = [];
    

    for (let yrI in this.state.organized) {
      jobs.push(this.renderYear(this.state.organized[yrI], yrI));
    }

    return (
      <div>
        <h1>Pending Extraction Request</h1>
        {jobs.length !== 0 ? <Button intent={Intent.SUCCESS} onClick={() => {
          for (let yrIdx in this.state.organized) {
            for (let moIdx in this.state.organized[yrIdx]) {
              let subs = this.state.organized[yrIdx][moIdx];
    
              this.props.postExtractSubreddits(subs, moIdx, yrIdx);
            }
          }
            
        }}>Submit </Button> : <div/> }
        {jobs}
      </div>
    );
  }

  renderYear(months, yr) {
    let objs = [];
    for (let moI in months) {
      objs.push(<div>{this.renderMonth(months[moI], moI)}</div>);
    }
    return (
      <Card>
        <h2>{yr}</h2>
        <div style={{ display: "flex", flexDirection: "row" }}>
          <div style={{ display: "flex", flexDirection: "column", width: '50%'}}>
            {objs.slice(0, 6)}
          </div>
          <div style={{ display: "flex", flexDirection: "column", width: '50%' }}>
            {objs.slice(6)}
          </div>
        </div>
      </Card>
    );
  }

  renderMonth(subs, month) {
    let objs = [];
    for (let subI in subs) {
      objs.push(<div>{subs[subI]}</div>);
    }
    return (
      <div>
        <h3>{month}</h3>
        {objs}
      </div>
    );
  }
}

const mapStateToProps = state => ({
  extractionJobs: state.extractionQueue
});

const mapDispatchToProps = dispatch => ({
  //addExtractionJob: (sub, mo, year) => dispatch(addExtractionJob(sub, mo, year))
  postExtractSubreddits: (subs, month, year) => dispatch(postExtractSubreddits(subs, month, year))
  
});

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(PendingJobsViewer);
