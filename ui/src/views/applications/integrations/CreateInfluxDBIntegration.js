import React, {Component} from "react";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";

import ApplicationStore from "../../../stores/ApplicationStore";
import InfluxDBIntegrationForm from "./InfluxDBIntegrationForm";
import InfluxDBV2IntegrationForm from "./InfluxDBV2IntegrationForm";
import Switch from '@material-ui/core/Switch';
import Form from "../../../components/Form";
import FormLabel from "@material-ui/core/FormLabel";


class CreateInfluxDBIntegration extends Component {
  constructor(props) {
    super(props)
    this.state ={
      V2: false
    }
  }

  onSubmit = (integration) => {
    let integr = integration;
    integr.applicationID = this.props.match.params.applicationID;

    ApplicationStore.createInfluxDBIntegration(integr, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
    }, this.state.V2);
  }

  handleChange=(ev)=>{
    this.setState({V2: !this.state.V2})
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card>
            <CardHeader title="Add InfluxDB integration" />
            <CardContent>
              <Form>
                <FormLabel>InfluxDB v2+</FormLabel>
                <Switch id="version" checked={this.state.V2} onChange={this.handleChange}/>
              </Form>
              <hr/>
              {this.state.V2 ?
                    <InfluxDBV2IntegrationForm submitLabel="Add integration" onSubmit={this.onSubmit}/>
                    : <InfluxDBIntegrationForm submitLabel="Add integration" onSubmit={this.onSubmit}/>
                }
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default CreateInfluxDBIntegration;
