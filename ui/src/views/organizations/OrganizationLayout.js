import React, { Component } from "react";
import { Route, Redirect, Switch, withRouter } from "react-router-dom";

import Grid from '@material-ui/core/Grid';

import Delete from "mdi-material-ui/Delete";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TitleBarButton from "../../components/TitleBarButton";
import OrganizationStore from "../../stores/OrganizationStore";
import UpdateOrganization from "./UpdateOrganization";


class OrganizationLayout extends Component {
  constructor() {
    super();
    this.state = {};
    this.loadData = this.loadData.bind(this);
    this.deleteOrganization = this.deleteOrganization.bind(this);
  }

  componentDidMount() {
    this.loadData();
  }

  componentDidUpdate(prevProps) {
    if (prevProps === this.props) {
      return;
    }

    this.loadData();
  }

  loadData() {
    OrganizationStore.get(this.props.match.params.organizationID, resp => {
      this.setState({
        organization: resp,
      });
    });
  }

  deleteOrganization() {
    if (window.confirm("Are you sure you want to delete this organization?")) {
      OrganizationStore.delete(this.props.match.params.organizationID, () => {
        this.props.history.push("/organizations");
      });
    }
  }

  render() {
    if (this.state.organization === undefined) {
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
              onClick={this.deleteOrganization}
            />,
          ]}
        >
          <TitleBarTitle to="/organizations" title="Organizations" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.organization.organization.name} />
        </TitleBar>

        <Grid item xs={12}>
          <Switch>
            <Route exact path={this.props.match.path} render={() => <Redirect to={`${this.props.match.url}/edit`} />} />
            <Route exact path={`${this.props.match.path}/edit`} render={props => <UpdateOrganization organization={this.state.organization.organization} {...props} />} />
          </Switch>
        </Grid>
      </Grid>
    );
  }
}


export default withRouter(OrganizationLayout);
