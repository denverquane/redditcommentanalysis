import React from "react";
import { Button, Intent } from "@blueprintjs/core";

export class CollapseExample extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      isOpen: false,
      component: props.component,
      typeLabel: props.typeLabel
    };
  }
  state = {
    isOpen: false,
    component: null,
    typeLabel: ""
  };

  componentWillReceiveProps(props) {
    this.setState({
      component: props.component,
      typeLabel: props.typeLabel
    });
  }

  render() {
    return (
      <div>
        {this.state.component && this.state.component.length > 0? <Button intent={Intent.PRIMARY} onClick={this.handleClick}>
          {this.state.isOpen ? "Hide" : "Show"} {this.state.typeLabel}
        </Button> : null}
        {this.state.isOpen ? this.state.component : <div />}
      </div>
    );
  }

  handleClick = () => {
    this.setState({ isOpen: !this.state.isOpen });
  };
}
