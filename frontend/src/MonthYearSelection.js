import React from 'react';
import { Checkbox, Button } from '@blueprintjs/core'

let Months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

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
    };

    render() {
        let arr = [];
        for (let mo in Months) {
            if (this.state.Months[Months[mo]]) {
                if (this.state.Months[Months[mo]] === -1) {
                    arr.push(
                        <Checkbox key= {Months[mo]} disabled={true}>{Months[mo]}: No Data</Checkbox>
                    )
                } else {
                    arr.push(<Checkbox key= {Months[mo]} checked={true}>{Months[mo]} : {this.state.Months[Months[mo]]} Comments</Checkbox>)
                }
            } else {
                arr.push(
                    <Checkbox key= {Months[mo]} onChange={(val) => {
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
                        this.state.ExtractFunc(this.props.Subreddit, this.state.PendingJobs[mo], this.state.Year);
                    }
                    
                }}
                disabled={this.state.PendingJobs.length === 0}>Confirm</Button>
            </div>
        )
    }
}