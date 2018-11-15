import React from "react";
import { connect } from "react-redux";
import { Card, Intent } from "@blueprintjs/core";

class PendingJobsViewer extends React.Component {
  constructor(props) {
    super(props);
    this.state = {};
  }

  render() {
    let jobs = [];
    let organized = [];

    for (let joI in this.props.extractionJobs) {
      let jo = this.props.extractionJobs[joI];
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

    for (let yrI in organized) {
      jobs.push(this.renderYear(organized[yrI], yrI));
    }

    return (
      <div>
        <h1>Pending Extraction Request</h1>
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
  extractionJobs: state
});

const mapDispatchToProps = dispatch => ({
  //addExtractionJob: (sub, mo, year) => dispatch(addExtractionJob(sub, mo, year))
});

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(PendingJobsViewer);
