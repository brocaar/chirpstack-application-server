import React, { Component } from "react";
import { Link } from "react-router";

import ChannelStore from "../../stores/ChannelStore";
import ChannelListForm from "../../components/ChannelListForm";

class CreateChannelList extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      list: {
        channels: [],
      },
    };
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(list) {
    ChannelStore.createChannelList(list, (responseData) => {
      this.context.router.push("/channels");
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/channels">Channel lists</Link></li>
          <li className="active">create channel-list</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <ChannelListForm list={this.state.list} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default CreateChannelList;
