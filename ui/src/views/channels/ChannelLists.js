import React, { Component } from 'react';
import { Link } from 'react-router';

import ChannelStore from "../../stores/ChannelStore";

class ChannelListRow extends Component {
  render() {
    return(
      <tr>
        <td><Link to={`/channels/${this.props.list.id}`}>{this.props.list.name}</Link></td>
        <td>{this.props.list.channels.join(", ")}</td>
      </tr>
    )
  }
}

class ChannelLists extends Component {
  constructor() {
    super();
    this.state = {
      lists:  [],
    };

    ChannelStore.getAllChannelLists((lists) => {
      this.setState({lists: lists});
    });
  }

  render() {
    const ChannelListRows = this.state.lists.map((list, i) => <ChannelListRow key={list.id} list={list} />); 

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li className="active">Channel lists</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to="/channels/create"><button type="button" className="btn btn-default">Create channel-list</button></Link>
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Channels</th>
                </tr>
              </thead>
              <tbody>
                {ChannelListRows}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    );
  }
}

export default ChannelLists;
