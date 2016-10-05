import React, { Component } from 'react';
import { Link } from 'react-router';

import NodeSessionStore from "../../stores/NodeSessionStore";
import NodeSessionForm from "../../components/NodeSessionForm";
import NodeStore from "../../stores/NodeStore";
import ChannelStore from "../../stores/ChannelStore";

class NodeSessionDetails extends Component {
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
    NodeSessionStore.getNodeSession(this.props.params.devEUI, (session) => {
      this.setState({
        session: session,
        sessionExists: true,
      });
    }); 

    NodeStore.getNode(this.props.params.devEUI, (node) => {
      this.setState({node: node}); 

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

    if (this.state.sessionExists) {
      NodeSessionStore.updateNodeSession(this.props.params.devEUI, session, (responseData) => {
        this.context.router.push("/nodes/"+this.props.params.devEUI);
      });
    } else {
      NodeSessionStore.createNodeSession(session, (responseData) => {
        this.context.router.push("/nodes/"+ this.props.params.devEUI);
      });
    }
  }

  deleteHandler() {
    if (confirm("Are you sure you want to delete this node-session (this does not delete the node)?")) {
      NodeSessionStore.deleteNodeSession(this.props.params.devEUI, (responseData) => {
        this.context.router.push("/nodes/"+this.props.params.devEUI);
      });
    }
  }

  render() {
    let deleteButton;

    if(this.state.sessionExists) {
      deleteButton = <Link><button type="button" className="btn btn-danger" onClick={this.deleteHandler}>Delete</button></Link>;
    }

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li>Nodes</li>
          <li><Link to={`/nodes/${this.props.params.devEUI}`}>{this.props.params.devEUI}</Link></li>
          <li className="active">node-session / ABP</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            {deleteButton}
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <NodeSessionForm session={this.state.session} onSubmit={this.submitHandler} />
          </div>
        </div>
      </div>
    );
  }
}

export default NodeSessionDetails;
