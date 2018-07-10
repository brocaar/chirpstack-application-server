import React, { Component } from "react";

import Grid from '@material-ui/core/Grid';
import { withStyles } from '@material-ui/core/styles';


const styles = {
  clear: {
    clear: "both",
  },

  left: {
    float: "left",
  },

  right: {
    float: "right",
  },
};


class TitleBar extends Component {
  render() {
    return(
        <Grid item xs className={this.props.classes.clear}>
          <div className={this.props.classes.left}>
              {this.props.children}
          </div>
          <div className={this.props.classes.right}>
              {this.props.buttons}
          </div>
        </Grid>
    );
  }
}

export default withStyles(styles)(TitleBar);
