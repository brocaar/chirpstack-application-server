import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";

import SessionStore from "../stores/SessionStore";
import theme from "../theme";

const styles = {
  footer: {
    paddingBottom: theme.spacing.unit,
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
    SessionStore.getBranding(resp => {
      if (resp.footer !== "") {
        this.setState({
          footer: resp.footer,
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
