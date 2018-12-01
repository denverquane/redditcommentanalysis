import React from "react";
import { connect } from "react-redux";
import { Months } from "./MonthYearSelection";
import { CollapseExample } from "./Collapse";
import Plot from "react-plotly.js";

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

        {this.state.YearMonthProcessSummaries ? (
          <div>
            {/* <CollapseExample
              component={this.getBoxplotsForYearAndMonthsByKey(
                this.state.YearMonthProcessSummaries,
                "WordLength"
              )}
              typeLabel="Word Length Boxplot"
            /> */}
            {this.getPlotlyBoxPlot(
              this.state.YearMonthProcessSummaries["2016"]
            )}
            {/* <CollapseExample
              component={this.getBoxplotsForYearAndMonthsByKey(
                this.state.YearMonthProcessSummaries,
                "Karma"
              )}
              typeLabel="Karma Boxplot"
            /> */}
            {/* <CollapseExample
              component={this.getBoxplotsForYearAndMonthsByKey(
                this.state.YearMonthProcessSummaries,
                "Sentiment"
              )}
              typeLabel="Sentiment Boxplot"
            /> */}
          </div>
        ) : null}
        {/* {this.state.YearMonthProcessSummaries
          ? this.getRechartsPlotAcrossMonths(
              this.state.YearMonthProcessSummaries
            )
          : null} */}
      </div>
    ) : null;
  }

  getPlotlyBoxPlot(boxplotdata) {
    let shapes = [];
    let karmaData = [];
    let commentData = [];
    let sentimentData = [];
    let wordlengthData = [];

    let min = 100000000;
    let max = -100000000;
    for (let moidx in boxplotdata) {
      if (boxplotdata[moidx]) {
        let datapt = boxplotdata[moidx]

        if (datapt["Karma"].Min < min) {
          min = datapt["Karma"].Min;
        }
        if (datapt["Karma"].Max > max) {
          max = datapt["Karma"].Max;
        }

        shapes.push({
          x0: moidx,
          x1: moidx + 1,
          y0: datapt["Karma"].Min,
          y1: datapt["Karma"].Min,
          'line': {
            'color': 'rgb(55, 128, 191)',
            'width': 3,
        },
          type: 'line'
        });
        karmaData.push(
          datapt["Karma"].Average,
        );
        // commentData.push(
        //   datapt["TotalComments"],
        // );
        sentimentData.push(
          datapt["Sentiment"].Average,
        );
        wordlengthData.push(
          datapt["WordLength"].Average,
        );
      }
    }
    return (
      <Plot
        data= {[
          {
            x : Months,
            y: karmaData,
            name: 'Karma',
            mode: 'lines',
            type: 'scatter'
          },
          {
            x : Months,
            y: wordlengthData,
            name: 'Comment Word Length',
            mode: 'lines',
            type: 'scatter'
          },
          {
            x : Months,
            y: sentimentData,
            name: 'Sentiment',
            mode: 'lines',
            type: 'scatter'
          }
        ]}
        layout={{
          width: 800,
          height: 600,
          title: "A Fancy Plot",
          
          shapes: shapes
        }}
      />
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
