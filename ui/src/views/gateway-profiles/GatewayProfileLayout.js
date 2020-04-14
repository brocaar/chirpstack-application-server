import React, { Component } from "react";
import { withRouter } from "react-router-dom";

import Grid from '@material-ui/core/Grid';

import Delete from "mdi-material-ui/Delete";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TitleBarButton from "../../components/TitleBarButton";
import GatewayProfileStore from "../../stores/GatewayProfileStore";
import UpdateGatewayProfile from "./UpdateGatewayProfile";


class GatewayProfileLayout extends Component {
  constructor() {
    super();

    this.state = {};

    this.deleteGatewayProfile = this.deleteGatewayProfile.bind(this);
  }

  componentDidMount() {
    GatewayProfileStore.get(this.props.match.params.gatewayProfileID, resp => {
      this.setState({
        gatewayProfile: resp,
      });
    });
  }

  deleteGatewayProfile() {
    if (window.confirm("Are you sure you want to delete this gateway-profile?")) {
      GatewayProfileStore.delete(this.props.match.params.gatewayProfileID, () => {
        this.props.history.push("/gateway-profiles");
      });
    }
  }

  render() {
    if (this.state.gatewayProfile === undefined) {
      return(<div></div>);
    }

    return(
      <Grid container spacing={4}>
        <TitleBar
          buttons={[
            <TitleBarButton
              key={1}
              label="Delete"
              icon={<Delete />}
              color="secondary"
              onClick={this.deleteGatewayProfile}
            />,
          ]}
        >
          <TitleBarTitle to="/gateway-profiles" title="Gateway-profiles" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.gatewayProfile.gatewayProfile.name} />
        </TitleBar>

        <Grid item xs={12}>
          <UpdateGatewayProfile gatewayProfile={this.state.gatewayProfile.gatewayProfile} />
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(GatewayProfileLayout);
