import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";

import MulticastGroupForm from "./MulticastGroupForm";

import ApplicationStore from "../../stores/ApplicationStore";
import MulticastGroupStore from "../../stores/MulticastGroupStore";


const styles = {
  card: {
    overflow: "visible",
  },
};


class CreateMulticastGroup extends Component {
  constructor() {
    super();
    this.state = {
      spDialog: false,
    };
    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    ApplicationStore.get(this.props.match.params.applicationID, resp => {
      this.setState({
        application: resp,
      });
    });
  }

  onSubmit = (multicastGroup) => {
    let mg = multicastGroup;
    mg.applicationID = this.props.match.params.applicationID;

    MulticastGroupStore.create(mg, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/multicast-groups`);
    });
  }

  render() {
    if (this.state.application === undefined) {
      return null;
    }

    return(
      <Grid container spacing={4}>
        <TitleBar>
          <TitleBarTitle title="Applications" to="/applications" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.application.application.name} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Multicast groups" to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/multicast-groups`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Create" />
        </TitleBar>

        <Grid item xs={12}>
          <Card className={this.props.classes.card}>
            <CardContent>
              <MulticastGroupForm
                submitLabel="Create multicast-group"
                onSubmit={this.onSubmit}
                match={this.props.match}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(CreateMulticastGroup));
