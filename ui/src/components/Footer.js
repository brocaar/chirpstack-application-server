import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";

import InternalStore from "../stores/InternalStore";
import theme from "../theme";

const styles = {
  footer: {
    paddingBottom: theme.spacing(1),
    "& a": {
      color: theme.palette.primary.main,
      textDecoration: "none",
    },
  },
};

class Footer extends Component {
  constructor() {
    super();
    this.state = {
      footer: null,
    };
  }

  componentDidMount() {
    InternalStore.settings(resp => {
      if (resp.branding.footer !== "") {
        this.setState({
          footer: resp.branding.footer,
        });
      }
    });
  }

  render() {
    if (this.state.footer === null) {
      return(null);
    }

    return(
      <footer className={this.props.classes.footer}>
        <Typography align="center" dangerouslySetInnerHTML={{__html: this.state.footer}}></Typography>
      </footer>
    );
  }
}

export default withStyles(styles)(Footer);
