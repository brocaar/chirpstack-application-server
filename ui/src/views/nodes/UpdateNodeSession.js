import React, { Component } from 'react';
import { Link } from 'react-router';

import ApplicationStore from "../../stores/ApplicationStore";
import NodeSessionStore from "../../stores/NodeSessionStore";
import NodeSessionForm from "../../components/NodeSessionForm";
import NodeStore from "../../stores/NodeStore";
import ChannelStore from "../../stores/ChannelStore";

class UpdateNodeSession extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      session: {},
      node: {},
      sessionExists: false,
      channels: [],
    };

    this.submitHandler = this.submitHandler.bind(this);
    this.deleteHandler = this.deleteHandler.bind(this);
  }

  componentWillMount() {
    ApplicationStore.getApplication(this.props.params.applicationName, (application) => {
      this.setState({application: application});
    });

    NodeSessionStore.getNodeSession(this.props.params.applicationName, this.props.params.nodeName, (session) => {
      this.setState({
        session: session,
        sessionExists: true,
      });
    });

    NodeStore.getNode(this.props.params.applicationName, this.props.params.nodeName, (node) => {
      this.setState({node: node});

      // eslint-disable-next-line
      if(node.channelListID != 0) {
        ChannelStore.getChannelList(node.channelListID, (list) => {
          this.setState({channels: list.channels});
        });
      }
    });
  }

  submitHandler(session) {
    session.devEUI = this.state.node.devEUI;
    session.appEUI = this.state.node.appEUI;
    session.rxWindow = this.state.node.rxWindow;
    session.rxDelay = this.state.node.rxDelay;
    session.rx1DROffset = this.state.node.rx1DROffset;
    session.rx2DR = this.state.node.rx2DR;
    session.cFList = this.state.channels;
    session.relaxFCnt = this.state.node.relaxFCnt;
    session.adrInterval = this.state.node.adrInterval;
    session.installationMargin = this.state.node.installationMargin;

    if (this.state.sessionExists) {
      NodeSessionStore.updateNodeSession(this.props.params.applicationName, this.props.params.nodeName, session, (responseData) => {
        this.context.router.push("/applications/"+this.props.params.applicationName+"/nodes/"+this.props.params.nodeName+"/edit");
      });
    } else {
      NodeSessionStore.createNodeSession(this.props.params.applicationName, this.props.params.nodeName, session, (responseData) => {
        this.context.router.push("/applications/"+this.props.params.applicationName+"/nodes/"+this.props.params.nodeName+"/edit");
      });
    }
  }

  deleteHandler() {
    if (confirm("Are you sure you want to delete this node-session (this does not delete the node)?")) {
      NodeSessionStore.deleteNodeSession(this.props.params.applicationName, this.props.params.nodeName, (responseData) => {
        this.context.router.push("/applications/"+this.props.params.applicationName+"/nodes/"+this.props.params.nodeName+"/edit");
      });
    }
  }

  render() {
    let deleteButton;

    if(this.state.sessionExists) {
      deleteButton = <Link><button type="button" className="btn btn-danger" onClick={this.deleteHandler}>Delete node-session</button></Link>;
    }

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/applications">Applications</Link></li>
          <li><Link to={`/applications/${this.props.params.applicationName}`}>{this.props.params.applicationName}</Link></li>
          <li><Link to={`/applications/${this.props.params.applicationName}/nodes/${this.props.params.nodeName}/edit`}>{this.props.params.nodeName}</Link></li>
          <li className="active">Edit node-session / ABP</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            {deleteButton}
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <NodeSessionForm application={this.state.application} node={this.state.node} session={this.state.session} onSubmit={this.submitHandler} />
          </div>
        </div>
      </div>
    );
  }
}

export default UpdateNodeSession;
