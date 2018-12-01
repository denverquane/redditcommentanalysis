import React from "react";
import { connect } from "react-redux";
import { Boxplot } from "react-boxplot";
import {Months} from './MonthYearSelection';
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

class SelectedSubredditViewer extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      YearMonthCommentCounts: [],
      YearMonthProcessSummaries: []
    };
  }

  componentWillReceiveProps(props) {
    if (props.subreddits && props.selectedSubreddit) {
      this.setState({
        YearMonthCommentCounts:
          props.subreddits[props.selectedSubreddit]
            .ExtractedYearMonthCommentCounts,
        YearMonthProcessSummaries:
          props.subreddits[props.selectedSubreddit]
            .ProcessedYearMonthCommentSummaries
      });
    }
  }

  render() {
    return (
      <div>
        <h1>Selected Subreddit:</h1>
        <h1>{this.props.selectedSubreddit}</h1>

        {this.state.YearMonthProcessSummaries
          ? this.getBoxplotsForYearAndMonthsByKey(
              this.state.YearMonthProcessSummaries,
              "WordLength"
            )
          : null}
        {this.state.YearMonthProcessSummaries
          ? this.getRechartsPlotAcrossMonths(
              this.state.YearMonthProcessSummaries
            )
          : null}
      </div>
    );
  }

  getBoxplotsForYearAndMonthsByKey(yearmonthdata, key) {
    let plots = [];

    for (let yr in yearmonthdata) {
      plots.push(<h5>{yr}</h5>);
      for (let mo in Months) {
        let data = yearmonthdata[yr][Months[mo]];
        if (data) {
          plots.push(this.getBoxplot(data[key], Months[mo]));
        }
      }
    }
    return plots;
  }

  getBoxplot(boxplotdata, month) {
    return (
      <div>
        {month}
        {": "}
        {boxplotdata.Min}
        <Boxplot
          width={400}
          height={10}
          orientation="horizontal"
          min={boxplotdata.Min}
          max={boxplotdata.Max}
          stats={{
            whiskerLow: boxplotdata.Min,
            quartile1: boxplotdata.FirstQuartile,
            quartile2: boxplotdata.Median,
            quartile3: boxplotdata.ThirdQuartile,
            whiskerHigh: boxplotdata.Max,
            outliers: []
          }}
        />
        {boxplotdata.Max}
      </div>
    );
  }
  getRechartsPlotAcrossMonths(rawData) {
    let data = [];

    for (let yr in rawData) {
        for (let mo in Months) {
            let month = rawData[yr][Months[mo]]
            if (month){
                data.push({
                    Month: Months[mo],
                    Comments: month.TotalComments,
                    Sentiment: month.Sentiment.Average,
                    Karma: month.Karma.Average,
                    WordLength: month.WordLength.Average
                })
            }
        }
    }
    return (
      <LineChart width={500} height={300} data={data}>
        <XAxis dataKey="Month" />
        {/* <YAxis dataKey="Sentiment" /> */}
        <CartesianGrid stroke="#eee" strokeDasharray="5 5" />
        <Line type="monotone" dataKey="Sentiment" stroke="#8884d8" />
        <Line type="monotone" dataKey="Karma" stroke="#8884d8" />
        <Line type="monotone" dataKey="WordLength" stroke="#8884d8" />
        {/* <ReferenceLine y={0.0} label="Neutral" stroke="blue" /> */}
        {/* <Area
          type="monotone"
          dataKey="Sentiment"
          stroke="#8884d8"
          fill="#8884d8"
        /> */}
        <Legend verticalAlign="top" />
        <Tooltip />
      </LineChart>
    );
  }
}

const mapStateToProps = state => ({
  selectedSubreddit: state.selectedSubreddit,
  subreddits: state.subreddits
});

const mapDispatchToProps = dispatch => ({});

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(SelectedSubredditViewer);
