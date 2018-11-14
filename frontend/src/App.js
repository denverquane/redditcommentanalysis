import React, { Component } from 'react';
import './App.css';
import { Button, Toaster, Position, Intent, Spinner, InputGroup, Card, Elevation, Tag} from '@blueprintjs/core'
import { Circle } from 'rc-progress'
// import { BarChart, Legend, XAxis, YAxis, Bar, CartesianGrid, Tooltip } from 'recharts'
import Sockette from 'sockette';
import { CollapseExample } from './Collapse' 
import {MonthYearSelector} from './MonthYearSelection'

import '@blueprintjs/core/lib/css/blueprint.css';

let IP = "dquane.tplinkdns.com";

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
                    <div style={{ width: '20%' }}><Button onClick={() => {
                        this.addSubredditEntry(this.state.TempExtractName);
                        console.log(this.state.TempExtractName);
                    }}>Add Subreddit Entry</Button></div>
                </div>
                {this.displaySubs()}
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
                    months.push(<div><h3>{month}</h3><p>{substatus["ExtractedMonthCommentCounts"][yr][month]}</p></div>);
                    if (substatus["ExtractedMonthCommentCounts"][yr][month] !== -1) {
                        totalComments += substatus["ExtractedMonthCommentCounts"][yr][month]
                    }
                }
                years.push(
                    <MonthYearSelector Year={yr} Months={substatus["ExtractedMonthCommentCounts"][yr]} Subreddit={key} ExtractFunc={this.extractSubreddit} />
                    // <div>
                    //     <h2>{yr}</h2>
                    //     {months}
                    // </div>
                )

            }
            arr.push(
                <Card key={key} style={{ display: 'flex', backgroundColor: '#ECEBEB', flexDirection: 'column', marginBottom: '1%' }} elevation={Elevation.TWO} interactive={true}>
                    <h1>{key}</h1>
                    ({totalComments.toLocaleString()} total comments)
                    <CollapseExample component={years} typeLabel={'Extraction Details'}/>
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

    addSubredditEntry(sub) {
        fetch('http://' + IP + ':5000/api/addSubEntry/' + sub, {
            method: 'post'
        })
            .then(results => {
                return results;
            }).then(data => {
                this.getSubs();
        });
    }
}

export default App;
