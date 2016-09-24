import React, { Component } from 'react';
import { Link } from 'react-router';

import ChannelStore from "../../stores/ChannelStore";
import ChannelListForm from "../../components/ChannelListForm";

class ChannelListDetails extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      list: {},
    };
    this.handleDelete = this.handleDelete.bind(this);
    this.handleChannelListSubmit = this.handleChannelListSubmit.bind(this);
  }

  componentWillMount() {
    ChannelStore.getChannelList(this.props.params.id, (list) => {
      this.setState({list: list});
    });
  }

  handleDelete() {
    if (confirm("Are you sure you want to delete this channel-list?")) {
      ChannelStore.deleteChannelList(this.props.params.id, (responseData) => {
        this.context.router.push("/channels");
      });
    }
  }

  handleChannelListSubmit(list) {
    ChannelStore.updateChannelList(this.props.params.id, list, (responseData) => {
      this.context.router.push("/channels");
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/channels">Channel lists</Link></li>
          <li className="active">{this.state.list.name}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <button type="button" className="btn btn-danger" onClick={this.handleDelete}>delete</button>
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <ChannelListForm list={this.state.list} onSubmit={this.handleChannelListSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default ChannelListDetails;
