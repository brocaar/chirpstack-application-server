import React, { Component } from "react";
import { Link } from "react-router-dom";

import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';

import theme from "../theme";


const styles = {
  button: {
    marginLeft: theme.spacing.unit,
  },
  icon: {
    marginRight: theme.spacing.unit,
  },
};


class TitleBarButton extends Component {
  render() {
    let component = "button";
    let icon = null;

    if (this.props.to !== undefined) {
      component = Link
    }

    if (this.props.icon !== undefined) {
      icon = React.cloneElement(this.props.icon, {
        className: this.props.classes.icon,
      })
    }

    return(
      <Button
        variant="outlined"
        color={this.props.color || "default"}
        className={this.props.classes.button}
        component={component}
        to={this.props.to}
        onClick={this.props.onClick}
      >
        {icon}
        {this.props.label}
      </Button>
    );
  }
}

export default withStyles(styles)(TitleBarButton);
