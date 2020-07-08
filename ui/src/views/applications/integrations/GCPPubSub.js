import React, { Component } from "react";
import { Link } from "react-router-dom";

import { withStyles } from '@material-ui/core/styles';
import Card from '@material-ui/core/Card';
import CardActions from '@material-ui/core/CardActions';
import CardContent from '@material-ui/core/CardContent';
import CardMedia from '@material-ui/core/CardMedia';
import Button from '@material-ui/core/Button';
import Typography from '@material-ui/core/Typography';

import ApplicationStore from "../../../stores/ApplicationStore";


const styles ={
  media: {
    paddingTop: '35%',
    backgroundSize: 'contain',
  },
};


class GCPPubSubCard extends Component {
  delete = () => {
    if (window.confirm("Are you sure you want to remove the GCP Pub/Sub integration?")) {
      ApplicationStore.deleteGCPPubSubIntegration(this.props.applicationID, () => {});
    }
  }

  render() {
    return (
      <Card className={this.props.classes.root}>
          <CardMedia
            className={this.props.classes.media}
            image="/integrations/gcp_pubsub.png"
            title="GCP Pub/Sub"
          />
          <CardContent>
            <Typography gutterBottom variant="h5" component="h2">
              GCP Pub/Sub
            </Typography>
            <Typography variant="body2" color="textSecondary" component="p">
              The Google Cloud Pub/Sub integration forwards events to a GCP Pub/Sub topic.
            </Typography>
          </CardContent>
        <CardActions>
          {!this.props.add && <Link to={`/organizations/${this.props.organizationID}/applications/${this.props.applicationID}/integrations/gcp-pubsub/edit`}>
            <Button size="small" color="primary">
              Edit
            </Button>
          </Link>}
          {!this.props.add && <Button size="small" color="primary" onClick={this.delete}>
            Remove
          </Button>}
          {!!this.props.add && <Link to={`/organizations/${this.props.organizationID}/applications/${this.props.applicationID}/integrations/gcp-pubsub/create`}>
            <Button size="small" color="primary">
              Add
            </Button>
          </Link>}
        </CardActions>
      </Card>
    );
  }
}

export default withStyles(styles)(GCPPubSubCard);
