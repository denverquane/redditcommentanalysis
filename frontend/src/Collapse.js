import React from 'react';
import { Button, Intent } from '@blueprintjs/core'

export class CollapseExample extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            isOpen: false,
            component: props.component,
            typeLabel: props.typeLabel
        }
    }
    state = {
        isOpen: false,
        component: React.Component,
        typeLabel: ''
    };

    render() {
        return (
            <div>
                <Button intent={Intent.PRIMARY} onClick={this.handleClick}>
                    {this.state.isOpen ? "Hide" : "Show"} {this.state.typeLabel}
                </Button>
                {this.state.isOpen ? this.state.component : <div/>}
            </div>
        );
    }

    handleClick = () => {
        this.setState({ isOpen: !this.state.isOpen });
    }
};