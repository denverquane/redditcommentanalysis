import React from "react";
import { Checkbox, Button } from "@blueprintjs/core";
import { connect } from "react-redux";

import { addExtractionJob } from "./reducer";

import {
  AreaChart,
  LineChart,
  XAxis,
  YAxis,
  Line,
  CartesianGrid,
  Tooltip,
  Legend,
  ReferenceLine,
  Area
} from "recharts";

let Months = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec"
];
let VerboseMonths = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December"
]
let state = [];

class MonthYearSelector extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      Months: props.Months,
      PendingJobs: state,
      Year: props.Year,
      Sentiments: props.Sentiments,
      ExtractFunc: props.ExtractFunc
    };
  }

  componentWillUnmount() {
    state = this.state.PendingJobs;
  }

  componentWillReceiveProps(newProps) {
    if (
      newProps.Months !== this.state.Months ||
      newProps.Sentiments !== this.state.Sentiments
    ) {
      this.forceUpdate();
    }
  }

  render() {
    let arr = [];
    let data = [];
    for (let mo in Months) {
      if (this.state.Months[Months[mo]] !== undefined) {
        if (this.state.Months[Months[mo]] === -1) {
          arr.push(
            <Checkbox key={Months[mo]} disabled={true}>
              {Months[mo]}: No Data
            </Checkbox>
          );
        } else {
          arr.push(
            <Checkbox key={Months[mo]} checked={true}>
              {Months[mo]} : {this.state.Months[Months[mo]]} Comments
            </Checkbox>
          );
          if (
            this.state.Sentiments &&
            this.state.Sentiments[Months[mo]] !== undefined
          ) {
            data.push({
              Month: Months[mo],
              Sentiment: this.state.Sentiments[Months[mo]]["AverageSentiment"],
              Comments: this.state.Months[Months[mo]]
            });
          } else {
            data.push({ Month: Months[mo], Sentiment: 0, Comments: this.state.Months[Months[mo]] });
          }
        }
      } else {
        arr.push(
          <Checkbox
            key={Months[mo]}
            onChange={val => {
              if (this.state.PendingJobs.indexOf(Months[mo]) === -1) {
                let joined = this.state.PendingJobs.concat(Months[mo]);
                this.setState({ PendingJobs: joined });
              } else {
                let arr = [...this.state.PendingJobs];
                let index = this.state.PendingJobs.indexOf(Months[mo]);
                if (index !== -1) {
                  arr.splice(index, 1);
                  this.setState({ PendingJobs: arr });
                }
              }
            }}
          >
            {Months[mo]}
          </Checkbox>
        );
      }
    }
    return (
      <div>
        <h2>{this.state.Year}</h2>
        <div style={{ display: "flex", flexDirection: "row" }}>
          <div
            style={{ display: "flex", flexDirection: "column", width: "33%" }}
          >
            {arr.slice(0, 4)}
          </div>
          <div
            style={{ display: "flex", flexDirection: "column", width: "33%" }}
          >
            {arr.slice(4, 8)}
          </div>
          <div
            style={{ display: "flex", flexDirection: "column", width: "33%" }}
          >
            {arr.slice(8, 12)}
          </div>
        </div>
        {this.state.Sentiments ? (
          <AreaChart width={500} height={300} data={data}>
            <XAxis dataKey="Month"/>
            <YAxis dataKey="Sentiment" />
            <CartesianGrid stroke="#eee" strokeDasharray="5 5" />
            <Line type="monotone" dataKey="Sentiment" stroke="#8884d8" />
            <ReferenceLine y={0.0} label="Neutral" stroke="blue"/>
            <Area type="monotone" dataKey="Sentiment" stroke="#8884d8" fill="#8884d8"/>
            <Legend verticalAlign="top"/>
            <Tooltip />
          </AreaChart>
        ) : (
          <div />
        )}
        <LineChart width={500} height={300} data={data}>
          <XAxis dataKey="Month"/>
          <YAxis dataKey="Comments" />
          <CartesianGrid stroke="#eee" strokeDasharray="5 5" />
          <Line type="monotone" dataKey="Comments" stroke="#8884d8" />
          <Tooltip />
          <Legend verticalAlign="top"/>
        </LineChart>
        <Button
          onClick={() => {
            for (let mo in this.state.PendingJobs) {
              this.props.addExtractionJob(
                this.props.Subreddit,
                this.state.PendingJobs[mo],
                this.state.Year
              );
            }
            this.setState({ PendingJobs: [] });
          }}
          disabled={this.state.PendingJobs.length === 0}
        >
          Confirm
        </Button>
      </div>
    );
  }
}
const mapStateToProps = state => ({
  extractionJobs: state
});

const mapDispatchToProps = dispatch => ({
  addExtractionJob: (sub, mo, year) => dispatch(addExtractionJob(sub, mo, year))
});

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(MonthYearSelector);
