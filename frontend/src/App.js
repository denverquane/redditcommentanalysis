import React, { Component } from "react";
import "./App.css";
import {
  Checkbox,
  Button,
  Toaster,
  Position,
  Intent,
  Spinner,
  InputGroup,
  Card,
  Elevation,
  Tag
} from "@blueprintjs/core";
import { Circle } from "rc-progress";
import Sockette from "sockette";
import { connect } from "react-redux";
import { CollapseExample } from "./Collapse";
import MonthYearSelector from "./MonthYearSelection";
import PendingJobsViewer from "./PendingJobsViewer";

import "@blueprintjs/core/lib/css/blueprint.css";
import { IP } from './index'

import { setSelectedSubreddit, fetchSubreddits, toggleCompareSubreddit } from "./reducer";
import SelectedSubredditViewer from "./SelectedSubredditViewer";
import ComparisonPlot from "./ComparisonPlot";

const AppToaster = Toaster.create({
  className: "notifyToaster",
  position: Position.TOP
});

class App extends Component {
  ws = new Sockette("ws://" + IP + ":5000/ws", {
    timeout: 5e3,
    maxAttempts: 10,
    onopen: e => {
      console.log("Connected!", e);
      AppToaster.show({
        message: "Connected to Backend!",
        intent: Intent.SUCCESS
      });
      this.props.fetchSubreddits();
      this.getStatus();
      this.setState({ ...this.state, Websocket: true });
    },
    onmessage: e => {
      console.log("Received:", e);
      console.log(String(e.data));
      if (String(e.data).includes("fetch")) {
        this.props.fetchSubreddits();
        this.getStatus();
        console.log("Fetch message");
      } else if (String(e.data).includes("status")) {
        this.getStatus();
        console.log("Status message");
      } else if (String(e.data).includes("error")) {
        console.log("Error message");
        AppToaster.show({ message: "Error!" + e.data, intent: Intent.DANGER });
      }
    },
    onreconnect: e => console.log("Reconnecting...", e),
    onmaximum: e => console.log("Stop Attempting!", e),
    onclose: e => {
      console.log("Closed!", e);
      AppToaster.show({
        message: "Disconnected from backend!",
        intent: Intent.DANGER
      });
      this.setState({ ...this.state, Websocket: false });
    },
    onerror: e => {
      console.log("Error!", e);
      AppToaster.show({
        message: "Error connecting to backend!",
        intent: Intent.DANGER
      });
      this.setState({ ...this.state, Websocket: false });
    }
  });
  state = {
    Websocket: Boolean,
    Status: Object,
    TempExtractName: String
  };
  constructor() {
    super();
    this.state.Websocket = false;
    this.getStatus();
  }

  componentDidMount() {
    this.props.fetchSubreddits();
  }

  render() {
    return (
      <div className="App">
        <header
          className="App-header"
          style={{ display: "flex", flexDirection: "row" }}
        >
          <div
            style={{ width: "35%", display: "flex", flexDirection: "column" }}
          >
            Extraction Queue:
            {this.getLabels(this.state.Status.ExtractQueue)}
          </div>
          <div style={{ width: "30%", display: "flex", flexDirection: "row" }}>
            {this.state.Status.Extracting ? (
              <div style={{ width: "20%", height: "20%" }}>
                <Circle
                  strokeWidth="10"
                  percent={this.state.Status.ExtractProgress}
                />
                {this.state.Status.ExtractQueue[0]["Month"]}{" "}
                {this.state.Status.ExtractQueue[0]["Year"]}{" "}
                {this.state.Status.ExtractProgress ? 
                    this.state.Status.ExtractProgress.toFixed(3) + '% Extracted ' : 'Extracting...'}
                {this.state.Status.ExtractTimeRem ? '~' + this.state.Status.ExtractTimeRem + ' Remaining' : ''}
              </div>
            ) : (
              <div style={{ width: "20%", height: "20%" }}>
                <Circle
                  strokeWidth="10"
                  percent={this.state.Status.ExtractProgress}
                />
                Nothing to Extract!
              </div>
            )}
            {
              <div style={{ width: "40%" }}>
                <Spinner
                  intent={this.state.Websocket ? Intent.SUCCESS : Intent.DANGER}
                  value={
                    this.state.Websocket
                      ? this.state.Status.Extracting ||
                        this.state.Status.Processing
                        ? null
                        : 1
                      : 1
                  }
                />{" "}
              </div>
            }
            {this.state.Status.Processing ? (
              <div style={{ width: "20%", height: "20%" }}>
                <Circle
                  strokeWidth="10"
                  percent={this.state.Status.ProcessProgress}
                />
                "{this.state.Status.ProcessQueue[0]["Month"]}/{this.state.Status.ProcessQueue[0]["Year"]}"{" "}
                {this.state.Status.ProcessProgress.toFixed(3)}% Processed
              </div>
            ) : (
              <div style={{ width: "20%", height: "20%" }}>
                <Circle
                  strokeWidth="10"
                  percent={this.state.Status.ProcessProgress}
                />
                Nothing to Process!
              </div>
            )}
          </div>
          <div
            style={{ width: "35%", display: "flex", flexDirection: "column" }}
          >
            Processing Queue:
            {this.getLabels(this.state.Status.ProcessQueue)}
          </div>
        </header>
        <div style={{ width: "100%", display: "flex", flexDirection: "row" }}>
          <div style={{ width: "16%" }}>
            <InputGroup
              onChange={event =>
                this.setState({
                  ...this.state,
                  TempExtractName: event.target.value
                })
              }
            />
          </div>
          <div style={{ width: "10%" }}>
            <Button
              onClick={() => {
                this.addSubredditEntry(this.state.TempExtractName);
                console.log(this.state.TempExtractName);
              }}
            >
              Add Subreddit Entry
            </Button>
          </div>
        </div>
        <div style={{ display: "flex", flexDirection: "row" }}>
          <div style={{ width: "25%" }}>{this.displaySubs()}</div>
          <div style={{ width: "60%" }}>
          <ComparisonPlot/>
          </div>
          <div style={{width: "15%"}}>
          <PendingJobsViewer />
          </div>
        </div>
      </div>
    );
  }

  displaySubs() {
    let arr = [];
    console.log(this.props.subreddits);
    for (let key in this.props.subreddits) {
      let years = [];
      let totalComments = 0;
      let substatus = this.props.subreddits[key];

      for (let yr in substatus["ExtractedYearMonthCommentCounts"]) {
        for (let month in substatus["ExtractedYearMonthCommentCounts"][yr]) {
          if (substatus["ExtractedYearMonthCommentCounts"][yr][month] !== -1) {
            totalComments +=
              substatus["ExtractedYearMonthCommentCounts"][yr][month];
          }
        }
        years.push(
          <MonthYearSelector
            Year={yr}
            Months={substatus["ExtractedYearMonthCommentCounts"][yr]}
            Sentiments={substatus["ProcessedYearMonthCommentSummaries"][yr]}
            Subreddit={key}
          />
        );
      }
      arr.push(
        <Card
          onClick={() => {this.props.setSelectedSubreddit(key)}}
          key={key}
          style={{
            display: "flex",
            backgroundColor: "#ECEBEB",
            flexDirection: "column",
            marginBottom: "1%"
          }}
          elevation={Elevation.TWO}
          interactive={true}
        >
          <h1>{key}</h1>
          <Checkbox 
          disabled={!this.props.subreddits[key].ProcessedYearMonthCommentSummaries["2016"]}
          onChange={() => this.props.toggleCompareSubreddit(key)}>
            Compare 2016
            </Checkbox>
          ({totalComments.toLocaleString()} total comments)
          <CollapseExample component={years} typeLabel={"Extraction Details"} />
        </Card>
      );
    }
    arr.sort(function(a, b) {
      var x = a.key.toLowerCase();
      var y = b.key.toLowerCase();
      if (x < y) {
        return -1;
      }
      if (x > y) {
        return 1;
      }
      return 0;
    });
    return arr;
  }

  getStatus() {
    fetch("http://" + IP + ":5000/api/status")
      .then(results => {
        return results.json();
      })
      .then(data => {
        if (JSON.stringify(this.state.Status) !== JSON.stringify(data)) {
          this.setState({ ...this.state, Status: data });
        }
      });
  }

  getLabels(queue) {
    let arr = [];
    if (!queue || queue.length === 0) {
      return <div>Empty Queue</div>;
    }
    for (let i in queue) {
      if (i === "0") {
        arr.push(
          <div>
            <div style={{ width: "10%" }} />
            <Tag
              key={i}
              intent={Intent.SUCCESS}
              style={{ width: "80%", marginBottom: "1%" }}
            >
              {i}. {queue[i]["Month"]}/{queue[i]["Year"]}{" "}{queue[i]["Subreddits"].join(',')}
            </Tag>
            <div style={{ width: "10%" }} />
          </div>
        );
      } else if (arr.length < 5){
        arr.push(
          <div>
            <div style={{ width: "10%" }} />
            <Tag key={i} intent={Intent.NONE} style={{ width: "80%" }}>
            {i}. {queue[i]["Month"]}/{queue[i]["Year"]}{" "}{queue[i]["Subreddits"].join(',')}
            </Tag>
            <div style={{ width: "10%" }} />
          </div>
        );
      } else if (i === "5"){
        arr.push(<div>
            <div style={{ width: "10%" }} />
            <Tag key={i} intent={Intent.NONE} style={{ width: "80%" }}>
            ...
            </Tag>
            <div style={{ width: "10%" }} />
          </div>);
      }
    }
    return arr;
  }

  processSubreddit(sub) {
    fetch("http://" + IP + ":5000/api/processSub/" + sub, {
      method: "post"
    })
      .then(results => {
        return results;
      })
      .then(data => {
        console.log(data);
        AppToaster.show({
          message: "Asked backend to process: " + sub,
          intent: Intent.NONE
        });
        this.getStatus();
      });
  }

  addSubredditEntry(sub) {
    fetch("http://" + IP + ":5000/api/addSubEntry/" + sub, {
      method: "post"
    })
      .then(results => {
        return results;
      })
      .then(data => {
        this.props.fetchSubreddits();
      });
  }
}

const mapStateToProps = state => ({
  selectedSubreddit: state.selectedSubreddit,
  subreddits: state.subreddits
});

const mapDispatchToProps = dispatch => ({
  setSelectedSubreddit: (sub) => dispatch(setSelectedSubreddit(sub)),
  toggleCompareSubreddit: sub => dispatch(toggleCompareSubreddit(sub)),
  fetchSubreddits: () => dispatch(fetchSubreddits())
});

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(App);
