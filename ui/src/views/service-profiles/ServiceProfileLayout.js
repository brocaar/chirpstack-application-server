import React, { Component } from "react";
import { withRouter } from "react-router-dom";

import Grid from '@material-ui/core/Grid';

import Delete from "mdi-material-ui/Delete";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TitleBarButton from "../../components/TitleBarButton";
import Admin from "../../components/Admin";
import ServiceProfileStore from "../../stores/ServiceProfileStore";
import SessionStore from "../../stores/SessionStore";
import UpdateServiceProfile from "./UpdateServiceProfile";


class ServiceProfileLayout extends Component {
  constructor() {
    super();
    this.state = {
      admin: false,
    };
    this.deleteServiceProfile = this.deleteServiceProfile.bind(this);
    this.setIsAdmin = this.setIsAdmin.bind(this);
  }

  componentDidMount() {
    ServiceProfileStore.get(this.props.match.params.serviceProfileID, resp => {
      this.setState({
        serviceProfile: resp,
      });
    });

    SessionStore.on("change", this.setIsAdmin);
    this.setIsAdmin();
  }

  componentWillUnmount() {
    SessionStore.removeListener("change", this.setIsAdmin);
  }

  setIsAdmin() {
    this.setState({
      admin: SessionStore.isAdmin(),
    });
  }

  deleteServiceProfile() {
    if (window.confirm("Are you sure you want to delete this service-profile?")) {
      ServiceProfileStore.delete(this.props.match.params.serviceProfileID, resp => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/service-profiles`);
      });
    }
  }

  render() {
    if (this.state.serviceProfile === undefined) {
      return(<div></div>);
    }

    return(
      <Grid container spacing={4}>
        <TitleBar
          buttons={
            <Admin>
              <TitleBarButton
                key={1}
                label="Delete"
                icon={<Delete />}
                color="secondary"
                onClick={this.deleteServiceProfile}
              />
            </Admin>
          }
        >
          <TitleBarTitle to={`/organizations/${this.props.match.params.organizationID}/service-profiles`} title="Service-profiles" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.serviceProfile.serviceProfile.name} />
        </TitleBar>

        <Grid item xs={12}>
          <UpdateServiceProfile serviceProfile={this.state.serviceProfile.serviceProfile} admin={this.state.admin} />
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(ServiceProfileLayout);
