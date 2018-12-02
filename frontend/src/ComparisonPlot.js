import React from "react";
import { connect } from "react-redux";
import { Months } from "./MonthYearSelection";
import { CollapseExample } from "./Collapse";
import { toggleCompareSubreddit } from "./reducer";
import Plot from "react-plotly.js";

const YEAR = "2010"

class ComparisonPlot extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
    };
  }

  componentWillReceiveProps(props) {

  }

  render() {
    let karmaArrayData = [];
    let polarityArrayData = [];
    let subjArrayData = [];
    let commentLengthArrayData = [];
    let commentsArrayData = [];

    for (let sub in this.props.compareSubreddits) {
      //loop over the subs we have to compare
      let index = this.props.compareSubreddits[sub];
      if (
        this.props.subreddits[index] &&
        this.props.subreddits[index].ProcessedYearMonthCommentSummaries
      ) {
        let extractedInfo = this.props.subreddits[index]
          .ProcessedYearMonthCommentSummaries[YEAR];
        if (extractedInfo) {
            let karmasubData = [];
            let polaritysubData = [];
            let subjsubData = [];
            let commentLengthsubData = [];
            let commentssubData = [];

          for (let mo in extractedInfo) {
            if (extractedInfo[mo]) {
              let monthData = extractedInfo[mo];
              if (monthData) {
                karmasubData.push(monthData["Karma"].Average);
                polaritysubData.push(monthData["Polarity"].Average);
                subjsubData.push(monthData["Subjectivity"].Average);
                commentLengthsubData.push(monthData["WordLength"].Average);
                commentssubData.push(monthData["TotalComments"]);

              }
            }
          }
          karmaArrayData.push({
            x: Months,
            y: karmasubData,
            name: index,
            mode: "lines",
            type: "scatter"
          });
          polarityArrayData.push({
            x: Months,
            y: polaritysubData,
            name: index,
            mode: "lines",
            type: "scatter"
          });
          subjArrayData.push({
            x: Months,
            y: subjsubData,
            name: index,
            mode: "lines",
            type: "scatter"
          });
          commentLengthArrayData.push({
            x: Months,
            y: commentLengthsubData,
            name: index,
            mode: "lines",
            type: "scatter"
          });
          commentsArrayData.push({
            x: Months,
            y: commentssubData,
            name: index,
            mode: "lines",
            type: "scatter"
          });
        }
      }
    }
    return (
      <div>
        <Plot
          data={commentsArrayData}
          layout={{
            width: 900,
            height: 450,
            title: YEAR+" Total Comments",

            shapes: {}
          }}
        />
        <Plot
          data={karmaArrayData}
          layout={{
            width: 900,
            height: 450,
            title: YEAR+ " Average Karma",

            shapes: {}
          }}
        />
        <Plot
          data={polarityArrayData}
          layout={{
            width: 900,
            height: 450,
            title: YEAR + " Average Sentiment",

            shapes: {}
          }}
        />
        <Plot
          data={subjArrayData}
          layout={{
            width: 900,
            height: 450,
            title: YEAR + " Average Subjectivity",

            shapes: {}
          }}
        />
        <Plot
          data={commentLengthArrayData}
          layout={{
            width: 900,
            height: 450,
            title: YEAR + " Average Comment Word Length",

            shapes: {}
          }}
        />
      </div>
    );
  }
}

const mapStateToProps = state => ({
  compareSubreddits: state.compareSubreddits,
  subreddits: state.subreddits
});

const mapDispatchToProps = dispatch => ({
  toggleCompareSubreddit: sub => dispatch(toggleCompareSubreddit(sub))
});

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(ComparisonPlot);
