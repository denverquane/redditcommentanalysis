import React from "react";
import { connect } from "react-redux";
import { Months } from "./MonthYearSelection";
import { Boxplot } from "react-boxplot";
import {
  Card
} from "@blueprintjs/core";

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
    return this.props.selectedSubreddit ? (
      <div>
        <h1>Selected Subreddit:</h1>
        <h1>{this.props.selectedSubreddit}</h1>

        {this.state.YearMonthProcessSummaries && this.state.YearMonthProcessSummaries["2016"] ? (
          <div>
            <Card>
              <h3>Karma:</h3>
              {this.getBoxplotsForMonthsByKey(
              this.state.YearMonthProcessSummaries["2016"],
              "Karma"
            )}
            </Card>
            <Card>
              <h3>Sentiment:</h3>
              {this.getBoxplotsForMonthsByKey(
              this.state.YearMonthProcessSummaries["2016"],
              "Polarity"
            )}
            </Card>
            <Card>
              <h3>Subjectivity:</h3>
              {this.getBoxplotsForMonthsByKey(
              this.state.YearMonthProcessSummaries["2016"],
              "Subjectivity"
            )}
            </Card>
            <Card>
              <h3>Comment Length:</h3>
              {this.getBoxplotsForMonthsByKey(
              this.state.YearMonthProcessSummaries["2016"],
              "WordLength"
            )}
            </Card>
          </div>
        ) : null}
      
      </div>
    ) : null;
  }
  getBoxplotsForMonthsByKey(yeardata, key) {
    let plots = [];
    let absMin = 10000000;
    let absMax = -10000000;

    for (let mo in Months) {
      let data = yeardata[Months[mo]];
      if (data) {
        if (data[key].Min < absMin) {
          absMin = data[key].Min;
        }
        if (data[key].Max > absMax) {
          absMax = data[key].Max;
        }
      }
    }

    for (let mo in Months) {
      let data = yeardata[Months[mo]];
      if (data) {
        plots.push(this.getBoxplot(data[key], Months[mo], absMin, absMax));
      }
    }

    return plots;
  }

  getBoxplot(boxplotdata, month, min, max) {
    return (
      <div style={{ display: "flex", flexDirection: "column" }}>
        <div style={{ display: "flex", flexDirection: "row" }}>
          <div style={{ backgroundColor: "#cccccc" }}>
            <Boxplot
              width={400}
              height={15}
              orientation="horizontal"
              min={min}
              max={max}
              stats={{
                whiskerLow: boxplotdata.Min,
                quartile1: boxplotdata.FirstQuartile,
                quartile2: boxplotdata.Median,
                quartile3: boxplotdata.ThirdQuartile,
                whiskerHigh: boxplotdata.Max,
                outliers: []
              }}
            />
          </div>
          {month}
        </div>
        Min:{boxplotdata.Min}, Q1: {boxplotdata.FirstQuartile}, Median: {boxplotdata.Median}, Q3: {boxplotdata.ThirdQuartile}, Max: {boxplotdata.Max}
      </div>
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
