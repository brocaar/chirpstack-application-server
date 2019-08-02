import React, { Component } from "react";
import { withRouter } from "react-router-dom";

import Grid from '@material-ui/core/Grid';

import Delete from "mdi-material-ui/Delete";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TitleBarButton from "../../components/TitleBarButton";

import NetworkServerStore from "../../stores/NetworkServerStore";
import UpdateNetworkServer from "./UpdateNetworkServer";


class NetworkServerLayout extends Component {
  constructor() {
    super();

    this.state = {};

    this.deleteNetworkServer = this.deleteNetworkServer.bind(this);
  }

  componentDidMount() {
    NetworkServerStore.get(this.props.match.params.networkServerID, (resp) => {
      this.setState({
        networkServer: resp,
      });
    });
  }

  deleteNetworkServer() {
    if (window.confirm("Are you sure you want to delete this network-server?")) {
      NetworkServerStore.delete(this.props.match.params.networkServerID, () => {
        this.props.history.push("/network-servers");
      });
    }
  }

  render() {
    if (this.state.networkServer === undefined) {
      return(<div></div>);
    }

    return(
      <Grid container spacing={4}>
        <TitleBar
          buttons={[
            <TitleBarButton
              key={1}
              icon={<Delete />}
              label="Delete"
              color="secondary"
              onClick={this.deleteNetworkServer}
            />,
          ]}
        >
          <TitleBarTitle to="/network-servers" title="Network-servers" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={`${this.state.networkServer.networkServer.name} (${this.state.networkServer.region} @ ${this.state.networkServer.version})`} />
        </TitleBar>

        <Grid item xs={12}>
          <UpdateNetworkServer networkServer={this.state.networkServer.networkServer} />
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(NetworkServerLayout);
