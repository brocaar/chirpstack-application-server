import React, { Component } from "react";
import { withRouter } from "react-router-dom";

import Snackbar from '@material-ui/core/Snackbar';
import IconButton from '@material-ui/core/IconButton';
import Close from "mdi-material-ui/Close";

import NotificationStore from "../stores/NotificationStore";
import dispatcher from "../dispatcher";


class Item extends Component {
  constructor() {
    super();
    this.onClose = this.onClose.bind(this);
  }

  onClose(event, reason) {
    dispatcher.dispatch({
      type: "DELETE_NOTIFICATION",
      id: this.props.id,
    });
  }

  render() {
    return(
      <Snackbar
        anchorOrigin={{
          vertical: "bottom",
          horizontal: "left",
        }}
        open={true}
        message={<span>{this.props.notification.message}</span>}
        autoHideDuration={3000}
        onClose={this.onClose}
        action={[
          <IconButton
            key="close"
            aria-label="Close"
            color="inherit"
            onClick={this.onClose}
          >
            <Close />
          </IconButton>
        ]}
      />
    );
  }
}


class Notifications extends Component {
  constructor() {
    super();

    this.state = {
      notifications: NotificationStore.getAll(),
    };
  }

  componentDidMount() {
    NotificationStore.on("change", () => {
      this.setState({
        notifications: NotificationStore.getAll(),
      });
    });
  }

  render() {
    const items = this.state.notifications.map((n, i) => <Item key={n.id} id={n.id} notification={n} />);

    return (items);
  }
}

export default withRouter(Notifications);
