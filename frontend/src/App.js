import React, { Component } from 'react';
import './App.css';
import { Button, Collapse, Toaster, Position, Intent, Spinner, InputGroup, Card, Elevation, Tag, Checkbox } from '@blueprintjs/core'
import { Circle } from 'rc-progress'
import { BarChart, Legend, XAxis, YAxis, Bar, CartesianGrid, Tooltip } from 'recharts'
import Sockette from 'sockette';

import '@blueprintjs/core/lib/css/blueprint.css';

let Months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
let IP = "localhost";

const AppToaster = Toaster.create({
    className: "notifyToaster",
    position: Position.TOP
});


class App extends Component {

    ws = new Sockette('ws://' + IP + ':5000/ws', {
        timeout: 5e3,
        maxAttempts: 10,
        onopen: (e) => {
            console.log('Connected!', e);
            AppToaster.show({ message: 'Connected to Backend!', intent: Intent.SUCCESS });
            this.getSubs();
            this.getStatus();
            this.setState({ ...this.state, Websocket: true });
        },
        onmessage: (e) => {
            console.log('Received:', e);
            console.log(String(e.data));
            if (String(e.data).includes("fetch")) {
                this.getSubs();
                this.getStatus();
                console.log('Fetch message')
            } else if (String(e.data).includes("status")) {
                this.getStatus();
                console.log('Status message')
            } else if (String(e.data).includes("error")) {
                console.log('Error message');
                AppToaster.show({ message: 'Error!' + e.data, intent: Intent.DANGER });
            }
        },
        onreconnect: e => console.log('Reconnecting...', e),
        onmaximum: e => console.log('Stop Attempting!', e),
        onclose: (e) => {
            console.log('Closed!', e);
            AppToaster.show({ message: 'Disconnected from backend!', intent: Intent.DANGER });
            this.setState({ ...this.state, Websocket: false });
        },
        onerror: (e) => {
            console.log('Error!', e);
            AppToaster.show({ message: 'Error connecting to backend!', intent: Intent.DANGER });
            this.setState({ ...this.state, Websocket: false });
        },
    });
    state = {
        Websocket: Boolean,
        Subs: Object,
        Status: Object,
        TempExtractName: String,
    };
    constructor() {
        super();
        this.state.Websocket = false;
        this.getStatus();
    }

    render() {
        return (
            <div className="App">
                <header className="App-header" style={{ display: 'flex', flexDirection: 'row' }}>
                    <div style={{ width: '33%', display: 'flex', flexDirection: 'column' }}>
                        Extraction Queue:
                {this.getLabels(this.state.Status.ExtractQueue)}
                    </div>
                    <div style={{ width: '33%', display: 'flex', flexDirection: 'row' }}>
                        {this.state.Status.Extracting ?
                            (
                                <div style={{ width: '20%', height: '20%' }}>
                                    <Circle strokeWidth="10" percent={this.state.Status.ExtractProgress} />
                                    "{this.state.Status.ExtractQueue[0]["Subreddit"]}" {this.state.Status.ExtractProgress.toFixed(3)}% Extracted
                        </div>
                            ) : <div style={{ width: '20%', height: '20%' }}>
                                <Circle strokeWidth="10" percent={this.state.Status.ExtractProgress} />
                                Nothing to Extract!
                    </div>
                        }
                        {<div style={{ width: '60%' }}><Spinner intent={this.state.Websocket ? Intent.SUCCESS : Intent.DANGER}
                            value={this.state.Websocket ? (this.state.Status.Extracting || this.state.Status.Processing) ? null : 1 : 1} /> </div>}
                        {this.state.Status.Processing ?
                            (
                                <div style={{ width: '20%', height: '20%' }}>
                                    <Circle strokeWidth="10" percent={this.state.Status.ProcessProgress} />
                                    "{this.state.Status.ProcessQueue[0]["Subreddit"]}" {this.state.Status.ProcessProgress.toFixed(3)}% Processed
                        </div>
                            ) : <div style={{ width: '20%', height: '20%' }}>
                                <Circle strokeWidth="10" percent={this.state.Status.ProcessProgress} />
                                Nothing to Process!
                    </div>
                        }
                    </div>
                    <div style={{ width: '33%', display: 'flex', flexDirection: 'column' }}>
                        Processing Queue:
                {this.getLabels(this.state.Status.ProcessQueue)}
                    </div>
                </header>
                <div style={{ width: '100%', display: 'flex', flexDirection: 'row' }}>
                    <div style={{ width: '25%' }} />
                    <div style={{ width: '35%' }}><InputGroup
                        onChange={(event) => (this.setState({ ...this.state, TempExtractName: event.target.value }))} /></div>
                    <div style={{ width: '15%' }}><Button onClick={() => {
                        this.extractSubreddit(this.state.TempExtractName);
                        console.log(this.state.TempExtractName);
                    }}>Extract Sub</Button></div>
                </div>
                {this.displaySubs()}
                {/* {this.state.Subs.EarthPorn ? (this.state.Subs.EarthPorn.Processing ? <p>'True'</p> : <p>'False'</p>) : ''} */}
            </div>
        );
    }
    displaySubs() {
        let arr = [];
        for (let key in this.state.Subs) {
            let years = [];
            let totalComments = 0;
            let substatus = this.state.Subs[key];

            for (let yr in substatus["ExtractedMonthCommentCounts"]) {
                let months = [];
                for (let month in substatus["ExtractedMonthCommentCounts"][yr]) {
                    months.push(<div><h3>{month}</h3><p>{substatus["ExtractedMonthCommentCounts"][yr][month]}</p></div>)
                }
                years.push(
                    <MonthYearSelector Year={yr} Months={substatus["ExtractedMonthCommentCounts"][yr]} Subreddit={key} ExtractFunc={this.extractSubreddit} />
                    // <div>
                    //     <h2>{yr}</h2>
                    //     {months}
                    // </div>
                )
            }
            // if (!substatus.ExtractedMonthCommentCounts || substatus.ExtractedMonthCommentCounts.length < 12){
            //     continue
            // }
            // // let all = true;
            // for (let mo in Months) {
            //     // if (!substatus.ExtractedMonthCommentCounts[Months[mo]]) {
            //     //     all = false
            //     //     break
            //     // }
            //     months.push({key: Months[mo], comments: substatus.ExtractedMonthCommentCounts[Months[mo]]});
            //     totalComments += substatus.ExtractedMonthCommentCounts[Months[mo]];
            // }
            // // if (!all) {
            // //     continue
            // // }
            // let words = [];
            // for (let wo in substatus.ProcessedSummary.KeywordCommentTallies) {
            //     words.push({key: wo, 'Percent of Comments Containing': substatus.ProcessedSummary.KeywordCommentTallies[wo],
            //         'Karma per Comment Containing': substatus.ProcessedSummary.KeywordCommentKarmas[wo]})
            // }
            arr.push(
                <Card key={key} style={{ display: 'flex', backgroundColor: '#ECEBEB', flexDirection: 'column', marginBottom: '1%' }} elevation={Elevation.TWO} interactive={true}>
                    <h1>{key}</h1>
                    ({totalComments.toLocaleString()} total comments)
              {years}
                    {/*<CollapseExample text={{label: 'Monthly Data', type: 'months', months: months}}/>*/}
                    {/*{this.state.Subs[key].Processed ? <div><CollapseExample text={{label: 'Keyword Data', type: 'karma', words: words}}/></div> : this.processButton(key)}*/}
                </Card>);
        }
        arr.sort(function (a, b) {
            var x = a.key.toLowerCase();
            var y = b.key.toLowerCase();
            if (x < y) { return -1; }
            if (x > y) { return 1; }
            return 0;
        });
        return arr
    }

    getSubs() {
        fetch('http://' + IP + ':5000/api/subs')
            .then(results => {
                return results.json();
            }).then(data => {

                if (JSON.stringify(this.state.Subs) !== JSON.stringify(data)) {
                    // console.log('Updated');
                    this.setState({ ...this.state, Subs: data });
                }

            });
    }

    getStatus() {
        fetch('http://' + IP + ':5000/api/status')
            .then(results => {
                return results.json();
            }).then(data => {
                if (JSON.stringify(this.state.Status) !== JSON.stringify(data)) {
                    this.setState({ ...this.state, Status: data });
                }
            });
    }

    getLabels(queue) {
        let arr = [];
        if (!queue || queue.length === 0) {
            return <div>Empty Queue</div>
        }
        for (let i in queue) {
            if (i === "0") {
                arr.push(
                    <div>
                        <div style={{ width: '33%' }} />
                        <Tag key={i} intent={Intent.SUCCESS} style={{ width: '33%', marginBottom: '1%' }}>{i}. {queue[i]["Subreddit"]} </Tag>
                        <div style={{ width: '33%' }} />
                    </div>
                )
            } else {
                arr.push(
                    <div>
                        <div style={{ width: '33%' }} />
                        <Tag key={i} intent={Intent.NONE} style={{ width: '33%' }}>{i}. {queue[i]["Subreddit"]}</Tag>
                        <div style={{ width: '33%' }} />
                    </div>
                )
            }

        }
        return arr
    }

    processButton(subreddit) {
        return <div><Button intent={Intent.WARNING} onClick={() => this.processSubreddit(subreddit)}>
            Process {subreddit}
        </Button></div>
    }

    processSubreddit(sub) {
        fetch('http://' + IP + ':5000/api/processSub/' + sub, {
            method: 'post'
        })
            .then(results => {
                return results;
            }).then(data => {
                console.log(data);
                AppToaster.show({ message: 'Asked backend to process: ' + sub, intent: Intent.NONE });
                this.getStatus();
            });
    }
    extractSubreddit(sub, month, year) {
        fetch('http://' + IP + ':5000/api/extractSub/' + sub + "/" + month + "/" + year, {
            method: 'post'
        })
            .then(results => {
                return results;
            }).then(data => {
                console.log(data);
                AppToaster.show({ message: 'Asked backend to extract: ' + sub, intent: Intent.NONE });
            
            });
    }
}

export class CollapseExample extends React.Component {
    constructor(text) {
        super(text);
        this.state = {
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
                <Button intent={Intent.PRIMARY} onClick={this.handleClick}>
                    {this.state.isOpen ? "Hide" : "Show"} {this.state.text.text.label}
                </Button>
                {this.state.text.text.type === 'months' ?
                    <Collapse isOpen={this.state.isOpen}>
                        {/*<LineChart width={width * 0.9} height={500} data={this.state.text.text.months} margin={{left: width*0.05}}>*/}
                        {/*<XAxis dataKey="key"/>*/}
                        {/*<YAxis />*/}
                        {/*<CartesianGrid stroke="#eee" strokeDasharray="5 5"/>*/}
                        {/*<Line dataKey="comments" fill="blue" />\*/}
                        {/*<Tooltip/>*/}
                        {/*<Legend/>*/}
                        {/*</LineChart>*/}

                    </Collapse>
                    :
                    <Collapse isOpen={this.state.isOpen}>
                        <BarChart width={width * 0.9} height={500} data={this.state.text.text.words} margin={{ left: width * 0.05 }}>
                            <XAxis dataKey="key" />
                            <YAxis />
                            <CartesianGrid stroke="#eee" strokeDasharray="5 5" />
                            <Bar dataKey="Percent of Comments Containing" fill="blue" />\
                            <Bar dataKey="Karma per Comment Containing" fill="red" />\
                            <Tooltip />
                            <Legend />
                        </BarChart>
                    </Collapse>
                }
            </div>
        );
    }

    handleClick = () => {
        this.setState({ isOpen: !this.state.isOpen });
    }
}

export class MonthYearSelector extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            Months: props.Months,
            PendingJobs: [],
            Year: props.Year,
            ExtractFunc: props.ExtractFunc
        }
    }
    state = {
        Months: [],
        PendingJobs: [],
        Year: "",
    }

    render() {
        let arr = [];
        for (let mo in Months) {
            if (this.state.Months[Months[mo]]) {
                arr.push(<Tag intent={Intent.SUCCESS}>{Months[mo]} : {this.state.Months[Months[mo]]} Comments</Tag>)
            } else {
                arr.push(
                    <Checkbox onChange={(val) => {
                        if (this.state.PendingJobs.indexOf(Months[mo]) === -1) {
                            let joined = this.state.PendingJobs.concat(Months[mo])
                            this.setState({ PendingJobs: joined });
                        } else {
                            let arr = [...this.state.PendingJobs];
                            let index = this.state.PendingJobs.indexOf(Months[mo])
                            if (index !== -1) {
                                arr.splice(index, 1);
                                this.setState({ PendingJobs: arr });
                            }
                        }

                    }}>{Months[mo]}</Checkbox>
                )
            }

        }
        return (
            <div>
                <h2>{this.state.Year}</h2>
                <div style={{ display: 'flex', flexDirection: 'row' }}>
                    <div style={{ display: 'flex', flexDirection: 'column', width: '25%' }}>{arr.slice(0, 3)}</div>
                    <div style={{ display: 'flex', flexDirection: 'column', width: '25%' }}>{arr.slice(3, 6)}</div>
                    <div style={{ display: 'flex', flexDirection: 'column', width: '25%' }}>{arr.slice(6, 9)}</div>
                    <div style={{ display: 'flex', flexDirection: 'column', width: '25%' }}>{arr.slice(9, 12)}</div>
                </div>
                <Button onClick={() => {
                    console.log(this.state.PendingJobs);
                    for (let mo in this.state.PendingJobs) {
                        this.state.ExtractFunc(this.props.Subreddit, mo, this.state.Year);
                    }
                    
                }}
                disabled={this.state.PendingJobs.length === 0}>Confirm</Button>
            </div>
        )
    }
}

export default App;
